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
func StoreApiKeyHashInDbRepository(ctx context.Context, hash string) error {
	query := insert_queries.GetInsertApiKeyHashQuery()

	created_at := time.Now().UnixMilli()
	updated_at := time.Now().UnixMilli()

	tenant_public_id := "880e2588-8b42-4a4d-8357-04bf6e808fb7"
	team_public_id := "e72115ca-8f6b-4d5a-b6d4-bd250f6f64fb"

	_, err := core.Get().Exec(ctx, query, tenant_public_id, team_public_id, hash, created_at, updated_at)

	if err != nil {
		fmt.Println("error inserting api key hash in DB: ", err)
		return err
	}

	return err
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
func RotateApiKeyRepository(ctx context.Context, hash string) error {

	query := update_queries.GetRotateApiKeyQuery()

	tenant_public_id := "880e2588-8b42-4a4d-8357-04bf6e808fb7"
	team_public_id := "e72115ca-8f6b-4d5a-b6d4-bd250f6f64fb"

	updated_at := time.Now().UnixMilli()

	_, err := core.Get().Exec(ctx, query, hash, tenant_public_id, team_public_id, updated_at)

	if err != nil {
		fmt.Println("error rotating the api key: ", err)
		return err
	}

	return err
}
