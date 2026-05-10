package core

import (
	"crypto/rand"
	"fmt"
)

// GetRandomBytes fills and returns a byte slice using cryptographically secure randomness.
func GetRandomBytes(b int) ([]byte, error) {

	random_bytes := make([]byte, b)

	_, err := rand.Read(random_bytes)

	if err != nil {
		fmt.Println("error generating random bytes: ", err)
		return random_bytes, err
	}

	return random_bytes, err
}
