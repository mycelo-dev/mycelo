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
func DeliverToHttp(topic string, external_http_endpoint string) (*http.Response, error) {
	var event_data []byte

	req, _ := http.NewRequest("POST", external_http_endpoint, bytes.NewBuffer(event_data))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		return resp, err
	}

	return resp, nil
}
