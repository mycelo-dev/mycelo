package outbound

import (
	"bytes"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func DeliverToHttp(endpoint string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
