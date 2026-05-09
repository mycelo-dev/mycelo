ALTER TABLE destination_topic_mapping
ADD COLUMN last_delivered_event_id BIGINT NOT NULL DEFAULT 0;

UPDATE destination_topic_mapping
SET last_delivered_event_id = COALESCE((SELECT MAX(id) FROM events), 0);
