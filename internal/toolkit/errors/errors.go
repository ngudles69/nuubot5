// Package errors owns shared Nuubot error construction.
package errors

import "fmt"

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers

// StateError returns a lifecycle-state error.
func StateError(owner, action string) error {
	return fmt.Errorf("%s cannot %s from current state", owner, action)
}
