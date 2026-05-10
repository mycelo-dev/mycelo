package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/stream"
)

func TestPublishRejectsNonJSONContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/publish", strings.NewReader(`{"topic":"orders"}`))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	stream.Publish(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Publish returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestPublishRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/publish", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	stream.Publish(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Publish returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetEventsRequiresTopic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rr := httptest.NewRecorder()

	stream.GetEvents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("GetEvents returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetEventsRejectsInvalidAfter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?topic=orders&after=bad", nil)
	rr := httptest.NewRecorder()

	stream.GetEvents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("GetEvents returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetEventsRejectsInvalidOffset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?topic=orders&offset=bad", nil)
	rr := httptest.NewRecorder()

	stream.GetEvents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("GetEvents returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
