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
	"golang.org/x/crypto/scrypt"
)

const (
	passwordHashAlgorithm       = "scrypt"
	passwordScryptN             = 32768
	passwordScryptR             = 8
	passwordScryptP             = 1
	passwordSaltBytes           = 16
	passwordHashBytes           = 32
	legacyPasswordHashAlgorithm = "pbkdf2_sha256"
	legacyPasswordIterations    = 210000
)

// HashPassword derives a salted scrypt password hash for storage.
func HashPassword(password string) (string, error) {
	salt, err := core.GetRandomBytes(passwordSaltBytes)
	if err != nil {
		return "", err
	}

	hash, err := scrypt.Key([]byte(password), salt, passwordScryptN, passwordScryptR, passwordScryptP, passwordHashBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s$%d$%d$%d$%s$%s",
		passwordHashAlgorithm,
		passwordScryptN,
		passwordScryptR,
		passwordScryptP,
		hex.EncodeToString(salt),
		hex.EncodeToString(hash),
	), nil
}

// VerifyPassword checks a password against a stored hash.
func VerifyPassword(password string, storedHash string) bool {
	parts := strings.Split(storedHash, "$")
	switch {
	case len(parts) == 6 && parts[0] == passwordHashAlgorithm:
		return verifyScryptPassword(password, parts)
	case len(parts) == 4 && parts[0] == legacyPasswordHashAlgorithm:
		return verifyLegacyPBKDF2Password(password, parts)
	default:
		return false
	}
}

func verifyScryptPassword(password string, parts []string) bool {
	n, err := strconv.Atoi(parts[1])
	if err != nil || n <= 1 {
		return false
	}

	r, err := strconv.Atoi(parts[2])
	if err != nil || r <= 0 {
		return false
	}

	p, err := strconv.Atoi(parts[3])
	if err != nil || p <= 0 {
		return false
	}

	salt, err := hex.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedHash, err := hex.DecodeString(parts[5])
	if err != nil {
		return false
	}

	actualHash, err := scrypt.Key([]byte(password), salt, n, r, p, len(expectedHash))
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}

func verifyLegacyPBKDF2Password(password string, parts []string) bool {
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
