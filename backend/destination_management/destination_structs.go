package destination_management

// CreateDestination carries the payload for destination creation.
type CreateDestination struct {
	Destination_name    string `json:"destination_name"`
	Destination_address string `json:"destination_address"`
}

// UpdateDestination carries the payload for destination updates.
type UpdateDestination struct {
	Destination_name     string  `json:"destination_name"`
	Destination_address  string  `json:"destination_address"`
	Id                   string  `json:"id"`
	WebhookSigningSecret *string `json:"webhook_signing_secret,omitempty"`
}

// DeliveryFlag wraps the persisted enablement flag for a destination.
type DeliveryFlag struct {
	Delivery_flag bool
}

// UpdateDeliveryFlag carries the payload for toggling outbound delivery.
type UpdateDeliveryFlag struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
	Delivery_flag  bool   `json:"delivery_flag"`
}

// DeleteDestination identifies the destination to delete.
type DeleteDestination struct {
	Id string `json:"id"`
}

// AssignTopicToDestination links a topic to a destination for outbound delivery.
type AssignTopicToDestination struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
	DestinationTopicMappingPolicyInput
}

// DestinationRecord is the read model returned by destination listing endpoints.
type DestinationRecord struct {
	Destination_id      string `json:"destination_id"`
	Destination_name    string `json:"destination_name"`
	Destination_address string `json:"destination_address"`
	Delivery_flag       bool   `json:"delivery_flag"`
}

// DestinationTopicMappingRecord is the read model for destination-topic mappings.
type DestinationTopicMappingRecord struct {
	Destination_id                       string `json:"destination_id"`
	Destination_name                     string `json:"destination_name"`
	Destination_address                  string `json:"destination_address"`
	Delivery_flag                        bool   `json:"delivery_flag"`
	Last_delivered_event_id              int64  `json:"last_delivered_event_id"`
	Retry_base_delay_ms                  int64  `json:"retry_base_delay_ms"`
	Retry_max_delay_ms                   int64  `json:"retry_max_delay_ms"`
	Max_consecutive_failures_before_skip int    `json:"max_consecutive_failures_before_skip"`
	Dead_letter_queue_enabled            bool   `json:"dead_letter_queue_enabled"`
	Skip_on_endpoint_4xx                 bool   `json:"skip_on_endpoint_4xx"`
	Skip_on_endpoint_5xx                 bool   `json:"skip_on_endpoint_5xx"`
	Skip_on_endpoint_transport_error     bool   `json:"skip_on_endpoint_transport_error"`
	Skip_on_event_payload_error          bool   `json:"skip_on_event_payload_error"`
	Last_attempted_event_id              int64  `json:"last_attempted_event_id"`
	Last_failed_event_id                 int64  `json:"last_failed_event_id"`
	Last_skipped_event_id                int64  `json:"last_skipped_event_id"`
	Consecutive_failure_count            int    `json:"consecutive_failure_count"`
	Last_attempted_at                    int64  `json:"last_attempted_at"`
	Last_succeeded_at                    int64  `json:"last_succeeded_at"`
	Last_failed_at                       int64  `json:"last_failed_at"`
	Last_skipped_at                      int64  `json:"last_skipped_at"`
	Next_attempt_at                      int64  `json:"next_attempt_at"`
	Last_error_category                  string `json:"last_error_category"`
	Last_error                           string `json:"last_error"`
	Topic_id                             string `json:"topic_id"`
	Topic_name                           string `json:"topic_name"`
}

// DeleteDestinationTopicMapping identifies the mapping to remove.
type DeleteDestinationTopicMapping struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
}
