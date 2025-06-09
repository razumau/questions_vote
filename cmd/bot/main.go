package main

import (
	"log"
	"questions-vote/internal/handlers"
)

func main() {
	log.Println("Starting questions-vote bot...")
	
	// TODO: Initialize database
	
	// Create and start bot handler
	botHandler, err := handlers.NewBotHandler()
	if err != nil {
		log.Fatalf("Failed to create bot handler: %v", err)
	}
	
	// Run the bot
	err = botHandler.Run()
	if err != nil {
		log.Fatalf("Bot failed to run: %v", err)
	}
}