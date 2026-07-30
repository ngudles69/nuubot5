package calendar

import (
	"testing"
	"time"
)

// Section 1 - Program Flow

func TestResolvePeriod(t *testing.T) {
	var tests = []struct {
		label string
		start string
		end   string
	}{
		{label: "2024", start: "2024-01-01", end: "2025-01-01"},
		{label: "2024-H1", start: "2024-01-01", end: "2024-07-01"},
		{label: "2024-H2", start: "2024-07-01", end: "2025-01-01"},
		{label: "2024-Q1", start: "2024-01-01", end: "2024-04-01"},
		{label: "2024-M01", start: "2024-01-01", end: "2024-02-01"},
		{label: "2024-W34", start: "2024-08-19", end: "2024-08-26"},
		{label: "2024-01-03", start: "2024-01-03", end: "2024-01-04"},
	}

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			var start, end, err = ResolvePeriod(test.label)
			if err != nil {
				t.Fatalf("ResolvePeriod() error = %v", err)
			}
			if start.Format(time.DateOnly) != test.start {
				t.Fatalf("start = %q, want %q", start.Format(time.DateOnly), test.start)
			}
			if end.Format(time.DateOnly) != test.end {
				t.Fatalf("end = %q, want %q", end.Format(time.DateOnly), test.end)
			}
			if start.Location() != time.UTC || end.Location() != time.UTC {
				t.Fatalf("locations = %v, %v, want UTC", start.Location(), end.Location())
			}
		})
	}
}

func TestResolvePeriodRejectsInvalidLabels(t *testing.T) {
	var labels = []string{
		"",
		"2024-H3",
		"2024-Q5",
		"2024-M13",
		"2024-W00",
		"2024-W53",
		"2024-02-30",
	}

	for _, label := range labels {
		t.Run(label, func(t *testing.T) {
			var _, _, err = ResolvePeriod(label)
			if err == nil {
				t.Fatal("invalid period label was accepted")
			}
		})
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
