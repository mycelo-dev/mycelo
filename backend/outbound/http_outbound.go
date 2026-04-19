package outbound

import (
	"bytes"
	"net/http"
	"time"
)

// read the specific topic from the stream and will deliver to the external endpoint
// here we can have two things as configurable - topic and external endpoint

func GetUrl() string {
	// for now, creating a local url variable where url is hardcoded
	// later we need to make it configurable
	url := "http://localhost:5000/events"
	return url
}

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
