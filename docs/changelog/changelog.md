# Changelog

## v0.0.3
Release date - TBD
Status - In progress

### Added

- **Production-grade outbound delivery (multi-instance, bounded reads, webhook semantics):**
    - **Database (`2026-05-09-005-outbound-production-hardening.sql`):**
        - `destinations.webhook_signing_secret` for per-destination HMAC signing of webhook POST bodies.
        - `destination_topic_mapping.delivery_lease_holder` and `delivery_lease_expires_at` for distributed lease ownership so only one app instance drives delivery per mapping at a time.
        - Composite index `idx_events_topic_created_id` on `events (topic, created_at, id)` aligned with cursor reads; drops standalone `idx_topic` / `idx_created_at`.
        - Unique index `uq_dead_letter_mapping_event` on `(destination_public_id, topic_public_id, source_event_id)` so DLQ rows dedupe per source event.
    - **Bounded event reads:** `GetEventsAfterCursor` SQL uses `LIMIT $4`; responses include `has_more`. Hot topics no longer load unbounded batches into memory.
    - **`GET /events`:** optional query parameter `limit` (capped by server maximum).
    - **Outbound runtime env vars:**
        - `OUTBOUND_INSTANCE_ID` — stable lease holder id across restarts (recommended in multi-instance deployments).
        - `OUTBOUND_LEASE_TTL_MS` — lease duration; server floors this to at least HTTP delivery timeout + margin so leases do not expire mid-flight under slow endpoints.
        - `OUTBOUND_EVENT_BATCH_LIMIT` — max events fetched per outbound poll (also capped globally).
    - **Webhook HTTP delivery:** headers include `X-Mycelo-Delivery-Id`, `Idempotency-Key` (same value), `X-Mycelo-Event-Id`, `X-Mycelo-Attempt`, `X-Mycelo-Timestamp`, `X-Mycelo-Topic`, and `X-Mycelo-Signature` (`t=<unix_ms>,v1=<hmac_sha256>` over `<ms>.<deliveryId>.<body>`) when `webhook_signing_secret` is set.
    - **Transactional skip + DLQ:** when using the real DB pool, dead-letter insert and mapping cursor/skip update run in one transaction (`ApplyDeadLetterSkipInTx`). DLQ insert uses `ON CONFLICT DO NOTHING` against the unique constraint for safe retries.
    - **Lease fencing:** delivery state updates (`UPDATE destination_topic_mapping …`) require `delivery_lease_holder` to match the writer; writers that lost the lease get `ErrOutboundLeaseLost` and stop mutating state for that batch.
    - **Cursor takeover safety:** after loading mapping state from the DB, the consumer reconciles the in-memory event cursor with `last_delivered_event_id` so a new lease holder does not replay from a stale local offset.
    - **`outbound.ErrOutboundLeaseLost`** exported for callers/tests.
    - **Destination API:** `POST /update_destination` accepts optional `webhook_signing_secret` (omit = unchanged; set string including empty to clear).
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

- **Outbound consumer lifecycle:** sync uses a mutex-protected consumer map; when a worker goroutine exits, its key is removed so a later sync can restart it (fixes “zombie” entries blocking restart after fatal errors).
- **Lease release on shutdown:** lease cleanup uses a fresh `context.WithTimeout` inside `defer` when the consumer exits, so the 15s budget applies at shutdown rather than expiring seconds after process start.
- **Delivery success ordering:** local cursor advances only after a successful lease-holding delivery-state write, avoiding optimistic cursor ahead of durable state on lease loss.
- **Semantics:** end-to-end delivery remains **at-least-once** if an endpoint accepts the HTTP request but the lease holder cannot persist success afterward; webhook headers (`X-Mycelo-Delivery-Id`, event id, signature) support idempotent receivers.
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
- `POST /update_destination_delivery_flag` now accepts both `destination_id` and `topic_id` so delivery flag updates are scoped to an existing destination-topic mapping.
- New destination-topic mappings now initialize `last_delivered_event_id` from the assigned topic's latest event instead of the global latest event across all topics.
- Normal outbound consumer shutdown via context cancellation is no longer logged as a delivery failure.

### Internal

- **Outbound packages:** `DefaultHTTPDeliveryTimeout` (10s) defines default HTTP client timeout; minimum lease TTL is derived from it plus margin. `MappingStore` now passes lease holder into fenced updates and transactional DLQ paths.
- **Query changes:** `GetUpdateDestinationTopicMappingDeliveryStateQuery` appends `AND delivery_lease_holder = $15`; `GetInsertDeadLetterEventQuery` includes `ON CONFLICT … DO NOTHING`; `GetEventsAfterCursorQuery` adds `LIMIT $4`; `GetOutboundMappingStateQuery` selects `webhook_signing_secret`; `GetUpdateDestinationQuery` can set `webhook_signing_secret` when the client opts in.
- **Tests (under `backend/tests` only):** webhook signature and delivery header tests in `backend/tests/outbound/webhook_sign_test.go`; cursor reconciliation vs DB `last_delivered_event_id` in `backend/tests/outbound/outbound_cursor_reconcile_test.go`; query builder assertions updated for new SQL fragments.
- Added retry policy helpers in `backend/internal/retrypolicy`.
- Added tests for retry policy behavior, outbound HTTP delivery, dead-letter and skip behavior, route validation, query builders, API key parsing, and core helpers.
- Added migrations:
    - `2026-05-09-001-destination-topic-mapping-cursor.sql`
    - `2026-05-09-002-destination-topic-mapping-retry-state.sql`
    - `2026-05-09-003-destination-topic-mapping-delivery-policy.sql`
    - `2026-05-09-004-dead-letter-events.sql`
    - `2026-05-09-005-outbound-production-hardening.sql`

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
