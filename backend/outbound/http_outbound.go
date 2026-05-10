package outbound

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"time"
)

// DefaultHTTPDeliveryTimeout bounds how long a single webhook POST waits for headers + body response.
const DefaultHTTPDeliveryTimeout = 10 * time.Second

var httpClient = &http.Client{Timeout: DefaultHTTPDeliveryTimeout}

// DeliveryResult captures the transport-level outcome needed by the consumer loop.
type DeliveryResult struct {
	StatusCode int
}

// WebhookDeliveryMeta carries signed-webhook semantics for receiver idempotency and verification.
type WebhookDeliveryMeta struct {
	EventID       int64
	DeliveryID    string
	Attempt       int
	SentAtUnixMs  int64
	SigningSecret string
	TopicName     string
}

// HTTPDoer is the minimal HTTP client contract needed by delivery code.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPDeliveryClient sends outbound payloads over HTTP.
type HTTPDeliveryClient struct {
	client HTTPDoer
}

// DeliveryClient abstracts the transport used to deliver outbound payloads (with optional webhook metadata).
type DeliveryClient interface {
	Deliver(ctx context.Context, endpoint string, payload []byte, meta *WebhookDeliveryMeta) (DeliveryResult, error)
}

// NewHTTPDeliveryClient wraps an HTTP client behind the DeliveryClient contract.
func NewHTTPDeliveryClient(client HTTPDoer) HTTPDeliveryClient {
	return HTTPDeliveryClient{client: client}
}

// DeliverToHttp preserves the legacy helper for direct HTTP delivery calls without webhook headers.
func DeliverToHttp(endpoint string, data []byte) (*http.Response, error) {
	return deliverHTTP(context.Background(), httpClient, endpoint, data, nil)
}

// Deliver sends an HTTP POST with Content-Type JSON and optional Mycelo webhook headers.
func (c HTTPDeliveryClient) Deliver(ctx context.Context, endpoint string, payload []byte, meta *WebhookDeliveryMeta) (DeliveryResult, error) {
	resp, err := deliverHTTP(ctx, c.client, endpoint, payload, meta)
	if err != nil {
		return DeliveryResult{}, err
	}
	defer resp.Body.Close()

	return DeliveryResult{StatusCode: resp.StatusCode}, nil
}

func deliverHTTP(ctx context.Context, client HTTPDoer, endpoint string, body []byte, meta *WebhookDeliveryMeta) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if meta != nil {
		if meta.DeliveryID != "" {
			req.Header.Set(webhookHeaderDeliveryID, meta.DeliveryID)
			req.Header.Set("Idempotency-Key", meta.DeliveryID)
		}

		req.Header.Set(webhookHeaderEventID, strconv.FormatInt(meta.EventID, 10))
		req.Header.Set(webhookHeaderAttempt, strconv.Itoa(meta.Attempt))
		req.Header.Set(webhookHeaderTimestamp, strconv.FormatInt(meta.SentAtUnixMs, 10))

		if meta.TopicName != "" {
			req.Header.Set("X-Mycelo-Topic", meta.TopicName)
		}

		sig := BuildWebhookSignatureValue(meta.SigningSecret, meta.DeliveryID, meta.SentAtUnixMs, body)
		if sig != "" {
			req.Header.Set(WebhookSignatureHeader, sig)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
