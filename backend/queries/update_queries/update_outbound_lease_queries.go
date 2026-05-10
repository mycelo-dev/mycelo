package update_queries

// GetClaimOutboundDeliveryLeaseQuery grants or renews a mapping lease when it is free or already held by the caller.
func GetClaimOutboundDeliveryLeaseQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET
			delivery_lease_holder = $3,
			delivery_lease_expires_at = $4
		WHERE destination_public_id = $1
		AND topic_public_id = $2
		AND (
			delivery_lease_expires_at <= $5
			OR delivery_lease_holder = $3
		)
	`
}

// GetReleaseOutboundDeliveryLeaseQuery clears a lease only if this instance owns it.
func GetReleaseOutboundDeliveryLeaseQuery() string {
	return `
		UPDATE destination_topic_mapping
		SET
			delivery_lease_holder = '',
			delivery_lease_expires_at = 0
		WHERE destination_public_id = $1
		AND topic_public_id = $2
		AND delivery_lease_holder = $3
	`
}
