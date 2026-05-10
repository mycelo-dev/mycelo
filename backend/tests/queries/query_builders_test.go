package tests

import (
	"strings"
	"testing"

	delete_queries "github.com/mycelo-dev/mycelo/backend/queries/delete_queries"
	insert_queries "github.com/mycelo-dev/mycelo/backend/queries/insert_queries"
	select_queries "github.com/mycelo-dev/mycelo/backend/queries/select_queries"
	update_queries "github.com/mycelo-dev/mycelo/backend/queries/update_queries"
)

func TestInsertQueries(t *testing.T) {
	assertQueryContainsAll(t, insert_queries.GetAssignTopicToDestinationQuery(), "INSERT INTO destination_topic_mapping", "last_delivered_event_id", "SELECT MAX(e.id)", "WHERE t.topic_public_id = $2")
	assertQueryContainsAll(t, insert_queries.GetInsertApiKeyHashQuery(), "INSERT INTO api_keys", "key_hash", "VALUES($1, $2, $3, $4, $5)")
	assertQueryContainsAll(t, insert_queries.GetInsertDeadLetterEventQuery(), "INSERT INTO dead_letter_events", "failure_category", "event_payload", "ON CONFLICT")
	assertQueryContainsAll(t, insert_queries.GetInsertDestinationQuery(), "INSERT INTO destinations", "destination_name", "destination_address")
	assertQueryContainsAll(t, insert_queries.GetInsertEventsQueries(), "INSERT INTO", "EVENTS", "event_data")
	assertQueryContainsAll(t, insert_queries.GetTopicsInsertQuery(), "INSERT INTO topics", "topic_name", "tenant_id")
}

func TestSelectQueries(t *testing.T) {
	assertQueryContainsAll(t, select_queries.GetApiKeyHashFromDbQuery(), "SELECT key_hash", "FROM", "api_keys")
	assertQueryContainsAll(t, select_queries.GetDeadLetterEventsQuery(), "FROM dead_letter_events dle", "failure_reason", "LIMIT $3")
	assertQueryContainsAll(t, select_queries.GetReadDeliveryFlagByPublicIdQuery(), "SELECT delivery_flag", "FROM destinations")
	assertQueryContainsAll(t, select_queries.GetDestinationsByTenantAndTeamQuery(), "FROM destinations", "delivery_flag", "ORDER BY destination_name ASC")
	assertQueryContainsAll(t, select_queries.GetDestinationTopicMappingsByTenantAndTeamQuery(), "FROM destination_topic_mapping dtm", "retry_base_delay_ms", "dead_letter_queue_enabled", "last_skipped_event_id", "last_error_category", "topic_name")
	assertQueryContainsAll(t, select_queries.GetDestinationTopicMappingPolicyQuery(), "FROM destination_topic_mapping", "retry_base_delay_ms", "skip_on_event_payload_error")
	assertQueryContainsAll(t, select_queries.GetEventsAfterCursorQuery(), "FROM events", "created_at > $2", "ORDER BY created_at ASC, id ASC", "LIMIT $4")
	assertQueryContainsAll(t, select_queries.GetDeadLetterEventsForReplayQuery(), "FROM dead_letter_events dle", "event_payload", "topic_name", "LIMIT $4")
	assertQueryContainsAll(t, select_queries.GetOutboundMappingsQuery(), "FROM destination_topic_mapping dtm", "last_delivered_event_id")
	assertQueryContainsAll(t, select_queries.GetOutboundMappingStateQuery(), "destination_address", "webhook_signing_secret", "retry_base_delay_ms", "skip_on_endpoint_transport_error", "last_skipped_event_id", "last_error_category", "last_error")
	assertQueryContainsAll(t, select_queries.GetTopicsByTenantAndTeamQuery(), "FROM topics", "topic_public_id", "ORDER BY topic_name ASC")
}

func TestUpdateQueries(t *testing.T) {
	assertQueryContainsAll(t, update_queries.GetRotateApiKeyQuery(), "UPDATE api_keys", "SET key_hash = $1", "WHERE tenant_public_id = $2")
	assertQueryContainsAll(t, update_queries.GetUpdateDestinationQuery(), "UPDATE destinations", "destination_address = $2", "webhook_signing_secret", "WHERE destination_public_id = $4")
	assertQueryContainsAll(t, update_queries.GetUpdateDeliveryFlagQuery(), "UPDATE destinations", "delivery_flag = $1", "updated_at = $2", "FROM destination_topic_mapping", "topic_public_id = $4")
	assertQueryContainsAll(t, update_queries.GetUpdateDestinationTopicMappingCursorQuery(), "UPDATE destination_topic_mapping", "last_delivered_event_id = $3")
	assertQueryContainsAll(t, update_queries.GetUpdateDestinationTopicMappingDeliveryStateQuery(), "last_attempted_event_id = $4", "last_skipped_event_id = $6", "last_skipped_at = $11", "last_error_category = $13", "last_error = $14", "delivery_lease_holder = $15")
	assertQueryContainsAll(t, update_queries.GetUpdateDestinationTopicMappingPolicyQuery(), "retry_base_delay_ms = $3", "dead_letter_queue_enabled = $6", "skip_on_event_payload_error = $10")
	assertQueryContainsAll(t, update_queries.GetClaimOutboundDeliveryLeaseQuery(), "UPDATE destination_topic_mapping", "delivery_lease_holder")
	assertQueryContainsAll(t, update_queries.GetReleaseOutboundDeliveryLeaseQuery(), "delivery_lease_expires_at")
	assertQueryContainsAll(t, update_queries.GetQueryToUpdateTopic(), "UPDATE TOPICS", "SET topic_name = $1", "WHERE topic_name = $2")
}

func TestDeleteQueries(t *testing.T) {
	assertQueryContainsAll(t, delete_queries.GetRevokeApiKeyQuery(), "DELETE FROM api_keys", "tenant_public_id = $1", "team_public_id = $2")
	assertQueryContainsAll(t, delete_queries.GetDeleteDestinationQuery(), "DELETE FROM destinations", "delivery_flag = false")
	assertQueryContainsAll(t, delete_queries.GetDeleteDestinationTopicMappingQuery(), "DELETE FROM destination_topic_mapping", "destination_public_id = $1", "topic_public_id = $2")
}

func assertQueryContainsAll(t *testing.T, query string, parts ...string) {
	t.Helper()

	for _, part := range parts {
		if !strings.Contains(query, part) {
			t.Fatalf("query %q does not contain %q", query, part)
		}
	}
}
