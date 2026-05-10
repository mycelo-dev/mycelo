package topics_management

// TopicRecord is the read model returned by topic listing endpoints.
type TopicRecord struct {
	Topic_id   string `json:"topic_id"`
	Topic_name string `json:"topic_name"`
}
