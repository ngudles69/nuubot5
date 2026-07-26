package btbot

import (
	"context"
	"fmt"
	stdruntime "runtime"
	"time"

	"github.com/shopspring/decimal"

	"nuubot/internal/botspec"
	"nuubot/internal/controller"
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

// BtBot owns one bounded historical replay.
type BtBot struct {
	log           *logging.Logger
	reader        replay.Reader
	clock         clock.Clock
	controller    controller.Controller
	symbol        string
	resultPath    string
	stats         stats
	telemetry     []telemetry.Sample
	report        report.Run
	stopRequested bool
	published     bool
	started       bool
	stopped       bool
}

// Section 1 - Program Flow

// Init prepares one bounded historical replay.
func (r *BtBot) Init(
	caller context.Context,
	log *logging.Logger,
	sweepID,
	botID uint64,
) error {
	r.log = log

	// prepare setup
	var admission, err = setup.Setup(caller, log, sweepID, botID)
	if err != nil {
		return fmt.Errorf("prepare setup: %w", err)
	}

	// set replay range
	var replayInput = admission.Bot.Replay
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

	// create clock
	r.clock, err = clock.Create(clock.Tick)
	if err != nil {
		return fmt.Errorf("create clock: %w", err)
	}

	// initialize clock
	err = r.clock.Init(log, startMS)
	if err != nil {
		return fmt.Errorf("initialize clock: %w", err)
	}

	// register Controller timer
	err = r.clock.RegisterTimer(clock.Timer{
		Name:       "controller",
		IntervalMS: admission.App.BtBot.TimerIntervalMS,
	}, r.controllerRun)
	if err != nil {
		return fmt.Errorf("register Controller timer: %w", err)
	}

	// initialize replay reader
	err = r.reader.Init(
		log,
		replayInput.TicksPath,
		start,
		end,
	)
	if err != nil {
		return fmt.Errorf("initialize replay reader: %w", err)
	}

	// build exact BotSpec
	var definition = botspec.BuildInput{
		Bot:  admission.Bot,
		Meta: admission.Meta,
		MinNotionalUSDC: decimal.NewFromInt(
			int64(admission.App.Hyperliquid.MinOrderNotionalUSDC),
		),
		ResultPath: admission.ResultPath,
	}
	var builtDefinition, buildErr = botspec.Build(log, definition)
	if buildErr != nil {
		return fmt.Errorf("build BotSpec: %w", buildErr)
	}

	// initialize Controller
	err = r.controller.Init(log, builtDefinition)
	if err != nil {
		builtDefinition.Signaler.Stop()
		for index := len(builtDefinition.Risks) - 1; index >= 0; index-- {
			builtDefinition.Risks[index].Stop()
		}
		return fmt.Errorf("initialize Controller: %w", err)
	}

	// create proof
	r.stats = stats{
		ticksExpected: durationMS / 1000,
		runsExpected: (durationMS + admission.App.BtBot.TimerIntervalMS - 1) /
			admission.App.BtBot.TimerIntervalMS,
		expectedFirstMS: startMS + 1000,
		expectedLastMS:  endMS,
	}

	r.symbol = replayInput.Symbol
	r.resultPath = admission.ResultPath
	log.Info(fmt.Sprintf(
		"btbot initialized bot_spec=%s symbol=%s",
		admission.Bot.BotSpecID,
		r.symbol,
	))
	return nil
}

// Start starts the owned Clock and Controller.
func (r *BtBot) Start() error {
	if r.started || r.stopped {
		return fmt.Errorf("btbot cannot start from current state")
	}

	// start clock
	var err = r.clock.Start()
	if err != nil {
		return fmt.Errorf("start clock: %w", err)
	}

	// start Controller
	err = r.controller.Start()
	if err != nil {
		r.clock.Stop()
		return fmt.Errorf("start Controller: %w", err)
	}
	r.started = true
	r.log.Info("btbot started")
	return nil
}

// Loop executes the complete bounded replay loop.
func (r *BtBot) Loop() error {
	if !r.started || r.stopped {
		return fmt.Errorf("btbot cannot loop from current state")
	}
	r.log.Info("btbot loop started")
	var started = time.Now()
	defer func() {
		r.stats.historicalDataLoopElapsed = time.Since(started)
	}()

	for {

		// read replay
		var bbo, ok, err = r.reader.Next()
		if err != nil {
			return fmt.Errorf("read replay: %w", err)
		}
		if !ok {
			break
		}

		// ingest Controller BBO
		bbo.Symbol = r.symbol
		err = r.controller.IngestBBO(bbo)
		if err != nil {
			return fmt.Errorf("ingest Controller BBO: %w", err)
		}

		// record proof
		if r.stats.firstMS == 0 {
			r.stats.firstMS = bbo.TimestampMS
		}
		r.stats.lastMS = bbo.TimestampMS
		r.stats.ticksServed++

		// advance clock
		err = r.clock.Advance(bbo.TimestampMS)
		if err != nil {
			return fmt.Errorf("advance clock: %w", err)
		}

		// check stop request
		if r.stopRequested {
			break
		}
	}

	// verify replay
	var err = r.verify()
	if err != nil {
		return fmt.Errorf("verify replay: %w", err)
	}
	return nil
}

// Stop releases owned resources and reports final proof.
func (r *BtBot) Stop() error {
	if r.stopped {
		return nil
	}
	r.started = false
	r.stopped = true

	// stop clock
	r.clock.Stop()

	// stop replay reader
	var readerErr = r.reader.Stop()
	if readerErr != nil {
		readerErr = fmt.Errorf("stop replay reader: %w", readerErr)
	}

	// stop Controller
	var controllerErr = r.controller.Stop("parent_stop")
	if controllerErr != nil {
		controllerErr = fmt.Errorf("stop Controller: %w", controllerErr)
	}

	// build and publish completed result
	var publishErr error
	if readerErr == nil && controllerErr == nil && r.stats.replayCompleted {
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
		r.report, publishErr = report.Build(input)
		if publishErr == nil {
			publishErr = resultpublisher.Publish(r.resultPath, input, r.report)
		}
		if publishErr != nil {
			publishErr = fmt.Errorf("build or publish result: %w", publishErr)
		} else {
			r.published = true
		}
	}

	// report proof
	var result = "failed"
	if readerErr == nil && controllerErr == nil && publishErr == nil &&
		r.stats.replayCompleted {
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

	// return stop errors
	if controllerErr != nil {
		return controllerErr
	}
	if readerErr != nil {
		return readerErr
	}
	if publishErr != nil {
		return publishErr
	}
	if !r.stats.replayCompleted {
		return fmt.Errorf("btbot replay did not complete")
	}

	return nil
}

// Section 2 - Domain Helpers

func (r *BtBot) controllerRun(nowMS uint64) error {
	// run Controller
	r.stats.runsTriggered++
	var stop, err = r.controller.Run(nowMS)
	if err != nil {
		return fmt.Errorf("run Controller: %w", err)
	}
	r.collectTelemetry(nowMS, false)
	// remember stop request
	if stop {
		r.stopRequested = true
	}
	return nil
}

// Result returns one complete terminal backtest result.
func (r *BtBot) Result() (Result, error) {
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

func (r *BtBot) collectTelemetry(nowMS uint64, terminal bool) {
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

func (r *BtBot) verify() error {
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
