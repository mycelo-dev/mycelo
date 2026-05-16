package api_key

import (
	"context"

	"github.com/mycelo-dev/mycelo/backend/internal/apikeytoken"
)

// CreateApiKeyServices generates a new API key, stores its hash, and returns the raw token once.
func CreateApiKeyServices(ctx context.Context, tenant_public_id string, team_public_id string) (CreateApiKeyResponse, error) {
	secret, hash_string, err := apikeytoken.CreateSecret()
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	err = StoreApiKeyHashInDbRepository(ctx, tenant_public_id, team_public_id, hash_string)
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	return CreateApiKeyResponse{
		ApiKey: apikeytoken.Build(tenant_public_id, team_public_id, secret),
	}, nil
}

// CreateApiKeyForTeamServices generates an API key for a team owned by a signed-up user.
func CreateApiKeyForTeamServices(ctx context.Context, tenant_public_id string, user_public_id string, team_public_id string) (CreateApiKeyResponse, error) {
	secret, hash_string, err := apikeytoken.CreateSecret()
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	tenantPublicId, teamPublicId, err := StoreApiKeyHashForTeamRepository(ctx, tenant_public_id, user_public_id, team_public_id, hash_string)
	if err != nil {
		return CreateApiKeyResponse{}, err
	}

	return CreateApiKeyResponse{
		ApiKey: apikeytoken.Build(tenantPublicId, teamPublicId, secret),
	}, nil
}

// RevokeApiKeyServices removes the stored key for the given tenant-team pair.
func RevokeApiKeyServices(ctx context.Context, tenant_public_id string, team_public_id string) error {

	return RevokeApiKeyRepository(ctx, tenant_public_id, team_public_id)
}

// RotateApiKeyServices replaces the stored key hash and returns the new raw token.
func RotateApiKeyServices(ctx context.Context, tenant_public_id string, team_public_id string) (RotateApiKeyResponse, error) {

	secret, hash_string, err := apikeytoken.CreateSecret()
	if err != nil {
		return RotateApiKeyResponse{}, err
	}

	err2 := RotateApiKeyRepository(ctx, tenant_public_id, team_public_id, hash_string)

	if err2 != nil {
		return RotateApiKeyResponse{}, err2
	}

	return RotateApiKeyResponse{
		ApiKey: apikeytoken.Build(tenant_public_id, team_public_id, secret),
	}, nil
}
