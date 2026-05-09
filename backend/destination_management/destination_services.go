package destination_management

import (
	"context"
	"fmt"
)

func CreateDestinationServices(ctx context.Context, destination_name string, destination_address string) error {
	return CreateDestinationRepository(ctx, destination_name, destination_address)
}

func UpdateDestinationServices(ctx context.Context, destination_name string, destination_address string, id string) error {
	return UpdateDestinationRepository(ctx, destination_name, destination_address, id)
}

func UpdateDeliveryFlagServices(ctx context.Context, id string, delivery_flag bool) error {
	return UpdateDeliveryFlagRepository(ctx, id, delivery_flag)
}

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

func AssignTopicToDestinationServices(ctx context.Context, destination_id string, topic_id string) error {

	return AssignTopicToDestinationRepository(ctx, destination_id, topic_id)
}

func ListDestinationsServices(ctx context.Context) ([]DestinationRecord, error) {
	return ListDestinationsRepository(ctx)
}

func ListDestinationTopicMappingsServices(ctx context.Context) ([]DestinationTopicMappingRecord, error) {
	return ListDestinationTopicMappingsRepository(ctx)
}

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
