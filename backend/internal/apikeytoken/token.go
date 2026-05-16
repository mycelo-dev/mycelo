package apikeytoken

import (
	"fmt"
	"strings"

	"github.com/mycelo-dev/mycelo/backend/core"
)

// CreateSecret returns the raw API-key secret and its persisted hash.
func CreateSecret() (string, string, error) {
	randomBytes, err := core.GetRandomBytes(32)
	if err != nil {
		fmt.Println("error generating random bytes")
		return "", "", err
	}

	secret := core.GetHexString(randomBytes)
	return secret, core.GetHashString(secret), nil
}

// Build renders the externally visible API key format.
func Build(tenantPublicID string, teamPublicID string, secret string) string {
	return strings.Join([]string{"mc", tenantPublicID, teamPublicID, secret}, "_")
}
