CREATE TABLE dead_letter_events (
    dead_letter_event_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    destination_public_id UUID NOT NULL REFERENCES destinations(destination_public_id),
    topic_public_id UUID NOT NULL REFERENCES topics(topic_public_id),
    source_event_id BIGINT NOT NULL REFERENCES events(id),
    endpoint VARCHAR(500) NOT NULL,
    failure_category TEXT NOT NULL,
    failure_reason TEXT NOT NULL,
    failure_count INTEGER NOT NULL,
    event_payload JSONB NOT NULL,
    dead_lettered_at BIGINT NOT NULL
)
;
