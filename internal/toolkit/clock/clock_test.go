package clock

import (
	"errors"
	"io"
	"testing"

	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestTickClockRunsRegisteredTimer(t *testing.T) {
	var callbackErr = errors.New("callback failed")
	var calls []uint64
	var tickClock = New(logging.New(io.Discard))
	var err = tickClock.RegisterTimer(10, func(nowMS uint64) error {
		calls = append(calls, nowMS)
		if len(calls) == 2 {
			return callbackErr
		}
		return nil
	})
	if err != nil {
		t.Fatalf("register timer: %v", err)
	}

	err = tickClock.Advance(100)
	if err != nil {
		t.Fatalf("advance first tick: %v", err)
	}
	err = tickClock.Advance(109)
	if err != nil {
		t.Fatalf("advance before interval: %v", err)
	}
	err = tickClock.Advance(110)
	if !errors.Is(err, callbackErr) {
		t.Fatalf("actual error %v, expected callback error", err)
	}

	var expected = []uint64{100, 110}
	if len(calls) != len(expected) {
		t.Fatalf("actual calls %v, expected %v", calls, expected)
	}
	for index := range expected {
		if calls[index] != expected[index] {
			t.Fatalf("actual calls %v, expected %v", calls, expected)
		}
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
