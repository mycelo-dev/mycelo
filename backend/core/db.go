// connect to a db

package db

import (
	"context"
	"fmt"

	"gitbub.com/mycelo-dev/mycelo/backend/configs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() {
	var db_url string
	db_url = configs.GetDBURL()
	pool, err := pgxpool.New(context.Background(), db_url)

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	var result int
	err = pool.QueryRow(context.Background(), "SELECT 1").Scan(&result)

	if err != nil {
		panic(err)
	}

	fmt.Println("Connected via pool: ", result)
}
