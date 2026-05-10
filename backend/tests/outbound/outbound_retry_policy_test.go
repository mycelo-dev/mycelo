package tests

import (
	"testing"
	"time"

	"github.com/mycelo-dev/mycelo/backend/internal/retrypolicy"
)

func TestComputeMaxDelay(t *testing.T) {
	testCases := []struct {
		name         string
		failureCount int
		want         time.Duration
	}{
		{name: "zero failures uses base delay", failureCount: 0, want: 2 * time.Second},
		{name: "first failure uses base delay", failureCount: 1, want: 2 * time.Second},
		{name: "second failure doubles delay", failureCount: 2, want: 4 * time.Second},
		{name: "third failure doubles again", failureCount: 3, want: 8 * time.Second},
		{name: "delay is capped", failureCount: 10, want: 1 * time.Minute},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := retrypolicy.ComputeMaxDelay(tc.failureCount, 2*time.Second, 1*time.Minute)
			if got != tc.want {
				t.Fatalf("ComputeMaxDelay(%d) = %s, want %s", tc.failureCount, got, tc.want)
			}
		})
	}
}

func TestComputeDelayWithFullJitter(t *testing.T) {
	originalRandomDurationUpTo := retrypolicy.RandomDurationUpTo
	t.Cleanup(func() {
		retrypolicy.RandomDurationUpTo = originalRandomDurationUpTo
	})

	called := false
	retrypolicy.RandomDurationUpTo = func(maxDelay time.Duration) time.Duration {
		called = true
		if maxDelay != 8*time.Second {
			t.Fatalf("RandomDurationUpTo received %s, want %s", maxDelay, 8*time.Second)
		}

		return 3 * time.Second
	}

	got := retrypolicy.ComputeDelayWithFullJitter(3, 2*time.Second, 1*time.Minute)
	if !called {
		t.Fatal("expected RandomDurationUpTo to be called")
	}

	if got != 3*time.Second {
		t.Fatalf("ComputeDelayWithFullJitter(3) = %s, want %s", got, 3*time.Second)
	}
}
