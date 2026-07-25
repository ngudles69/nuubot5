package risk

import (
	"bytes"
	"testing"

	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestBalancedAllowsImmutableInput(t *testing.T) {
	var policy, err = Create(logging.Create(&bytes.Buffer{}), 1, "balanced")
	if err != nil {
		t.Fatal(err)
	}
	var decision = policy.Assess(Input{
		TimestampMS:     1,
		ActiveCycle:     true,
		CompletedCycles: 2,
	})
	if decision != Allow {
		t.Fatalf("actual decision %q, expected %q", decision, Allow)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
