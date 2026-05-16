package insert_queries

// GetInsertEventsQueries inserts an event into the stream table.
func GetInsertEventsQueries() string {

	return `
			INSERT INTO 
			EVENTS 
			(tenant_public_id, team_public_id, topic, event_data, created_at)
			VALUES
			($1, $2, $3, $4, $5)
	`
}
