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
	var actual = policy.(*balanced)
	if actual.assessments != 1 {
		t.Fatalf("actual assessments %d, expected 1", actual.assessments)
	}
	policy.Stop()
	policy.Stop()
	if !actual.stopped {
		t.Fatal("BalancedRisk did not stop")
	}
}

func TestCreateRejectsUnknownRisk(t *testing.T) {
	var _, err = Create(logging.Create(&bytes.Buffer{}), 1, "unknown")
	if err == nil {
		t.Fatal("unknown Risk was accepted")
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
