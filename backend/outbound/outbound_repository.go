package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mycelo-dev/mycelo/backend/auth"
	"github.com/mycelo-dev/mycelo/backend/core"
	"github.com/mycelo-dev/mycelo/backend/queries/insert_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/select_queries"
	"github.com/mycelo-dev/mycelo/backend/queries/update_queries"
)

// DestinationTopicMapping represents the cursor state for one destination-topic pair.
type DestinationTopicMapping struct {
	DestinationID        string
	TopicID              string
	TenantPublicID       string
	TeamPublicID         string
	LastDeliveredEventID int64
}

// OutboundMappingState contains the runtime delivery configuration for one mapping.
type OutboundMappingState struct {
	TopicName                        string
	Endpoint                         string
	WebhookSigningSecret             string
	DeliveryFlag                     bool
	LastDeliveredEventID             int64
	RetryBaseDelayMs                 int64
	RetryMaxDelayMs                  int64
	MaxConsecutiveFailuresBeforeSkip int
	DeadLetterQueueEnabled           bool
	SkipOnEndpoint4xx                bool
	SkipOnEndpoint5xx                bool
	SkipOnEndpointTransportError     bool
	SkipOnEventPayloadError          bool
	DeliveryMode                     string
	UnorderedMaxInFlight             int
	UnorderedLastEnqueuedEventID     int64
	LastAttemptedEventID             int64
	LastFailedEventID                int64
	LastSkippedEventID               int64
	ConsecutiveFailureCount          int
	LastSucceededAt                  int64
	LastFailedAt                     int64
	LastSkippedAt                    int64
	NextAttemptAt                    int64
	LastErrorCategory                string
	LastError                        string
	MappingExists                    bool
}

// DeliveryStateUpdate describes the delivery metadata to persist after an attempt.
type DeliveryStateUpdate struct {
	LastDeliveredEventID    int64
	LastAttemptedEventID    int64
	LastFailedEventID       int64
	LastSkippedEventID      int64
	ConsecutiveFailureCount int
	LastAttemptedAt         int64
	LastSucceededAt         int64
	LastFailedAt            int64
	LastSkippedAt           int64
	NextAttemptAt           int64
	LastErrorCategory       string
	LastError               string
}

// DeadLetterEventInsert describes one event that should be recorded in the dead-letter queue.
type DeadLetterEventInsert struct {
	DestinationID   string
	TopicID         string
	SourceEventID   int64
	Endpoint        string
	FailureCategory string
	FailureReason   string
	FailureCount    int
	EventPayload    []byte
	DeadLetteredAt  int64
}

// DeliveryFailureEventInsert describes a failed delivery attempt to expose in operator views.
type DeliveryFailureEventInsert struct {
	DestinationID   string
	TopicID         string
	SourceEventID   int64
	Endpoint        string
	FailureCategory string
	FailureReason   string
	FailureCount    int
	FailedAt        int64
}

// DeadLetterEventRecord is the API-facing shape returned by dead-letter reads.
type DeadLetterEventRecord struct {
	DeadLetterEventID int64           `json:"dead_letter_event_id"`
	DestinationID     string          `json:"destination_id"`
	DestinationName   string          `json:"destination_name"`
	TopicID           string          `json:"topic_id"`
	TopicName         string          `json:"topic_name"`
	SourceEventID     int64           `json:"source_event_id"`
	Endpoint          string          `json:"endpoint"`
	FailureCategory   string          `json:"failure_category"`
	FailureReason     string          `json:"failure_reason"`
	FailureCount      int             `json:"failure_count"`
	EventPayload      json.RawMessage `json:"event_payload"`
	DeadLetteredAt    int64           `json:"dead_lettered_at"`
}

// DeliveryFailureEventRecord is the API-facing shape returned by delivery failure reads.
type DeliveryFailureEventRecord struct {
	DeliveryFailureID int64  `json:"delivery_failure_id"`
	DestinationID     string `json:"destination_id"`
	DestinationName   string `json:"destination_name"`
	TopicID           string `json:"topic_id"`
	TopicName         string `json:"topic_name"`
	SourceEventID     int64  `json:"source_event_id"`
	Endpoint          string `json:"endpoint"`
	FailureCategory   string `json:"failure_category"`
	FailureReason     string `json:"failure_reason"`
	FailureCount      int    `json:"failure_count"`
	FirstFailedAt     int64  `json:"first_failed_at"`
	LastFailedAt      int64  `json:"last_failed_at"`
}

type deadLetterReplayRecord struct {
	DeadLetterEventID int64
	TopicName         string
	EventPayload      []byte
}

// DeadLetterReplayResult reports how many DLQ records were re-enqueued.
type DeadLetterReplayResult struct {
	ReplayedCount int `json:"replayed_count"`
}

// UnorderedDeliveryEventInsert describes an event discovered for unordered delivery.
type UnorderedDeliveryEventInsert struct {
	SourceEventID int64
}

// UnorderedDeliveryClaim is one independently claimable unordered delivery.
type UnorderedDeliveryClaim struct {
	DeliveryID    int64
	SourceEventID int64
	AttemptCount  int
	Topic         string
	EventPayload  json.RawMessage
	CreatedAt     int64
}

type outboundQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Repository reads and updates outbound delivery state using the configured database handle.
type Repository struct {
	db   outboundQueryer
	pool *pgxpool.Pool
}

// NewOutboundRepository builds a repository backed by the shared application database pool.
func NewOutboundRepository() *Repository {
	p := core.Get()
	return &Repository{db: p, pool: p}
}

// NewOutboundRepositoryWithDB builds a repository with an injected database dependency (tests / mocks).
func NewOutboundRepositoryWithDB(db outboundQueryer) *Repository {
	return &Repository{db: db, pool: nil}
}

// ApplyDeadLetterSkipInTx records a DLQ insert and advances the mapping cursor atomically when a pool is available.
func (r *Repository) ApplyDeadLetterSkipInTx(ctx context.Context, leaseHolder string, insert DeadLetterEventInsert, update DeliveryStateUpdate) error {
	if r.pool != nil {
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		if err := execInsertDeadLetterEvent(ctx, tx, insert); err != nil {
			return err
		}
		if err := execUpdateOutboundDeliveryState(ctx, tx, insert.DestinationID, insert.TopicID, leaseHolder, update); err != nil {
			return err
		}

		return tx.Commit(ctx)
	}

	if err := execInsertDeadLetterEvent(ctx, r.db, insert); err != nil {
		return err
	}

	return execUpdateOutboundDeliveryState(ctx, r.db, insert.DestinationID, insert.TopicID, leaseHolder, update)
}

// ClaimOutboundDeliveryLease attempts to steal a free/expired lease or renew a lease owned by holderID.
func (r *Repository) ClaimOutboundDeliveryLease(ctx context.Context, destinationID string, topicID string, holderID string, nowMillis int64, leaseExpiresAtMillis int64) (bool, error) {
	tag, err := r.db.Exec(
		ctx,
		update_queries.GetClaimOutboundDeliveryLeaseQuery(),
		destinationID,
		topicID,
		holderID,
		leaseExpiresAtMillis,
		nowMillis,
	)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() > 0, nil
}

// ReleaseOutboundDeliveryLease clears lease metadata when this holder still owns the mapping.
func (r *Repository) ReleaseOutboundDeliveryLease(ctx context.Context, destinationID string, topicID string, holderID string) error {
	_, err := r.db.Exec(
		ctx,
		update_queries.GetReleaseOutboundDeliveryLeaseQuery(),
		destinationID,
		topicID,
		holderID,
	)

	return err
}

// GetDestinationTopicMappings returns all active mapping cursors.
func (r *Repository) GetDestinationTopicMappings(ctx context.Context) ([]DestinationTopicMapping, error) {
	query := select_queries.GetOutboundMappingsQuery()

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := make([]DestinationTopicMapping, 0)

	for rows.Next() {
		var mapping DestinationTopicMapping

		if err := rows.Scan(&mapping.DestinationID, &mapping.TopicID, &mapping.TenantPublicID, &mapping.TeamPublicID, &mapping.LastDeliveredEventID); err != nil {
			return nil, err
		}

		mappings = append(mappings, mapping)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mappings, nil
}

// GetOutboundMappingState returns the current delivery state for one mapping.
func (r *Repository) GetOutboundMappingState(ctx context.Context, destinationID string, topicID string) (OutboundMappingState, error) {
	query := select_queries.GetOutboundMappingStateQuery()

	row := r.db.QueryRow(ctx, query, destinationID, topicID)

	var state OutboundMappingState

	err := row.Scan(
		&state.TopicName,
		&state.Endpoint,
		&state.WebhookSigningSecret,
		&state.DeliveryFlag,
		&state.LastDeliveredEventID,
		&state.RetryBaseDelayMs,
		&state.RetryMaxDelayMs,
		&state.MaxConsecutiveFailuresBeforeSkip,
		&state.DeadLetterQueueEnabled,
		&state.SkipOnEndpoint4xx,
		&state.SkipOnEndpoint5xx,
		&state.SkipOnEndpointTransportError,
		&state.SkipOnEventPayloadError,
		&state.DeliveryMode,
		&state.UnorderedMaxInFlight,
		&state.UnorderedLastEnqueuedEventID,
		&state.LastAttemptedEventID,
		&state.LastFailedEventID,
		&state.LastSkippedEventID,
		&state.ConsecutiveFailureCount,
		&state.LastSucceededAt,
		&state.LastFailedAt,
		&state.LastSkippedAt,
		&state.NextAttemptAt,
		&state.LastErrorCategory,
		&state.LastError,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundMappingState{MappingExists: false}, nil
		}

		return OutboundMappingState{}, err
	}

	state.MappingExists = true

	return state, nil
}

// UpdateOutboundMappingDeliveryState persists the latest delivery attempt state for a mapping when leaseHolder owns the outbound lease.
func (r *Repository) UpdateOutboundMappingDeliveryState(ctx context.Context, destinationID string, topicID string, leaseHolder string, update DeliveryStateUpdate) error {
	return execUpdateOutboundDeliveryState(ctx, r.db, destinationID, topicID, leaseHolder, update)
}

// InsertDeadLetterEvent persists a skipped event for later inspection or reprocessing.
func (r *Repository) InsertDeadLetterEvent(ctx context.Context, event DeadLetterEventInsert) error {
	return execInsertDeadLetterEvent(ctx, r.db, event)
}

// RecordDeliveryFailureEvent persists the latest failed delivery attempt for user-facing inspection.
func (r *Repository) RecordDeliveryFailureEvent(ctx context.Context, event DeliveryFailureEventInsert) error {
	return execInsertDeliveryFailureEvent(ctx, r.db, event)
}

func execInsertDeliveryFailureEvent(ctx context.Context, db outboundQueryer, event DeliveryFailureEventInsert) error {
	query := insert_queries.GetInsertOutboundDeliveryFailureQuery()

	_, err := db.Exec(
		ctx,
		query,
		event.DestinationID,
		event.TopicID,
		event.SourceEventID,
		event.Endpoint,
		event.FailureCategory,
		event.FailureReason,
		event.FailureCount,
		event.FailedAt,
	)

	return err
}

func execInsertDeadLetterEvent(ctx context.Context, db outboundQueryer, event DeadLetterEventInsert) error {
	query := insert_queries.GetInsertDeadLetterEventQuery()

	_, err := db.Exec(
		ctx,
		query,
		event.DestinationID,
		event.TopicID,
		event.SourceEventID,
		event.Endpoint,
		event.FailureCategory,
		event.FailureReason,
		event.FailureCount,
		event.EventPayload,
		event.DeadLetteredAt,
	)

	return err
}

// EnqueueUnorderedDeliveryEvents creates pending unordered delivery records for newly discovered events.
func (r *Repository) EnqueueUnorderedDeliveryEvents(ctx context.Context, destinationID string, topicID string, events []UnorderedDeliveryEventInsert) error {
	if len(events) == 0 {
		return nil
	}

	query := insert_queries.GetInsertOutboundEventDeliveryQuery()
	nowMillis := nowMillis()
	sourceEventIDs := make([]int64, 0, len(events))

	for _, event := range events {
		sourceEventIDs = append(sourceEventIDs, event.SourceEventID)
	}

	_, err := r.db.Exec(ctx, query, destinationID, topicID, sourceEventIDs, nowMillis)
	return err
}

// UpdateUnorderedEnqueueCursor advances the discovery cursor once events have been enqueued.
func (r *Repository) UpdateUnorderedEnqueueCursor(ctx context.Context, destinationID string, topicID string, leaseHolder string, cursor int64) error {
	tag, err := r.db.Exec(ctx, update_queries.GetUpdateUnorderedEnqueueCursorQuery(), destinationID, topicID, cursor, leaseHolder)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOutboundLeaseLost
	}

	return nil
}

// ClaimUnorderedDeliveryEvents claims independently retryable unordered deliveries.
func (r *Repository) ClaimUnorderedDeliveryEvents(ctx context.Context, destinationID string, topicID string, holderID string, nowMillis int64, lockExpiresAtMillis int64, limit int) ([]UnorderedDeliveryClaim, error) {
	rows, err := r.db.Query(
		ctx,
		select_queries.GetClaimOutboundEventDeliveriesQuery(),
		destinationID,
		topicID,
		nowMillis,
		holderID,
		lockExpiresAtMillis,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := make([]UnorderedDeliveryClaim, 0)
	for rows.Next() {
		var claim UnorderedDeliveryClaim

		if err := rows.Scan(
			&claim.DeliveryID,
			&claim.SourceEventID,
			&claim.AttemptCount,
			&claim.Topic,
			&claim.EventPayload,
			&claim.CreatedAt,
		); err != nil {
			return nil, err
		}

		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return claims, nil
}

// MarkUnorderedDeliveryDelivered records one unordered delivery success and aggregate state.
func (r *Repository) MarkUnorderedDeliveryDelivered(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, deliveredAt int64) error {
	tag, err := r.db.Exec(ctx, update_queries.GetMarkOutboundEventDeliveryDeliveredQuery(), deliveryID, holderID, deliveredAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOutboundLeaseLost
	}

	if _, err := r.db.Exec(ctx, update_queries.GetUpdateUnorderedDeliverySuccessStateQuery(), destinationID, topicID, sourceEventID, deliveredAt); err != nil {
		return err
	}

	return nil
}

// MarkUnorderedDeliveryFailed records one unordered delivery failure and retry schedule.
func (r *Repository) MarkUnorderedDeliveryFailed(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, nextAttemptAt int64, failureCategory string, failureReason string, failedAt int64) error {
	tag, err := r.db.Exec(ctx, update_queries.GetMarkOutboundEventDeliveryFailedQuery(), deliveryID, holderID, nextAttemptAt, failureCategory, failureReason, failedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOutboundLeaseLost
	}

	_, err = r.db.Exec(ctx, update_queries.GetUpdateUnorderedDeliveryFailureStateQuery(), destinationID, topicID, sourceEventID, failedAt, nextAttemptAt, failureCategory, failureReason)
	return err
}

// MarkUnorderedDeliverySkipped records one unordered delivery skip.
func (r *Repository) MarkUnorderedDeliverySkipped(ctx context.Context, deliveryID int64, destinationID string, topicID string, holderID string, sourceEventID int64, failureCategory string, failureReason string, skippedAt int64) error {
	tag, err := r.db.Exec(ctx, update_queries.GetMarkOutboundEventDeliverySkippedQuery(), deliveryID, holderID, skippedAt, failureCategory, failureReason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOutboundLeaseLost
	}

	return nil
}

// AdvanceUnorderedContiguousCursor advances last_delivered_event_id across contiguous delivered or skipped unordered rows.
func (r *Repository) AdvanceUnorderedContiguousCursor(ctx context.Context, destinationID string, topicID string) error {
	_, err := r.db.Exec(ctx, update_queries.GetAdvanceUnorderedContiguousCursorQuery(), destinationID, topicID)
	return err
}

func execUpdateOutboundDeliveryState(ctx context.Context, db outboundQueryer, destinationID string, topicID string, leaseHolder string, update DeliveryStateUpdate) error {
	query := update_queries.GetUpdateDestinationTopicMappingDeliveryStateQuery()

	tag, err := db.Exec(
		ctx,
		query,
		destinationID,
		topicID,
		update.LastDeliveredEventID,
		update.LastAttemptedEventID,
		update.LastFailedEventID,
		update.LastSkippedEventID,
		update.ConsecutiveFailureCount,
		update.LastAttemptedAt,
		update.LastSucceededAt,
		update.LastFailedAt,
		update.LastSkippedAt,
		update.NextAttemptAt,
		update.LastErrorCategory,
		update.LastError,
		leaseHolder,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrOutboundLeaseLost
	}

	return nil
}

// ListDeliveryFailureEvents returns recent failed delivery attempts, optionally filtered by mapping.
func (r *Repository) ListDeliveryFailureEvents(ctx context.Context, destinationID string, topicID string, limit int) ([]DeliveryFailureEventRecord, error) {
	authContext, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	query := select_queries.GetOutboundDeliveryFailuresQuery()

	rows, err := r.db.Query(ctx, query, destinationID, topicID, limit, authContext.TenantPublicId, authContext.TeamPublicId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]DeliveryFailureEventRecord, 0)
	for rows.Next() {
		var record DeliveryFailureEventRecord

		if err := rows.Scan(
			&record.DeliveryFailureID,
			&record.DestinationID,
			&record.DestinationName,
			&record.TopicID,
			&record.TopicName,
			&record.SourceEventID,
			&record.Endpoint,
			&record.FailureCategory,
			&record.FailureReason,
			&record.FailureCount,
			&record.FirstFailedAt,
			&record.LastFailedAt,
		); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// ListDeadLetterEvents returns recent dead-letter records, optionally filtered by mapping.
func (r *Repository) ListDeadLetterEvents(ctx context.Context, destinationID string, topicID string, limit int) ([]DeadLetterEventRecord, error) {
	authContext, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	query := select_queries.GetDeadLetterEventsQuery()

	rows, err := r.db.Query(ctx, query, destinationID, topicID, limit, authContext.TenantPublicId, authContext.TeamPublicId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]DeadLetterEventRecord, 0)
	for rows.Next() {
		var record DeadLetterEventRecord

		if err := rows.Scan(
			&record.DeadLetterEventID,
			&record.DestinationID,
			&record.DestinationName,
			&record.TopicID,
			&record.TopicName,
			&record.SourceEventID,
			&record.Endpoint,
			&record.FailureCategory,
			&record.FailureReason,
			&record.FailureCount,
			&record.EventPayload,
			&record.DeadLetteredAt,
		); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// ReplayDeadLetterEvents re-enqueues bounded DLQ payloads as fresh events on their original topics.
func (r *Repository) ReplayDeadLetterEvents(ctx context.Context, deadLetterEventID int64, destinationID string, topicID string, limit int) (DeadLetterReplayResult, error) {
	if limit <= 0 {
		limit = defaultDeadLetterEventsLimit
	}

	records, err := r.getDeadLetterEventsForReplay(ctx, deadLetterEventID, destinationID, topicID, limit)
	if err != nil {
		return DeadLetterReplayResult{}, err
	}

	for _, record := range records {
		if err := r.insertReplayedEvent(ctx, record.TopicName, record.EventPayload); err != nil {
			return DeadLetterReplayResult{}, err
		}
		incrementOutboundMetric("dead_letter_replay_total")
	}

	return DeadLetterReplayResult{ReplayedCount: len(records)}, nil
}

func (r *Repository) getDeadLetterEventsForReplay(ctx context.Context, deadLetterEventID int64, destinationID string, topicID string, limit int) ([]deadLetterReplayRecord, error) {
	authContext, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, select_queries.GetDeadLetterEventsForReplayQuery(), deadLetterEventID, destinationID, topicID, limit, authContext.TenantPublicId, authContext.TeamPublicId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]deadLetterReplayRecord, 0)
	for rows.Next() {
		var record deadLetterReplayRecord
		if err := rows.Scan(&record.DeadLetterEventID, &record.TopicName, &record.EventPayload); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) insertReplayedEvent(ctx context.Context, topicName string, eventPayload []byte) error {
	authContext, err := auth.FromContext(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, insert_queries.GetInsertEventsQueries(), authContext.TenantPublicId, authContext.TeamPublicId, topicName, eventDataForReplay(eventPayload), nowMillis())
	return err
}

func eventDataForReplay(eventPayload []byte) []byte {
	var envelope struct {
		EventData json.RawMessage `json:"event_data"`
	}
	if err := json.Unmarshal(eventPayload, &envelope); err == nil && len(envelope.EventData) > 0 {
		return envelope.EventData
	}

	return eventPayload
}

func nowMillis() int64 {
	return time.Now().UnixMilli()
}
