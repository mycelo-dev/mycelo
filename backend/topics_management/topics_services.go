package topics_management

import "context"

// CreateTopicServices creates a topic in the current tenant-team scope.
func CreateTopicServices(ctx context.Context, topic_name string) error {
	err := CreateTopicRepository(ctx, topic_name)
	return err
}

// UpdateTopicServices renames a topic within the current tenant-team scope.
func UpdateTopicServices(ctx context.Context, old_topic_name string, new_topic_name string) error {
	err := UpdateTopicRepository(ctx, old_topic_name, new_topic_name)
	return err
}

// ListTopicsServices returns all topics for the current tenant-team scope.
func ListTopicsServices(ctx context.Context) ([]TopicRecord, error) {
	return ListTopicsRepository(ctx)
}
