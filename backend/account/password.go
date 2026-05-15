package account

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mycelo-dev/mycelo/backend/core"
)

const (
	passwordHashAlgorithm  = "pbkdf2_sha256"
	passwordHashIterations = 210000
	passwordSaltBytes      = 16
	passwordHashBytes      = 32
)

// HashPassword derives a salted password hash for storage.
func HashPassword(password string) (string, error) {
	salt, err := core.GetRandomBytes(passwordSaltBytes)
	if err != nil {
		return "", err
	}

	hash := pbkdf2SHA256([]byte(password), salt, passwordHashIterations, passwordHashBytes)
	return fmt.Sprintf(
		"%s$%d$%s$%s",
		passwordHashAlgorithm,
		passwordHashIterations,
		hex.EncodeToString(salt),
		hex.EncodeToString(hash),
	), nil
}

// VerifyPassword checks a password against a stored hash.
func VerifyPassword(password string, storedHash string) bool {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 4 || parts[0] != passwordHashAlgorithm {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expectedHash, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}

	actualHash := pbkdf2SHA256([]byte(password), salt, iterations, len(expectedHash))
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}

func pbkdf2SHA256(password []byte, salt []byte, iterations int, keyLength int) []byte {
	hashLength := sha256.Size
	blockCount := (keyLength + hashLength - 1) / hashLength
	derived := make([]byte, 0, blockCount*hashLength)

	for block := 1; block <= blockCount; block++ {
		derived = append(derived, pbkdf2Block(password, salt, iterations, uint32(block))...)
	}

	return derived[:keyLength]
}

func pbkdf2Block(password []byte, salt []byte, iterations int, blockNumber uint32) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)

	blockBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(blockBytes, blockNumber)
	mac.Write(blockBytes)

	sum := mac.Sum(nil)
	result := make([]byte, len(sum))
	copy(result, sum)

	for i := 1; i < iterations; i++ {
		mac = hmac.New(sha256.New, password)
		mac.Write(sum)
		sum = mac.Sum(nil)
		for j := range result {
			result[j] ^= sum[j]
		}
	}

	return result
}

var errInvalidPassword = errors.New("invalid password")
var errMissingSession = errors.New("missing session token")
