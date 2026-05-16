export type Topic = {
  topic_id: string;
  topic_name: string;
};

export type Destination = {
  destination_id: string;
  destination_name: string;
  destination_address: string;
  delivery_flag: boolean;
};

export type Mapping = {
  destination_id: string;
  destination_name: string;
  destination_address: string;
  delivery_flag: boolean;
  last_delivered_event_id: number;
  retry_base_delay_ms: number;
  retry_max_delay_ms: number;
  max_consecutive_failures_before_skip: number;
  dead_letter_queue_enabled: boolean;
  skip_on_endpoint_4xx: boolean;
  skip_on_endpoint_5xx: boolean;
  skip_on_endpoint_transport_error: boolean;
  skip_on_event_payload_error: boolean;
  last_attempted_event_id: number;
  last_failed_event_id: number;
  last_skipped_event_id: number;
  consecutive_failure_count: number;
  last_attempted_at: number;
  last_succeeded_at: number;
  last_failed_at: number;
  last_skipped_at: number;
  next_attempt_at: number;
  last_error_category: string;
  last_error: string;
  topic_id: string;
  topic_name: string;
};

export type DeadLetterEvent = {
  dead_letter_event_id: number;
  destination_id: string;
  destination_name: string;
  topic_id: string;
  topic_name: string;
  source_event_id: number;
  endpoint: string;
  failure_category: string;
  failure_reason: string;
  failure_count: number;
  event_payload: unknown;
  dead_lettered_at: number;
};

export type StreamEvent = {
  id?: number;
  topic: string;
  event_data: unknown;
  created_at: number;
};

export type EventTopic = {
  topic_name: string;
};

export type EventsResponse = {
  events: StreamEvent[];
  count: number;
  cursor: number;
  has_more: boolean;
};

export type DurationMetrics = {
  count: number;
  total: number;
  max: number;
  last: number;
  average: number;
};

export type OutboundMetrics = {
  delivery_success_total: number;
  delivery_success_last_at: number;
  delivery_failure_total: Record<string, number>;
  dead_letter_write_total: number;
  dead_letter_replay_total: number;
  circuit_opened_total: Record<string, number>;
  circuit_blocked_total: Record<string, number>;
  delivery_lag_ms: DurationMetrics;
  delivery_attempt_duration_ms: DurationMetrics;
};

export type ReplayResult = {
  replayed_count: number;
};

export type ApiKeyResponse = {
  api_key: string;
};

export type SignUpResponse = {
  tenant_public_id: string;
  user_public_id: string;
  tenant_name: string;
  user_name: string;
  email: string;
  team_public_id?: string;
  team_name?: string;
};

export type AccountContext = SignUpResponse;

export type Team = {
  team_public_id: string;
  team_name: string;
};
