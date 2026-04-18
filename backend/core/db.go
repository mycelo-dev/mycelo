// connect to a db

package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool
	once sync.Once //Ensure the pool is created only once
	err  error
)

func ConnectDB(ctx context.Context) (*pgxpool.Pool, error) {
	db_url := os.Getenv("DB_URL")

	once.Do(func() {
		pool, err = pgxpool.New(ctx, db_url)

		if err != nil {
			return
		}

		//verify connectivity during initialization
		var result int
		err = pool.QueryRow(ctx, "select 1").Scan(&result)

		if err != nil {
			pool.Close()
			pool = nil
		}
	})

	var result int
	var err2 error

	err2 = pool.QueryRow(ctx, "select 1").Scan(&result)

	if err2 != nil {
		log.Fatal("query execution failed: ", err2)
	}

	fmt.Println("connected to the pool: ", result)

	return pool, nil
}

func Get() *pgxpool.Pool {
	return pool
}
