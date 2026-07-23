package signaler

import (
	"fmt"

	"nuubot5/internal/bars"
	"nuubot5/internal/common"
	"nuubot5/internal/config"
)

type Side string

const (
	Long  Side = "long"
	Short Side = "short"
)

type Signal struct {
	SignalMS    uint64
	AvailableMS uint64
	Side        Side
	Price       float64
}

type calculator interface {
	BarsNeeded() []bars.Requirement
	Calculate([]bars.Data) ([]Signal, error)
}

type Signaler struct {
	log        *common.Logger
	calculator calculator
	bars       []bars.Data
	signals    []Signal
	next       int
	started    bool
	prepared   bool
	stopped    bool
}

func New(log *common.Logger, cfg config.Signaler) (*Signaler, error) {
	var implementation calculator
	var err error
	switch cfg.Kind {
	case "macross":
		implementation, err = newMacross(cfg)
	case "rsi":
		implementation, err = newRSI(cfg)
	default:
		err = fmt.Errorf("unknown signaler: %s", cfg.Kind)
	}
	if err != nil {
		return nil, err
	}
	log.Info("signaler", "init kind=%s", cfg.Kind)
	return &Signaler{log: log, calculator: implementation}, nil
}

func (s *Signaler) BarsNeeded() []bars.Requirement {
	return s.calculator.BarsNeeded()
}

func (s *Signaler) Prepare(loaded []bars.Data) error {
	if s.prepared || s.started || s.stopped {
		return common.StateError("Signaler", "prepare")
	}
	signals, err := s.calculator.Calculate(loaded)
	if err != nil {
		return err
	}
	for index, signal := range signals {
		if signal.SignalMS >= signal.AvailableMS ||
			(index > 0 && signals[index-1].AvailableMS >= signal.AvailableMS) {
			return fmt.Errorf("signaler produced invalid timestamp order")
		}
	}
	barCount := 0
	for _, data := range loaded {
		barCount += len(data.Close)
	}
	s.bars = loaded
	s.signals = signals
	s.prepared = true
	s.log.Info("signaler", "prepare timeframes=%d bars_loaded=%d signals_calculated=%d", len(loaded), barCount, len(signals))
	return nil
}

func (s *Signaler) Start() error {
	if !s.prepared || s.started || s.stopped {
		return common.StateError("Signaler", "start")
	}
	s.started = true
	s.log.Info("signaler", "start")
	return nil
}

func (s *Signaler) Next(nowMS uint64) (Signal, bool, error) {
	if !s.started || s.stopped {
		return Signal{}, false, common.StateError("Signaler", "release signal")
	}
	if s.next == len(s.signals) || s.signals[s.next].AvailableMS >= nowMS {
		return Signal{}, false, nil
	}
	signal := s.signals[s.next]
	s.next++
	return signal, true, nil
}

func (s *Signaler) Stop() {
	if s.stopped {
		return
	}
	s.started = false
	s.stopped = true
	s.log.Info(
		"signaler",
		"stop status=success timeframes=%d signals_calculated=%d signals_released=%d signals_pending=%d",
		len(s.bars), len(s.signals), s.next, len(s.signals)-s.next,
	)
}
