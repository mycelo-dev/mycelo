ALTER TABLE destination_topic_mapping
ADD COLUMN retry_base_delay_ms BIGINT NOT NULL DEFAULT 2000,
ADD COLUMN retry_max_delay_ms BIGINT NOT NULL DEFAULT 60000,
ADD COLUMN max_consecutive_failures_before_skip INTEGER NOT NULL DEFAULT 0,
ADD COLUMN dead_letter_queue_enabled BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN skip_on_endpoint_4xx BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN skip_on_endpoint_5xx BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN skip_on_endpoint_transport_error BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN skip_on_event_payload_error BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN last_skipped_event_id BIGINT NOT NULL DEFAULT 0,
ADD COLUMN last_skipped_at BIGINT NOT NULL DEFAULT 0;
