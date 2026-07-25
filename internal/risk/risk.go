package risk

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/toolkit/logging"
)

// Decision reports one Controller gate or exit request.
type Decision string

const (
	Allow           Decision = "allow"
	BlockCycleStart Decision = "block_cycle_start"
	StopCycle       Decision = "stop_cycle"
	StopController  Decision = "stop_controller"
)

// Input contains one immutable Controller risk view.
type Input struct {
	TimestampMS     uint64
	ActiveCycle     bool
	CompletedCycles uint64
	Accounts        []account.Snapshot
	BotCapital      decimal.Decimal
	NetPnL          decimal.Decimal
	BotEquity       decimal.Decimal
	PeakEquity      decimal.Decimal
	CurrentDrawdown decimal.Decimal
	MaximumDrawdown decimal.Decimal
}

// Risk defines one Controller-owned risk policy.
type Risk interface {
	Assess(Input) Decision
	Stop()
}

// Section 1 - Program Flow

// Create constructs the configured Risk.
func Create(log *logging.Logger, number int, kind string) (Risk, error) {
	// select implementation
	switch kind {
	case "balanced":
		return createBalanced(log, number), nil
	default:
		return nil, fmt.Errorf("unknown risk: %s", kind)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
