package risk

import (
	"fmt"

	"nuubot5/internal/common"
	"nuubot5/internal/config"
)

type Risk interface {
	Assess() bool
	Stop()
}

func New(log *common.Logger, number int, cfg config.Risk) (Risk, error) {
	switch cfg.Kind {
	case "balanced":
		return newBalanced(log, number), nil
	default:
		return nil, fmt.Errorf("unknown risk: %s", cfg.Kind)
	}
}
