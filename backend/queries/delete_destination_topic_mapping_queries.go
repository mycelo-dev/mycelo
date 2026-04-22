package queries

func GetDeleteDestinationTopicMappingQuery() string {
	return `
			DELETE FROM destination_topic_mapping
			WHERE 
			destination_public_id = $1 and topic_public_id = $2
	`
}
