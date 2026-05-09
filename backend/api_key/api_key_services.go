package api_key

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycelo-dev/mycelo/backend/core"
)

// CreateApiKeyServices generates a new API key, stores its hash, and returns the raw token once.
func CreateApiKeyServices(ctx context.Context) (CreateApiKeyResponse, error) {

	random_bytes, err := core.GetRandomBytes(32)

	if err != nil {
		fmt.Println("error generating random bytes")
		return CreateApiKeyResponse{}, err
	}

	hex_string := core.GetHexString(random_bytes)

	hash_string := core.GetHashString(hex_string)

	err2 := StoreApiKeyHashInDbRepository(ctx, hash_string)

	if err2 != nil {
		return CreateApiKeyResponse{}, err2
	}

	tenant_public_id := "880e2588-8b42-4a4d-8357-04bf6e808fb7"
	team_public_id := "e72115ca-8f6b-4d5a-b6d4-bd250f6f64fb"

	api_key := strings.Join([]string{"mc", tenant_public_id, team_public_id, hex_string}, "_")

	return CreateApiKeyResponse{
		ApiKey: api_key,
	}, nil
}

// RevokeApiKeyServices removes the stored key for the given tenant-team pair.
func RevokeApiKeyServices(ctx context.Context, tenant_public_id string, team_public_id string) error {

	return RevokeApiKeyRepository(ctx, tenant_public_id, team_public_id)
}

// RotateApiKeyServices replaces the stored key hash and returns the new raw token.
func RotateApiKeyServices(ctx context.Context) (RotateApiKeyResponse, error) {

	random_bytes, err := core.GetRandomBytes(32)

	if err != nil {
		fmt.Println("error generating random bytes")
		return RotateApiKeyResponse{}, err
	}

	hex_string := core.GetHexString(random_bytes)

	hash_string := core.GetHashString(hex_string)

	err2 := RotateApiKeyRepository(ctx, hash_string)

	if err2 != nil {
		return RotateApiKeyResponse{}, err2
	}

	tenant_public_id := "880e2588-8b42-4a4d-8357-04bf6e808fb7"
	team_public_id := "e72115ca-8f6b-4d5a-b6d4-bd250f6f64fb"

	api_key := strings.Join([]string{"mc", tenant_public_id, team_public_id, hex_string}, "_")

	return RotateApiKeyResponse{
		ApiKey: api_key,
	}, nil
}
