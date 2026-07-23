package executor

import (
	"fmt"

	"nuubot5/internal/common"
	"nuubot5/internal/config"
	"nuubot5/internal/market"
	"nuubot5/internal/signaler"
)

type Executor interface {
	Start() error
	OnBBO(market.BBO)
	MainLoop(uint64) bool
	Stop(string) error
	Terminal() bool
	ExitReason() string
}

func New(
	log *common.Logger,
	cycleNumber int,
	executorNumber int,
	signal signaler.Signal,
	cfg config.Executor,
) (Executor, error) {
	switch cfg.Kind {
	case "observer":
		return newObserver(log, cycleNumber, executorNumber, signal, cfg)
	default:
		return nil, fmt.Errorf("unknown executor: %s", cfg.Kind)
	}
}
