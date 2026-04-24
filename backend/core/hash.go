package core

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetHash(s string) []byte {

	hash := sha256.Sum256([]byte(s))

	return hash[:]
}

func GetHashString(s string) string {

	hash := GetHash(s)

	hash_string := hex.EncodeToString(hash)
	return hash_string
}
