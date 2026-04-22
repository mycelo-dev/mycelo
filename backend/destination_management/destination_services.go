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
