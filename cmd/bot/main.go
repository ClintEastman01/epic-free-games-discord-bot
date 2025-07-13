package main

import (
	"log"

	"free-games-scrape/internal/app"
)

func main() {
	// Create and run the application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}