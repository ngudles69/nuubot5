package telemetry

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store appends Live telemetry to one Bot execution database.
type Store struct {
	db      *sql.DB
	started bool
	stopped bool
}

// Section 1 - Program Flow

// Init opens one telemetry store and returns its last persisted sequence.
func (s *Store) Init(path string) (uint64, error) {
	// Step 1: validate Store input
	if path == "" || s.started || s.stopped {
		return 0, fmt.Errorf("initialize telemetry Store: invalid state or path")
	}

	// Step 2: create runtime directory
	var err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return 0, fmt.Errorf("initialize telemetry Store: create directory: %w", err)
	}

	// Step 3: open runtime database
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	s.db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return 0, fmt.Errorf("initialize telemetry Store: open database: %v", err)
	}
	s.db.SetMaxOpenConns(1)

	// Step 4: prepare telemetry schema
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS telemetry_sample (
			sequence INTEGER PRIMARY KEY,
			timestamp_ms INTEGER NOT NULL,
			terminal INTEGER NOT NULL,
			ticks_served INTEGER NOT NULL,
			controller_runs INTEGER NOT NULL,
			signal_packages INTEGER NOT NULL,
			start_actions_skipped INTEGER NOT NULL,
			cycles_started INTEGER NOT NULL,
			cycles_rejected INTEGER NOT NULL,
			cycles_closed INTEGER NOT NULL,
			active_cycle INTEGER NOT NULL,
			bot_capital TEXT NOT NULL,
			bot_balance TEXT NOT NULL,
			bot_equity TEXT NOT NULL,
			net_pnl TEXT NOT NULL,
			peak_equity TEXT NOT NULL,
			drawdown TEXT NOT NULL,
			max_drawdown TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS telemetry_sample_timestamp
			ON telemetry_sample(timestamp_ms);
	`)
	if err != nil {
		var primary = fmt.Errorf("initialize telemetry Store: prepare schema: %v", err)
		var closeErr = s.db.Close()
		if closeErr != nil {
			return 0, errors.Join(primary, fmt.Errorf("close telemetry Store: %v", closeErr))
		}
		return 0, primary
	}

	// Step 5: load persisted sequence
	var lastSequence uint64
	err = s.db.QueryRow(`SELECT COALESCE(MAX(sequence), 0) FROM telemetry_sample`).Scan(&lastSequence)
	if err != nil {
		var primary = fmt.Errorf("initialize telemetry Store: load sequence: %v", err)
		var closeErr = s.db.Close()
		if closeErr != nil {
			return 0, errors.Join(primary, fmt.Errorf("close telemetry Store: %v", closeErr))
		}
		return 0, primary
	}
	s.started = true
	return lastSequence, nil
}

// Write appends one telemetry sample.
func (s *Store) Write(sample Sample) error {
	if !s.started || s.stopped || sample.Sequence == 0 || sample.TimestampMS == 0 {
		return fmt.Errorf("write telemetry sample: invalid state or sample")
	}
	_, err := s.db.Exec(`
		INSERT INTO telemetry_sample VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`,
		sample.Sequence,
		sample.TimestampMS,
		sample.Terminal,
		sample.TicksServed,
		sample.ControllerRuns,
		sample.SignalPackages,
		sample.StartActionsSkipped,
		sample.CyclesStarted,
		sample.CyclesRejected,
		sample.CyclesClosed,
		sample.ActiveCycle,
		sample.BotCapital.String(),
		sample.BotBalance.String(),
		sample.BotEquity.String(),
		sample.NetPnL.String(),
		sample.PeakEquity.String(),
		sample.Drawdown.String(),
		sample.MaxDrawdown.String(),
	)
	if err != nil {
		return fmt.Errorf("write telemetry sample: insert: %v", err)
	}
	return nil
}

// Stop closes the telemetry store.
func (s *Store) Stop() error {
	if s.stopped {
		return nil
	}
	s.stopped = true
	if !s.started {
		return nil
	}
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("stop telemetry Store: %v", err)
	}
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
