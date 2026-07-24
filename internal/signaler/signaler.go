package signaler

import (
	"fmt"
	"time"

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
	log      *logging.Logger
	rows     []Series
	signals  []Signal
	next     int
	started  bool
	prepared bool
	stopped  bool
}

// Section 1 - Program Flow

// Init prepares the configured Signaler and its complete Signal range.
func (s *Signaler) Init(
	log *logging.Logger,
	cfg config.Signaler,
	source string,
	start time.Time,
	end time.Time,
) error {
	s.log = log

	// select calculator
	var implementation calculator
	var err error
	switch cfg.Kind {
	case "macross":
		implementation, err = createMacross(cfg)
	case "rsi":
		implementation, err = createRSI(cfg)
	default:
		err = fmt.Errorf("unknown signaler: %s", cfg.Kind)
	}
	if err != nil {
		return err
	}

	// resolve requirements
	var requirements = implementation.Requirements()
	s.rows = make([]Series, 0, len(requirements))
	for _, requirement := range requirements {
		var duration, durationErr = requirement.Interval.Duration()
		if durationErr != nil {
			return fmt.Errorf("resolve signaler interval: %w", durationErr)
		}
		var loadStart = start.Add(-duration * time.Duration(requirement.PriorRows))

		// load ohlcv
		var rows, loadErr = ohlcv.Load(source, requirement.Interval, loadStart, end)
		if loadErr != nil {
			return fmt.Errorf("load signaler OHLCV: %w", loadErr)
		}
		s.rows = append(s.rows, Series{Data: rows, PriorRows: requirement.PriorRows})
	}

	// calculate signals
	s.signals, err = implementation.Calculate(s.rows)
	if err != nil {
		return fmt.Errorf("calculate signals: %w", err)
	}

	// validate signals
	for index, signal := range s.signals {
		if signal.SignalMS >= signal.AvailableMS ||
			(index > 0 && s.signals[index-1].AvailableMS >= signal.AvailableMS) {
			return fmt.Errorf("signaler produced invalid timestamp order")
		}
	}
	rowCount := 0
	for _, data := range s.rows {
		rowCount += len(data.Close)
	}
	s.prepared = true

	// initialize signaler
	s.log.Info(fmt.Sprintf(
		"signaler initialized timeframes=%d rows_loaded=%d signals_calculated=%d",
		len(s.rows),
		rowCount,
		len(s.signals),
	))
	return nil
}

// Start starts signal release.
func (s *Signaler) Start() error {
	if !s.prepared || s.started || s.stopped {
		return fmt.Errorf("signaler cannot start from current state")
	}
	// start signaler
	s.started = true
	s.log.Info("signaler started")
	return nil
}

// Stop stops signal release and reports final statistics.
func (s *Signaler) Stop() {
	if s.stopped {
		return
	}
	// stop signaler
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

// Run releases one available signal.
func (s *Signaler) Run(nowMS uint64) (Signal, bool, error) {
	if !s.started || s.stopped {
		return Signal{}, false, fmt.Errorf("signaler cannot release signal from current state")
	}

	// release signal
	if s.next == len(s.signals) || s.signals[s.next].AvailableMS >= nowMS {
		return Signal{}, false, nil
	}
	signal := s.signals[s.next]
	signal.AvailableMS = nowMS
	s.next++
	return signal, true, nil
}

// Section 3 - Generic Helpers
