# Changelog

## v0.0.3
Release date - TBD
Status - In progress

### Added

- Customer-configurable outbound delivery policy on each destination-topic mapping.
- Per-mapping policy fields:
    - `retry_base_delay_ms`
    - `retry_max_delay_ms`
    - `max_consecutive_failures_before_skip`
    - `dead_letter_queue_enabled`
    - `skip_on_endpoint_4xx`
    - `skip_on_endpoint_5xx`
    - `skip_on_endpoint_transport_error`
    - `skip_on_event_payload_error`
- Retry and delivery state on each mapping so operators can see what the consumer is doing and why it is blocked:
    - `last_delivered_event_id`
    - `last_attempted_event_id`
    - `last_failed_event_id`
    - `last_skipped_event_id`
    - `consecutive_failure_count`
    - `last_attempted_at`
    - `last_succeeded_at`
    - `last_failed_at`
    - `last_skipped_at`
    - `next_attempt_at`
    - `last_error_category`
    - `last_error`
- Dead-letter storage in `dead_letter_events` for skipped events, including:
    - source event id
    - destination and topic ids
    - endpoint
    - failure category and failure reason
    - failure count at time of skip
    - original event payload
    - dead-letter timestamp
- New API endpoints for managing and inspecting delivery policy:
    - `POST /update_destination_topic_mapping_policy`
    - `GET /dead_letter_events`
- Policy-aware request payloads in `POST /assign_topic_to_destination`, so customers can set mapping behavior when the mapping is created.
- Expanded `/destination_topic_mappings` responses with both delivery state and customer-configurable policy fields.
- A dedicated backend test suite under `backend/tests`, grouped into `auth`, `core`, `outbound`, `queries`, and `routes`.

### Changed

- Outbound consumers now use configurable capped exponential backoff with full jitter instead of hardcoded retry timing.
- Outbound delivery is still ordered per destination-topic mapping, but customers can now choose when a repeatedly failing event should be skipped so later events can continue.
- When the configured skip threshold is reached for an allowed failure category, the consumer can:
- write the failed event to the dead-letter queue
- advance the cursor past that event
- continue delivering later events for the same mapping
- Success, retry, failure, and skip attempts now update mapping state in the database instead of only advancing the cursor.
- Operators can now distinguish between endpoint-side failures and event-specific failures by checking `last_error_category` together with the last attempted, failed, and skipped event ids.
- Default policy behavior for new mappings is now driven by environment-backed configuration rather than only by constants in the consumer.
- Invalid JSON now returns `400` for the revoke API key route and the delete destination-topic mapping route.
- API key parsing helpers now safely handle malformed keys and read the hash segment from the correct position.

### Internal

- Added retry policy helpers in `backend/internal/retrypolicy`.
- Added tests for retry policy behavior, outbound HTTP delivery, dead-letter and skip behavior, route validation, query builders, API key parsing, and core helpers.
- Added migrations:
    - `2026-05-09-001-destination-topic-mapping-cursor.sql`
    - `2026-05-09-002-destination-topic-mapping-retry-state.sql`
    - `2026-05-09-003-destination-topic-mapping-delivery-policy.sql`
    - `2026-05-09-004-dead-letter-events.sql`

## v0.0.2
Release date - 09 may 2026

### Added

- Topics data model and topic management repository/service/route layers.
- Topic endpoints:
- `POST /create_topic`
- `POST /update_topic`
- `GET /topics`
- Destination data model and destination management repository/service/route layers.
- Destination endpoints:
- `POST /create_destination`
- `POST /update_destination`
- `POST /update_destination_delivery_flag`
- `POST /delete_destination`
- `GET /destinations`
- Destination-topic mapping endpoints:
- `POST /assign_topic_to_destination`
- `GET /destination_topic_mappings`
- `POST /delete_topic_for_destination`
- API key data model plus API key create, revoke, and rotate flows.
- API key endpoints:
- `POST /create_api_key`
- `POST /revoke_api_key`
- `POST /rotate_api_key`
- API key authentication helpers for parsing keys, hashing incoming values, reading stored hashes, and comparing them.

### Changed

- Added `delivery_flag` to destinations so delivery can be enabled or disabled without deleting the destination.
- Added a unique constraint on `destination_topic_mapping(destination_public_id, topic_public_id)`.
- Prevented destination deletion and destination-topic mapping deletion while `delivery_flag` is still active.
- Standardized destination and topic listing queries to return sorted results.

### Internal

- Added migrations:
- `2026-04-20-001-topics-data-modelling.sql`
- `2026-04-21-001-destination-data-modelling.sql`
- `2026-04-22-001-api-key-data-model.sql`
- `2026-04-22-002-destinations-delivery-flag.sql`
- `2026-04-22-003-uq-constraint-destination-topic.sql`
- Auth is scaffolded in code but is not yet wired into the HTTP request flow.

## v0.0.1
Release date - 19 april 2026

### Added

- Initial event ingestion flow backed by the `events` table.
- Publish endpoint:
- `POST /publish`
- Event read endpoint with cursor-style query parameters:
- `GET /events`
- Event payload storage as JSON with `topic`, `event_data`, `created_at`, and database `id`.
- Ordered event reads using `created_at` and `id`.
- Initial outbound HTTP delivery support for sending stored events to external destinations.
- Database indexes for event lookup by topic and creation time.

### Changed

- Event publishing now marshals request payloads before persisting them.
- Event reads now support `topic`, `after`, and `offset` query parameters for incremental consumption.

### Internal

- Added migrations:
- `2026-04-09-001-events-table.sql`
- `2026-04-19-001-events-table-index.sql`
- Added the first route registration flow in `backend/routes/all_route_handle_requests.go`.
