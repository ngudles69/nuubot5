package cloid

import "testing"

// Section 1 - Program Flow

func TestEncode(t *testing.T) {
	var encoded, err = Encode(1, 2)
	if err != nil {
		t.Fatalf("encode cloid: %v", err)
	}
	var expected = "0x00000000000000010000000000000002"
	if encoded != expected {
		t.Fatalf("actual cloid %q, expected %q", encoded, expected)
	}
}

func TestEncodeRejectsInvalidIdentity(t *testing.T) {
	for _, identity := range [][2]uint64{{0, 1}, {1, 0}} {
		if _, err := Encode(identity[0], identity[1]); err == nil {
			t.Fatalf("Encode(%d, %d) error nil", identity[0], identity[1])
		}
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
