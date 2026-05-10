package tests

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/mycelo-dev/mycelo/backend/outbound"
)

func TestBuildWebhookSignatureValueDeterministic(t *testing.T) {
	secret := "whsec_unit_test"
	body := []byte(`{"hello":"world"}`)
	deliveryID := "dly_4829ab"
	ts := int64(1_700_000_000_000)

	a := outbound.BuildWebhookSignatureValue(secret, deliveryID, ts, body)
	b := outbound.BuildWebhookSignatureValue(secret, deliveryID, ts, body)
	if a != b {
		t.Fatalf("signature not stable")
	}

	if a == "" {
		t.Fatal("expected non-empty signature")
	}

	if strings.Count(a, "t=") != 1 || !strings.Contains(a, ",v1=") {
		t.Fatalf("unexpected format: %s", a)
	}
}

func TestBuildWebhookSignatureValueEmptySecret(t *testing.T) {
	got := outbound.BuildWebhookSignatureValue("", "id", 1, []byte("{}"))
	if got != "" {
		t.Fatalf("expected empty signature, got %q", got)
	}
}

func TestHTTPDeliveryClientSignatureMatchesBodySent(t *testing.T) {
	secret := "whsec_hmac_test"
	body := []byte(`{"count":7}`)
	meta := &outbound.WebhookDeliveryMeta{
		EventID:       90210,
		DeliveryID:    "d01ffeed",
		Attempt:       2,
		SentAtUnixMs:  1700000999000,
		SigningSecret: secret,
		TopicName:     "orders.ready",
	}

	sig := outbound.BuildWebhookSignatureValue(secret, meta.DeliveryID, meta.SentAtUnixMs, body)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatInt(meta.SentAtUnixMs, 10)))
	_, _ = mac.Write([]byte{'.'})
	_, _ = mac.Write([]byte(meta.DeliveryID))
	_, _ = mac.Write([]byte{'.'})
	_, _ = mac.Write(body)
	want := "t=" + strconv.FormatInt(meta.SentAtUnixMs, 10) + ",v1=" + hex.EncodeToString(mac.Sum(nil))
	if sig != want {
		t.Fatalf("signature mismatch:\nwant %s\n got %s", want, sig)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		readBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !bytes.Equal(readBody, body) {
			t.Fatalf("unexpected body %s", string(readBody))
		}

		gotHeader := r.Header.Get(outbound.WebhookSignatureHeader)
		if gotHeader != want {
			t.Fatalf("header mismatch: got %q want %q", gotHeader, want)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cli := outbound.NewHTTPDeliveryClient(http.DefaultClient)
	_, err := cli.Deliver(context.Background(), srv.URL, body, meta)
	if err != nil {
		t.Fatalf("Deliver: %v", err)
	}
}
