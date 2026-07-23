package executor

import (
	"fmt"
	"log/slog"

	"nuubot5/internal/config"
	"nuubot5/internal/market"
	"nuubot5/internal/signaler"
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

// Program Flow

// New constructs the configured Executor.
func New(
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
