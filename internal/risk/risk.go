package risk

import (
	"fmt"
	"log/slog"

	"nuubot5/internal/config"
)

// Risk defines one Runtime-owned risk policy.
type Risk interface {
	Assess() bool
	Stop()
}

// Program Flow

// New constructs the configured Risk.
func New(logger *slog.Logger, number int, cfg config.Risk) (Risk, error) {
	switch cfg.Kind {
	case "balanced":
		return newBalanced(logger, number), nil
	default:
		return nil, fmt.Errorf("unknown risk: %s", cfg.Kind)
	}
}
