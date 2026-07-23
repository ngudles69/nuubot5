package executor

import (
	"fmt"
	"log/slog"

	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
)

// Executor defines one BotCycle-owned execution policy.
type Executor interface {
	Start() error
	OnBBO(market.BBO)
	Pass(uint64) bool
	Stop(string) error
	Terminal() bool
	ExitReason() string
}

// Section 1 - Program Flow

// Create constructs the configured Executor.
func Create(
	logger *slog.Logger,
	cycleNumber int,
	executorNumber int,
	signal signaler.Signal,
	cfg config.Executor,
) (Executor, error) {
	switch cfg.Kind {
	case "observer":
		return newObserver(logger, cycleNumber, executorNumber, signal, cfg)
	default:
		return nil, fmt.Errorf("unknown executor: %s", cfg.Kind)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
