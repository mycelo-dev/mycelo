package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/outbound"
)

func TestBulkDeadLetterReplayRequiresConfirmation(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/dead_letter_events", strings.NewReader(`{"limit":25}`))
	rec := httptest.NewRecorder()

	outbound.ReplayDeadLetterEventsRoute(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestBulkDeadLetterReplayRejectsLimitAboveCap(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/dead_letter_events", strings.NewReader(`{"limit":26,"confirmation":"REPLAY_FILTERED_DLQ"}`))
	rec := httptest.NewRecorder()

	outbound.ReplayDeadLetterEventsRoute(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
