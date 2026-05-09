package update_queries

func GetUpdateDestinationTopicMappingCursorQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET last_delivered_event_id = $3
		WHERE destination_public_id = $1
		AND topic_public_id = $2
	`
}
