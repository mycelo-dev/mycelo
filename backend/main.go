package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	db "github.com/mycelo-dev/mycelo/backend/core"
	http_outbound "github.com/mycelo-dev/mycelo/backend/outbound"
	stream_routes "github.com/mycelo-dev/mycelo/backend/routes/stream"
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

	fmt.Println("successfully connected to the DB")
	fmt.Println("Now executing the handleRequests function")

	go http_outbound.ConsumeEvents(ctx, "my_topic", 0)
	stream_routes.HandleRequests()

	defer pool.Close() // keep it open until the app stops
}
