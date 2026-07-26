// Package report builds terminal run and suite reports.
package report

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/controller"
	"nuubot/internal/order"
	"nuubot/internal/telemetry"
)

// Memory contains one pre-publication Go memory observation.
type Memory struct {
	HeapMB       float64 `json:"heap_before_publication_mb"`
	TotalAllocMB float64 `json:"total_alloc_before_publication_mb"`
	GCRuns       uint32  `json:"gc_runs_before_publication"`
	GCPauseMS    float64 `json:"gc_pause_before_publication_ms"`
}

// Replay contains one immutable historical-data-loop proof.
type Replay struct {
	Symbol                      string `json:"symbol"`
	TicksExpected               uint64 `json:"ticks_expected"`
	TicksServed                 uint64 `json:"ticks_served"`
	RunsExpected                uint64 `json:"runs_expected"`
	RunsTriggered               uint64 `json:"runs_triggered"`
	FirstMS                     uint64 `json:"first_ms"`
	LastMS                      uint64 `json:"last_ms"`
	HistoricalDataLoopElapsedMS int64  `json:"historical_data_loop_elapsed_ms"`
	Completed                   bool   `json:"completed"`
}

// Input contains one import-safe terminal RunReport input.
type Input struct {
	Controller controller.Result
	Replay     Replay
	Telemetry  []telemetry.Sample
	Memory     Memory
}

// Run contains one exact terminal BtBot report.
type Run struct {
	SweepID                     uint64          `json:"sweep_id"`
	BotID                       uint64          `json:"bot_id"`
	BotSpecID                   string          `json:"bot_spec_id"`
	ConfigHash                  string          `json:"config_hash"`
	Symbol                      string          `json:"symbol"`
	FirstMS                     uint64          `json:"first_ms"`
	LastMS                      uint64          `json:"last_ms"`
	Status                      string          `json:"status"`
	Ticks                       uint64          `json:"ticks"`
	ControllerRuns              uint64          `json:"controller_runs"`
	SignalPackages              uint64          `json:"signal_packages"`
	StartActionsSkipped         uint64          `json:"start_actions_skipped"`
	CyclesStarted               uint64          `json:"cycles_started"`
	CyclesRejected              uint64          `json:"cycles_rejected"`
	CyclesClosed                uint64          `json:"cycles_closed"`
	Trades                      uint64          `json:"trades"`
	Orders                      uint64          `json:"orders"`
	Fills                       uint64          `json:"fills"`
	Cancellations               uint64          `json:"cancellations"`
	StopOrders                  uint64          `json:"stop_orders"`
	Retries                     uint64          `json:"retries"`
	RoundTrips                  uint64          `json:"round_trips"`
	BotCapital                  decimal.Decimal `json:"bot_capital"`
	GrossPnL                    decimal.Decimal `json:"gross_pnl"`
	Fees                        decimal.Decimal `json:"fees"`
	NetPnL                      decimal.Decimal `json:"net_pnl"`
	EndingEquity                decimal.Decimal `json:"ending_equity"`
	MaxDrawdown                 decimal.Decimal `json:"max_drawdown"`
	HistoricalDataLoopElapsedMS int64           `json:"historical_data_loop_elapsed_ms"`
	Memory                      Memory          `json:"memory"`
	TelemetrySamples            int             `json:"telemetry_samples"`
}

// Attempt contains one suite-owned fresh-process outcome.
type Attempt struct {
	Run            int    `json:"run"`
	Exit           int    `json:"exit"`
	BtBotElapsedMS int64  `json:"btbot_elapsed_ms"`
	Report         *Run   `json:"report,omitempty"`
	Error          string `json:"error,omitempty"`
}

// SuiteInput contains one complete suite aggregation request.
type SuiteInput struct {
	Requested      int       `json:"requested"`
	SweepID        uint64    `json:"sweep_id"`
	BotID          uint64    `json:"bot_id"`
	SuiteElapsedMS int64     `json:"suite_elapsed_ms"`
	Attempts       []Attempt `json:"attempts"`
}

// Metric contains one calculated table row.
type Metric struct {
	Item       string   `json:"item"`
	Unit       string   `json:"unit"`
	Samples    int      `json:"samples"`
	Cumulative *float64 `json:"cumulative,omitempty"`
	Average    *float64 `json:"average,omitempty"`
	Minimum    *float64 `json:"minimum,omitempty"`
	Maximum    *float64 `json:"maximum,omitempty"`
}

// Suite contains one calculated multi-process report.
type Suite struct {
	Requested int       `json:"requested"`
	Attempted int       `json:"attempted"`
	Passed    int       `json:"passed"`
	Failed    int       `json:"failed"`
	SweepID   uint64    `json:"sweep_id"`
	BotID     uint64    `json:"bot_id"`
	BotSpecID string    `json:"bot_spec_id"`
	Symbol    string    `json:"symbol"`
	Status    string    `json:"status"`
	Timing    []Metric  `json:"timing"`
	Memory    []Metric  `json:"memory"`
	GCRuns    []Metric  `json:"gc_runs"`
	GCPause   []Metric  `json:"gc_pause"`
	Execution []Metric  `json:"execution"`
	PnL       []Metric  `json:"pnl"`
	Attempts  []Attempt `json:"attempts"`
}

// Section 1 - Program Flow

// Build creates one terminal RunReport from immutable input.
func Build(input Input) (Run, error) {
	if input.Controller.Identity.SweepID == 0 ||
		input.Controller.Identity.BotID == 0 ||
		input.Controller.Identity.BotSpecID == "" ||
		input.Replay.Symbol == "" ||
		!input.Replay.Completed ||
		len(input.Telemetry) == 0 ||
		!input.Telemetry[len(input.Telemetry)-1].Terminal {
		return Run{}, fmt.Errorf("build RunReport: input is incomplete")
	}

	var trades uint64
	var orders uint64
	var fills uint64
	var cancellations uint64
	var stopOrders uint64
	var retries uint64
	var roundTrips uint64
	var grossPnL = decimal.Zero
	var fees = decimal.Zero
	for _, cycle := range input.Controller.Cycles {
		for _, current := range cycle.Executors {
			retries += current.Retries
			roundTrips += current.RoundTrips
			if current.Account == nil {
				continue
			}
			grossPnL = grossPnL.Add(current.Account.Snapshot.GrossPnL)
			fees = fees.Add(current.Account.Snapshot.Fees)
			trades += uint64(len(current.Account.Ledger.Trades))
			for _, ownedTrade := range current.Account.Ledger.Trades {
				orders += uint64(len(ownedTrade.Orders))
				for _, ownedOrder := range ownedTrade.Orders {
					fills += uint64(len(ownedOrder.Fills))
					if ownedOrder.Status == order.Canceled {
						cancellations++
					}
					if ownedOrder.Role == order.Stop {
						stopOrders++
					}
				}
			}
		}
	}

	var last = input.Telemetry[len(input.Telemetry)-1]
	return Run{
		SweepID:                     input.Controller.Identity.SweepID,
		BotID:                       input.Controller.Identity.BotID,
		BotSpecID:                   input.Controller.Identity.BotSpecID,
		ConfigHash:                  input.Controller.Identity.ConfigHash,
		Symbol:                      input.Replay.Symbol,
		FirstMS:                     input.Replay.FirstMS,
		LastMS:                      input.Replay.LastMS,
		Status:                      "complete",
		Ticks:                       input.Replay.TicksServed,
		ControllerRuns:              input.Replay.RunsTriggered,
		SignalPackages:              last.SignalPackages,
		StartActionsSkipped:         last.StartActionsSkipped,
		CyclesStarted:               last.CyclesStarted,
		CyclesRejected:              last.CyclesRejected,
		CyclesClosed:                last.CyclesClosed,
		Trades:                      trades,
		Orders:                      orders,
		Fills:                       fills,
		Cancellations:               cancellations,
		StopOrders:                  stopOrders,
		Retries:                     retries,
		RoundTrips:                  roundTrips,
		BotCapital:                  input.Controller.BotCapital,
		GrossPnL:                    grossPnL,
		Fees:                        fees,
		NetPnL:                      input.Controller.NetPnL,
		EndingEquity:                input.Controller.BotEquity,
		MaxDrawdown:                 input.Controller.MaxDrawdown,
		HistoricalDataLoopElapsedMS: input.Replay.HistoricalDataLoopElapsedMS,
		Memory:                      input.Memory,
		TelemetrySamples:            len(input.Telemetry),
	}, nil
}

// BuildSuite creates one standardized suite report.
func BuildSuite(input SuiteInput) (Suite, error) {
	if input.Requested <= 0 ||
		input.SweepID == 0 ||
		input.BotID == 0 ||
		input.SuiteElapsedMS < 0 ||
		len(input.Attempts) == 0 ||
		len(input.Attempts) > input.Requested {
		return Suite{}, fmt.Errorf("build SuiteReport: input is incomplete")
	}

	var passed int
	var botSpecID string
	var symbol string
	var btbotValues []float64
	var loopValues []float64
	var heapValues []float64
	var allocationValues []float64
	var gcRunValues []float64
	var gcPauseValues []float64
	var ticks []float64
	var controllerRuns []float64
	var telemetrySamples []float64
	var signals []float64
	var startActionsSkipped []float64
	var cyclesStarted []float64
	var cyclesRejected []float64
	var cyclesClosed []float64
	var trades []float64
	var orders []float64
	var fills []float64
	var cancellations []float64
	var stopOrders []float64
	var retries []float64
	var roundTrips []float64
	var capital []float64
	var grossPnL []float64
	var fees []float64
	var netPnL []float64
	var endingEquity []float64
	var maxDrawdown []float64
	for _, attempt := range input.Attempts {
		if attempt.Run <= 0 || attempt.BtBotElapsedMS < 0 {
			return Suite{}, fmt.Errorf("build SuiteReport: attempt is invalid")
		}
		btbotValues = append(
			btbotValues,
			float64(attempt.BtBotElapsedMS),
		)
		if attempt.Exit != 0 || attempt.Report == nil {
			continue
		}
		if attempt.Report.SweepID != input.SweepID ||
			attempt.Report.BotID != input.BotID ||
			attempt.Report.Status != "complete" {
			return Suite{}, fmt.Errorf("build SuiteReport: report identity is invalid")
		}
		if botSpecID == "" {
			botSpecID = attempt.Report.BotSpecID
			symbol = attempt.Report.Symbol
		}
		if attempt.Report.BotSpecID != botSpecID ||
			attempt.Report.Symbol != symbol {
			return Suite{}, fmt.Errorf("build SuiteReport: report identity changed")
		}
		passed++
		var report = attempt.Report
		loopValues = append(loopValues, float64(report.HistoricalDataLoopElapsedMS))
		heapValues = append(heapValues, report.Memory.HeapMB)
		allocationValues = append(allocationValues, report.Memory.TotalAllocMB)
		gcRunValues = append(gcRunValues, float64(report.Memory.GCRuns))
		gcPauseValues = append(gcPauseValues, report.Memory.GCPauseMS)
		ticks = append(ticks, float64(report.Ticks))
		controllerRuns = append(controllerRuns, float64(report.ControllerRuns))
		telemetrySamples = append(
			telemetrySamples,
			float64(report.TelemetrySamples),
		)
		signals = append(signals, float64(report.SignalPackages))
		startActionsSkipped = append(
			startActionsSkipped,
			float64(report.StartActionsSkipped),
		)
		cyclesStarted = append(cyclesStarted, float64(report.CyclesStarted))
		cyclesRejected = append(cyclesRejected, float64(report.CyclesRejected))
		cyclesClosed = append(cyclesClosed, float64(report.CyclesClosed))
		trades = append(trades, float64(report.Trades))
		orders = append(orders, float64(report.Orders))
		fills = append(fills, float64(report.Fills))
		cancellations = append(cancellations, float64(report.Cancellations))
		stopOrders = append(stopOrders, float64(report.StopOrders))
		retries = append(retries, float64(report.Retries))
		roundTrips = append(roundTrips, float64(report.RoundTrips))
		capital = append(capital, decimalFloat(report.BotCapital))
		grossPnL = append(grossPnL, decimalFloat(report.GrossPnL))
		fees = append(fees, decimalFloat(report.Fees))
		netPnL = append(netPnL, decimalFloat(report.NetPnL))
		endingEquity = append(endingEquity, decimalFloat(report.EndingEquity))
		maxDrawdown = append(maxDrawdown, decimalFloat(report.MaxDrawdown))
	}

	var status = "pass"
	if passed != len(input.Attempts) || len(input.Attempts) != input.Requested {
		status = "fail"
	}
	var suiteValue = float64(input.SuiteElapsedMS)
	return Suite{
		Requested: input.Requested,
		Attempted: len(input.Attempts),
		Passed:    passed,
		Failed:    len(input.Attempts) - passed,
		SweepID:   input.SweepID,
		BotID:     input.BotID,
		BotSpecID: botSpecID,
		Symbol:    symbol,
		Status:    status,
		Timing: []Metric{
			singleMetric("Suite (total)", "ms", suiteValue),
			summarizeMetric("BtBot", "ms", btbotValues, true),
			summarizeMetric(
				"Historical Data Loop",
				"ms",
				loopValues,
				true,
			),
		},
		Memory: []Metric{
			summarizeMetric("Heap Before Publication", "MB", heapValues, false),
			summarizeMetric(
				"Total Allocation Before Publication",
				"MB",
				allocationValues,
				true,
			),
		},
		GCRuns: []Metric{
			summarizeMetric(
				"GC Runs",
				"count",
				gcRunValues,
				true,
			),
		},
		GCPause: []Metric{
			summarizeMetric(
				"GC Pause",
				"ms",
				gcPauseValues,
				true,
			),
		},
		Execution: []Metric{
			summarizeMetric("Ticks", "count", ticks, true),
			summarizeMetric("Controller Runs", "count", controllerRuns, true),
			summarizeMetric(
				"Telemetry Samples",
				"count",
				telemetrySamples,
				true,
			),
			summarizeMetric("Signal Packages", "count", signals, true),
			summarizeMetric(
				"Start Actions Skipped",
				"count",
				startActionsSkipped,
				true,
			),
			summarizeMetric(
				"BotCycles Started",
				"count",
				cyclesStarted,
				true,
			),
			summarizeMetric(
				"BotCycles Rejected",
				"count",
				cyclesRejected,
				true,
			),
			summarizeMetric(
				"BotCycles Closed",
				"count",
				cyclesClosed,
				true,
			),
			summarizeMetric("Trades", "count", trades, true),
			summarizeMetric("Orders", "count", orders, true),
			summarizeMetric("Fills", "count", fills, true),
			summarizeMetric("Cancellations", "count", cancellations, true),
			summarizeMetric("Stop Orders", "count", stopOrders, true),
			summarizeMetric("Submission Retries", "count", retries, true),
			summarizeMetric(
				"Completed Round Trips",
				"count",
				roundTrips,
				true,
			),
		},
		PnL: []Metric{
			summarizeMetric("Starting Capital", "USDC", capital, false),
			summarizeMetric("Gross PnL", "USDC", grossPnL, true),
			summarizeMetric("Fees", "USDC", fees, true),
			summarizeMetric("Net PnL", "USDC", netPnL, true),
			summarizeMetric("Ending Capital", "USDC", endingEquity, false),
			summarizeMetric("Maximum Drawdown", "USDC", maxDrawdown, false),
		},
		Attempts: append([]Attempt(nil), input.Attempts...),
	}, nil
}

// Section 2 - Domain Helpers

func singleMetric(item, unit string, value float64) Metric {
	return Metric{
		Item:       item,
		Unit:       unit,
		Samples:    1,
		Cumulative: &value,
	}
}

func summarizeMetric(
	item,
	unit string,
	values []float64,
	cumulative bool,
) Metric {
	var metric = Metric{
		Item:    item,
		Unit:    unit,
		Samples: len(values),
	}
	if len(values) == 0 {
		return metric
	}
	var total float64
	var minimum = values[0]
	var maximum = values[0]
	for _, value := range values {
		total += value
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
	}
	var average = total / float64(len(values))
	metric.Average = &average
	metric.Minimum = &minimum
	metric.Maximum = &maximum
	if cumulative {
		metric.Cumulative = &total
	}
	return metric
}

// Section 3 - Generic Helpers

func decimalFloat(value decimal.Decimal) float64 {
	var result, _ = value.Float64()
	return result
}
