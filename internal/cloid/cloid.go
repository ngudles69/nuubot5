// Package cloid encodes one Order identity for Venue use.
package cloid

import "fmt"

// Section 1 - Program Flow

// Encode encodes one canonical Ledger and Order key as a 128-bit CLOID.
func Encode(ledgerID uint64, orderID uint64) (string, error) {
	if ledgerID == 0 || orderID == 0 {
		return "", fmt.Errorf("encode cloid: ledger and order identities must be positive")
	}
	return fmt.Sprintf("0x%016x%016x", ledgerID, orderID), nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
