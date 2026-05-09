package outbound

import (
	"bytes"
	"context"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// DeliveryResult captures the transport-level outcome needed by the consumer loop.
type DeliveryResult struct {
	StatusCode int
}

// HTTPDoer is the minimal HTTP client contract needed by delivery code.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPDeliveryClient sends outbound payloads over HTTP.
type HTTPDeliveryClient struct {
	client HTTPDoer
}

// NewHTTPDeliveryClient wraps an HTTP client behind the DeliveryClient contract.
func NewHTTPDeliveryClient(client HTTPDoer) HTTPDeliveryClient {
	return HTTPDeliveryClient{client: client}
}

// DeliverToHttp preserves the original helper for direct HTTP delivery calls.
func DeliverToHttp(endpoint string, data []byte) (*http.Response, error) {
	return deliverToHTTP(context.Background(), httpClient, endpoint, data)
}

// Deliver sends a JSON payload to the configured endpoint and returns its status code.
func (c HTTPDeliveryClient) Deliver(ctx context.Context, endpoint string, data []byte) (DeliveryResult, error) {
	resp, err := deliverToHTTP(ctx, c.client, endpoint, data)
	if err != nil {
		return DeliveryResult{}, err
	}
	defer resp.Body.Close()

	return DeliveryResult{StatusCode: resp.StatusCode}, nil
}

func deliverToHTTP(ctx context.Context, client HTTPDoer, endpoint string, data []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
