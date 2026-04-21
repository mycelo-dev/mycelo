package destination_management

import "context"

func CreateDestinationServices(ctx context.Context, destination_name string, destination_address string) error {
	return CreateDestinationRepository(ctx, destination_name, destination_address)
}
