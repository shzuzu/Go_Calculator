package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/shzuzu/Go_Calculator/internal/auth"
	"github.com/shzuzu/Go_Calculator/internal/database/repo"
	"github.com/shzuzu/Go_Calculator/internal/grpc"
	"github.com/shzuzu/Go_Calculator/internal/middleware"
	"github.com/shzuzu/Go_Calculator/pkg/calc"
)

type Request struct {
	Expression string `json:"expression"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Id struct {
	Id string `json:"id"`
}

type Token struct {
	Token string `json:"token"`
}

type Error struct {
	Error string `json:"error"`
}

type Orchestrator struct {
	mu               sync.Mutex
	expressionRepo   *repo.Repository
	authService      *auth.AuthService
	calculatorClient *grpc.CalculatorClient
}

func NewOrchestrator(db *sql.DB, calcClient *grpc.CalculatorClient) *Orchestrator {
	return &Orchestrator{
		expressionRepo:   repo.NewRepository(db),
		authService:      auth.NewAuthService(db),
		calculatorClient: calcClient,
	}
}

func (o *Orchestrator) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Error: "Bad request"})
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "", http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Error: "Login and password are required"})
		return
	}

	err := o.authService.RegisterUser(req.Login, req.Password)
	if err != nil {
		if err == auth.ErrUserAlreadyExists {
			http.Error(w, "", http.StatusConflict)
			json.NewEncoder(w).Encode(Error{Error: "User already exists"})
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Error: "Internal server error register"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (o *Orchestrator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Error: "Bad request"})
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "", http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Error: "Login and password are required"})
		return
	}

	token, err := o.authService.LoginUser(req.Login, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCreds {
			http.Error(w, "", http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Error{Error: "Invalid credentials"})
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Error: "Internal server error login"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Token{Token: token})
}
func (o *Orchestrator) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expressions, err := o.expressionRepo.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Internal server error expression", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(expressions); err != nil {
		http.Error(w, "Something went wrong..", http.StatusInternalServerError)
		return
	}

}
func (o *Orchestrator) ExpressionFromID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	expr, err := o.expressionRepo.GetByID(id)
	if err != nil {
		http.Error(w, "Internal server error expression by id", http.StatusInternalServerError)
		return
	}

	if expr == nil || expr.UserID != userID {
		http.Error(w, fmt.Sprintf("Expression with ID %s not found", idStr), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(expr); err != nil {
		http.Error(w, "Something went wrong..", http.StatusInternalServerError)
		return
	}
}

func (o *Orchestrator) CreateExpressionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateExpressionHandler: started")
	request := &Request{}
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		log.Println("CreateExpressionHandler: method not allowed")
		http.Error(w, "Can't complete that method", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("CreateExpressionHandler: error decoding request: %v", err)
		http.Error(w, "", http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Error{Error: "Unprocessable Entity"})
		return
	}

	log.Printf("CreateExpressionHandler: received expression: %s", request.Expression)

	if err := o.calculatorClient.ValidateExpression(request.Expression); err != nil {
		log.Printf("CreateExpressionHandler: error validating expression: %v", err)
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
			json.NewEncoder(w).Encode(Error{Error: "Internal server error validate"})
			return
		}
	}

	id, err := o.expressionRepo.Create(userID, request.Expression)
	if err != nil {
		log.Printf("CreateExpressionHandler: error creating expression: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Error: "Internal server error db"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Id{Id: strconv.FormatInt(id, 10)})

	go func() {
		log.Printf("CreateExpressionHandler: calculating expression: %s", request.Expression)
		result, err := o.calculatorClient.Calculate(request.Expression)

		if err != nil {
			log.Printf("CreateExpressionHandler: calculation error: %v", err)
			o.expressionRepo.UpdateStatus(id, "error", nil)
		} else {
			log.Printf("CreateExpressionHandler: calculation result: %v", result)
			o.expressionRepo.UpdateStatus(id, "done", &result)
		}
	}()
}
