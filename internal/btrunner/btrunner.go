package btrunner

import (
	"fmt"
	"log/slog"
	stdruntime "runtime"
	"time"

	"nuubot5/internal/bars"
	"nuubot5/internal/clock"
	"nuubot5/internal/common"
	"nuubot5/internal/config"
	"nuubot5/internal/replay"
	"nuubot5/internal/runtime"
	"nuubot5/internal/setup"
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

// Program Flow

// New constructs one bounded historical replay.
func New(
	logger *slog.Logger,
	root string,
	cfg config.Config,
	sweepID uint64,
	botID uint64,
) (*BtRunner, error) {
	ctx, err := setup.Init(logger, root, cfg, sweepID, botID)
	if err != nil {
		return nil, fmt.Errorf("initialize setup: %w", err)
	}
	end := ctx.Bot.ReplayEnd
	if ctx.Bot.EndAt != nil && ctx.Bot.EndAt.Before(end) {
		end = *ctx.Bot.EndAt
	}
	if !ctx.Bot.ReplayStart.Before(end) {
		return nil, fmt.Errorf("bot end must follow replay start")
	}

	log := logger.With("component", "btrunner")
	log.Info(
		"btrunner initialized",
		"event", "init",
		"status", "success",
		"symbol", ctx.Bot.Symbol,
	)
	tickClock := clock.New(logger, cfg.BtRunner.TimerIntervalMS)
	tickReader, err := replay.NewReader(logger, ctx.Bot.TicksPath, ctx.Bot.ReplayStart, end)
	if err != nil {
		return nil, fmt.Errorf("create replay reader: %w", err)
	}
	runTime, err := runtime.New(logger, cfg.Runtime, &end)
	if err != nil {
		return nil, fmt.Errorf("create runtime: %w", err)
	}
	loadedBars, err := bars.Load(logger, ctx.Bot.TicksPath, ctx.Bot.ReplayStart, end, runTime.BarsNeeded())
	if err != nil {
		return nil, fmt.Errorf("load bars: %w", err)
	}
	if err := runTime.PrepareBars(loadedBars); err != nil {
		return nil, fmt.Errorf("prepare runtime bars: %w", err)
	}

	startMS := uint64(ctx.Bot.ReplayStart.UnixMilli())
	endMS := uint64(end.UnixMilli())
	durationMS := endMS - startMS
	return &BtRunner{
		log: log, reader: tickReader, clock: tickClock, runtime: runTime,
		stats: stats{
			ticksExpected: durationMS / 1000,
			passesExpected: (durationMS + cfg.BtRunner.TimerIntervalMS - 1) /
				cfg.BtRunner.TimerIntervalMS,
			expectedFirstMS: startMS + 1000,
			expectedLastMS:  endMS,
		},
	}, nil
}

// Start starts the owned Runtime.
func (r *BtRunner) Start() error {
	if r.started || r.stopped {
		return common.StateError("btrunner", "start")
	}
	if err := r.runtime.Start(); err != nil {
		return fmt.Errorf("start runtime: %w", err)
	}
	r.started = true
	r.log.Info("btrunner started", "event", "start", "status", "success")
	return nil
}

// Run executes the complete bounded replay.
func (r *BtRunner) Run() error {
	if !r.started || r.stopped {
		return common.StateError("btrunner", "run")
	}
	r.log.Info("btrunner running", "event", "run", "status", "started")
	started := time.Now()
	defer func() { r.stats.elapsed = time.Since(started) }()

	for {
		bbo, ok, err := r.reader.Next()
		if err != nil {
			return fmt.Errorf("read replay: %w", err)
		}
		if !ok {
			break
		}
		if err := r.runtime.Ingest(bbo); err != nil {
			return fmt.Errorf("ingest runtime bbo: %w", err)
		}
		if r.stats.firstMS == 0 {
			r.stats.firstMS = bbo.TimestampMS
		}
		r.stats.lastMS = bbo.TimestampMS
		r.stats.ticksServed++

		due, err := r.clock.Advance(bbo.TimestampMS)
		if err != nil {
			return fmt.Errorf("advance tick clock: %w", err)
		}
		if due {
			r.stats.passesTriggered++
			stop, err := r.runtime.Pass(bbo.TimestampMS)
			if err != nil {
				return fmt.Errorf("pass runtime: %w", err)
			}
			if stop {
				break
			}
		}
	}
	if err := r.runtime.Stop("end_date"); err != nil {
		return fmt.Errorf("stop runtime at end date: %w", err)
	}
	if err := r.verify(); err != nil {
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
	runtimeErr := r.runtime.Stop("parent_stop")
	if runtimeErr != nil {
		runtimeErr = fmt.Errorf("stop runtime: %w", runtimeErr)
	}
	readerErr := r.reader.Stop()
	if readerErr != nil {
		readerErr = fmt.Errorf("stop replay reader: %w", readerErr)
	}
	r.clock.Stop()

	status := "failed"
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

// Domain Helpers

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
