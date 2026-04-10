// connect to a db

package db

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
)

func ConnectDB(ctx context.Context) (*pgxpool.Pool, error) {
	var err error
	var db_url = os.Getenv("DB_URL")

	once.Do(func() {
		pool, err = pgxpool.New(ctx, db_url)
	})

	var result int
	var err2 error

	err2 = pool.QueryRow(ctx, "select 1").Scan(&result)

	if err2 != nil {
		log.Fatal("query execution failed: ", err2)
	}

	fmt.Println("connected to the pool: ", result)

	return pool, err
}

func Get() *pgxpool.Pool {
	return pool
}
