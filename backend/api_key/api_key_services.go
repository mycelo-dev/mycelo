package api_key

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycelo-dev/mycelo/backend/core"
)

// CreateApiKeyServices generates a new API key, stores its hash, and returns the raw token once.
func CreateApiKeyServices(ctx context.Context, tenant_public_id string, team_public_id string) (CreateApiKeyResponse, error) {
	hex_string, hash_string, err := createApiKeySecret()
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	err = StoreApiKeyHashInDbRepository(ctx, tenant_public_id, team_public_id, hash_string)
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	return CreateApiKeyResponse{
		ApiKey: buildApiKey(tenant_public_id, team_public_id, hex_string),
	}, nil
}

// CreateApiKeyForTeamServices generates an API key for a team owned by a signed-up user.
func CreateApiKeyForTeamServices(ctx context.Context, tenant_public_id string, user_public_id string, team_public_id string) (CreateApiKeyResponse, error) {
	hex_string, hash_string, err := createApiKeySecret()
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	tenantPublicId, teamPublicId, err := StoreApiKeyHashForTeamRepository(ctx, tenant_public_id, user_public_id, team_public_id, hash_string)
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	return CreateApiKeyResponse{
		ApiKey: buildApiKey(tenantPublicId, teamPublicId, hex_string),
	}, nil
}

func createApiKeySecret() (string, string, error) {
	random_bytes, err := core.GetRandomBytes(32)

	if err != nil {
		fmt.Println("error generating random bytes")
		return "", "", err
	}

	hex_string := core.GetHexString(random_bytes)

	hash_string := core.GetHashString(hex_string)

	return hex_string, hash_string, nil
}

func buildApiKey(tenant_public_id string, team_public_id string, hex_string string) string {
	return strings.Join([]string{"mc", tenant_public_id, team_public_id, hex_string}, "_")
}

// RevokeApiKeyServices removes the stored key for the given tenant-team pair.
func RevokeApiKeyServices(ctx context.Context, tenant_public_id string, team_public_id string) error {

	return RevokeApiKeyRepository(ctx, tenant_public_id, team_public_id)
}

// RotateApiKeyServices replaces the stored key hash and returns the new raw token.
func RotateApiKeyServices(ctx context.Context, tenant_public_id string, team_public_id string) (RotateApiKeyResponse, error) {

	random_bytes, err := core.GetRandomBytes(32)

	if err != nil {
		fmt.Println("error generating random bytes")
		return RotateApiKeyResponse{}, err
	}

	hex_string := core.GetHexString(random_bytes)

	hash_string := core.GetHashString(hex_string)

	err2 := RotateApiKeyRepository(ctx, tenant_public_id, team_public_id, hash_string)

	if err2 != nil {
		return RotateApiKeyResponse{}, err2
	}

	return RotateApiKeyResponse{
		ApiKey: buildApiKey(tenant_public_id, team_public_id, hex_string),
	}, nil
}
