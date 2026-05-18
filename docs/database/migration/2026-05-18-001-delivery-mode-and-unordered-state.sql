ALTER TABLE destination_topic_mapping
ADD COLUMN IF NOT EXISTS delivery_mode TEXT NOT NULL DEFAULT 'ordered',
ADD COLUMN IF NOT EXISTS unordered_last_enqueued_event_id BIGINT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS unordered_max_in_flight INTEGER NOT NULL DEFAULT 8;

UPDATE destination_topic_mapping
SET unordered_last_enqueued_event_id = last_delivered_event_id
WHERE unordered_last_enqueued_event_id = 0;

ALTER TABLE destination_topic_mapping
DROP CONSTRAINT IF EXISTS chk_destination_topic_mapping_delivery_mode;

ALTER TABLE destination_topic_mapping
ADD CONSTRAINT chk_destination_topic_mapping_delivery_mode
CHECK (delivery_mode IN ('ordered', 'unordered'));

CREATE TABLE IF NOT EXISTS outbound_event_deliveries (
    delivery_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    destination_public_id UUID NOT NULL REFERENCES destinations(destination_public_id),
    topic_public_id UUID NOT NULL REFERENCES topics(topic_public_id),
    source_event_id BIGINT NOT NULL REFERENCES events(id),
    status TEXT NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at BIGINT NOT NULL DEFAULT 0,
    locked_by TEXT NOT NULL DEFAULT '',
    lock_expires_at BIGINT NOT NULL DEFAULT 0,
    last_attempted_at BIGINT NOT NULL DEFAULT 0,
    delivered_at BIGINT NOT NULL DEFAULT 0,
    skipped_at BIGINT NOT NULL DEFAULT 0,
    last_error_category TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

ALTER TABLE outbound_event_deliveries
DROP CONSTRAINT IF EXISTS chk_outbound_event_deliveries_status;

ALTER TABLE outbound_event_deliveries
ADD CONSTRAINT chk_outbound_event_deliveries_status
CHECK (status IN ('pending', 'in_flight', 'delivered', 'failed', 'skipped'));

CREATE UNIQUE INDEX IF NOT EXISTS uq_outbound_event_delivery_mapping_event
ON outbound_event_deliveries (destination_public_id, topic_public_id, source_event_id);

CREATE INDEX IF NOT EXISTS idx_outbound_event_deliveries_claim
ON outbound_event_deliveries (destination_public_id, topic_public_id, status, next_attempt_at, lock_expires_at, source_event_id);

CREATE INDEX IF NOT EXISTS idx_events_tenant_team_topic_id
ON events (tenant_public_id, team_public_id, topic, id);
