package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shzuzu/Go_Calculator/internal/application"
)

func main() {
	mode := flag.String("mode", "console", "Application operating mode: console or server")
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Please choose the mode, use --mode=console or --mode=server")
		os.Exit(1)
	}

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
	envPath := filepath.Join("..", ".env")

	file, err := os.Create(envPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	envVars :=
		`TIME_ADDITION_MS=0
		TIME_SUBTRACTION_MS=0
		TIME_MULTIPLICATION_MS=0
		TIME_DIVISION_MS=0
	`
	_, err = file.WriteString(envVars)
	if err != nil {
		panic(err)
	}
}
