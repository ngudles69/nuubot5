// Package resultpublisher atomically publishes successful BtRunner evidence.
package resultpublisher

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"nuubot/internal/controller"
	"nuubot/internal/ledger"
	"nuubot/internal/simulator"

	_ "modernc.org/sqlite"
)

// ReplayProof contains terminal replay evidence for publication.
type ReplayProof struct {
	Symbol        string
	TicksExpected uint64
	TicksServed   uint64
	RunsExpected  uint64
	RunsTriggered uint64
	FirstMS       uint64
	LastMS        uint64
	ReplayMS      int64
	Completed     bool
}

// Section 1 - Program Flow

// Publish writes one complete Controller and replay result atomically.
func Publish(path string, result controller.Result, replay ReplayProof) error {
	if path == "" {
		return fmt.Errorf("publish result: path is empty")
	}
	for _, cycle := range result.Cycles {
		for _, executorResult := range cycle.Executors {
			if executorResult.Account == nil {
				continue
			}
			var current = *executorResult.Account
			if current.ResultPath != path {
				return fmt.Errorf("publish result: Accounts use different result paths")
			}
		}
	}

	// prepare temporary result path
	var err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("publish result: prepare directory: %v", err)
	}
	var partial = path + ".partial"
	err = os.Remove(partial)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("publish result: clear partial file: %v", err)
	}
	var published = false
	defer func() {
		if !published {
			os.Remove(partial)
		}
	}()

	// publish Account children
	for _, cycle := range result.Cycles {
		for _, executorResult := range cycle.Executors {
			if executorResult.Account == nil {
				continue
			}
			var current = *executorResult.Account
			err = ledger.Publish(partial, current.Ledger)
			if err != nil {
				return fmt.Errorf("publish result: %w", err)
			}
			if current.Simulator != nil {
				err = simulator.Publish(partial, *current.Simulator)
				if err != nil {
					return fmt.Errorf("publish result: %w", err)
				}
			}
		}
	}

	// publish Controller and replay evidence
	err = publishResult(partial, result, replay)
	if err != nil {
		return err
	}

	// publish completed result
	err = os.Rename(partial, path)
	if err != nil {
		return fmt.Errorf("publish result: replace completed database: %v", err)
	}
	published = true
	return nil
}

// Section 2 - Domain Helpers

func publishResult(
	path string,
	result controller.Result,
	replay ReplayProof,
) error {
	var db, err = sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("publish result: open summary database: %w", err)
	}
	defer db.Close()
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return fmt.Errorf("publish result: begin summary: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
		CREATE TABLE backtest_result (
			sweep_id INTEGER NOT NULL,
			bot_id INTEGER NOT NULL,
			bot_spec_id TEXT NOT NULL,
			config_toml TEXT NOT NULL,
			config_hash TEXT NOT NULL,
			exit_reason TEXT NOT NULL,
			symbol TEXT NOT NULL,
			ticks_expected INTEGER NOT NULL,
			ticks_served INTEGER NOT NULL,
			runs_expected INTEGER NOT NULL,
			runs_triggered INTEGER NOT NULL,
			first_ms INTEGER NOT NULL,
			last_ms INTEGER NOT NULL,
			replay_ms INTEGER NOT NULL,
			completed INTEGER NOT NULL,
			bot_capital TEXT NOT NULL,
			net_pnl TEXT NOT NULL,
			bot_equity TEXT NOT NULL,
			peak_equity TEXT NOT NULL,
			drawdown TEXT NOT NULL,
			max_drawdown TEXT NOT NULL
		);
		CREATE TABLE botcycle_result (
			cycle_number INTEGER PRIMARY KEY
		);
		CREATE TABLE executor_result (
			cycle_number INTEGER NOT NULL,
			executor_id TEXT NOT NULL,
			role TEXT NOT NULL,
			kind TEXT NOT NULL,
			side TEXT NOT NULL,
			venue TEXT NOT NULL,
			network TEXT NOT NULL,
			physical_account_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			status TEXT NOT NULL,
			exit_reason TEXT NOT NULL,
			capital_usdc TEXT NOT NULL,
			order_size_usdc TEXT NOT NULL,
			PRIMARY KEY (cycle_number, executor_id),
			FOREIGN KEY (cycle_number) REFERENCES botcycle_result(cycle_number)
		);
		CREATE TABLE signal_decision (
			sequence INTEGER PRIMARY KEY,
			timestamp_ms INTEGER NOT NULL,
			action TEXT NOT NULL
		);
		CREATE TABLE risk_decision (
			sequence INTEGER PRIMARY KEY,
			timestamp_ms INTEGER NOT NULL,
			policy INTEGER NOT NULL,
			decision TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("publish result: create summary schema: %w", err)
	}
	_, err = tx.Exec(`
		INSERT INTO backtest_result VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`,
		result.Identity.SweepID,
		result.Identity.BotID,
		result.Identity.BotSpecID,
		result.Identity.ConfigTOML,
		result.Identity.ConfigHash,
		result.ExitReason,
		replay.Symbol,
		replay.TicksExpected,
		replay.TicksServed,
		replay.RunsExpected,
		replay.RunsTriggered,
		replay.FirstMS,
		replay.LastMS,
		replay.ReplayMS,
		replay.Completed,
		result.BotCapital.String(),
		result.NetPnL.String(),
		result.BotEquity.String(),
		result.PeakEquity.String(),
		result.Drawdown.String(),
		result.MaxDrawdown.String(),
	)
	if err != nil {
		return fmt.Errorf("publish result: insert backtest summary: %w", err)
	}
	for _, cycle := range result.Cycles {
		_, err = tx.Exec(
			`INSERT INTO botcycle_result (cycle_number) VALUES (?)`,
			cycle.CycleNumber,
		)
		if err != nil {
			return fmt.Errorf("publish result: insert BotCycle: %w", err)
		}
		for _, current := range cycle.Executors {
			_, err = tx.Exec(
				`INSERT INTO executor_result VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				cycle.CycleNumber,
				current.ID,
				current.Role,
				current.Kind,
				current.Side,
				current.Resource.Venue,
				current.Resource.Network,
				current.Resource.PhysicalAccountID,
				current.Resource.Symbol,
				current.Status,
				current.ExitReason,
				current.CapitalUSDC.String(),
				current.OrderSizeUSDC.String(),
			)
			if err != nil {
				return fmt.Errorf("publish result: insert Executor: %w", err)
			}
		}
	}
	for index, current := range result.Signals {
		_, err = tx.Exec(
			`INSERT INTO signal_decision VALUES (?, ?, ?)`,
			index+1,
			current.TimestampMS,
			current.Action,
		)
		if err != nil {
			return fmt.Errorf("publish result: insert Signal decision: %w", err)
		}
	}
	for index, current := range result.Risks {
		_, err = tx.Exec(
			`INSERT INTO risk_decision VALUES (?, ?, ?, ?)`,
			index+1,
			current.TimestampMS,
			current.Policy,
			current.Decision,
		)
		if err != nil {
			return fmt.Errorf("publish result: insert Risk decision: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("publish result: commit summary: %w", err)
	}
	return nil
}

// Section 3 - Generic Helpers
