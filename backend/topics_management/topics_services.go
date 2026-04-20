package topics_management

import "context"

func CreateTopicServices(ctx context.Context, topic_name string) error {
	err := CreateTopicRepository(ctx, topic_name)
	return err
}
