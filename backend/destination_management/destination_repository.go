package destination_management

import (
	"context"
	"fmt"
	"time"

	"github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries"
)

func CreateDestinationRepository(ctx context.Context, destination_name string, destination_address string) error {
	query := queries.GetInsertDestinationQuery()

	created_at := time.Now().UnixMilli()
	updated_at := time.Now().UnixMilli()

	_, err := core.Get().Exec(ctx, query, 1, 1, destination_name, destination_address, created_at, updated_at)

	if err != nil {
		fmt.Println("failed to create destination: ", err)
		return err
	}

	return err
}
