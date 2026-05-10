-- Outbound delivery: signing secret, distributed lease, event read index, DLQ deduplication.

ALTER TABLE destinations
ADD COLUMN IF NOT EXISTS webhook_signing_secret TEXT NOT NULL DEFAULT '';

ALTER TABLE destination_topic_mapping
ADD COLUMN IF NOT EXISTS delivery_lease_holder TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS delivery_lease_expires_at BIGINT NOT NULL DEFAULT 0;

-- Replace single-column indexes with one that matches fan-out reads: topic + ordering.
DROP INDEX IF EXISTS idx_topic;
DROP INDEX IF EXISTS idx_created_at;
CREATE INDEX IF NOT EXISTS idx_events_topic_created_id ON events (topic, created_at, id);

-- One DLQ row per (destination, topic, source event); retries use ON CONFLICT DO NOTHING.
CREATE UNIQUE INDEX IF NOT EXISTS uq_dead_letter_mapping_event
ON dead_letter_events (destination_public_id, topic_public_id, source_event_id);
