package insert_queries

// GetInsertEventsQueries inserts an event into the stream table.
func GetInsertEventsQueries() string {

	return `
			INSERT INTO 
			EVENTS 
			(topic, event_data, created_at)
			VALUES
			($1, $2, $3)							
	`
}
