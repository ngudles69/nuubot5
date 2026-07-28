package report

import (
	"bytes"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/account/ledger"
	"nuubot/internal/bot"
	"nuubot/internal/botcycle"
	"nuubot/internal/controller"
	"nuubot/internal/executor"
	"nuubot/internal/telemetry"
)

// Section 1 - Program Flow

func TestBuildAndRenderSuite(t *testing.T) {
	var input = Input{
		Controller: controller.Result{
			Identity: bot.Identity{
				SweepID:    1,
				BotID:      2,
				BotSpecID:  "test_bot",
				ConfigHash: "hash",
			},
			Cycles: []botcycle.Result{{
				CycleNumber: 1,
				Executors: []executor.Result{{
					ID:         "trade",
					Retries:    3,
					RoundTrips: 4,
					Account: &account.Result{
						Snapshot: account.Snapshot{
							GrossPnL: decimal.NewFromInt(5),
							Fees:     decimal.NewFromInt(1),
						},
						Ledger: ledger.Result{
							Trades: 1, Orders: 2, Fills: 2,
							Cancellations: 1, StopOrders: 1,
						},
					},
				}},
			}},
			BotCapital:  decimal.NewFromInt(100),
			NetPnL:      decimal.NewFromInt(4),
			BotEquity:   decimal.NewFromInt(104),
			MaxDrawdown: decimal.NewFromInt(2),
		},
		Replay: Replay{
			Symbol:                      "BTC",
			TicksServed:                 10,
			RunsTriggered:               2,
			FirstMS:                     1000,
			LastMS:                      2000,
			HistoricalDataLoopElapsedMS: 50,
			Completed:                   true,
		},
		Telemetry: []telemetry.Sample{{
			Sequence:       1,
			Terminal:       true,
			SignalPackages: 3,
			CyclesStarted:  1,
			CyclesClosed:   1,
		}},
		Memory: Memory{
			HeapMB:       10,
			TotalAllocMB: 20,
			GCRuns:       2,
			GCPauseMS:    3,
		},
	}
	var report, err = Build(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Trades != 1 || report.Orders != 2 || report.Fills != 2 ||
		report.Cancellations != 1 || report.StopOrders != 1 ||
		report.Retries != 3 || report.RoundTrips != 4 ||
		!report.GrossPnL.Equal(decimal.NewFromInt(5)) ||
		!report.Fees.Equal(decimal.NewFromInt(1)) ||
		!report.NetPnL.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("unexpected RunReport: %+v", report)
	}
	var suite Suite
	suite, err = BuildSuite(SuiteInput{
		Requested:      2,
		SweepID:        1,
		BotID:          2,
		SuiteElapsedMS: 120,
		Attempts: []Attempt{
			{Run: 1, BtBotElapsedMS: 60, Report: &report},
			{Run: 2, BtBotElapsedMS: 70, Report: &report},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = WriteTable(&output, suite)
	if err != nil {
		t.Fatal(err)
	}
	var expected = `2x BtBot — Sweep 1, Bot 2

BotSpec: test_bot             Symbol: BTC
Status: PASS                  Requested: 2
Attempted: 2                  Passed: 2    Failed: 0

Timing (ms)
Item                  #  Cumulative  Avg  Min  Max
Suite (total)         1         120    —    —    —
BtBot                 2         130   65   60   70
Historical Data Loop  2         100   50   50   50

Memory (MB)
Item                                 #  Cumulative  Avg  Min  Max
Heap Before Publication              2           —   10   10   10
Total Allocation Before Publication  2          40   20   20   20

Garbage Collection (#)
Item     #  Cumulative  Avg  Min  Max
GC Runs  2           4    2    2    2

Garbage Collection Pause (ms)
Item      #  Cumulative  Avg  Min  Max
GC Pause  2           6    3    3    3

Replay and Execution (#)
Item                   #  Cumulative  Avg  Min  Max
Ticks                  2          20   10   10   10
Controller Runs        2           4    2    2    2
Telemetry Samples      2           2    1    1    1
Signal Packages        2           6    3    3    3
Start Actions Skipped  2           0    0    0    0
BotCycles Started      2           2    1    1    1
BotCycles Rejected     2           0    0    0    0
BotCycles Closed       2           2    1    1    1
Trades                 2           2    1    1    1
Orders                 2           4    2    2    2
Fills                  2           4    2    2    2
Cancellations          2           2    1    1    1
Stop Orders            2           2    1    1    1
Submission Retries     2           6    3    3    3
Completed Round Trips  2           8    4    4    4

Financial Results (USDC)
Item              #  Cumulative     Avg     Min     Max
Starting Capital  2           —  100.00  100.00  100.00
Gross PnL         2       10.00    5.00    5.00    5.00
Fees              2        2.00    1.00    1.00    1.00
Net PnL           2        8.00    4.00    4.00    4.00
Ending Capital    2           —  104.00  104.00  104.00
Maximum Drawdown  2           —    2.00    2.00    2.00

`
	if suite.Status != "pass" || output.String() != expected {
		t.Fatalf(
			"unexpected SuiteReport:\nwant:\n%s\ngot:\n%s",
			expected,
			output.String(),
		)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
