package backtest

import (
	"context"
	"fmt"
	"os"
	stdruntime "runtime"
	"time"

	"nuubot/internal/controller"
	"nuubot/internal/datastore"
	"nuubot/internal/market"
	"nuubot/internal/replay"
	"nuubot/internal/report"
	"nuubot/internal/resultpublisher"
	"nuubot/internal/setup"
	"nuubot/internal/telemetry"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
)

type stats struct {
	ticksExpected             uint64
	ticksServed               uint64
	runsExpected              uint64
	runsTriggered             uint64
	expectedFirstMS           uint64
	expectedLastMS            uint64
	firstMS                   uint64
	lastMS                    uint64
	replayCompleted           bool
	historicalDataLoopElapsed time.Duration
}

// ReplayResult contains immutable replay and publication proof.
type ReplayResult struct {
	Symbol                      string
	TicksExpected               uint64
	TicksServed                 uint64
	RunsExpected                uint64
	RunsTriggered               uint64
	FirstMS                     uint64
	LastMS                      uint64
	HistoricalDataLoopElapsedMS int64
	Completed                   bool
	Published                   bool
}

// Result contains one complete immutable backtest result.
type Result struct {
	Controller controller.Result
	Replay     ReplayResult
	Report     report.Run
}

// Run owns one bounded historical replay.
type Run struct {
	log                 *logging.Logger
	reader              replay.Reader
	clock               clock.Clock
	marketData          *market.MarketData
	marketKeys          []market.Key
	controller          controller.Controller
	symbol              string
	resultPath          string
	stats               stats
	telemetry           []telemetry.Sample
	telemetryIntervalMS uint64
	lastTelemetryMS     uint64
	report              report.Run
	stopRequested       bool
	published           bool
	started             bool
	stopped             bool
}

// Section 1 - Program Flow

// Init prepares one bounded historical replay.
func (r *Run) Init(
	caller context.Context,
	log *logging.Logger,
	sweepID,
	botID uint64,
) error {
	r.log = log

	// Step 1: general app global setup
	var nuubot, err = setup.Setup(caller, log, sweepID, botID)
	if err != nil {
		return fmt.Errorf("prepare setup: %w", err)
	}

	// Step 2: select Backtest runtime policy
	nuubot.Runtime = nuubot.App.Backtest
	r.telemetryIntervalMS = nuubot.Runtime.TelemetryIntervalMS

	// Step 3: reset Bot status for fresh replay
	nuubot.Bot.Status = datastore.BotConfigured

	// Step 4: clear replay data
	nuubot.RuntimePath = nuubot.ResultPath + ".partial"
	err = clearData(nuubot.RuntimePath)
	if err != nil {
		return err
	}

	// Step 5: retain runtime inputs
	var replayInput = nuubot.Bot.Replay
	r.symbol = replayInput.Symbol
	r.resultPath = nuubot.ResultPath

	// Step 6: set replay range
	var start = replayInput.ReplayStart
	var end = replayInput.ReplayEnd
	if replayInput.EndAt != nil && replayInput.EndAt.Before(end) {
		end = *replayInput.EndAt
	}
	if !start.Before(end) {
		return fmt.Errorf("bot end must follow replay start")
	}
	var startMS = uint64(start.UnixMilli())
	var endMS = uint64(end.UnixMilli())
	var durationMS = endMS - startMS

	// Step 7: initialize replay reader
	err = r.reader.Init(
		log,
		replayInput.TicksPath,
		start,
		end,
	)
	if err != nil {
		return fmt.Errorf("initialize replay reader: %w", err)
	}

	// Step 8: create clock
	r.clock, err = clock.Create(clock.Tick)
	if err != nil {
		return fmt.Errorf("create clock: %w", err)
	}

	// Step 9: initialize clock
	err = r.clock.Init(log, startMS)
	if err != nil {
		return fmt.Errorf("initialize clock: %w", err)
	}

	// Step 10: attach clock to Nuubot
	nuubot.Clock = r.clock

	// Step 11: create and attach MarketData to Nuubot
	r.marketData = market.CreateMarketData()
	r.marketKeys = marketKeys(nuubot)
	nuubot.MarketData = r.marketData

	// Step 12: initialize Controller
	err = r.controller.Init(nuubot)
	if err != nil {
		return fmt.Errorf("initialize Controller: %w", err)
	}

	// Step 13: register Controller timer
	err = r.clock.RegisterTimer(clock.Timer{
		Name:       "controller",
		IntervalMS: nuubot.Runtime.ControllerIntervalMS,
	}, r.controllerRun)
	if err != nil {
		return fmt.Errorf("register Controller timer: %w", err)
	}

	// Step 14: initialize replay stats
	r.stats = stats{
		ticksExpected: durationMS / 1000,
		runsExpected: (durationMS + nuubot.Runtime.ControllerIntervalMS - 1) /
			nuubot.Runtime.ControllerIntervalMS,
		expectedFirstMS: startMS + 1000,
		expectedLastMS:  endMS,
	}

	// Step 15: log init completed
	log.Info(fmt.Sprintf(
		"btbot initialized bot_spec=%s symbol=%s",
		nuubot.Bot.BotSpecID,
		r.symbol,
	))
	return nil
}

// Start starts the owned Clock and Controller.
func (r *Run) Start() error {
	if r.started || r.stopped {
		return fmt.Errorf("btbot cannot start from current state")
	}

	// Step 1: start clock
	var err = r.clock.Start()
	if err != nil {
		return fmt.Errorf("start clock: %w", err)
	}

	// Step 2: start Controller
	err = r.controller.Start()
	if err != nil {
		r.clock.Stop()
		return fmt.Errorf("start Controller: %w", err)
	}

	// Step 3: log start completed
	r.started = true
	r.log.Info("btbot started")
	return nil
}

// Loop executes the complete bounded replay loop.
func (r *Run) Loop() error {
	if !r.started || r.stopped {
		return fmt.Errorf("btbot cannot loop from current state")
	}
	r.log.Info("btbot loop started")
	var started = time.Now()
	var bbo market.BBO
	var more bool
	var err error
	defer func() {
		r.stats.historicalDataLoopElapsed = time.Since(started)
	}()

	for {
		// Step 1: read replay input
		bbo, more, err = r.reader.Next()
		if err != nil {
			return fmt.Errorf("read replay: %w", err)
		}
		if !more {
			break
		}

		// Step 2: publish BBO to MarketData
		bbo.Symbol = r.symbol
		for _, key := range r.marketKeys {
			err = r.marketData.IngestBBO(key, bbo)
			if err != nil {
				return fmt.Errorf("ingest MarketData BBO: %w", err)
			}
		}

		// Step 3: update replay stats
		if r.stats.firstMS == 0 {
			r.stats.firstMS = bbo.TimestampMS
		}
		r.stats.lastMS = bbo.TimestampMS
		r.stats.ticksServed++

		// Step 4: advance clock (will trigger controllerRun)
		err = r.clock.Advance(bbo.TimestampMS)
		if err != nil {
			return fmt.Errorf("advance clock: %w", err)
		}

		// Step 5: check stop request
		if r.stopRequested {
			break
		}
	}

	// Step 6: verify replay completion
	err = r.verify()
	if err != nil {
		return fmt.Errorf("verify replay: %w", err)
	}
	return nil
}

// Stop releases owned resources and reports final results.
func (r *Run) Stop() error {
	// Step 1: log stop started
	r.log.Info("btbot stop started")

	// Step 2: ignore repeated stop request
	if r.stopped {
		r.log.Info("btbot stopping - ignoring stop request")
		return nil
	}

	// Step 3: mark Run stopped
	r.started = false
	r.stopped = true

	// Step 4: stop clock
	r.clock.Stop()

	// Step 5: stop replay reader
	var readerErr = r.reader.Stop()
	if readerErr != nil {
		readerErr = fmt.Errorf("stop replay reader: %w", readerErr)
	}

	// Step 6: stop Controller
	var controllerErr = r.controller.Stop("parent_stop")
	if controllerErr != nil {
		controllerErr = fmt.Errorf("stop Controller: %w", controllerErr)
	}

	// Step 7: stop MarketData
	var marketDataErr = r.marketData.Stop()
	if marketDataErr != nil {
		marketDataErr = fmt.Errorf("stop MarketData: %w", marketDataErr)
	}

	// Step 8: prepare completed result
	var publishErr error
	if readerErr == nil && controllerErr == nil && marketDataErr == nil &&
		r.stats.replayCompleted {
		r.collectTelemetry(r.stats.lastMS, true)
		var memory stdruntime.MemStats
		stdruntime.ReadMemStats(&memory)
		var input = report.Input{
			Controller: r.controller.Result(),
			Replay: report.Replay{
				Symbol:                      r.symbol,
				TicksExpected:               r.stats.ticksExpected,
				TicksServed:                 r.stats.ticksServed,
				RunsExpected:                r.stats.runsExpected,
				RunsTriggered:               r.stats.runsTriggered,
				FirstMS:                     r.stats.firstMS,
				LastMS:                      r.stats.lastMS,
				HistoricalDataLoopElapsedMS: r.stats.historicalDataLoopElapsed.Milliseconds(),
				Completed:                   r.stats.replayCompleted,
			},
			Telemetry: append([]telemetry.Sample(nil), r.telemetry...),
			Memory: report.Memory{
				HeapMB:       float64(memory.HeapAlloc) / (1 << 20),
				TotalAllocMB: float64(memory.TotalAlloc) / (1 << 20),
				GCRuns:       memory.NumGC,
				GCPauseMS:    float64(memory.PauseTotalNs) / 1e6,
			},
		}

		// Step 9: build terminal report
		r.report, publishErr = report.Build(input)

		// Step 10: publish completed result
		if publishErr == nil {
			publishErr = resultpublisher.Publish(r.resultPath, input, r.report)
		}
		if publishErr != nil {
			publishErr = fmt.Errorf("build or publish result: %w", publishErr)
		} else {
			r.published = true
		}
	}

	// Step 11: log stop results and stats
	var result = "failed"
	if readerErr == nil && controllerErr == nil && marketDataErr == nil &&
		publishErr == nil && r.stats.replayCompleted {
		result = "complete"
	}
	r.log.Info(fmt.Sprintf(
		"btbot stopped loader=parquet ticks_served=%d ticks_expected=%d "+
			"runs_triggered=%d runs_expected=%d first_ts_ms=%d last_ts_ms=%d "+
			"replay_completed=%t btbot_historical_data_loop_elapsed_ms=%d "+
			"heap_before_publication_mb=%f total_alloc_before_publication_mb=%f "+
			"gc_runs_before_publication=%d gc_pause_before_publication_ms=%f "+
			"result=%s",
		r.stats.ticksServed,
		r.stats.ticksExpected,
		r.stats.runsTriggered,
		r.stats.runsExpected,
		r.stats.firstMS,
		r.stats.lastMS,
		r.stats.replayCompleted,
		r.stats.historicalDataLoopElapsed.Milliseconds(),
		r.report.Memory.HeapMB,
		r.report.Memory.TotalAllocMB,
		r.report.Memory.GCRuns,
		r.report.Memory.GCPauseMS,
		result,
	))

	// Step 12: return stop errors
	if controllerErr != nil {
		return controllerErr
	}
	if readerErr != nil {
		return readerErr
	}
	if marketDataErr != nil {
		return marketDataErr
	}
	if publishErr != nil {
		return publishErr
	}
	if !r.stats.replayCompleted {
		return fmt.Errorf("btbot replay did not complete")
	}

	// Step 13: log stop completed
	r.log.Info("btbot stopped.")
	return nil
}

// Section 2 - Domain Helpers

func marketKeys(nuubot *setup.Nuubot) []market.Key {
	var seen = make(map[market.Key]struct{})
	var keys = make([]market.Key, 0, len(nuubot.BotSpec.Executors))
	for _, spec := range nuubot.BotSpec.Executors {
		var key = market.Key{
			Venue:   spec.Resource.Venue,
			Network: spec.Resource.Network,
			Symbol:  spec.Resource.Symbol,
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}

func clearData(path string) error {
	for _, current := range []string{path, path + "-wal", path + "-shm"} {
		var err = os.Remove(current)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("clear Run data %s: %w", current, err)
		}
	}
	return nil
}

func (r *Run) controllerRun(_ uint64) error {
	// run Controller - triggered by Timer
	r.stats.runsTriggered++
	var stop, err = r.controller.Run()
	if err != nil {
		return fmt.Errorf("run Controller: %w", err)
	}
	var nowMS = r.clock.NowMS()
	if r.lastTelemetryMS == 0 || nowMS-r.lastTelemetryMS >= r.telemetryIntervalMS {
		r.collectTelemetry(nowMS, false)
		r.lastTelemetryMS = nowMS
	}
	// remember stop request
	if stop {
		r.stopRequested = true
	}
	return nil
}

// Result returns one complete terminal backtest result.
func (r *Run) Result() (Result, error) {
	if !r.stopped || !r.stats.replayCompleted {
		return Result{}, fmt.Errorf("btbot result is unavailable")
	}
	return Result{
		Controller: r.controller.Result(),
		Replay: ReplayResult{
			Symbol:                      r.symbol,
			TicksExpected:               r.stats.ticksExpected,
			TicksServed:                 r.stats.ticksServed,
			RunsExpected:                r.stats.runsExpected,
			RunsTriggered:               r.stats.runsTriggered,
			FirstMS:                     r.stats.firstMS,
			LastMS:                      r.stats.lastMS,
			HistoricalDataLoopElapsedMS: r.stats.historicalDataLoopElapsed.Milliseconds(),
			Completed:                   r.stats.replayCompleted,
			Published:                   r.published,
		},
		Report: r.report,
	}, nil
}

func (r *Run) collectTelemetry(nowMS uint64, terminal bool) {
	var current = r.controller.Telemetry()
	r.telemetry = append(r.telemetry, telemetry.Sample{
		Sequence:            uint64(len(r.telemetry) + 1),
		TimestampMS:         nowMS,
		Terminal:            terminal,
		TicksServed:         r.stats.ticksServed,
		ControllerRuns:      current.Runs,
		SignalPackages:      current.SignalPackagesRead,
		StartActionsSkipped: current.StartActionsSkipped,
		CyclesStarted:       current.CyclesStarted,
		CyclesRejected:      current.CyclesRejected,
		CyclesClosed:        current.CyclesClosed,
		ActiveCycle:         current.ActiveCycle,
		BotCapital:          current.BotCapital,
		BotBalance:          current.BotBalance,
		BotEquity:           current.BotEquity,
		NetPnL:              current.NetPnL,
		PeakEquity:          current.PeakEquity,
		Drawdown:            current.Drawdown,
		MaxDrawdown:         current.MaxDrawdown,
	})
}

func (r *Run) verify() error {
	if r.stats.ticksServed != r.stats.ticksExpected ||
		r.stats.runsTriggered != r.stats.runsExpected ||
		r.stats.firstMS != r.stats.expectedFirstMS ||
		r.stats.lastMS != r.stats.expectedLastMS {
		return fmt.Errorf(
			"replay mismatch ticks=%d/%d runs=%d/%d range=%d..%d/%d..%d",
			r.stats.ticksServed, r.stats.ticksExpected,
			r.stats.runsTriggered, r.stats.runsExpected,
			r.stats.firstMS, r.stats.lastMS,
			r.stats.expectedFirstMS, r.stats.expectedLastMS,
		)
	}
	r.stats.replayCompleted = true
	return nil
}

// Section 3 - Generic Helpers
