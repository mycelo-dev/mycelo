package insert_queries

func GetAssignTopicToDestinationQuery() string {

	return `
			INSERT INTO destination_topic_mapping
			(destination_public_id, topic_public_id, last_delivered_event_id)
			VALUES
			($1, $2, COALESCE((SELECT MAX(id) FROM events), 0))
	`
}
