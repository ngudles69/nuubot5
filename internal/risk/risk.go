package risk

import (
	"fmt"
	"log/slog"

	"nuubot/internal/config"
)

// Risk defines one Runtime-owned risk policy.
type Risk interface {
	Assess() bool
	Stop()
}

// Section 1 - Program Flow

// Create constructs the configured Risk.
func Create(logger *slog.Logger, number int, cfg config.Risk) (Risk, error) {
	switch cfg.Kind {
	case "balanced":
		return newBalanced(logger, number), nil
	default:
		return nil, fmt.Errorf("unknown risk: %s", cfg.Kind)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
