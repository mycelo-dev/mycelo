package queries

func GetInsertEventsQueries() string {

	return `
			INSERT INTO 
			EVENTS 
			(topic, event_data, created_at)
			VALUES
			($1, $2, $3)							
	`
}
