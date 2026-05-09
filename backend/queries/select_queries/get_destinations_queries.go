package select_queries

// GetDestinationsByTenantAndTeamQuery lists destinations for the current tenant-team scope.
func GetDestinationsByTenantAndTeamQuery() string {
	return `
		SELECT
			destination_public_id,
			destination_name,
			destination_address,
			delivery_flag
		FROM destinations
		WHERE tenant_id = $1
		AND team_id = $2
		ORDER BY destination_name ASC
	`
}
