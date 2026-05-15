package api_key_auth

import (
	"context"
	"errors"

	"github.com/mycelo-dev/mycelo/backend/auth"
	"github.com/mycelo-dev/mycelo/backend/core"
)

var ErrInvalidApiKey = errors.New("invalid api key")

// ApiKeyAuthenticator validates an incoming API key and returns its auth context.
func ApiKeyAuthenticator(ctx context.Context, api_key string) (auth.AuthContext, error) {

	tenant_public_id := GetTenantPublicIdFromApiKey(api_key)
	team_public_id := GetTeamPublicIdFromApiKey(api_key)
	incoming_hash_string := GetHashStringFromApiKey(api_key)
	if tenant_public_id == "" || team_public_id == "" || incoming_hash_string == "" {
		return auth.AuthContext{}, ErrInvalidApiKey
	}

	incoming_hash := core.GetHashString(incoming_hash_string)

	stored_hash, err := GetApiKeyHashFromDbRepository(ctx, tenant_public_id, team_public_id)

	if err != nil {
		return auth.AuthContext{}, err
	}

	is_api_key_valid := CompareApiKeyHash(incoming_hash, stored_hash)

	if is_api_key_valid == false {

		return auth.AuthContext{}, ErrInvalidApiKey
	}

	return auth.AuthContext{
		TenantPublicId: tenant_public_id,
		TeamPublicId:   team_public_id,
	}, nil
}
