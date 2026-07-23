package main

import "testing"

// Section 1 - Program Flow

func TestParseInput(t *testing.T) {
	sweepID, botID, err := parseInput([]string{"6", "9"})
	if err != nil {
		t.Fatal(err)
	}
	if sweepID != 6 {
		t.Fatalf("actual sweep ID %d, expected 6", sweepID)
	}
	if botID != 9 {
		t.Fatalf("actual bot ID %d, expected 9", botID)
	}
}

func TestParseInputRejectsInvalidInput(t *testing.T) {
	tests := [][]string{
		nil,
		{"0", "9"},
		{"6", "invalid"},
	}
	for _, args := range tests {
		_, _, err := parseInput(args)
		if err == nil {
			t.Fatalf("actual error nil for %v, expected error", args)
		}
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
