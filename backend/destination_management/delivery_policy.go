package destination_management

import "github.com/mycelo-dev/mycelo/backend/configs"

// PolicyValidationError reports an invalid customer-supplied delivery control.
type PolicyValidationError struct {
	Message string
}

func (e PolicyValidationError) Error() string {
	return e.Message
}

// DestinationTopicMappingPolicy contains the resolved per-mapping delivery controls.
type DestinationTopicMappingPolicy struct {
	Retry_base_delay_ms                  int64 `json:"retry_base_delay_ms"`
	Retry_max_delay_ms                   int64 `json:"retry_max_delay_ms"`
	Max_consecutive_failures_before_skip int   `json:"max_consecutive_failures_before_skip"`
	Dead_letter_queue_enabled            bool  `json:"dead_letter_queue_enabled"`
	Skip_on_endpoint_4xx                 bool  `json:"skip_on_endpoint_4xx"`
	Skip_on_endpoint_5xx                 bool  `json:"skip_on_endpoint_5xx"`
	Skip_on_endpoint_transport_error     bool  `json:"skip_on_endpoint_transport_error"`
	Skip_on_event_payload_error          bool  `json:"skip_on_event_payload_error"`
}

// DestinationTopicMappingPolicyInput carries optional per-field policy overrides from API requests.
type DestinationTopicMappingPolicyInput struct {
	Retry_base_delay_ms                  *int64 `json:"retry_base_delay_ms"`
	Retry_max_delay_ms                   *int64 `json:"retry_max_delay_ms"`
	Max_consecutive_failures_before_skip *int   `json:"max_consecutive_failures_before_skip"`
	Dead_letter_queue_enabled            *bool  `json:"dead_letter_queue_enabled"`
	Skip_on_endpoint_4xx                 *bool  `json:"skip_on_endpoint_4xx"`
	Skip_on_endpoint_5xx                 *bool  `json:"skip_on_endpoint_5xx"`
	Skip_on_endpoint_transport_error     *bool  `json:"skip_on_endpoint_transport_error"`
	Skip_on_event_payload_error          *bool  `json:"skip_on_event_payload_error"`
}

// UpdateDestinationTopicMappingPolicy identifies the mapping and policy fields to update.
type UpdateDestinationTopicMappingPolicy struct {
	Destination_id string `json:"destination_id"`
	Topic_id       string `json:"topic_id"`
	DestinationTopicMappingPolicyInput
}

// DefaultDestinationTopicMappingPolicy returns the system defaults applied to new mappings.
func DefaultDestinationTopicMappingPolicy() DestinationTopicMappingPolicy {
	return DestinationTopicMappingPolicy{
		Retry_base_delay_ms:                  configs.GetOutboundRetryBaseDelayMilliseconds(),
		Retry_max_delay_ms:                   configs.GetOutboundRetryMaxDelayMilliseconds(),
		Max_consecutive_failures_before_skip: configs.GetOutboundMaxFailuresBeforeSkip(),
		Dead_letter_queue_enabled:            configs.GetOutboundDeadLetterQueueEnabled(),
		Skip_on_endpoint_4xx:                 configs.GetOutboundSkipOnEndpoint4xx(),
		Skip_on_endpoint_5xx:                 configs.GetOutboundSkipOnEndpoint5xx(),
		Skip_on_endpoint_transport_error:     configs.GetOutboundSkipOnEndpointTransportError(),
		Skip_on_event_payload_error:          configs.GetOutboundSkipOnEventPayloadError(),
	}
}

// Apply merges optional policy overrides onto an existing policy.
func (p DestinationTopicMappingPolicy) Apply(input DestinationTopicMappingPolicyInput) DestinationTopicMappingPolicy {
	if input.Retry_base_delay_ms != nil {
		p.Retry_base_delay_ms = *input.Retry_base_delay_ms
	}
	if input.Retry_max_delay_ms != nil {
		p.Retry_max_delay_ms = *input.Retry_max_delay_ms
	}
	if input.Max_consecutive_failures_before_skip != nil {
		p.Max_consecutive_failures_before_skip = *input.Max_consecutive_failures_before_skip
	}
	if input.Dead_letter_queue_enabled != nil {
		p.Dead_letter_queue_enabled = *input.Dead_letter_queue_enabled
	}
	if input.Skip_on_endpoint_4xx != nil {
		p.Skip_on_endpoint_4xx = *input.Skip_on_endpoint_4xx
	}
	if input.Skip_on_endpoint_5xx != nil {
		p.Skip_on_endpoint_5xx = *input.Skip_on_endpoint_5xx
	}
	if input.Skip_on_endpoint_transport_error != nil {
		p.Skip_on_endpoint_transport_error = *input.Skip_on_endpoint_transport_error
	}
	if input.Skip_on_event_payload_error != nil {
		p.Skip_on_event_payload_error = *input.Skip_on_event_payload_error
	}

	return p
}

// Validate enforces safe, non-contradictory delivery controls.
func (p DestinationTopicMappingPolicy) Validate() error {
	if p.Retry_base_delay_ms <= 0 {
		return PolicyValidationError{Message: "retry_base_delay_ms must be greater than zero"}
	}
	if p.Retry_max_delay_ms <= 0 {
		return PolicyValidationError{Message: "retry_max_delay_ms must be greater than zero"}
	}
	if p.Retry_max_delay_ms < p.Retry_base_delay_ms {
		return PolicyValidationError{Message: "retry_max_delay_ms must be greater than or equal to retry_base_delay_ms"}
	}
	if p.Max_consecutive_failures_before_skip < 0 {
		return PolicyValidationError{Message: "max_consecutive_failures_before_skip cannot be negative"}
	}

	return nil
}
