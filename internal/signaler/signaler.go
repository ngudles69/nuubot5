package signaler

import (
	"fmt"

	"nuubot/internal/config"
	"nuubot/internal/ohlcv"
	"nuubot/internal/toolkit/logging"
)

// Side identifies the signal direction.
type Side string

const (
	// Long identifies a long signal.
	Long Side = "long"
	// Short identifies a short signal.
	Short Side = "short"
)

// Signal describes one ordered trading signal.
type Signal struct {
	SignalMS    uint64
	AvailableMS uint64
	Side        Side
	Price       float64
}

type calculator interface {
	Requirements() []Requirement
	Calculate([]Series) ([]Signal, error)
}

type Requirement struct {
	Interval  ohlcv.Interval
	PriorRows int
}

type Series struct {
	ohlcv.Data
	PriorRows int
}

// Signaler calculates and releases ordered signals.
type Signaler struct {
	log        *logging.Logger
	calculator calculator
	rows       []Series
	signals    []Signal
	next       int
	started    bool
	prepared   bool
	stopped    bool
}

// Section 1 - Program Flow

// Create constructs the configured Signaler.
func Create(log *logging.Logger, cfg config.Signaler) (*Signaler, error) {
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
	log.Info(fmt.Sprintf("signaler initialized kind=%s", cfg.Kind))
	return &Signaler{log: log, calculator: implementation}, nil
}

// Prepare calculates and validates all signals.
func (s *Signaler) Prepare(loaded []Series) error {
	if s.prepared || s.started || s.stopped {
		return fmt.Errorf("signaler cannot prepare from current state")
	}
	signals, err := s.calculator.Calculate(loaded)
	if err != nil {
		return fmt.Errorf("calculate signals: %w", err)
	}
	for index, signal := range signals {
		if signal.SignalMS >= signal.AvailableMS ||
			(index > 0 && signals[index-1].AvailableMS >= signal.AvailableMS) {
			return fmt.Errorf("signaler produced invalid timestamp order")
		}
	}
	rowCount := 0
	for _, data := range loaded {
		rowCount += len(data.Close)
	}
	s.rows = loaded
	s.signals = signals
	s.prepared = true
	s.log.Info(fmt.Sprintf(
		"signaler prepared timeframes=%d rows_loaded=%d signals_calculated=%d",
		len(loaded),
		rowCount,
		len(signals),
	))
	return nil
}

// Start starts signal release.
func (s *Signaler) Start() error {
	if !s.prepared || s.started || s.stopped {
		return fmt.Errorf("signaler cannot start from current state")
	}
	s.started = true
	s.log.Info("signaler started")
	return nil
}

// Stop stops signal release and reports final statistics.
func (s *Signaler) Stop() {
	if s.stopped {
		return
	}
	s.started = false
	s.stopped = true
	s.log.Info(fmt.Sprintf(
		"signaler stopped timeframes=%d signals_calculated=%d "+
			"signals_released=%d signals_pending=%d",
		len(s.rows),
		len(s.signals),
		s.next,
		len(s.signals)-s.next,
	))
}

// Section 2 - Domain Helpers

// Requirements returns the calculator OHLCV requirements.
func (s *Signaler) Requirements() []Requirement {
	return s.calculator.Requirements()
}

// Next releases the next available signal.
func (s *Signaler) Next(nowMS uint64) (Signal, bool, error) {
	if !s.started || s.stopped {
		return Signal{}, false, fmt.Errorf("signaler cannot release signal from current state")
	}
	if s.next == len(s.signals) || s.signals[s.next].AvailableMS >= nowMS {
		return Signal{}, false, nil
	}
	signal := s.signals[s.next]
	signal.AvailableMS = nowMS
	s.next++
	return signal, true, nil
}

// Section 3 - Generic Helpers
