package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	topics_management "github.com/mycelo-dev/mycelo/backend/topics_management"
)

func TestCreateTopicRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/create_topic", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	topics_management.CreateTopicRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("CreateTopicRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateTopicRouteRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update_topic", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	topics_management.UpdateTopicRoute(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("UpdateTopicRoute returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
