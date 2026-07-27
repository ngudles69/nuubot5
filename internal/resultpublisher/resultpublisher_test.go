package resultpublisher

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"

	"nuubot/internal/account"
	"nuubot/internal/bot"
	"nuubot/internal/botcycle"
	"nuubot/internal/botspec"
	"nuubot/internal/controller"
	"nuubot/internal/executor"
	"nuubot/internal/ledger"
	"nuubot/internal/report"
	"nuubot/internal/setup"
	"nuubot/internal/telemetry"
)

// Section 1 - Program Flow

func TestPublishWritesControllerAndReplayResult(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var err = publishTestResult(t, path, controller.Result{
		Identity: bot.Identity{
			SweepID: 1, BotID: 2, BotSpecID: "test_bot",
			ConfigTOML: "bot_spec = \"test_bot\"\n",
			ConfigHash: "hash",
		},
		ExitReason: "parent_stop",
		BotCapital: decimal.NewFromInt(1000),
		NetPnL:     decimal.NewFromInt(25),
		BotEquity:  decimal.NewFromInt(1025),
	}, report.Replay{
		Symbol: "BTC", TicksServed: 10, RunsTriggered: 1,
		HistoricalDataLoopElapsedMS: 123,
		Completed:                   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var botSpecID string
	var configTOML string
	var botEquity string
	var ticks uint64
	var historicalDataLoopElapsedMS int64
	var telemetrySamples int
	var storedReportSamples int
	var telemetryEvents int
	var botEndEvents int
	err = db.QueryRow(`
		SELECT bot_spec_id, config_toml, bot_equity, ticks_served,
		btbot_historical_data_loop_elapsed_ms,
		       (SELECT COUNT(*) FROM telemetry_sample),
		       (SELECT telemetry_samples FROM run_report),
		       (SELECT COUNT(*) FROM telemetry_event),
		       (SELECT COUNT(*) FROM telemetry_event
		        WHERE kind='bot' AND frequency='end')
		FROM backtest_result
	`).Scan(
		&botSpecID,
		&configTOML,
		&botEquity,
		&ticks,
		&historicalDataLoopElapsedMS,
		&telemetrySamples,
		&storedReportSamples,
		&telemetryEvents,
		&botEndEvents,
	)
	if err != nil {
		t.Fatal(err)
	}
	if botSpecID != "test_bot" ||
		configTOML != "bot_spec = \"test_bot\"\n" ||
		botEquity != "1025" ||
		ticks != 10 ||
		historicalDataLoopElapsedMS != 123 ||
		telemetrySamples != 1 ||
		storedReportSamples != 1 ||
		telemetryEvents != 1 ||
		botEndEvents != 1 {
		t.Fatalf(
			"unexpected result bot_spec=%s config=%q equity=%s ticks=%d",
			botSpecID,
			configTOML,
			botEquity,
			ticks,
		)
	}
}

func TestPublishPreservesMaximumAccountEvidence(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var runtimePath = path + ".partial"
	var accountResult = account.Result{
		Config: account.Config{
			Nuubot: &setup.Nuubot{ResultPath: path, RuntimePath: runtimePath},
			Name:   "sim", Venue: "simulator", Network: "simnet", Symbol: "BTC",
			PersistMode: ledger.Max,
		},
		Ledger: ledger.Result{Config: ledger.Config{
			ID: 1, CycleNumber: 1, ExecutorNumber: 1,
			Account: "sim", Network: "simnet", Symbol: "BTC",
			PersistMode: ledger.Max, Path: runtimePath,
		}},
	}
	var ownedLedger ledger.Ledger
	var err = ownedLedger.Init(accountResult.Ledger.Config)
	if err != nil {
		t.Fatal(err)
	}
	if err = ownedLedger.Stop(); err != nil {
		t.Fatal(err)
	}
	err = publishTestResult(t, path, controller.Result{
		Identity: bot.Identity{
			SweepID: 1, BotID: 2, BotSpecID: "test_bot",
			ConfigTOML: "bot_spec = \"test_bot\"\n",
			ConfigHash: "hash",
		},
		Cycles: []botcycle.Result{{
			CycleNumber: 1,
			Executors: []executor.Result{{
				ID: "trade", Kind: "trade", Side: executor.Long,
				Resource: botspec.Resource{
					Venue:             "simulator",
					Network:           "simnet",
					PhysicalAccountID: "sim",
					Symbol:            "BTC",
				},
				Status:  executor.Stopped,
				Account: &accountResult,
			}},
		}},
		ExitReason: "parent_stop",
		BotCapital: decimal.NewFromInt(1000),
		BotEquity:  decimal.NewFromInt(1000),
	}, report.Replay{Symbol: "BTC", Completed: true})
	if err != nil {
		t.Fatal(err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var summary, ledgers, simulatorEvents int
	err = db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM backtest_result),
			(SELECT COUNT(*) FROM account_ledger),
			(SELECT COUNT(*) FROM telemetry_event WHERE kind = 'simulator')
	`).Scan(&summary, &ledgers, &simulatorEvents)
	if err != nil {
		t.Fatal(err)
	}
	if summary != 1 || ledgers != 1 || simulatorEvents != 0 {
		t.Fatalf(
			"summary=%d ledgers=%d simulator_events=%d",
			summary,
			ledgers,
			simulatorEvents,
		)
	}
}

func TestPublishWritesGridLevels(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var err = publishTestResult(t, path, controller.Result{
		Identity: bot.Identity{
			SweepID: 1, BotID: 2, BotSpecID: "macross_grid_bot",
			ConfigTOML: "bot_spec = \"macross_grid_bot\"\n",
			ConfigHash: "hash",
		},
		Cycles: []botcycle.Result{{
			CycleNumber: 1,
			Executors: []executor.Result{{
				ID: "grid", Role: "grid", Kind: "grid", Side: executor.Long,
				Resource: botspec.Resource{
					Venue:             "simulator",
					Network:           "simnet",
					PhysicalAccountID: "sim",
					Symbol:            "BTC",
				},
				Status: executor.Stopped,
				Levels: []executor.GridLevel{{
					Level:                      1,
					GridPrice:                  decimal.NewFromInt(100),
					InitialEntryPrice:          decimal.NewFromInt(100),
					ReentryPrice:               decimal.NewFromInt(100),
					ExitPrice:                  decimal.NewFromInt(101),
					Quantity:                   decimal.RequireFromString("0.11"),
					InitialNotional:            decimal.NewFromInt(11),
					ReentryNotional:            decimal.NewFromInt(11),
					IntendedAction:             "enter_long",
					Status:                     "completed",
					InitialSubmissionCompleted: true,
					SubmissionAttempts:         1,
				}},
			}},
		}},
		ExitReason: "parent_stop",
	}, report.Replay{Symbol: "BTC", Completed: true})
	if err != nil {
		t.Fatal(err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var levels int
	var action string
	err = db.QueryRow(`
		SELECT COUNT(*), intended_action
		FROM grid_level_result
	`).Scan(&levels, &action)
	if err != nil {
		t.Fatal(err)
	}
	if levels != 1 || action != "enter_long" {
		t.Fatalf("actual levels=%d action=%s", levels, action)
	}
}

// Section 2 - Domain Helpers

func publishTestResult(
	t *testing.T,
	path string,
	result controller.Result,
	replay report.Replay,
) error {
	t.Helper()
	var input = report.Input{
		Controller: result,
		Replay:     replay,
		Telemetry: []telemetry.Sample{{
			Sequence:       1,
			Terminal:       true,
			ControllerRuns: replay.RunsTriggered,
			BotCapital:     result.BotCapital,
			BotBalance:     result.BotEquity,
			BotEquity:      result.BotEquity,
			NetPnL:         result.NetPnL,
			PeakEquity:     result.PeakEquity,
			Drawdown:       result.Drawdown,
			MaxDrawdown:    result.MaxDrawdown,
			TicksServed:    replay.TicksServed,
			SignalPackages: uint64(len(result.Signals)),
			CyclesClosed:   uint64(len(result.Cycles)),
			CyclesStarted:  uint64(len(result.Cycles)),
		}},
	}
	var report, err = report.Build(input)
	if err != nil {
		return err
	}
	return Publish(path, input, report)
}

// Section 3 - Generic Helpers
