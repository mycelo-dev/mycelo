package outbound

import (
	"expvar"
	"strconv"
	"sync"
	"time"
)

var outboundMetrics = expvar.NewMap("mycelo_outbound")

var outboundMetricKeys sync.Map
var outboundMetricMaxMu sync.Mutex

func init() {
	for _, name := range []string{
		"delivery_success_total",
		"delivery_failure_total.category." + failureCategoryEndpoint4xx,
		"delivery_failure_total.category." + failureCategoryEndpoint5xx,
		"delivery_failure_total.category." + failureCategoryEndpointOther,
		"delivery_failure_total.category." + failureCategoryEndpointClient,
		"dead_letter_write_total",
		"dead_letter_replay_total",
		"delivery_lag_ms_count",
		"delivery_lag_ms_total",
		"delivery_lag_ms_max",
		"delivery_lag_ms_last",
		"delivery_attempt_duration_ms_count",
		"delivery_attempt_duration_ms_total",
		"delivery_attempt_duration_ms_max",
		"delivery_attempt_duration_ms_last",
		"delivery_success_last_at",
	} {
		outboundMetrics.Set(name, new(expvar.Int))
	}

	help := expvar.NewMap("mycelo_outbound_help")
	setOutboundMetricHelp(help, "where_to_look", "GET /debug/vars returns runtime data; outbound metrics live under the mycelo_outbound object")
	setOutboundMetricHelp(help, "delivery_success_total", "successful outbound deliveries")
	setOutboundMetricHelp(help, "delivery_failure_total.category.endpoint_response_4xx", "non-2xx endpoint responses with HTTP 4xx status")
	setOutboundMetricHelp(help, "delivery_failure_total.category.endpoint_response_5xx", "non-2xx endpoint responses with HTTP 5xx status")
	setOutboundMetricHelp(help, "delivery_failure_total.category.endpoint_response_other", "non-2xx endpoint responses outside 4xx/5xx")
	setOutboundMetricHelp(help, "delivery_failure_total.category.endpoint_transport", "HTTP client or transport errors")
	setOutboundMetricHelp(help, "dead_letter_write_total", "events written to the dead-letter queue during skip")
	setOutboundMetricHelp(help, "dead_letter_replay_total", "dead-letter records re-enqueued as fresh stream events")
	setOutboundMetricHelp(help, "delivery_lag_ms_count", "number of successful deliveries included in event-age lag aggregates")
	setOutboundMetricHelp(help, "delivery_lag_ms_total", "sum of source event created_at to successful delivery time in milliseconds; includes queue/backlog time")
	setOutboundMetricHelp(help, "delivery_lag_ms_max", "maximum observed source event created_at to successful delivery time in milliseconds; includes queue/backlog time")
	setOutboundMetricHelp(help, "delivery_lag_ms_last", "most recent successful delivery lag in milliseconds; this is the recovery/freshness signal")
	setOutboundMetricHelp(help, "delivery_attempt_duration_ms_count", "number of HTTP delivery attempts included in attempt-duration aggregates")
	setOutboundMetricHelp(help, "delivery_attempt_duration_ms_total", "sum of outbound HTTP delivery attempt durations in milliseconds")
	setOutboundMetricHelp(help, "delivery_attempt_duration_ms_max", "maximum observed outbound HTTP delivery attempt duration in milliseconds")
	setOutboundMetricHelp(help, "delivery_attempt_duration_ms_last", "most recent outbound HTTP delivery attempt duration in milliseconds")
	setOutboundMetricHelp(help, "delivery_success_last_at", "Unix milliseconds timestamp of the most recent successful outbound delivery")
	setOutboundMetricHelp(help, "circuit_opened_total.endpoint.<endpoint>", "dynamic key created when an endpoint circuit opens")
	setOutboundMetricHelp(help, "circuit_blocked_total.endpoint.<endpoint>", "dynamic key created when delivery is skipped because an endpoint circuit is open")
}

func setOutboundMetricHelp(help *expvar.Map, name string, description string) {
	v := new(expvar.String)
	v.Set(description)
	help.Set(name, v)
}

func incrementOutboundMetric(name string) {
	outboundMetrics.Add(name, 1)
}

func incrementOutboundMetricFor(name string, labels ...string) {
	incrementOutboundMetric(metricKey(name, labels...))
}

func observeOutboundDeliveryLag(createdAtMillis int64) {
	if createdAtMillis <= 0 {
		return
	}

	lag := time.Now().UnixMilli() - createdAtMillis
	if lag < 0 {
		return
	}

	observeOutboundDuration("delivery_lag_ms", lag)
}

func observeOutboundAttemptDuration(start time.Time) {
	if start.IsZero() {
		return
	}

	duration := time.Since(start).Milliseconds()
	if duration < 0 {
		return
	}

	observeOutboundDuration("delivery_attempt_duration_ms", duration)
}

func observeOutboundDuration(prefix string, durationMillis int64) {
	outboundMetrics.Add(prefix+"_count", 1)
	outboundMetrics.Add(prefix+"_total", durationMillis)
	setOutboundMetric(prefix+"_last", durationMillis)
	updateOutboundMetricMax(prefix+"_max", durationMillis)
}

func observeOutboundDeliverySuccess() {
	setOutboundMetric("delivery_success_last_at", time.Now().UnixMilli())
}

func setOutboundMetric(name string, value int64) {
	currentVar := outboundMetrics.Get(name)
	if currentVar == nil {
		currentVar = new(expvar.Int)
		outboundMetrics.Set(name, currentVar)
	}

	currentVar.(*expvar.Int).Set(value)
}

func updateOutboundMetricMax(name string, value int64) {
	outboundMetricMaxMu.Lock()
	defer outboundMetricMaxMu.Unlock()

	currentVar := outboundMetrics.Get(name)
	if currentVar == nil {
		currentVar = new(expvar.Int)
		outboundMetrics.Set(name, currentVar)
	}

	current, err := strconv.ParseInt(currentVar.String(), 10, 64)
	if err != nil || value <= current {
		return
	}

	currentVar.(*expvar.Int).Set(value)
}

func metricKey(name string, labels ...string) string {
	key := name
	for i := 0; i+1 < len(labels); i += 2 {
		key += "." + labels[i] + "." + labels[i+1]
	}

	if _, loaded := outboundMetricKeys.LoadOrStore(key, struct{}{}); !loaded {
		outboundMetrics.Set(key, new(expvar.Int))
	}

	return key
}
