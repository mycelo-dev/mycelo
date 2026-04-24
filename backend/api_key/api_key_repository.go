package api_key

import (
	"context"
	"fmt"
	"time"

	"github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries"
)

func StoreApiKeyHashInDbRepository(ctx context.Context, hash string) error {
	query := queries.GetInsertApiKeyHashQuery()

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
