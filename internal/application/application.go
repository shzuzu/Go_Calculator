package application

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/shzuzu/Go_Calculator/pkg/calc"
)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
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
			log.Println(text, "=", result)
		}
	}
}

type Request struct {
	Expression string `json:"expression"`
}

type Result struct {
	Result float64 `json:"result"`
}

type Error struct {
	Error string `json:"error"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Can`t complete that method", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Error: "Internal server error"})
		return
	}

	result, err := calc.Calc(request.Expression)
	if err != nil {
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
			json.NewEncoder(w).Encode(Error{Error: "You shold enter an expression"})
			return

		default:
			http.Error(w, "", http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{Error: "Internal server error"})
			return

		}
	}
	json.NewEncoder(w).Encode(Result{Result: result})

}
func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
