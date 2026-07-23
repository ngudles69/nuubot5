package btrunner

import (
	"errors"
	"fmt"
	"log/slog"
	stdruntime "runtime"
	"time"

	"nuubot/internal/replay"
	"nuubot/internal/runtime"
	"nuubot/internal/setup"
	"nuubot/internal/toolkit/clock"
	nuuerrors "nuubot/internal/toolkit/errors"
)

type stats struct {
	ticksExpected   uint64
	ticksServed     uint64
	passesExpected  uint64
	passesTriggered uint64
	expectedFirstMS uint64
	expectedLastMS  uint64
	firstMS         uint64
	lastMS          uint64
	replayCompleted bool
	elapsed         time.Duration
}

// BtRunner owns one bounded historical replay.
type BtRunner struct {
	log     *slog.Logger
	reader  *replay.Reader
	clock   *clock.TickClock
	runtime *runtime.Runtime
	stats   stats
	started bool
	stopped bool
}

// Section 1 - Program Flow

// Run executes one complete BtRunner command.
func Run(logger *slog.Logger, sweepID, botID uint64) error {
	var runner, err = Init(logger, sweepID, botID)
	if err != nil {
		return fmt.Errorf("create btrunner: %w", err)
	}
	err = runner.Start()
	if err != nil {
		return fmt.Errorf("start btrunner: %w", err)
	}
	var runErr = runner.Run()
	if runErr != nil {
		runErr = fmt.Errorf("run btrunner: %w", runErr)
	}
	var stopErr = runner.Stop()
	if stopErr != nil {
		stopErr = fmt.Errorf("stop btrunner: %w", stopErr)
	}
	return errors.Join(runErr, stopErr)
}

// Init constructs one bounded historical replay.
func Init(logger *slog.Logger, sweepID, botID uint64) (*BtRunner, error) {
	var ctx, err = setup.Init(logger, sweepID, botID)
	if err != nil {
		return nil, fmt.Errorf("initialize setup: %w", err)
	}
	var end = ctx.Bot.ReplayEnd
	if ctx.Bot.EndAt != nil && ctx.Bot.EndAt.Before(end) {
		end = *ctx.Bot.EndAt
	}
	if !ctx.Bot.ReplayStart.Before(end) {
		return nil, fmt.Errorf("bot end must follow replay start")
	}

	var log = logger.With("component", "btrunner")
	log.Info(
		"btrunner initialized",
		"event", "init",
		"status", "success",
		"symbol", ctx.Bot.Symbol,
	)
	var tickClock = clock.New(logger, ctx.Config.BtRunner.TimerIntervalMS)
	var tickReader, readerErr = replay.NewReader(
		logger,
		ctx.Bot.TicksPath,
		ctx.Bot.ReplayStart,
		end,
	)
	if readerErr != nil {
		return nil, fmt.Errorf("create replay reader: %w", readerErr)
	}
	var runTime, runtimeErr = runtime.Init(logger, ctx, end)
	if runtimeErr != nil {
		return nil, fmt.Errorf("initialize runtime: %w", runtimeErr)
	}

	var startMS = uint64(ctx.Bot.ReplayStart.UnixMilli())
	var endMS = uint64(end.UnixMilli())
	var durationMS = endMS - startMS
	return &BtRunner{
		log: log, reader: tickReader, clock: tickClock, runtime: runTime,
		stats: stats{
			ticksExpected: durationMS / 1000,
			passesExpected: (durationMS + ctx.Config.BtRunner.TimerIntervalMS - 1) /
				ctx.Config.BtRunner.TimerIntervalMS,
			expectedFirstMS: startMS + 1000,
			expectedLastMS:  endMS,
		},
	}, nil
}

// Start starts the owned Runtime.
func (r *BtRunner) Start() error {
	if r.started || r.stopped {
		return nuuerrors.StateError("btrunner", "start")
	}
	var err = r.runtime.Start()
	if err != nil {
		return fmt.Errorf("start runtime: %w", err)
	}
	r.started = true
	r.log.Info("btrunner started", "event", "start", "status", "success")
	return nil
}

// Run executes the complete bounded replay.
func (r *BtRunner) Run() error {
	if !r.started || r.stopped {
		return nuuerrors.StateError("btrunner", "run")
	}
	r.log.Info("btrunner running", "event", "run", "status", "started")
	var started = time.Now()
	defer func() { r.stats.elapsed = time.Since(started) }()

	for {
		var bbo, ok, err = r.reader.Next()
		if err != nil {
			return fmt.Errorf("read replay: %w", err)
		}
		if !ok {
			break
		}
		err = r.runtime.Ingest(bbo)
		if err != nil {
			return fmt.Errorf("ingest runtime bbo: %w", err)
		}
		if r.stats.firstMS == 0 {
			r.stats.firstMS = bbo.TimestampMS
		}
		r.stats.lastMS = bbo.TimestampMS
		r.stats.ticksServed++

		var due bool
		due, err = r.clock.Advance(bbo.TimestampMS)
		if err != nil {
			return fmt.Errorf("advance tick clock: %w", err)
		}
		if due {
			r.stats.passesTriggered++
			var stop bool
			stop, err = r.runtime.Pass(bbo.TimestampMS)
			if err != nil {
				return fmt.Errorf("pass runtime: %w", err)
			}
			if stop {
				break
			}
		}
	}
	var err = r.verify()
	if err != nil {
		return fmt.Errorf("verify replay: %w", err)
	}
	return nil
}

// Stop releases owned resources and reports final proof.
func (r *BtRunner) Stop() error {
	if r.stopped {
		return nil
	}
	r.started = false
	r.stopped = true
	var runtimeErr = r.runtime.Stop("parent_stop")
	if runtimeErr != nil {
		runtimeErr = fmt.Errorf("stop runtime: %w", runtimeErr)
	}
	var readerErr = r.reader.Stop()
	if readerErr != nil {
		readerErr = fmt.Errorf("stop replay reader: %w", readerErr)
	}
	r.clock.Stop()

	var status = "failed"
	if r.stats.replayCompleted && runtimeErr == nil && readerErr == nil {
		status = "success"
	}
	var memory stdruntime.MemStats
	stdruntime.ReadMemStats(&memory)
	r.log.Info(
		"btrunner stopped",
		"event", "stop",
		"status", status,
		"loader", "parquet",
		"ticks_served", r.stats.ticksServed,
		"ticks_expected", r.stats.ticksExpected,
		"passes_triggered", r.stats.passesTriggered,
		"passes_expected", r.stats.passesExpected,
		"first_ts_ms", r.stats.firstMS,
		"last_ts_ms", r.stats.lastMS,
		"replay_completed", r.stats.replayCompleted,
		"replay_ms", r.stats.elapsed.Milliseconds(),
		"heap_mb", float64(memory.HeapAlloc)/(1<<20),
		"total_alloc_mb", float64(memory.TotalAlloc)/(1<<20),
		"gc_runs", memory.NumGC,
		"gc_pause_ms", float64(memory.PauseTotalNs)/1e6,
	)
	if runtimeErr != nil {
		return runtimeErr
	}
	if readerErr != nil {
		return readerErr
	}
	if !r.stats.replayCompleted {
		return fmt.Errorf("btrunner replay did not complete")
	}
	return nil
}

// Section 2 - Domain Helpers

func (r *BtRunner) verify() error {
	if r.stats.ticksServed != r.stats.ticksExpected ||
		r.stats.passesTriggered != r.stats.passesExpected ||
		r.stats.firstMS != r.stats.expectedFirstMS ||
		r.stats.lastMS != r.stats.expectedLastMS {
		return fmt.Errorf(
			"replay mismatch ticks=%d/%d passes=%d/%d range=%d..%d/%d..%d",
			r.stats.ticksServed, r.stats.ticksExpected,
			r.stats.passesTriggered, r.stats.passesExpected,
			r.stats.firstMS, r.stats.lastMS,
			r.stats.expectedFirstMS, r.stats.expectedLastMS,
		)
	}
	r.stats.replayCompleted = true
	return nil
}

// Section 3 - Generic Helpers
