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

func UpdateDestinationRepository(ctx context.Context, destination_name string, destination_address string, id string) error {

	query := queries.GetUpdateDestinationQuery()

	updated_at := time.Now().UnixMilli()

	_, err := core.Get().Exec(ctx, query, destination_name, destination_address, updated_at, id)

	if err != nil {
		fmt.Println("failed to update the destination: ", err)
		return err
	}

	return err
}

func DeleteDestinationRepository(ctx context.Context, id string) error {

	query := queries.GetDeleteDestinationQuery()

	_, err := core.Get().Exec(ctx, query, id)

	if err != nil {
		fmt.Println("failed to delete the destination: ", err)
		return err
	}

	return err
}

func GetDeliveryFlagByPublicId(ctx context.Context, id string) (bool, error) {

	query := queries.GetReadDeliveryFlagByPublicIdQuery()

	row := core.Get().QueryRow(ctx, query, id)

	var df DeliveryFlag
	var err error

	if err := row.Scan(&df.Delivery_flag); err != nil {
		fmt.Println("failed to put delivery flag value in the struct: ", err)
		return true, err
	}
	return df.Delivery_flag, err
}

func AssignTopicToDestinationRepository(ctx context.Context, destination_id string, topic_id string) error {

	query := queries.GetAssignTopicToDestinationQuery()

	_, err := core.Get().Exec(ctx, query, destination_id, topic_id)

	if err != nil {
		fmt.Println("failed to assign topic to destination: ", err)
		return err
	}

	return err
}
