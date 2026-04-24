package core

import (
	"crypto/rand"
	"fmt"
)

func GetRandomBytes(b int) ([]byte, error) {

	random_bytes := make([]byte, b)

	_, err := rand.Read(random_bytes)

	if err != nil {
		fmt.Println("error generating random bytes: ", err)
		return random_bytes, err
	}

	return random_bytes, err
}
