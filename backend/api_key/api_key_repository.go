package api_key

import (
	"context"
	"fmt"
	"time"

	"github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries/delete_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/insert_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/update_queries"
)

// StoreApiKeyHashInDbRepository persists a newly generated API key hash.
func StoreApiKeyHashInDbRepository(ctx context.Context, tenant_public_id string, team_public_id string, hash string) error {
	query := insert_queries.GetInsertApiKeyHashQuery()

	created_at := time.Now().UnixMilli()
	updated_at := time.Now().UnixMilli()

	_, err := core.Get().Exec(ctx, query, tenant_public_id, team_public_id, hash, created_at, updated_at)

	if err != nil {
		fmt.Println("error inserting api key hash in DB: ", err)
		return err
	}

	return err
}

// StoreApiKeyHashForTeamRepository persists a team key for a tenant user and returns the verified scope.
func StoreApiKeyHashForTeamRepository(ctx context.Context, tenant_public_id string, user_public_id string, team_public_id string, hash string) (string, string, error) {
	query := insert_queries.GetInsertApiKeyHashForTeamQuery()

	created_at := time.Now().UnixMilli()
	updated_at := created_at

	var storedTenantPublicId string
	var storedTeamPublicId string
	err := core.Get().QueryRow(
		ctx,
		query,
		tenant_public_id,
		user_public_id,
		team_public_id,
		hash,
		created_at,
		updated_at,
	).Scan(&storedTenantPublicId, &storedTeamPublicId)

	if err != nil {
		fmt.Println("error inserting api key hash for team: ", err)
		return "", "", err
	}

	return storedTenantPublicId, storedTeamPublicId, nil
}

// RevokeApiKeyRepository deletes the stored API key for a tenant-team pair.
func RevokeApiKeyRepository(ctx context.Context, tenant_public_id string, team_public_id string) error {

	query := delete_queries.GetRevokeApiKeyQuery()

	_, err := core.Get().Exec(ctx, query, tenant_public_id, team_public_id)

	if err != nil {
		fmt.Println("error revoking API key: ", err)
		return err
	}

	return err
}

// RotateApiKeyRepository updates the stored hash for the current API key record.
func RotateApiKeyRepository(ctx context.Context, tenant_public_id string, team_public_id string, hash string) error {

	query := update_queries.GetRotateApiKeyQuery()

	updated_at := time.Now().UnixMilli()

	_, err := core.Get().Exec(ctx, query, hash, tenant_public_id, team_public_id, updated_at)

	if err != nil {
		fmt.Println("error rotating the api key: ", err)
		return err
	}

	return err
}
