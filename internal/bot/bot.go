// Package bot defines one admitted Bot composition.
package bot

import (
	"nuubot/internal/executor"
	"nuubot/internal/risk"
	"nuubot/internal/signaler"
)

// Identity identifies one exact admitted Bot.
type Identity struct {
	SweepID    uint64
	BotID      uint64
	BotSpecID  string
	ConfigTOML string
	ConfigHash string
}

// Definition contains one immutable Controller composition.
type Definition struct {
	Identity     Identity
	SignalSymbol string
	MaxCycles    uint64
	Signaler     signaler.Signaler
	Risks        []risk.Risk
	Executors    []executor.Spec
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
