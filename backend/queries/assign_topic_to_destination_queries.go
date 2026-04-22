package queries

func GetAssignTopicToDestinationQuery() string {

	return `
			INSERT INTO destination_topic_mapping
			(destination_public_id, topic_public_id)
			VALUES
			($1, $2)
	`
}
