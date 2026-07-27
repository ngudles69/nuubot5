// Package resultpublisher atomically publishes successful BtBot evidence.
package resultpublisher

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"nuubot/internal/report"

	_ "modernc.org/sqlite"
)

// Section 1 - Program Flow

// Publish writes one complete backtest result atomically.
func Publish(path string, input report.Input, report report.Run) error {
	if path == "" {
		return fmt.Errorf("publish result: path is empty")
	}
	var domainPersisted bool
	for _, cycle := range input.Controller.Cycles {
		for _, executorResult := range cycle.Executors {
			if executorResult.Account == nil {
				continue
			}
			var current = *executorResult.Account
			if current.Infrastructure.ResultPath != path {
				return fmt.Errorf("publish result: Accounts use different result paths")
			}
			if current.PersistMode == "max" {
				domainPersisted = true
			}
		}
	}

	// prepare result directory
	var err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("publish result: prepare directory: %v", err)
	}

	// append terminal proof to domain persistence
	if domainPersisted {
		return publishResult(path, input, report)
	}

	// prepare temporary result path
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

	// publish terminal proof only
	err = publishResult(partial, input, report)
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
	input report.Input,
	report report.Run,
) error {
	var result = input.Controller
	var replay = input.Replay
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
			btbot_historical_data_loop_elapsed_ms INTEGER NOT NULL,
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
			cancellations INTEGER NOT NULL,
			closure_orders INTEGER NOT NULL,
			retries INTEGER NOT NULL,
			round_trips INTEGER NOT NULL,
			PRIMARY KEY (cycle_number, executor_id),
			FOREIGN KEY (cycle_number) REFERENCES botcycle_result(cycle_number)
		);
		CREATE TABLE grid_level_result (
			cycle_number INTEGER NOT NULL,
			executor_id TEXT NOT NULL,
			level INTEGER NOT NULL,
			boundary INTEGER NOT NULL,
			grid_price TEXT NOT NULL,
			initial_entry_price TEXT NOT NULL,
			reentry_price TEXT NOT NULL,
			exit_price TEXT NOT NULL,
			quantity TEXT NOT NULL,
			initial_notional TEXT NOT NULL,
			reentry_notional TEXT NOT NULL,
			initial_entry_commission TEXT NOT NULL,
			reentry_commission TEXT NOT NULL,
			exit_commission TEXT NOT NULL,
			initial_expected_pnl TEXT NOT NULL,
			reentry_expected_pnl TEXT NOT NULL,
			intended_action TEXT NOT NULL,
			current_trade_id INTEGER NOT NULL,
			current_trade_no INTEGER NOT NULL,
			current_trade_status TEXT NOT NULL,
			status TEXT NOT NULL,
			initial_submission_completed INTEGER NOT NULL,
			submission_attempts INTEGER NOT NULL,
			last_submitted_ms INTEGER NOT NULL,
			last_completed_ms INTEGER NOT NULL,
			PRIMARY KEY (cycle_number, executor_id, level),
			FOREIGN KEY (cycle_number, executor_id)
				REFERENCES executor_result(cycle_number, executor_id)
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
		);
		CREATE TABLE telemetry_sample (
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
		CREATE INDEX telemetry_sample_timestamp
			ON telemetry_sample(timestamp_ms);
		CREATE TABLE telemetry_event (
			sequence INTEGER PRIMARY KEY,
			timestamp_ms INTEGER NOT NULL,
			kind TEXT NOT NULL,
			frequency TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			parent_id TEXT NOT NULL,
			payload_json TEXT NOT NULL
		);
		CREATE INDEX telemetry_event_kind_frequency
			ON telemetry_event(kind, frequency);
		CREATE TABLE run_report (
			sweep_id INTEGER NOT NULL,
			bot_id INTEGER NOT NULL,
			bot_spec_id TEXT NOT NULL,
			config_hash TEXT NOT NULL,
			symbol TEXT NOT NULL,
			first_ms INTEGER NOT NULL,
			last_ms INTEGER NOT NULL,
			status TEXT NOT NULL,
			ticks INTEGER NOT NULL,
			controller_runs INTEGER NOT NULL,
			signal_packages INTEGER NOT NULL,
			start_actions_skipped INTEGER NOT NULL,
			cycles_started INTEGER NOT NULL,
			cycles_rejected INTEGER NOT NULL,
			cycles_closed INTEGER NOT NULL,
			trades INTEGER NOT NULL,
			orders INTEGER NOT NULL,
			fills INTEGER NOT NULL,
			cancellations INTEGER NOT NULL,
			stop_orders INTEGER NOT NULL,
			retries INTEGER NOT NULL,
			round_trips INTEGER NOT NULL,
			bot_capital TEXT NOT NULL,
			gross_pnl TEXT NOT NULL,
			fees TEXT NOT NULL,
			net_pnl TEXT NOT NULL,
			ending_equity TEXT NOT NULL,
			max_drawdown TEXT NOT NULL,
			historical_data_loop_elapsed_ms INTEGER NOT NULL,
			heap_before_publication_mb REAL NOT NULL,
			total_alloc_before_publication_mb REAL NOT NULL,
			gc_runs_before_publication INTEGER NOT NULL,
			gc_pause_before_publication_ms REAL NOT NULL,
			telemetry_samples INTEGER NOT NULL
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
		replay.HistoricalDataLoopElapsedMS,
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
				`INSERT INTO executor_result VALUES (
					?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
				)`,
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
				current.Cancellations,
				current.ClosureOrders,
				current.Retries,
				current.RoundTrips,
			)
			if err != nil {
				return fmt.Errorf("publish result: insert Executor: %w", err)
			}
			for _, level := range current.Levels {
				_, err = tx.Exec(
					`INSERT INTO grid_level_result VALUES (
						?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
						?, ?, ?, ?, ?, ?, ?, ?
					)`,
					cycle.CycleNumber,
					current.ID,
					level.Level,
					level.Boundary,
					level.GridPrice.String(),
					level.InitialEntryPrice.String(),
					level.ReentryPrice.String(),
					level.ExitPrice.String(),
					level.Quantity.String(),
					level.InitialNotional.String(),
					level.ReentryNotional.String(),
					level.InitialEntryCommission.String(),
					level.ReentryCommission.String(),
					level.ExitCommission.String(),
					level.InitialExpectedPnL.String(),
					level.ReentryExpectedPnL.String(),
					level.IntendedAction,
					level.CurrentTradeID,
					level.CurrentTradeNo,
					level.CurrentTradeStatus,
					level.Status,
					level.InitialSubmissionCompleted,
					level.SubmissionAttempts,
					level.LastSubmittedMS,
					level.LastCompletedMS,
				)
				if err != nil {
					return fmt.Errorf("publish result: insert Grid level: %w", err)
				}
			}
		}
	}
	err = publishTelemetryEvents(tx, input)
	if err != nil {
		return err
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
	var telemetryStatement *sql.Stmt
	telemetryStatement, err = tx.Prepare(`
		INSERT INTO telemetry_sample VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`)
	if err != nil {
		return fmt.Errorf("publish result: prepare telemetry samples: %w", err)
	}
	defer telemetryStatement.Close()
	for _, current := range input.Telemetry {
		_, err = telemetryStatement.Exec(
			current.Sequence,
			current.TimestampMS,
			current.Terminal,
			current.TicksServed,
			current.ControllerRuns,
			current.SignalPackages,
			current.StartActionsSkipped,
			current.CyclesStarted,
			current.CyclesRejected,
			current.CyclesClosed,
			current.ActiveCycle,
			current.BotCapital.String(),
			current.BotBalance.String(),
			current.BotEquity.String(),
			current.NetPnL.String(),
			current.PeakEquity.String(),
			current.Drawdown.String(),
			current.MaxDrawdown.String(),
		)
		if err != nil {
			return fmt.Errorf("publish result: insert telemetry sample: %w", err)
		}
	}
	_, err = tx.Exec(
		`INSERT INTO run_report VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)`,
		report.SweepID,
		report.BotID,
		report.BotSpecID,
		report.ConfigHash,
		report.Symbol,
		report.FirstMS,
		report.LastMS,
		report.Status,
		report.Ticks,
		report.ControllerRuns,
		report.SignalPackages,
		report.StartActionsSkipped,
		report.CyclesStarted,
		report.CyclesRejected,
		report.CyclesClosed,
		report.Trades,
		report.Orders,
		report.Fills,
		report.Cancellations,
		report.StopOrders,
		report.Retries,
		report.RoundTrips,
		report.BotCapital.String(),
		report.GrossPnL.String(),
		report.Fees.String(),
		report.NetPnL.String(),
		report.EndingEquity.String(),
		report.MaxDrawdown.String(),
		report.HistoricalDataLoopElapsedMS,
		report.Memory.HeapMB,
		report.Memory.TotalAllocMB,
		report.Memory.GCRuns,
		report.Memory.GCPauseMS,
		report.TelemetrySamples,
	)
	if err != nil {
		return fmt.Errorf("publish result: insert RunReport: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("publish result: commit summary: %w", err)
	}
	return nil
}

func publishTelemetryEvents(transaction *sql.Tx, input report.Input) error {
	var sequence uint64 = 1
	var result = input.Controller
	for _, cycle := range result.Cycles {
		var err = insertTelemetryEvent(
			transaction,
			sequence,
			cycle.EndMS,
			"botcycle",
			"end",
			fmt.Sprint(cycle.CycleNumber),
			fmt.Sprint(result.Identity.BotID),
			map[string]any{
				"start_ms": cycle.StartMS, "end_ms": cycle.EndMS,
				"duration_ms": cycle.DurationMS, "recon": cycle.Recon,
			},
		)
		if err != nil {
			return err
		}
		sequence++
		for _, current := range cycle.Executors {
			err = insertTelemetryEvent(
				transaction,
				sequence,
				cycle.EndMS,
				"executor",
				"end",
				current.ID,
				fmt.Sprint(cycle.CycleNumber),
				map[string]any{
					"status": current.Status, "exit_reason": current.ExitReason,
					"cancellations":  current.Cancellations,
					"closure_orders": current.ClosureOrders,
					"retries":        current.Retries, "round_trips": current.RoundTrips,
				},
			)
			if err != nil {
				return err
			}
			sequence++
			if current.Account == nil {
				continue
			}
			var accountResult = *current.Account
			err = insertTelemetryEvent(
				transaction,
				sequence,
				cycle.EndMS,
				"account",
				"end",
				accountResult.Name,
				current.ID,
				map[string]any{"snapshot": accountResult.Snapshot, "recon": accountResult.Recon},
			)
			if err != nil {
				return err
			}
			sequence++
			err = insertTelemetryEvent(
				transaction,
				sequence,
				cycle.EndMS,
				"ledger",
				"end",
				fmt.Sprint(accountResult.Ledger.ID),
				accountResult.Name,
				map[string]any{
					"fills_through_ms": accountResult.Ledger.FillsThroughMS,
					"last_recon_ms":    accountResult.Ledger.LastReconMS,
					"trades":           accountResult.Ledger.Trades,
				},
			)
			if err != nil {
				return err
			}
			sequence++
		}
	}
	return insertTelemetryEvent(
		transaction,
		sequence,
		result.LastMS,
		"bot",
		"end",
		fmt.Sprint(result.Identity.BotID),
		fmt.Sprint(result.Identity.SweepID),
		map[string]any{
			"first_ms": result.FirstMS, "last_ms": result.LastMS,
			"time_in_cycles_ms":  result.TimeInCyclesMS,
			"time_out_cycles_ms": result.TimeOutCyclesMS,
			"cycles":             len(result.Cycles), "recon": result.Recon,
			"net_pnl": result.NetPnL, "bot_equity": result.BotEquity,
		},
	)
}

// Section 3 - Generic Helpers

func insertTelemetryEvent(
	transaction *sql.Tx,
	sequence uint64,
	timestampMS uint64,
	kind string,
	frequency string,
	ownerID string,
	parentID string,
	payload any,
) error {
	var encoded, err = json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("publish result: encode %s telemetry: %v", kind, err)
	}
	_, err = transaction.Exec(
		`INSERT INTO telemetry_event VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sequence,
		timestampMS,
		kind,
		frequency,
		ownerID,
		parentID,
		string(encoded),
	)
	if err != nil {
		return fmt.Errorf("publish result: insert %s telemetry: %v", kind, err)
	}
	return nil
}
