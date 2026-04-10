package main

import (
	"log"

	db "gitbub.com/mycelo-dev/mycelo/backend/core"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to database
	db.ConnectDB()
}
