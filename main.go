package main

import (
	"free-games-scrape/internal/app"
	"log"
)

func main() {
	log.Println("Note: main.go is deprecated. Please use 'go run cmd/bot/main.go' instead.")
	log.Println("Running application with new modular structure...")

	// Create and run the application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
