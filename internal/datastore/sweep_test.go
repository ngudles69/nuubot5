package datastore

import "testing"

// Section 1 - Program Flow

func TestOptionalBotTime(t *testing.T) {
	for _, value := range []string{"2026-07-23", "2026-07-23T12:30:00Z"} {
		parsed, err := parseOptionalTime(value)
		if err != nil || parsed == nil {
			t.Fatalf("parse %q: %v", value, err)
		}
	}
	parsed, err := parseOptionalTime("")
	if err != nil || parsed != nil {
		t.Fatalf("empty time: parsed=%v err=%v", parsed, err)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
