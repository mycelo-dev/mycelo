package destination_management

import (
	"context"
	"fmt"
	"time"

	"github.com/mycelo-dev/mycelo/backend/core"
	delete_queries "github.com/mycelo-dev/mycelo/backend/queries/delete_queries"
	insert_queries "github.com/mycelo-dev/mycelo/backend/queries/insert_queries"
	select_queries "github.com/mycelo-dev/mycelo/backend/queries/select_queries"
	udpate_queries "github.com/mycelo-dev/mycelo/backend/queries/update_queries"
)

func CreateDestinationRepository(ctx context.Context, destination_name string, destination_address string) error {
	query := insert_queries.GetInsertDestinationQuery()

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

	query := udpate_queries.GetUpdateDestinationQuery()

	updated_at := time.Now().UnixMilli()

	_, err := core.Get().Exec(ctx, query, destination_name, destination_address, updated_at, id)

	if err != nil {
		fmt.Println("failed to update the destination: ", err)
		return err
	}

	return err
}

func UpdateDeliveryFlagRepository(ctx context.Context, id string, delivery_flag bool) error {

	query := udpate_queries.GetUpdateDeliveryFlagQuery()

	updated_at := time.Now().UnixMilli()

	_, err := core.Get().Exec(ctx, query, delivery_flag, updated_at, id)

	if err != nil {
		fmt.Println("failed to update the delivery flag: ", err)
		return err
	}

	return err
}

func DeleteDestinationRepository(ctx context.Context, id string) error {

	query := delete_queries.GetDeleteDestinationQuery()

	_, err := core.Get().Exec(ctx, query, id)

	if err != nil {
		fmt.Println("failed to delete the destination: ", err)
		return err
	}

	return err
}

func GetDeliveryFlagByPublicId(ctx context.Context, id string) (bool, error) {

	query := select_queries.GetReadDeliveryFlagByPublicIdQuery()

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

	query := insert_queries.GetAssignTopicToDestinationQuery()

	_, err := core.Get().Exec(ctx, query, destination_id, topic_id)

	if err != nil {
		fmt.Println("failed to assign topic to destination: ", err)
		return err
	}

	return err
}

func ListDestinationsRepository(ctx context.Context) ([]DestinationRecord, error) {

	query := select_queries.GetDestinationsByTenantAndTeamQuery()

	rows, err := core.Get().Query(ctx, query, 1, 1)
	if err != nil {
		fmt.Println("failed to read destinations: ", err)
		return nil, err
	}
	defer rows.Close()

	destinations := make([]DestinationRecord, 0)

	for rows.Next() {
		var destination DestinationRecord

		if err := rows.Scan(
			&destination.Destination_id,
			&destination.Destination_name,
			&destination.Destination_address,
			&destination.Delivery_flag,
		); err != nil {
			fmt.Println("failed to scan destination: ", err)
			return nil, err
		}

		destinations = append(destinations, destination)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("failed while reading destination rows: ", err)
		return nil, err
	}

	return destinations, nil
}

func ListDestinationTopicMappingsRepository(ctx context.Context) ([]DestinationTopicMappingRecord, error) {

	query := select_queries.GetDestinationTopicMappingsByTenantAndTeamQuery()

	rows, err := core.Get().Query(ctx, query, 1, 1)
	if err != nil {
		fmt.Println("failed to read destination topic mappings: ", err)
		return nil, err
	}
	defer rows.Close()

	mappings := make([]DestinationTopicMappingRecord, 0)

	for rows.Next() {
		var mapping DestinationTopicMappingRecord

		if err := rows.Scan(
			&mapping.Destination_id,
			&mapping.Destination_name,
			&mapping.Destination_address,
			&mapping.Delivery_flag,
			&mapping.Last_delivered_event_id,
			&mapping.Topic_id,
			&mapping.Topic_name,
		); err != nil {
			fmt.Println("failed to scan destination topic mapping: ", err)
			return nil, err
		}

		mappings = append(mappings, mapping)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("failed while reading destination topic mappings rows: ", err)
		return nil, err
	}

	return mappings, nil
}

func DeleteDestinationTopicMappingRepository(ctx context.Context, destination_id string, topic_id string) error {

	query := delete_queries.GetDeleteDestinationTopicMappingQuery()

	_, err := core.Get().Exec(ctx, query, destination_id, topic_id)

	if err != nil {
		fmt.Println("failed to delete the topic for the destination: ", err)
		return err
	}

	return err
}
