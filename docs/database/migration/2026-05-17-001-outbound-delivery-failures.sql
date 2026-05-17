CREATE TABLE IF NOT EXISTS outbound_delivery_failures (
    delivery_failure_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    destination_public_id UUID NOT NULL REFERENCES destinations(destination_public_id),
    topic_public_id UUID NOT NULL REFERENCES topics(topic_public_id),
    source_event_id BIGINT NOT NULL REFERENCES events(id),
    endpoint VARCHAR(500) NOT NULL,
    failure_category TEXT NOT NULL,
    failure_reason TEXT NOT NULL,
    failure_count INTEGER NOT NULL,
    first_failed_at BIGINT NOT NULL,
    last_failed_at BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_outbound_delivery_failure_event
ON outbound_delivery_failures (destination_public_id, topic_public_id, source_event_id);

CREATE INDEX IF NOT EXISTS idx_outbound_delivery_failures_mapping_last_failed
ON outbound_delivery_failures (destination_public_id, topic_public_id, last_failed_at DESC);
