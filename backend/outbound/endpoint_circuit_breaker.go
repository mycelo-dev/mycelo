package outbound

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultOutboundCircuitBreakerFailureThreshold = 5
	defaultOutboundCircuitBreakerCooldownMs       = int64(30_000)
)

type endpointCircuitState struct {
	consecutiveFailures int
	openUntil           time.Time
}

// EndpointCircuitBreaker tracks endpoint health across all topic mappings for a destination URL.
type EndpointCircuitBreaker struct {
	mu               sync.Mutex
	failureThreshold int
	cooldown         time.Duration
	now              func() time.Time
	stateByEndpoint  map[string]endpointCircuitState
}

func newEndpointCircuitBreaker(failureThreshold int, cooldown time.Duration) *EndpointCircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = defaultOutboundCircuitBreakerFailureThreshold
	}
	if cooldown <= 0 {
		cooldown = time.Duration(defaultOutboundCircuitBreakerCooldownMs) * time.Millisecond
	}

	return &EndpointCircuitBreaker{
		failureThreshold: failureThreshold,
		cooldown:         cooldown,
		now:              time.Now,
		stateByEndpoint:  make(map[string]endpointCircuitState),
	}
}

func (b *EndpointCircuitBreaker) Allow(endpoint string) bool {
	if b == nil || endpoint == "" {
		return true
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.stateByEndpoint[endpoint]
	if state.openUntil.IsZero() {
		return true
	}
	if !b.now().Before(state.openUntil) {
		delete(b.stateByEndpoint, endpoint)
		return true
	}

	return false
}

func (b *EndpointCircuitBreaker) RemainingCooldown(endpoint string) time.Duration {
	if b == nil || endpoint == "" {
		return 0
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.stateByEndpoint[endpoint]
	if state.openUntil.IsZero() {
		return 0
	}

	remaining := time.Until(state.openUntil)
	if remaining <= 0 {
		delete(b.stateByEndpoint, endpoint)
		return 0
	}

	return remaining
}

func (b *EndpointCircuitBreaker) RecordSuccess(endpoint string) {
	if b == nil || endpoint == "" {
		return
	}

	b.mu.Lock()
	delete(b.stateByEndpoint, endpoint)
	b.mu.Unlock()
}

func (b *EndpointCircuitBreaker) RecordFailure(endpoint string) {
	if b == nil || endpoint == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.stateByEndpoint[endpoint]
	state.consecutiveFailures++
	if state.consecutiveFailures >= b.failureThreshold {
		state.openUntil = b.now().Add(b.cooldown)
		incrementOutboundMetricFor("circuit_opened_total", "endpoint", endpoint)
	}

	b.stateByEndpoint[endpoint] = state
}

func outboundCircuitBreakerFailureThreshold() int {
	if v := strings.TrimSpace(os.Getenv("OUTBOUND_CIRCUIT_BREAKER_FAILURE_THRESHOLD")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}

	return defaultOutboundCircuitBreakerFailureThreshold
}

func outboundCircuitBreakerCooldown() time.Duration {
	if v := strings.TrimSpace(os.Getenv("OUTBOUND_CIRCUIT_BREAKER_COOLDOWN_MS")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return time.Duration(n) * time.Millisecond
		}
	}

	return time.Duration(defaultOutboundCircuitBreakerCooldownMs) * time.Millisecond
}
