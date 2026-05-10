// connect to a db

package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool *pgxpool.Pool
	once sync.Once //Ensure the pool is created only once
	err  error
)

// ConnectDB initializes the shared connection pool and verifies it can execute queries.
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

	if err != nil {
		return nil, err
	}

	if pool == nil {
		return nil, errors.New("database pool is not initialized")
	}

	var result int
	err = pool.QueryRow(ctx, "select 1").Scan(&result)

	if err != nil {
		return nil, err
	}

	fmt.Println("connected to the pool: ", result)

	return pool, nil
}

// Get returns the shared database pool after startup has initialized it.
func Get() *pgxpool.Pool {
	return pool
}
