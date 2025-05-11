package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/shzuzu/Go_Calculator/internal/application"
	"github.com/shzuzu/Go_Calculator/internal/database/database"
	calcGrpc "github.com/shzuzu/Go_Calculator/internal/grpc"
)

func main() {
	mode := flag.String("mode", "console", "Application operating mode: console, server, or calc-server")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Please choose the mode, use --mode=console, --mode=server, or --mode=calc-server")
		os.Exit(1)
	}

	// Load environment variables
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	envPath := filepath.Join(dir, "../../.env")
	createEnv(envPath)

	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize database
	db, err := database.InitDB("./calculator.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	switch *mode {
	case "":
		fmt.Println("Please choose the mode, use --mode=console, --mode=server, or --mode=calc-server")
		os.Exit(1)
	case "console":
		fmt.Println("Starting calculator in console mode...")
		app := application.New(db, nil)
		app.Run()
	case "server":
		fmt.Println("Starting HTTP-server...")
		// Start gRPC client
		grpcClient, err := calcGrpc.NewCalculatorClient("localhost:50051")
		if err != nil {
			log.Fatalf("Failed to create gRPC client: %v", err)
		}
		defer grpcClient.Close()

		app := application.New(db, grpcClient)
		err = app.RunServer()
		if err != nil {
			fmt.Println("Error via starting the server:", err)
			os.Exit(1)
		}
	case "calc-server":
		fmt.Println("Starting gRPC calculator server...")
		err := calcGrpc.StartServer(":50051")
		if err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	default:
		fmt.Println("Unknown mode. Use --mode=console, --mode=server, or --mode=calc-server")
		os.Exit(1)
	}
}

func createEnv(envPath string) {
	if _, err := os.Stat(envPath); err == nil {
		fmt.Println("It's OK! .env file already exists")
		return
	}
	envVars :=
		`
	PORT=8080
	TIME_ADDITION_MS=0
	TIME_SUBTRACTION_MS=0
	TIME_MULTIPLICATION_MS=0
	TIME_DIVISION_MS=0
	COMPUTING_POWER=3
	GRPC_SERVER_ADDRESS=localhost:50051
	`
	d1 := []byte(envVars)
	err := os.WriteFile(envPath, d1, 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
}
