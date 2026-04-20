package queries

func GetTopicsInsertQuery() string {
	return `
			INSERT INTO topics
			(tenant_id, team_id, topic_name, created_at, updated_at)
			VALUES( $1, $2, $3, $4, $5)
	`
}
