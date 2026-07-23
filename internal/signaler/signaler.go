package signaler

import (
	"fmt"
	"log/slog"

	"nuubot5/internal/bars"
	"nuubot5/internal/common"
	"nuubot5/internal/config"
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
	BarsNeeded() []bars.Requirement
	Calculate([]bars.Data) ([]Signal, error)
}

// Signaler calculates and releases ordered signals.
type Signaler struct {
	log        *slog.Logger
	calculator calculator
	bars       []bars.Data
	signals    []Signal
	next       int
	started    bool
	prepared   bool
	stopped    bool
}

// Program Flow

// New constructs the configured Signaler.
func New(logger *slog.Logger, cfg config.Signaler) (*Signaler, error) {
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
	log := logger.With("component", "signaler")
	log.Info(
		"signaler initialized",
		"event", "init",
		"status", "success",
		"kind", cfg.Kind,
	)
	return &Signaler{log: log, calculator: implementation}, nil
}

// Prepare calculates and validates all signals.
func (s *Signaler) Prepare(loaded []bars.Data) error {
	if s.prepared || s.started || s.stopped {
		return common.StateError("signaler", "prepare")
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
	barCount := 0
	for _, data := range loaded {
		barCount += len(data.Close)
	}
	s.bars = loaded
	s.signals = signals
	s.prepared = true
	s.log.Info(
		"signaler prepared",
		"event", "prepare",
		"status", "success",
		"timeframes", len(loaded),
		"bars_loaded", barCount,
		"signals_calculated", len(signals),
	)
	return nil
}

// Start starts signal release.
func (s *Signaler) Start() error {
	if !s.prepared || s.started || s.stopped {
		return common.StateError("signaler", "start")
	}
	s.started = true
	s.log.Info("signaler started", "event", "start", "status", "success")
	return nil
}

// Stop stops signal release and reports final statistics.
func (s *Signaler) Stop() {
	if s.stopped {
		return
	}
	s.started = false
	s.stopped = true
	s.log.Info(
		"signaler stopped",
		"event", "stop",
		"status", "success",
		"timeframes", len(s.bars),
		"signals_calculated", len(s.signals),
		"signals_released", s.next,
		"signals_pending", len(s.signals)-s.next,
	)
}

// Domain Helpers

// BarsNeeded returns the calculator bar requirements.
func (s *Signaler) BarsNeeded() []bars.Requirement {
	return s.calculator.BarsNeeded()
}

// Next releases the next available signal.
func (s *Signaler) Next(nowMS uint64) (Signal, bool, error) {
	if !s.started || s.stopped {
		return Signal{}, false, common.StateError("signaler", "release signal")
	}
	if s.next == len(s.signals) || s.signals[s.next].AvailableMS >= nowMS {
		return Signal{}, false, nil
	}
	signal := s.signals[s.next]
	s.next++
	return signal, true, nil
}
