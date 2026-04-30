package api_key_auth

import (
	"context"
	"fmt"

	"github.com/mycelo-dev/mycelo/backend/core"
	select_queries "github.com/mycelo-dev/mycelo/backend/queries/select_queries"
)

func GetApiKeyHashFromDbRepository(ctx context.Context, tenant_public_id string, team_public_id string) (string, error) {

	var api_key_hash string

	query := select_queries.GetApiKeyHashFromDbQuery()

	row := core.Get().QueryRow(ctx, query, tenant_public_id, team_public_id)

	err := row.Scan(&api_key_hash)

	if err != nil {
		fmt.Println("error reading api key hash from DB: ", err)
		return "", err
	}

	return api_key_hash, nil

}
