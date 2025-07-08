package main

import (
	"log"
	"questions-vote/internal/db"
	"questions-vote/internal/handlers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	log.Println("Starting questions-vote bot...")

	err = db.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	botHandler, err := handlers.NewBotHandler()
	if err != nil {
		log.Fatalf("Failed to create bot handler: %v", err)
	}

	err = botHandler.Run()
	if err != nil {
		log.Fatalf("Bot failed to run: %v", err)
	}
}
