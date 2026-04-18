package queries

func GetEventsAfterCursorQuery() string {
	return `
		SELECT topic, event_data, created_at
		FROM events
		WHERE topic = $1
		AND created_at > $2
		ORDER BY created_at ASC
	`
}
