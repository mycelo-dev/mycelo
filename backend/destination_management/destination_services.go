package destination_management

import (
	"context"
	"fmt"
)

// CreateDestinationServices creates a destination record.
func CreateDestinationServices(ctx context.Context, destination_name string, destination_address string) error {
	return CreateDestinationRepository(ctx, destination_name, destination_address)
}

// UpdateDestinationServices updates a destination by public ID.
func UpdateDestinationServices(ctx context.Context, destination_name string, destination_address string, id string) error {
	return UpdateDestinationRepository(ctx, destination_name, destination_address, id)
}

// UpdateDeliveryFlagServices enables or disables delivery for a destination-topic mapping.
func UpdateDeliveryFlagServices(ctx context.Context, destination_id string, topic_id string, delivery_flag bool) error {
	return UpdateDeliveryFlagRepository(ctx, destination_id, topic_id, delivery_flag)
}

// DeleteDestinationServices blocks deletion while delivery is still active.
func DeleteDestinationServices(ctx context.Context, id string) error {

	var df DeliveryFlag
	var err error
	df.Delivery_flag, err = GetDeliveryFlagByPublicId(ctx, id)

	if err != nil {
		return err
	}

	if df.Delivery_flag {
		return fmt.Errorf("cannot delete: delivery flag is active for ID %s", id)
	}

	err2 := DeleteDestinationRepository(ctx, id)

	return err2

}

// AssignTopicToDestinationServices links a topic to a destination.
func AssignTopicToDestinationServices(ctx context.Context, destination_id string, topic_id string, input DestinationTopicMappingPolicyInput) error {
	policy := DefaultDestinationTopicMappingPolicy().Apply(input)
	if err := policy.Validate(); err != nil {
		return err
	}

	return AssignTopicToDestinationRepository(ctx, destination_id, topic_id, policy)
}

// ListDestinationsServices returns all destinations for the active tenant-team scope.
func ListDestinationsServices(ctx context.Context) ([]DestinationRecord, error) {
	return ListDestinationsRepository(ctx)
}

// ListDestinationTopicMappingsServices returns all current destination-topic mappings.
func ListDestinationTopicMappingsServices(ctx context.Context) ([]DestinationTopicMappingRecord, error) {
	return ListDestinationTopicMappingsRepository(ctx)
}

// UpdateDestinationTopicMappingPolicyServices updates configurable delivery policy for a mapping.
func UpdateDestinationTopicMappingPolicyServices(ctx context.Context, destination_id string, topic_id string, input DestinationTopicMappingPolicyInput) error {
	currentPolicy, err := GetDestinationTopicMappingPolicyRepository(ctx, destination_id, topic_id)
	if err != nil {
		return err
	}

	nextPolicy := currentPolicy.Apply(input)
	if err := nextPolicy.Validate(); err != nil {
		return err
	}

	return UpdateDestinationTopicMappingPolicyRepository(ctx, destination_id, topic_id, nextPolicy)
}

// DeleteDestinationTopicMappingServices blocks mapping deletion while delivery is active.
func DeleteDestinationTopicMappingServices(ctx context.Context, destination_id string, topic_id string) error {

	var df DeliveryFlag
	var err error

	df.Delivery_flag, err = GetDeliveryFlagByPublicId(ctx, destination_id)

	if err != nil {
		return err
	}

	if df.Delivery_flag {
		return fmt.Errorf("cannot delete the topic for this destination as delivery flag is still active for ID %s", destination_id)
	}

	err2 := DeleteDestinationTopicMappingRepository(ctx, destination_id, topic_id)

	return err2
}
