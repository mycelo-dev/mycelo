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

// CreateDestinationRepository inserts a destination record for the current tenant-team scope.
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

// UpdateDestinationRepository updates a destination's name and address.
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

// UpdateDeliveryFlagRepository persists the delivery enablement flag for a destination.
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

// DeleteDestinationRepository deletes a destination by public ID.
func DeleteDestinationRepository(ctx context.Context, id string) error {

	query := delete_queries.GetDeleteDestinationQuery()

	_, err := core.Get().Exec(ctx, query, id)

	if err != nil {
		fmt.Println("failed to delete the destination: ", err)
		return err
	}

	return err
}

// GetDeliveryFlagByPublicId reads the delivery flag for a destination.
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

// AssignTopicToDestinationRepository creates a destination-topic mapping.
func AssignTopicToDestinationRepository(ctx context.Context, destination_id string, topic_id string, policy DestinationTopicMappingPolicy) error {

	query := insert_queries.GetAssignTopicToDestinationQuery()

	_, err := core.Get().Exec(
		ctx,
		query,
		destination_id,
		topic_id,
		policy.Retry_base_delay_ms,
		policy.Retry_max_delay_ms,
		policy.Max_consecutive_failures_before_skip,
		policy.Dead_letter_queue_enabled,
		policy.Skip_on_endpoint_4xx,
		policy.Skip_on_endpoint_5xx,
		policy.Skip_on_endpoint_transport_error,
		policy.Skip_on_event_payload_error,
	)

	if err != nil {
		fmt.Println("failed to assign topic to destination: ", err)
		return err
	}

	return err
}

// ListDestinationsRepository lists destinations for the current tenant-team scope.
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

// ListDestinationTopicMappingsRepository lists destination-topic mappings with delivery state.
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
			&mapping.Retry_base_delay_ms,
			&mapping.Retry_max_delay_ms,
			&mapping.Max_consecutive_failures_before_skip,
			&mapping.Dead_letter_queue_enabled,
			&mapping.Skip_on_endpoint_4xx,
			&mapping.Skip_on_endpoint_5xx,
			&mapping.Skip_on_endpoint_transport_error,
			&mapping.Skip_on_event_payload_error,
			&mapping.Last_attempted_event_id,
			&mapping.Last_failed_event_id,
			&mapping.Last_skipped_event_id,
			&mapping.Consecutive_failure_count,
			&mapping.Last_attempted_at,
			&mapping.Last_succeeded_at,
			&mapping.Last_failed_at,
			&mapping.Last_skipped_at,
			&mapping.Next_attempt_at,
			&mapping.Last_error_category,
			&mapping.Last_error,
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

// GetDestinationTopicMappingPolicyRepository reads current policy for a mapping.
func GetDestinationTopicMappingPolicyRepository(ctx context.Context, destination_id string, topic_id string) (DestinationTopicMappingPolicy, error) {
	query := select_queries.GetDestinationTopicMappingPolicyQuery()

	row := core.Get().QueryRow(ctx, query, destination_id, topic_id)

	var policy DestinationTopicMappingPolicy

	if err := row.Scan(
		&policy.Retry_base_delay_ms,
		&policy.Retry_max_delay_ms,
		&policy.Max_consecutive_failures_before_skip,
		&policy.Dead_letter_queue_enabled,
		&policy.Skip_on_endpoint_4xx,
		&policy.Skip_on_endpoint_5xx,
		&policy.Skip_on_endpoint_transport_error,
		&policy.Skip_on_event_payload_error,
	); err != nil {
		fmt.Println("failed to read destination topic mapping policy: ", err)
		return DestinationTopicMappingPolicy{}, err
	}

	return policy, nil
}

// UpdateDestinationTopicMappingPolicyRepository persists delivery policy for a mapping.
func UpdateDestinationTopicMappingPolicyRepository(ctx context.Context, destination_id string, topic_id string, policy DestinationTopicMappingPolicy) error {
	query := udpate_queries.GetUpdateDestinationTopicMappingPolicyQuery()

	_, err := core.Get().Exec(
		ctx,
		query,
		destination_id,
		topic_id,
		policy.Retry_base_delay_ms,
		policy.Retry_max_delay_ms,
		policy.Max_consecutive_failures_before_skip,
		policy.Dead_letter_queue_enabled,
		policy.Skip_on_endpoint_4xx,
		policy.Skip_on_endpoint_5xx,
		policy.Skip_on_endpoint_transport_error,
		policy.Skip_on_event_payload_error,
	)
	if err != nil {
		fmt.Println("failed to update destination topic mapping policy: ", err)
		return err
	}

	return nil
}

// DeleteDestinationTopicMappingRepository deletes a destination-topic mapping.
func DeleteDestinationTopicMappingRepository(ctx context.Context, destination_id string, topic_id string) error {

	query := delete_queries.GetDeleteDestinationTopicMappingQuery()

	_, err := core.Get().Exec(ctx, query, destination_id, topic_id)

	if err != nil {
		fmt.Println("failed to delete the topic for the destination: ", err)
		return err
	}

	return err
}
