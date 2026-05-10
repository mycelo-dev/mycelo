package outbound

import (
	"encoding/json"
	"expvar"
	"net/http"
	"strconv"
	"strings"
)

type OutboundObservabilityResponse struct {
	DeliverySuccessTotal      int64            `json:"delivery_success_total"`
	DeliverySuccessLastAt     int64            `json:"delivery_success_last_at"`
	DeliveryFailureTotal      map[string]int64 `json:"delivery_failure_total"`
	DeadLetterWriteTotal      int64            `json:"dead_letter_write_total"`
	DeadLetterReplayTotal     int64            `json:"dead_letter_replay_total"`
	CircuitOpenedTotal        map[string]int64 `json:"circuit_opened_total"`
	CircuitBlockedTotal       map[string]int64 `json:"circuit_blocked_total"`
	DeliveryLagMs             DurationMetrics  `json:"delivery_lag_ms"`
	DeliveryAttemptDurationMs DurationMetrics  `json:"delivery_attempt_duration_ms"`
}

type DurationMetrics struct {
	Count   int64   `json:"count"`
	Total   int64   `json:"total"`
	Max     int64   `json:"max"`
	Last    int64   `json:"last"`
	Average float64 `json:"average"`
}

// OutboundObservabilityRoute returns only Mycelo outbound runtime metrics.
func OutboundObservabilityRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := buildOutboundObservabilityResponse()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func buildOutboundObservabilityResponse() OutboundObservabilityResponse {
	response := OutboundObservabilityResponse{
		DeliverySuccessTotal:  outboundMetricInt("delivery_success_total"),
		DeliverySuccessLastAt: outboundMetricInt("delivery_success_last_at"),
		DeliveryFailureTotal: map[string]int64{
			failureCategoryEndpoint4xx:    outboundMetricInt("delivery_failure_total.category." + failureCategoryEndpoint4xx),
			failureCategoryEndpoint5xx:    outboundMetricInt("delivery_failure_total.category." + failureCategoryEndpoint5xx),
			failureCategoryEndpointOther:  outboundMetricInt("delivery_failure_total.category." + failureCategoryEndpointOther),
			failureCategoryEndpointClient: outboundMetricInt("delivery_failure_total.category." + failureCategoryEndpointClient),
		},
		DeadLetterWriteTotal:      outboundMetricInt("dead_letter_write_total"),
		DeadLetterReplayTotal:     outboundMetricInt("dead_letter_replay_total"),
		CircuitOpenedTotal:        make(map[string]int64),
		CircuitBlockedTotal:       make(map[string]int64),
		DeliveryLagMs:             outboundDurationMetrics("delivery_lag_ms"),
		DeliveryAttemptDurationMs: outboundDurationMetrics("delivery_attempt_duration_ms"),
	}

	outboundMetrics.Do(func(kv expvar.KeyValue) {
		switch {
		case strings.HasPrefix(kv.Key, "circuit_opened_total.endpoint."):
			endpoint := strings.TrimPrefix(kv.Key, "circuit_opened_total.endpoint.")
			response.CircuitOpenedTotal[endpoint] = expvarIntValue(kv.Value)
		case strings.HasPrefix(kv.Key, "circuit_blocked_total.endpoint."):
			endpoint := strings.TrimPrefix(kv.Key, "circuit_blocked_total.endpoint.")
			response.CircuitBlockedTotal[endpoint] = expvarIntValue(kv.Value)
		}
	})

	return response
}

func outboundDurationMetrics(prefix string) DurationMetrics {
	count := outboundMetricInt(prefix + "_count")
	total := outboundMetricInt(prefix + "_total")
	average := float64(0)
	if count > 0 {
		average = float64(total) / float64(count)
	}

	return DurationMetrics{
		Count:   count,
		Total:   total,
		Max:     outboundMetricInt(prefix + "_max"),
		Last:    outboundMetricInt(prefix + "_last"),
		Average: average,
	}
}

func outboundMetricInt(name string) int64 {
	return expvarIntValue(outboundMetrics.Get(name))
}

func expvarIntValue(v expvar.Var) int64 {
	if v == nil {
		return 0
	}

	n, err := strconv.ParseInt(v.String(), 10, 64)
	if err != nil {
		return 0
	}

	return n
}
