# Changelog

## v0.0.4
Release date - 16 may 2026
Status - completed

### Added

- Next.js frontend scaffold under `frontend/`, ready for Vercel deployment.
- Next.js frontend runtime:
    - Uses Next.js `15.5.x`, React `19.0.0`, TypeScript `5.8.x`, and a private `mycelo-frontend` package with `dev`, `build`, `start`, and `lint` scripts.
    - Runs local development on `http://localhost:3001` so the Go API can keep using `http://localhost:3000`.
    - Adds `frontend/README.md` with local development and proxy setup instructions.
- Same-origin API proxy at `/api/mycelo/*`, backed by `MYCELO_API_BASE_URL`, so the console can call the Go API without browser CORS coupling.
    - Preserves the incoming request path and query string when forwarding to the Go backend.
    - Supports `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`.
    - Removes hop-by-hop request/response headers such as `host`, `connection`, `content-encoding`, and `transfer-encoding`.
    - Uses `cache: "no-store"` so console reads are not served from stale browser or server cache.
    - Returns `502 Backend unavailable at <base_url>` when the configured backend cannot be reached.
- Operator console screens:
    - Delivery state dashboard for destination-topic mappings, emphasizing `last_error`, failure category, backoff, failure count, cursor, and recent activity.
    - DLQ viewer with destination/topic filters and replay actions for one DLQ event or the filtered DLQ set.
    - Observability dashboard for `GET /observability/outbound`, including success/failure counters, DLQ write/replay counters, circuit metrics, delivery lag, attempt duration, and latest success freshness.
    - Topics list and create form.
    - Destinations list, create form, and edit form with webhook signing secret support.
    - Mapping management for assigning topics, toggling delivery, and editing core retry/DLQ policy fields.
    - Mapping management now exposes deletion for destination-topic mappings from the console table.
    - Event log per topic with cursor pagination through `GET /events`.
    - Account signup, team creation, and team API key create/revoke workflows.
- Production-grade operator console UI shell:
    - Replaced the original demo-like single-page surface with a two-region SaaS application layout: persistent operator navigation on authenticated console screens and a dedicated workspace for the active operational view.
    - Added grouped navigation metadata for monitor/configuration areas, with stable view identifiers, labels, section kickers, summaries, and short navigation tokens.
    - Added an authenticated overview dashboard as the default console landing view. The overview summarizes successful deliveries, failure totals, mappings that need attention, DLQ record count, retry-due count, disabled route count, and control-plane setup state.
    - Added overview navigation actions that deep-link operators into delivery health, DLQ recovery, topics, destinations, and mappings without requiring a single crammed page.
    - Added consistent empty states for tables and lists so zero-data states remain informative instead of rendering blank panels.
    - Added responsive behavior for desktop and mobile layouts, including compact mobile navigation, wrapped empty table messages, and full-width mobile actions.
    - Added clearer toast/status severity handling, including negative status detection for messages containing `failed`, `error`, `invalid`, `unauthorized`, or `denied`.
- Dedicated authentication UI:
    - Unauthenticated users now see a standalone auth screen instead of the operator console shell.
    - Login and signup are no longer rendered side-by-side.
    - The auth screen uses a single centered card with a segmented `Log in` / `Sign up` switch and only renders the active form.
    - The operator navbar, workspace status bar, delivery/dashboard panels, and account management surfaces are hidden until an account session is loaded.
- Request credential handling in the frontend:
    - Removed the visible sidebar API-key input and manual "apply key" workflow.
    - Team request keys are now stored only when issued from the account/team workflow and are used internally by the frontend request layer for stream/data-plane calls.
    - Operator-console session JWTs are stored only in an `httpOnly`, `Secure`, `SameSite=Strict` cookie; the frontend no longer stores or forwards session tokens from localStorage.
    - The console chrome displays request authorization status without exposing request credentials as a general user-entered setting.
    - Account copy now distinguishes team request credentials from operator login/session state.
- Event read API and console event log details:
    - `GET /events` now accepts `order=desc` to read newest-first event pages for operator event-log views.
    - `GET /console/events` exposes the same event-log read behavior through the JWT cookie plus `X-Mycelo-Team` console auth path.
    - `GET /console/event_topics` lists distinct topic names with stored events so the console event log can show events even when the topic was first seen through `POST /publish`.
    - Descending reads use an id cursor: `offset=0` returns the latest page, and the response `cursor` can be passed as the next `offset` to continue reading older events.
    - The default ascending read path still supports `after` plus `offset`, ordered by `created_at ASC, id ASC`.
    - Event reads are tenant/team scoped through `auth.AuthContext`, with SQL filters on `tenant_public_id`, `team_public_id`, and `topic`.
    - Event responses include `events`, `count`, `cursor`, and `has_more`.
    - `limit` is validated as a positive integer and capped by `MaxEventsFetchUpperBound`.
    - The operator console uses `order=desc`, `limit=50`, live refresh while viewing the latest page, and pauses live refresh when the operator pages into history.
- Authenticated HTTP flow:
    - Stream/data-plane routes (`POST /publish` and external `GET /events`) require an API key through `Authorization: Bearer <api_key>` or `X-API-Key`.
    - Console/control-plane routes use the `mycelo_session` JWT cookie plus `X-Mycelo-Team` active-team scope, validate that the selected team belongs to the signed-in user, and then populate `auth.AuthContext` for repository calls.
    - The backend parses `tenant_public_id` and `team_public_id` from API keys only on stream/data-plane routes.
    - `POST /signup` accepts `tenant_name`, `user_name`, `email`, and `password`, then creates the tenant, first user, default team, and first team request key.
    - `POST /signup` returns the first request key once so the console can start immediately after onboarding.
    - `POST /signup` and `POST /login` issue a 1-hour operator-console JWT in the `mycelo_session` cookie instead of returning a session token in JSON.
    - JWT signing uses `MYCELO_SESSION_JWT_SECRET`; rotating the secret invalidates existing operator-console sessions.
    - `POST /login` accepts `email` and `password`, verifies the stored password hash, and restores the existing tenant/user account context.
    - `POST /create_team` creates teams for the signed-in tenant user using the `mycelo_session` cookie.
    - `GET /teams` lists teams for the signed-in tenant user using the `mycelo_session` cookie.
    - `POST /create_api_key` creates an API key for a selected team using the `mycelo_session` cookie; it now returns `409 Conflict` if a key already exists instead of silently replacing it.
    - `POST /rotate_api_key` and `POST /revoke_api_key` operate on the current API key's tenant-team scope instead of accepting hardcoded or body-provided tenant/team values.
- Migration `2026-05-11-002-users-account-signup.sql`:
    - Adds `users` for signup records with user public ids, user names, and email addresses.
    - Stores only a salted password KDF hash for users.
    - Adds a tenant-scoped unique team-name constraint so each tenant can create multiple named teams cleanly.
- Migration `2026-05-11-004-users-password-hash.sql`:
    - Adds `users.password_hash` for environments that already applied the users migration before password support.
- Migration `2026-05-15-001-drop-account-sessions.sql`:
    - Drops `account_sessions` now that operator-console sessions are signed JWT cookies and no longer require per-request session table reads.
- Migration `2026-05-11-003-backfill-internal-scope-ids.sql`:
    - Backfills internal `tenant_id` and `team_id` on topics and destinations from their public scope columns.
- Public-ID multitenancy:
    - Topics, destinations, events, destination-topic mapping reads, DLQ reads, DLQ replay, and outbound event consumption now scope by `tenant_public_id` and `team_public_id`.
    - The application no longer passes internal `tenant_id` or `team_id` through auth or API code paths.
- Migration `2026-05-11-001-events-tenant-team-scope.sql`:
    - Adds `tenant_public_id` and `team_public_id` to `topics`, `destinations`, and `events`.
    - Backfills topic and destination public scope columns from existing internal ids for compatibility.
    - Recreates topic and destination uniqueness constraints on public tenant/team ids.
    - Adds public-scope indexes for topics, destinations, and ordered event reads.

### Changed

- The API key repository, service, and route layers no longer use hardcoded tenant/team public ids.
- The operator console signup flow now asks only for tenant name, user name, email, and password.
- Scoped topic, destination, event, DLQ, and outbound queries now filter directly on public tenant/team ids.
- `GET /events` changed from a single ascending cursor path to dual read modes:
    - ascending stream-style reads for forward consumption,
    - descending newest-first reads for operator inspection.
- The operator console now treats unauthenticated access as a separate auth experience, not as a console view.
- The account console no longer renders login/signup forms after the console shell has loaded. It is now limited to tenant identity, team creation, team listing, request-key issuance, request-key revocation, and sign-out.
- The frontend no longer lets an operator paste or manually apply an API key from the console chrome.
- Issued team request keys are stored locally only as stream/data-plane request credentials.
- Account/team/request-key creation and console control-plane calls now use the `mycelo_session` JWT cookie; external stream/data-plane access continues to use API keys.
- Topic and destination creation now populate both internal ids and public ids for tenant/team scope.
- Filtered DLQ replay now requires an explicit `REPLAY_FILTERED_DLQ` confirmation, caps bulk replay at 25 records, and rate-limits repeated bulk replay attempts per tenant-team scope.
- The event-log control now labels paused historical pagination as "live paused on history" and changes the latest-page action to "Resume live" when live refresh is paused.
- New password hashes use scrypt with per-password salts. Existing PBKDF2-SHA256 password hashes remain verifiable for compatibility.

### Auth Boundary Notes

- API keys now represent external/programmatic stream credentials.
- Operator console routes are authorized by the signed operator session plus an active team selection.
- The session-team guard reads the `mycelo_session` JWT cookie, requires `X-Mycelo-Team`, validates that the signed-in user belongs to the selected tenant/team, and constructs the existing `auth.AuthContext`.

### Documentation

- Updated the Postman collection so bearer auth is scoped to stream requests and console/control-plane requests carry `X-Mycelo-Team` from the captured team id.
- Added Postman variables for `{{base_url}}`, `{{api_key}}`, and generated account/team ids captured from setup responses.
- Added Postman requests for `POST /signup`, `POST /login`, `GET /teams`, `POST /create_team`, `POST /create_api_key`, `POST /rotate_api_key`, and `POST /revoke_api_key` matching the current account flow.

### Internal

- Added auth context helpers for carrying tenant/team public scope through request handling.
- Added request API-key extraction from `Authorization: Bearer` and `X-API-Key`.
- Added session-team route guard coverage so console/control-plane routes no longer depend on API-key authentication.
- Added account/session internals:
    - scrypt password hashing with legacy PBKDF2-SHA256 verification,
    - signed operator-console JWT creation,
    - session restoration from the hardened `mycelo_session` cookie.
- Query builder tests now assert public tenant/team scoping fragments for topics, destinations, mappings, events, DLQ, API keys, account lookup, team creation, and active-team validation.
- Route tests now cover signup validation, login validation, session requirements for API-key creation, session-team requirements for console routes, revoke auth-context requirements, and mux registration for account/team/key endpoints.
- Stream read helpers now expose tenant-scoped variants for both ascending and descending event reads, so route handlers and tests can exercise the query behavior without bypassing tenant/team scope.
- Playwright snapshots and screenshots were generated during frontend verification. These are verification artifacts, not runtime product features.

## v0.0.3
Release date - 10 May 2026
Status - Completed

### Added

- **Production-grade outbound delivery (multi-instance, bounded reads, webhook semantics):**
    - **Database (`2026-05-09-005-outbound-production-hardening.sql`):**
        - `destinations.webhook_signing_secret` stores an optional per-destination secret used only when sending outbound webhooks. If the secret is empty, the consumer still sends the webhook but omits the HMAC signature header.
        - `destination_topic_mapping.delivery_lease_holder` stores the current outbound worker identity for a destination-topic mapping. `delivery_lease_expires_at` stores the lease expiry in Unix milliseconds. Together these fields allow multiple application instances to run consumers while only one instance actively delivers a given mapping at any moment.
        - Composite index `idx_events_topic_created_id` on `events (topic, created_at, id)` matches the outbound cursor query shape. The migration drops the old standalone `idx_topic` and `idx_created_at` indexes so the database can use one ordered index for topic-filtered cursor scans.
        - Unique index `uq_dead_letter_mapping_event` on `(destination_public_id, topic_public_id, source_event_id)` prevents duplicate DLQ rows for the same source event on the same mapping. Retries and lease failovers can safely attempt the same insert because the query uses `ON CONFLICT DO NOTHING`.
    - **Distributed lease acquisition and renewal:**
        - Every consumer loop calls `ClaimOutboundDeliveryLease` before reading mapping state or delivering events.
        - A lease can be claimed when it is free, expired, or already held by the same `OUTBOUND_INSTANCE_ID`.
        - During long batches, the worker renews the lease before each event delivery by writing a fresh `delivery_lease_expires_at`.
        - If another worker has taken over the mapping, lease renewal returns `held=false`; the old worker stops mutating delivery state for that batch.
    - **Lease fencing on state writes:**
        - `UpdateOutboundMappingDeliveryState` writes success, retry, failure, and skip state only when `delivery_lease_holder = $15`.
        - If a worker loses the lease between HTTP delivery and database update, the update affects zero rows and returns `outbound.ErrOutboundLeaseLost`.
        - On lease loss, the worker stops processing the batch. This keeps stale workers from advancing cursors or clearing failures after another instance has taken ownership.
    - **Cursor reconciliation on takeover:**
        - Each loop reloads `GetOutboundMappingState` from the database after claiming the lease.
        - The in-memory cursor is reconciled upward to `last_delivered_event_id` before reading events.
        - This prevents a new lease holder from replaying from a stale local offset after failover or worker restart.
    - **Bounded event reads:**
        - `GetEventsAfterCursor` SQL now accepts `LIMIT $4` and orders by `created_at ASC, id ASC`.
        - `GET /events` accepts an optional `limit` query parameter. The server validates and caps this value.
        - The outbound consumer uses `OUTBOUND_EVENT_BATCH_LIMIT` to bound every poll. Invalid, missing, or out-of-range values fall back to the default.
        - Responses include `has_more` so clients can tell whether additional events remain after the returned page.
    - **Outbound runtime env vars:**
        - `OUTBOUND_INSTANCE_ID` — stable lease holder id across restarts (recommended in multi-instance deployments).
        - `OUTBOUND_REQUIRE_INSTANCE_ID` — fail startup when `OUTBOUND_INSTANCE_ID` is empty, for production/multi-instance deployments.
        - `OUTBOUND_LEASE_TTL_MS` — lease duration; server floors this to at least HTTP delivery timeout + margin so leases do not expire mid-flight under slow endpoints.
        - `OUTBOUND_EVENT_BATCH_LIMIT` — max events fetched per outbound poll (also capped globally).
        - `OUTBOUND_CIRCUIT_BREAKER_FAILURE_THRESHOLD` and `OUTBOUND_CIRCUIT_BREAKER_COOLDOWN_MS` — endpoint-level circuit breaker controls.
    - **Webhook HTTP delivery:**
        - Each delivery receives a generated delivery id.
        - The delivery id is sent in both `X-Mycelo-Delivery-Id` and `Idempotency-Key`, so receivers can deduplicate repeated attempts.
        - The consumer also sends `X-Mycelo-Event-Id`, `X-Mycelo-Attempt`, `X-Mycelo-Timestamp`, and `X-Mycelo-Topic`.
        - When `webhook_signing_secret` is set, `X-Mycelo-Signature` is added as `t=<unix_ms>,v1=<hmac_sha256>`.
        - The signature input is `<timestamp_ms>.<delivery_id>.<raw_body>`, binding the timestamp, delivery id, and exact payload bytes.
    - **Retry, failure classification, and skip policy:**
        - Endpoint HTTP 4xx responses are classified as `endpoint_response_4xx`.
        - Endpoint HTTP 5xx responses are classified as `endpoint_response_5xx`.
        - Non-2xx, non-4xx, non-5xx responses are classified as `endpoint_response_other`.
        - HTTP client/transport failures are classified as `endpoint_transport`.
        - Event serialization failures are classified as `event_payload`.
        - Retry delay uses capped exponential backoff with full jitter, using each mapping's `retry_base_delay_ms` and `retry_max_delay_ms`.
        - A mapping can skip a repeatedly failing event only when `max_consecutive_failures_before_skip` is reached and the matching `skip_on_*` flag is enabled.
    - **Transactional skip + DLQ:**
        - When the real database pool is available, `ApplyDeadLetterSkipInTx` inserts the DLQ row and advances the mapping cursor in one transaction.
        - If the DLQ insert succeeds but the fenced cursor update loses the lease, the transaction rolls back.
        - If the same source event is skipped more than once, `ON CONFLICT DO NOTHING` avoids duplicate DLQ rows while still allowing the cursor update to be retried safely.
        - If `dead_letter_queue_enabled=false`, the consumer skips the event by advancing delivery state without inserting a DLQ row.
    - **DLQ replay:**
        - `GET /dead_letter_events` lists recent DLQ records, optionally filtered by `destination_id`, `topic_id`, and `limit`.
        - `POST /dead_letter_events` re-enqueues DLQ payloads as fresh events on their original topic.
        - Request body supports `dead_letter_event_id`, `destination_id`, `topic_id`, and `limit`.
        - If `dead_letter_event_id > 0`, replay is limited to exactly that DLQ record.
        - If replaying a DLQ row whose payload is the outbound event envelope, replay extracts `event_data` so the stream does not double-wrap the event.
        - Replayed events are inserted into `events` with a new `created_at`, so normal outbound consumers pick them up through the existing stream pipeline.
        - Replay returns `202 Accepted` with `replayed_count`.
    - **Outbound metrics:**
        - `GET /observability/outbound` exposes a clean Mycelo-only JSON response for outbound delivery metrics.
        - `/debug/vars` remains available as raw Go `expvar` diagnostics and includes Go runtime fields such as `memstats`.
        - `delivery_success_total` increments after a successful HTTP response and successful durable delivery-state update path.
        - `delivery_failure_total.category.<category>` increments for transport and non-2xx endpoint failures by failure category.
        - `dead_letter_write_total` increments when a skipped event is written to the DLQ path.
        - `dead_letter_replay_total` increments for each DLQ row re-enqueued as a fresh event.
        - `circuit_opened_total.endpoint.<endpoint>` increments when an endpoint circuit opens.
        - `circuit_blocked_total.endpoint.<endpoint>` increments when a delivery attempt is blocked because the endpoint circuit is open.
        - `delivery_lag_ms_count`, `delivery_lag_ms_total`, `delivery_lag_ms_max`, and `delivery_lag_ms_last` track event age from source event `created_at` to successful outbound delivery. This includes queue/backlog time.
        - `delivery_attempt_duration_ms_count`, `delivery_attempt_duration_ms_total`, `delivery_attempt_duration_ms_max`, and `delivery_attempt_duration_ms_last` track the actual outbound HTTP call duration.
        - `delivery_success_last_at` records the Unix milliseconds timestamp of the most recent successful outbound delivery.
    - **Observability:**
        - Runtime outbound telemetry is now available through `GET /observability/outbound`.
        - The response intentionally excludes raw Go runtime fields such as `memstats`, `cmdline`, and other `expvar` internals.
        - The response includes `delivery_success_total`, `delivery_failure_total`, `dead_letter_write_total`, `dead_letter_replay_total`, `circuit_opened_total`, `circuit_blocked_total`, `delivery_lag_ms`, and `delivery_attempt_duration_ms`.
        - `delivery_failure_total` is grouped by category: `endpoint_response_4xx`, `endpoint_response_5xx`, `endpoint_response_other`, and `endpoint_transport`.
        - `delivery_lag_ms` includes `count`, `total`, `max`, `last`, and `average`; `last` is the clearest signal that deliveries are fresh again after backlog recovery.
        - `delivery_attempt_duration_ms` includes `count`, `total`, `max`, `last`, and `average`; `last` is the clearest signal that the endpoint is currently responding quickly.
        - `delivery_success_last_at` lets operators tell when the latest successful delivery happened.
        - Circuit metrics are grouped by endpoint URL. Endpoint keys appear after an endpoint circuit has opened or blocked a delivery.
        - `/debug/vars` remains registered for low-level Go runtime diagnostics and exposes the raw `mycelo_outbound` / `mycelo_outbound_help` expvar objects alongside Go's default runtime data.
        - Mapping-level delivery state remains in `destination_topic_mapping`, while process-level counters live in memory through `expvar`.
        - This release provides scrapeable runtime metrics and database-visible delivery state, but does not yet include a bundled alert manager, dashboard, or external metrics backend integration.
        - Suggested alerts can now be built around sustained growth in `delivery_failure_total`, rising `dead_letter_write_total`, non-zero `circuit_opened_total`, increasing `circuit_blocked_total`, high `delivery_lag_ms.last`, stale `delivery_success_last_at`, or high `delivery_attempt_duration_ms.last`.
    - **Endpoint circuit breaker:**
        - Circuit state is held in memory per application instance and keyed by destination endpoint URL.
        - Failures that count toward the circuit are transport errors, 5xx responses, and other non-2xx/non-4xx endpoint responses.
        - 4xx responses do not open the circuit because they usually represent request or receiver validation issues for a specific event.
        - The default failure threshold is 5 consecutive counted endpoint failures.
        - The default cooldown is 30 seconds.
        - Once open, the circuit prevents additional HTTP delivery attempts to the same endpoint across all topic mappings in that process.
        - Blocked attempts are recorded as `endpoint_circuit_open` in mapping delivery state and scheduled for the remaining circuit cooldown.
        - Circuit-open blocks do not increment the mapping's consecutive failure count, so an endpoint recovering from downtime is not held behind an extra exponential backoff delay.
        - A successful delivery clears the endpoint circuit state.
        - Because circuit state is in memory, each app instance keeps its own breaker state; the distributed lease still controls mapping ownership.
    - **`outbound.ErrOutboundLeaseLost`** exported for callers/tests.
    - **Destination API:** `POST /update_destination` accepts optional `webhook_signing_secret` (omit = unchanged; set string including empty to clear).
- Customer-configurable outbound delivery policy on each destination-topic mapping.
- Per-mapping policy fields:
    - `retry_base_delay_ms` - base retry window used by capped exponential backoff.
    - `retry_max_delay_ms` - maximum retry window after backoff growth.
    - `max_consecutive_failures_before_skip` - number of consecutive failures required before a mapping is allowed to skip the blocking event. `0` disables threshold-based skipping.
    - `dead_letter_queue_enabled` - when true, skipped events are persisted to `dead_letter_events` before the cursor advances.
    - `skip_on_endpoint_4xx` - allows skip after threshold for endpoint 4xx responses.
    - `skip_on_endpoint_5xx` - allows skip after threshold for endpoint 5xx responses.
    - `skip_on_endpoint_transport_error` - allows skip after threshold for HTTP client/transport failures.
    - `skip_on_event_payload_error` - allows skip after threshold for event serialization/payload failures.
- Retry and delivery state on each mapping so operators can see what the consumer is doing and why it is blocked:
    - `last_delivered_event_id` - durable cursor; events at or below this id are considered delivered or intentionally skipped for this mapping.
    - `last_attempted_event_id` - most recent event id the consumer tried to deliver or skip.
    - `last_failed_event_id` - most recent event id that failed and remains blocking.
    - `last_skipped_event_id` - most recent event id skipped by policy.
    - `consecutive_failure_count` - number of consecutive failures for the currently blocking event; reset on success and after skip.
    - `last_attempted_at` - Unix milliseconds for the last attempt.
    - `last_succeeded_at` - Unix milliseconds for the last successful delivery.
    - `last_failed_at` - Unix milliseconds for the last failure.
    - `last_skipped_at` - Unix milliseconds for the last skip.
    - `next_attempt_at` - Unix milliseconds before which the consumer will wait instead of retrying.
    - `last_error_category` - machine-readable reason category such as `endpoint_response_4xx`, `endpoint_response_5xx`, `endpoint_transport`, `endpoint_circuit_open`, or `event_payload`.
    - `last_error` - human-readable error string from the last failed or skipped attempt.
- Dead-letter storage in `dead_letter_events` for skipped events, including:
    - source event id from the original `events` row
    - destination and topic ids for the affected mapping
    - endpoint that was being delivered to
    - failure category and failure reason
    - failure count at time of skip
    - original outbound payload, with replay support that extracts the original `event_data` when possible
    - dead-letter timestamp in Unix milliseconds
- New API endpoints for managing and inspecting delivery policy:
    - `POST /update_destination_topic_mapping_policy`
    - `GET /dead_letter_events`
    - `POST /dead_letter_events` (DLQ replay)
- Policy-aware request payloads in `POST /assign_topic_to_destination`, so customers can set mapping behavior when the mapping is created.
- Expanded `/destination_topic_mappings` responses with both delivery state and customer-configurable policy fields.
- A dedicated backend test suite under `backend/tests`, grouped into `auth`, `core`, `outbound`, `queries`, and `routes`.

### Changed

- **Outbound consumer lifecycle details:**
    - The service syncs active destination-topic mappings on a timer and starts one goroutine per mapping.
    - Active consumer cancel functions are stored in a mutex-protected map keyed by `destination_id:topic_id`.
    - When a mapping disappears, the service cancels its consumer and removes the key.
    - When a worker goroutine exits, it removes its own key so a later sync can restart it. This fixes stale in-memory zombie entries that previously could block restart after fatal errors.
- **Lease release on shutdown details:**
    - Consumer shutdown releases the mapping lease with a fresh `context.WithTimeout(context.Background(), 15*time.Second)`.
    - The timeout is created inside `defer`, so the 15-second budget starts at shutdown time instead of being consumed while the worker is running.
    - Release is fenced by holder id, so an old worker cannot clear another instance's lease.
- **Delivery success ordering details:**
    - The local in-memory cursor advances only after the success update is durably written while the worker still owns the lease.
    - If the HTTP request succeeds but the database update loses the lease, the cursor does not advance locally and the worker stops processing.
    - This avoids a stale worker believing it delivered an event that it did not persist.
- **Delivery semantics details:**
    - Outbound delivery remains at-least-once.
    - A receiver may accept the HTTP request and then the worker may fail to persist success because of lease loss or database failure.
    - In that case, a later worker can retry the same event.
    - Receivers should use `Idempotency-Key`, `X-Mycelo-Delivery-Id`, `X-Mycelo-Event-Id`, and `X-Mycelo-Signature` to deduplicate and verify retries.
- **Retry scheduling details:**
    - Outbound consumers now use configurable capped exponential backoff with full jitter instead of hardcoded retry timing.
    - The next retry timestamp is persisted in `next_attempt_at`, allowing operators to see when a blocked mapping will be retried.
    - On every loop, if `next_attempt_at` is in the future, the consumer sleeps until that time instead of hammering the endpoint.
- **Ordered delivery with policy-based skip details:**
    - Outbound delivery remains ordered per destination-topic mapping.
    - A failing event blocks later events for that mapping until it succeeds or the configured skip policy permits advancing past it.
    - When the configured skip threshold is reached for an enabled failure category, the consumer can write the failed event to DLQ, advance `last_delivered_event_id` to that event id, reset `consecutive_failure_count`, and continue later events.
- **Delivery state persistence details:**
    - Success, retry, failure, circuit-open, and skip attempts now update mapping state in the database instead of only advancing an in-memory cursor.
    - Operators can distinguish endpoint-side failures from event-specific failures by checking `last_error_category` together with the attempted, failed, skipped, and delivered event ids.
- **Default delivery policy details:**
    - Default policy behavior for new mappings is now driven by environment-backed configuration instead of only by constants in the consumer.
    - Explicit policy values in `POST /assign_topic_to_destination` override defaults for that mapping.
- Invalid JSON now returns `400` for the revoke API key route and the delete destination-topic mapping route.
- API key parsing helpers now safely handle malformed keys and read the hash segment from the correct position.
- `POST /update_destination_delivery_flag` now accepts both `destination_id` and `topic_id` so delivery flag updates are scoped to an existing destination-topic mapping.
- New destination-topic mappings now initialize `last_delivered_event_id` from the assigned topic's latest event instead of the global latest event across all topics.
- Normal outbound consumer shutdown via context cancellation is no longer logged as a delivery failure.

### Internal

- **Outbound package implementation details:**
    - `DefaultHTTPDeliveryTimeout` (10s) defines the default HTTP client timeout.
    - Minimum lease TTL is derived from HTTP timeout plus margin, so the configured lease cannot be shorter than a normal slow delivery attempt.
    - `MappingStore` now passes lease holder into fenced updates and transactional DLQ paths.
    - `ConsumerService` owns retry scheduling, lease renewal, cursor reconciliation, delivery state writes, DLQ skip handling, metrics updates, and endpoint circuit checks.
    - `EndpointCircuitBreaker` is process-local and keyed by endpoint URL. It protects the destination from repeated attempts within the same app instance while keeping the database lease model focused on mapping ownership.
    - `outbound_metrics.go` uses Go `expvar` rather than adding an external metrics dependency.
- **Query implementation details:**
    - `GetUpdateDestinationTopicMappingDeliveryStateQuery` appends `AND delivery_lease_holder = $15` for fencing.
    - `GetClaimOutboundDeliveryLeaseQuery` claims free/expired leases and renews leases already held by the caller.
    - `GetReleaseOutboundDeliveryLeaseQuery` clears lease fields only when the caller still owns the lease.
    - `GetInsertDeadLetterEventQuery` includes `ON CONFLICT (destination_public_id, topic_public_id, source_event_id) DO NOTHING`.
    - `GetEventsAfterCursorQuery` adds `LIMIT $4` for bounded reads.
    - `GetOutboundMappingStateQuery` selects `webhook_signing_secret`, retry policy, skip policy, delivery state, and the destination endpoint for each mapping.
    - `GetDeadLetterEventsForReplayQuery` reads bounded DLQ records in oldest-first order and joins topics to recover the original topic name for re-enqueue.
    - `GetUpdateDestinationQuery` can set `webhook_signing_secret` only when the client explicitly sends the optional field.
- **Route and documentation details:**
    - `/dead_letter_events` now supports `GET` for inspection and `POST` for replay.
    - `/observability/outbound` is registered as the operator-facing Mycelo outbound metrics endpoint.
    - `/debug/vars` is registered on the main mux using the standard `expvar` handler.
    - The Postman collection includes `GET /events?limit=100`, `POST /dead_letter_events`, `GET /observability/outbound`, `GET /debug/vars`, and `webhook_signing_secret` in the update destination example.
- **Test coverage details:**
    - Webhook signature and delivery header coverage lives in `backend/tests/outbound/webhook_sign_test.go`.
    - Cursor reconciliation vs database `last_delivered_event_id` is covered in `backend/tests/outbound/outbound_cursor_reconcile_test.go`.
    - Retry, failure recording, skip, DLQ write, and transport failure behavior are covered in `backend/tests/outbound/outbound_consumer_test.go`.
    - Endpoint circuit behavior is covered through the public consumer path in `backend/tests/outbound/outbound_circuit_breaker_test.go`.
    - Query builder assertions cover the new SQL fragments for bounded reads, DLQ replay reads, lease fencing, and policy fields.
    - Route registration tests cover `/dead_letter_events`, `/observability/outbound`, and `/debug/vars`.
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
