package queries

func GetEventsAfterCursorQuery() string {
	return `
		SELECT topic, event_data, created_at, id
		FROM events
		WHERE topic = $1
		AND (created_at, id) > ($2, $3)
		ORDER BY created_at ASC, id ASC
	`
}
