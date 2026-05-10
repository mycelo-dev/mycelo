package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	db "github.com/mycelo-dev/mycelo/backend/core"
	http_outbound "github.com/mycelo-dev/mycelo/backend/outbound"
	all_routes "github.com/mycelo-dev/mycelo/backend/routes"
)

func main() {

	// load the environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("could not load .env file: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	pool, err := db.ConnectDB(ctx)

	if err != nil {
		log.Fatal("could not connect to DB: ", err)
	}
	defer pool.Close()

	fmt.Println("successfully connected to the DB")
	fmt.Println("Now executing the outbound consumers and HTTP server")

	if err := http_outbound.StartConsumers(ctx); err != nil {
		log.Fatal("could not start outbound consumers: ", err)
	}

	server := &http.Server{
		Addr:    ":3000",
		Handler: all_routes.NewMux(),
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("could not start HTTP server: ", err)
	}
}
