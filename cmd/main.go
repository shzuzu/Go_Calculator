package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shzuzu/Go_Calculator/internal/application"
)

func main() {
	mode := flag.String("mode", "console", "Application operating mode: console or server")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Please choose the mode, use --mode=console or --mode=server")
		os.Exit(1)
	}

	createEnv()
	app := application.New()
	switch *mode {
	case "":
		fmt.Println("Please choose the mode, use --mode=console or --mode=server")
		os.Exit(1)
	case "console":
		fmt.Println("Starting calculator in console mode...")
		app.Run()
	case "server":
		fmt.Println("Starting HTTP-server...")
		err := app.RunServer()
		if err != nil {
			fmt.Println("Error via starting the server:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown mode. Use --mode=console or --mode=server")
		os.Exit(1)
	}
}

func createEnv() {
	if _, err := os.Stat(".env"); err == nil {
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
	`
	d1 := []byte(envVars)
	err := os.WriteFile(".env", d1, 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
}
