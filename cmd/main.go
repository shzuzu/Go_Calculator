package main

import "github.com/shzuzu/Go_Calculator/internal/application"

func main() {
	app := application.New()
	// app.Run()
	app.RunServer()
}
