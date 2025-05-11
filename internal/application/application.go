package application

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/shzuzu/Go_Calculator/internal/auth"
	calcGrpc "github.com/shzuzu/Go_Calculator/internal/grpc"
	"github.com/shzuzu/Go_Calculator/internal/middleware"
	"github.com/shzuzu/Go_Calculator/pkg/calc"
)

type Config struct {
	Addr           string
	GrpcServerAddr string
}

func ConfigFromEnv() *Config {
	config := new(Config)

	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}

	config.GrpcServerAddr = os.Getenv("GRPC_SERVER_ADDRESS")
	if config.GrpcServerAddr == "" {
		config.GrpcServerAddr = "localhost:50051"
	}

	return config
}

type Application struct {
	config           *Config
	db               *sql.DB
	calculatorClient *calcGrpc.CalculatorClient
}

func New(db *sql.DB, calculatorClient *calcGrpc.CalculatorClient) *Application {
	return &Application{
		config:           ConfigFromEnv(),
		db:               db,
		calculatorClient: calculatorClient,
	}
}

func (a *Application) Run() error {
	// Use direct calculation in console mode
	for {
		fmt.Println(`Input expression (enter "exit" to exit):`)
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Failed to read expression from console!")
		}

		text = strings.TrimSpace(text)
		if text == "exit" {
			log.Println("Application was successfully closed!")
			return nil
		}

		// In console mode, we calculate directly without using gRPC
		result, err := calc.Calc(text)
		if err != nil {
			log.Println(text, "<-- you've entered \nCalculation failed with error: ", err)
		} else {
			log.Println(result)
		}
	}
}

func (a *Application) RunServer() error {
	orchestrator := NewOrchestrator(a.db, a.calculatorClient)

	authService := auth.NewAuthService(a.db)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", orchestrator.RegisterHandler)
	mux.HandleFunc("/api/v1/login", orchestrator.LoginHandler)

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/calculate", orchestrator.CreateExpressionHandler)
	// protectedMux.HandleFunc("/api/v1/expressions", orchestrator.GetExpressionsHandler)
	protectedMux.HandleFunc("/api/v1/expressions/{id}", orchestrator.ExpressionFromID)

	authMiddleware := middleware.AuthMiddleware(authService)
	protectedHandler := authMiddleware(protectedMux)

	mux.Handle("/api/v1/calculate", protectedHandler)
	mux.Handle("/api/v1/expressions", protectedHandler)
	mux.Handle("/api/v1/expressions/{id}", protectedHandler)

	serverAddr := ":" + a.config.Addr
	log.Printf("HTTP server listening on %s", serverAddr)
	return http.ListenAndServe(serverAddr, mux)
}
