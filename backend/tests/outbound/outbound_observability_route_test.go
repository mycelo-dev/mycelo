package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/outbound"
)

func TestOutboundObservabilityRouteReturnsMyceloOnlyMetrics(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/observability/outbound", nil)
	rec := httptest.NewRecorder()

	outbound.OutboundObservabilityRoute(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, exists := body["delivery_success_total"]; !exists {
		t.Fatal("missing delivery_success_total")
	}
	if _, exists := body["delivery_failure_total"]; !exists {
		t.Fatal("missing delivery_failure_total")
	}
	if _, exists := body["delivery_lag_ms"]; !exists {
		t.Fatal("missing delivery_lag_ms")
	}
	if _, exists := body["delivery_attempt_duration_ms"]; !exists {
		t.Fatal("missing delivery_attempt_duration_ms")
	}
	if _, exists := body["memstats"]; exists {
		t.Fatal("response should not include raw Go memstats")
	}
}
