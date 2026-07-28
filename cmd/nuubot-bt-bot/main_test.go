package main

import "testing"

// Section 1 - Program Flow

func TestParseArguments(t *testing.T) {
	var options, err = parseArguments([]string{"6", "9"})
	if err != nil {
		t.Fatal(err)
	}
	if options.SweepID != 6 {
		t.Fatalf("actual Sweep ID %d, expected 6", options.SweepID)
	}
	if options.BotID != 9 {
		t.Fatalf("actual Bot ID %d, expected 9", options.BotID)
	}
	if options.ProfilePrefix != "" {
		t.Fatalf("actual profile prefix %q, expected empty", options.ProfilePrefix)
	}
}

func TestParseArgumentsAcceptsProfilePrefix(t *testing.T) {
	var options, err = parseArguments([]string{"6", "9", "-pp", "profiles/run-001"})
	if err != nil {
		t.Fatal(err)
	}
	if options.SweepID != 6 || options.BotID != 9 {
		t.Fatalf("actual identity %d/%d, expected 6/9", options.SweepID, options.BotID)
	}
	if options.ProfilePrefix != "profiles/run-001" {
		t.Fatalf("actual profile prefix %q, expected profiles/run-001", options.ProfilePrefix)
	}
}

func TestParseArgumentsRejectsInvalidInput(t *testing.T) {
	tests := [][]string{
		nil,
		{"0", "9"},
		{"6", "invalid"},
		{"6", "9", "-pp"},
		{"6", "9", "invalid", "prefix"},
		{"6", "9", "-pp", ""},
		{"6", "9", "-pp", "prefix", "extra"},
	}
	for _, args := range tests {
		_, err := parseArguments(args)
		if err == nil {
			t.Fatalf("actual error nil for %v, expected error", args)
		}
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
