package configs

import (
	"os"
	"strconv"
)

const (
	defaultOutboundRetryBaseDelayMilliseconds   = int64(2000)
	defaultOutboundRetryMaxDelayMilliseconds    = int64(60000)
	defaultOutboundMaxFailuresBeforeSkip        = 0
	defaultOutboundDeadLetterQueueEnabled       = true
	defaultOutboundSkipOnEndpoint4xx            = true
	defaultOutboundSkipOnEndpoint5xx            = false
	defaultOutboundSkipOnEndpointTransportError = false
	defaultOutboundSkipOnEventPayloadError      = true
)

// GetDBURL returns the database connection string from the environment.
func GetDBURL() string {
	return os.Getenv("DB_URL")
}

// GetOutboundRetryBaseDelayMilliseconds returns the default outbound base retry delay.
func GetOutboundRetryBaseDelayMilliseconds() int64 {
	return getEnvInt64("OUTBOUND_RETRY_BASE_DELAY_MS", defaultOutboundRetryBaseDelayMilliseconds)
}

// GetOutboundRetryMaxDelayMilliseconds returns the default outbound max retry delay.
func GetOutboundRetryMaxDelayMilliseconds() int64 {
	return getEnvInt64("OUTBOUND_RETRY_MAX_DELAY_MS", defaultOutboundRetryMaxDelayMilliseconds)
}

// GetOutboundMaxFailuresBeforeSkip returns the default max failure threshold before skipping.
func GetOutboundMaxFailuresBeforeSkip() int {
	return getEnvInt("OUTBOUND_MAX_FAILURES_BEFORE_SKIP", defaultOutboundMaxFailuresBeforeSkip)
}

// GetOutboundDeadLetterQueueEnabled returns whether DLQ is enabled by default.
func GetOutboundDeadLetterQueueEnabled() bool {
	return getEnvBool("OUTBOUND_DEAD_LETTER_QUEUE_ENABLED", defaultOutboundDeadLetterQueueEnabled)
}

// GetOutboundSkipOnEndpoint4xx returns whether endpoint 4xx failures are skippable by default.
func GetOutboundSkipOnEndpoint4xx() bool {
	return getEnvBool("OUTBOUND_SKIP_ON_ENDPOINT_4XX", defaultOutboundSkipOnEndpoint4xx)
}

// GetOutboundSkipOnEndpoint5xx returns whether endpoint 5xx failures are skippable by default.
func GetOutboundSkipOnEndpoint5xx() bool {
	return getEnvBool("OUTBOUND_SKIP_ON_ENDPOINT_5XX", defaultOutboundSkipOnEndpoint5xx)
}

// GetOutboundSkipOnEndpointTransportError returns whether transport failures are skippable by default.
func GetOutboundSkipOnEndpointTransportError() bool {
	return getEnvBool("OUTBOUND_SKIP_ON_ENDPOINT_TRANSPORT_ERROR", defaultOutboundSkipOnEndpointTransportError)
}

// GetOutboundSkipOnEventPayloadError returns whether event payload failures are skippable by default.
func GetOutboundSkipOnEventPayloadError() bool {
	return getEnvBool("OUTBOUND_SKIP_ON_EVENT_PAYLOAD_ERROR", defaultOutboundSkipOnEventPayloadError)
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
