package configs

import "os"

func GetDBURL() string {
	return os.Getenv("DB_URL")
}
