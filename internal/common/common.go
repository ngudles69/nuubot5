package common

import (
	"fmt"
)

// Generic Helpers

// StateError returns a shared lifecycle-state error.
func StateError(owner, action string) error {
	return fmt.Errorf("%s cannot %s from current state", owner, action)
}

// Duration returns the non-negative difference between two millisecond timestamps.
func Duration(start, end uint64) uint64 {
	if end < start {
		return 0
	}
	return end - start
}
