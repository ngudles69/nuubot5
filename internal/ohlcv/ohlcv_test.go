package ohlcv

import "testing"

// Section 1 - Program Flow

func TestNormalizeStartRejectsMisalignment(t *testing.T) {
	var timestamp, err = normalizeStart(1_735_689_600_000_000, 1000)
	if err != nil || timestamp != 1_735_689_600_000 {
		t.Fatalf("unexpected timestamp: %d %v", timestamp, err)
	}
	if _, err = normalizeStart(1_735_689_600_000_001, 1000); err == nil {
		t.Fatal("fractional millisecond was accepted")
	}
	if _, err = normalizeStart(1_735_689_601_000_000, 3_600_000); err == nil {
		t.Fatal("misaligned interval start was accepted")
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
