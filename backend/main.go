package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	db "github.com/mycelo-dev/mycelo/backend/core"
)

var (
	pool *pgxpool.Pool
	err  error
)

func main() {

	// load the environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("could not load .env file: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	pool, err = db.ConnectDB(ctx)

	if err != nil {
		log.Fatal("could not connect to DB: ", err)
	}

	defer pool.Close() // keep it open until the app stops
}
