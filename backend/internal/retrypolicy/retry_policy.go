package retrypolicy

import (
	"math"
	"math/rand/v2"
	"time"
)

// RandomDurationUpTo selects a duration between zero and the provided maximum.
var RandomDurationUpTo = func(maxDelay time.Duration) time.Duration {
	if maxDelay <= 0 {
		return 0
	}

	return time.Duration(rand.Int64N(int64(maxDelay) + 1))
}

// ComputeDelayWithFullJitter returns a jittered retry delay bounded by the computed max delay.
func ComputeDelayWithFullJitter(failureCount int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	return RandomDurationUpTo(ComputeMaxDelay(failureCount, baseDelay, maxDelay))
}

// ComputeMaxDelay returns the capped exponential backoff window for a failure count.
func ComputeMaxDelay(failureCount int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	if failureCount < 1 {
		return baseDelay
	}

	delay := float64(baseDelay) * math.Pow(2, float64(failureCount-1))
	if delay > float64(maxDelay) {
		return maxDelay
	}

	return time.Duration(delay)
}
