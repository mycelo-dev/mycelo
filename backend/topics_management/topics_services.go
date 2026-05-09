package topics_management

import "context"

func CreateTopicServices(ctx context.Context, topic_name string) error {
	err := CreateTopicRepository(ctx, topic_name)
	return err
}

func UpdateTopicServices(ctx context.Context, old_topic_name string, new_topic_name string) error {
	err := UpdateTopicRepository(ctx, old_topic_name, new_topic_name)
	return err
}

func ListTopicsServices(ctx context.Context) ([]TopicRecord, error) {
	return ListTopicsRepository(ctx)
}
