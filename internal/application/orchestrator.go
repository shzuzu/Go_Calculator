package application

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/shzuzu/Go_Calculator/pkg/calc"
)

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file in application")
	}
	config := new(Config)

	config.Addr = os.Getenv("PORT")

	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config

}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{config: ConfigFromEnv()}
}

func (a *Application) Run() error {
	for {
		// читаем выражение для вычисления из командной строки
		log.Println(`Input expression (enter "exit" to exit):`)
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Failed to read expression from console!")
		}
		// убираем пробелы, чтобы оставить только вычислемое выражение
		text = strings.TrimSpace(text)
		// выходим, если ввели команду "exit"
		if text == "exit" {
			log.Println("Application was successfully closed!")
			return nil
		}
		//вычисляем выражение
		result, err := calc.Calc(text)
		if err != nil {
			log.Println(text, "<-- you've entered \nCalculation failed with error: ", err)
		} else {
			log.Println(result)
		}
	}
}

type Request struct {
	Expression string `json:"expression"`
}
type Id struct {
	Id string `json:"id"`
}

type Expression struct {
	Id     string   `json:"id"`
	Status string   `json:"status"`
	Result *float64 `json:"result"`
}

type Result struct {
	Expressions []Expression `json:"expressions"`
}

type Error struct {
	Error string `json:"error"`
}
type Orchestrator struct {
	mu    sync.Mutex
	Exprs []Expression
	ID    int
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		Exprs: make([]Expression, 0),
		ID:    0,
	}
}

// w.WriteHeader(http.StatusInternalServerError) <-- статус код
func (o *Orchestrator) ExpressionFromID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	for _, e := range o.Exprs {
		if e.Id == id {
			if err := json.NewEncoder(w).Encode(e); err != nil {
				http.Error(w, "Something went wrong..", http.StatusInternalServerError)
				return
			}
			return
		}
	}
	http.Error(w, fmt.Sprintf("Expression with ID %s not found", id), http.StatusNotFound)
}

func (o *Orchestrator) CreateExpressionHandler(w http.ResponseWriter, r *http.Request) {
	request := &Request{}
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Can't complete that method", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "", http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Error{Error: "Unprocessable Entity"})
		return
	}

	o.mu.Lock()
	o.ID += 1
	id := strconv.Itoa(o.ID)
	o.mu.Unlock()
	expr := Expression{
		Id:     id,
		Status: "pending",
		Result: nil,
	}
	o.mu.Lock()
	o.Exprs = append(o.Exprs, expr)
	o.mu.Unlock()

	errChan := make(chan error, 1)
	go func() {
		result, err := calc.Calc(request.Expression)

		// Обновляем статус и результат выражения
		o.mu.Lock()
		defer o.mu.Unlock()
		for i, e := range o.Exprs {
			if e.Id == id {
				if err != nil {
					o.Exprs[i].Status = "error"
					o.Exprs[i].Result = nil
					errChan <- err
				} else {
					o.Exprs[i].Status = "done"
					o.Exprs[i].Result = &result
					errChan <- nil
				}
				break
			}
		}
	}()
	if err := <-errChan; err != nil {
		switch err {
		case calc.ErrInvalidExpression:
			http.Error(w, "", http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(Error{Error: "Expression is not valid"})
			return
		case calc.ErrDivisionByZero:
			http.Error(w, "", http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(Error{Error: "Division by zero"})
			return
		case calc.ErrEOF:
			http.Error(w, "", http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(Error{Error: "You should enter an expression"})
			return
		default:
			http.Error(w, "", http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{Error: "Internal server error"})
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Id{Id: id})

}

func (o *Orchestrator) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	o.mu.Lock()
	if err := json.NewEncoder(w).Encode(Result{Expressions: o.Exprs}); err != nil {
		http.Error(w, "Something went wrong..", http.StatusInternalServerError)
	}
	defer o.mu.Unlock()
}

func (a *Application) RunServer() error {
	orchestrator := NewOrchestrator()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", orchestrator.CreateExpressionHandler)
	mux.HandleFunc("/api/v1/expressions", orchestrator.GetExpressionsHandler)
	mux.HandleFunc("/api/v1/expressions/{id}", orchestrator.ExpressionFromID)

	return http.ListenAndServe(":"+a.config.Addr, mux)
}
