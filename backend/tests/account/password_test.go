package tests

import (
	"strings"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/account"
)

func TestHashPasswordUsesScryptAndVerifies(t *testing.T) {
	hash, err := account.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !strings.HasPrefix(hash, "scrypt$") {
		t.Fatalf("HashPassword used unexpected algorithm: %s", hash)
	}
	if !account.VerifyPassword("correct horse battery staple", hash) {
		t.Fatal("VerifyPassword rejected the original password")
	}
	if account.VerifyPassword("wrong password", hash) {
		t.Fatal("VerifyPassword accepted the wrong password")
	}
}
