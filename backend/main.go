package main

import (
	"context"
	"log"

	db "gitbub.com/mycelo-dev/mycelo/backend/core"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	pool *pgxpool.Pool
	err  error
)

func main() {

	// load the environment variables
	godotenv.Load()

	// Connect to database
	ctx := context.Background()
	pool, err = db.ConnectDB(ctx)

	if err != nil {
		log.Fatal("could not connect to DB: ", err)
	}

	defer pool.Close() // keep it open until the app stops
}
