package tests

import (
	"testing"

	"github.com/mycelo-dev/mycelo/backend/configs"
	"github.com/mycelo-dev/mycelo/backend/core"
)

func TestGetHashString(t *testing.T) {
	got := core.GetHashString("hello")

	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("GetHashString returned %q, want %q", got, want)
	}
}

func TestGetHexString(t *testing.T) {
	got := core.GetHexString([]byte{0xde, 0xad, 0xbe, 0xef})

	if got != "deadbeef" {
		t.Fatalf("GetHexString returned %q, want %q", got, "deadbeef")
	}
}

func TestGetRandomBytes(t *testing.T) {
	got, err := core.GetRandomBytes(32)
	if err != nil {
		t.Fatalf("GetRandomBytes returned error: %v", err)
	}

	if len(got) != 32 {
		t.Fatalf("GetRandomBytes returned %d bytes, want 32", len(got))
	}
}

func TestGetDBURL(t *testing.T) {
	t.Setenv("DB_URL", "postgres://example")

	got := configs.GetDBURL()
	if got != "postgres://example" {
		t.Fatalf("GetDBURL returned %q, want %q", got, "postgres://example")
	}
}

func TestOutboundConfigDefaultsAndOverrides(t *testing.T) {
	t.Setenv("OUTBOUND_RETRY_BASE_DELAY_MS", "3000")
	t.Setenv("OUTBOUND_RETRY_MAX_DELAY_MS", "45000")
	t.Setenv("OUTBOUND_MAX_FAILURES_BEFORE_SKIP", "7")
	t.Setenv("OUTBOUND_DEAD_LETTER_QUEUE_ENABLED", "false")
	t.Setenv("OUTBOUND_SKIP_ON_ENDPOINT_4XX", "false")
	t.Setenv("OUTBOUND_SKIP_ON_ENDPOINT_5XX", "true")
	t.Setenv("OUTBOUND_SKIP_ON_ENDPOINT_TRANSPORT_ERROR", "true")
	t.Setenv("OUTBOUND_SKIP_ON_EVENT_PAYLOAD_ERROR", "false")

	if got := configs.GetOutboundRetryBaseDelayMilliseconds(); got != 3000 {
		t.Fatalf("GetOutboundRetryBaseDelayMilliseconds returned %d, want %d", got, 3000)
	}
	if got := configs.GetOutboundRetryMaxDelayMilliseconds(); got != 45000 {
		t.Fatalf("GetOutboundRetryMaxDelayMilliseconds returned %d, want %d", got, 45000)
	}
	if got := configs.GetOutboundMaxFailuresBeforeSkip(); got != 7 {
		t.Fatalf("GetOutboundMaxFailuresBeforeSkip returned %d, want %d", got, 7)
	}
	if got := configs.GetOutboundDeadLetterQueueEnabled(); got != false {
		t.Fatalf("GetOutboundDeadLetterQueueEnabled returned %t, want %t", got, false)
	}
	if got := configs.GetOutboundSkipOnEndpoint4xx(); got != false {
		t.Fatalf("GetOutboundSkipOnEndpoint4xx returned %t, want %t", got, false)
	}
	if got := configs.GetOutboundSkipOnEndpoint5xx(); got != true {
		t.Fatalf("GetOutboundSkipOnEndpoint5xx returned %t, want %t", got, true)
	}
	if got := configs.GetOutboundSkipOnEndpointTransportError(); got != true {
		t.Fatalf("GetOutboundSkipOnEndpointTransportError returned %t, want %t", got, true)
	}
	if got := configs.GetOutboundSkipOnEventPayloadError(); got != false {
		t.Fatalf("GetOutboundSkipOnEventPayloadError returned %t, want %t", got, false)
	}
}
