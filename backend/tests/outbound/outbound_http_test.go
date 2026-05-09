package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/outbound"
)

func TestDeliverToHttp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected application/json content type, got %q", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if string(body) != `{"hello":"world"}` {
			t.Fatalf("unexpected request body %q", string(body))
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	resp, err := outbound.DeliverToHttp(server.URL, []byte(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("DeliverToHttp returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("DeliverToHttp returned status %d, want %d", resp.StatusCode, http.StatusAccepted)
	}
}

func TestDeliverToHttpInvalidEndpoint(t *testing.T) {
	resp, err := outbound.DeliverToHttp("://bad-endpoint", []byte(`{}`))
	if err == nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		t.Fatal("DeliverToHttp expected error for invalid endpoint")
	}
}
