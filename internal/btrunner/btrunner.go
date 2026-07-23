package btrunner

import (
	"fmt"
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

type BtRunner struct {
	log     *common.Logger
	reader  *replay.Reader
	clock   *clock.TickClock
	runtime *runtime.Runtime
	stats   stats
	started bool
	stopped bool
}

func New(
	log *common.Logger,
	root string,
	cfg config.Config,
	sweepID uint64,
	botID uint64,
) (*BtRunner, error) {
	ctx, err := setup.Init(log, root, cfg, sweepID, botID)
	if err != nil {
		return nil, err
	}
	end := ctx.Bot.ReplayEnd
	if ctx.Bot.EndAt != nil && ctx.Bot.EndAt.Before(end) {
		end = *ctx.Bot.EndAt
	}
	if !ctx.Bot.ReplayStart.Before(end) {
		return nil, fmt.Errorf("Bot end must follow replay start")
	}

	log.Info("btrunner", "init sweep_id=%d bot_id=%d symbol=%s", sweepID, botID, ctx.Bot.Symbol)
	tickClock := clock.New(log, cfg.BtRunner.TimerIntervalMS)
	tickReader, err := replay.NewReader(log, ctx.Bot.TicksPath, ctx.Bot.ReplayStart, end)
	if err != nil {
		return nil, err
	}
	runTime, err := runtime.New(log, cfg.Runtime, &end)
	if err != nil {
		return nil, err
	}
	loadedBars, err := bars.Load(log, ctx.Bot.TicksPath, ctx.Bot.ReplayStart, end, runTime.BarsNeeded())
	if err != nil {
		return nil, err
	}
	if err := runTime.PrepareBars(loadedBars); err != nil {
		return nil, err
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

func (r *BtRunner) Start() error {
	if r.started || r.stopped {
		return common.StateError("BtRunner", "start")
	}
	if err := r.runtime.Start(); err != nil {
		return err
	}
	r.started = true
	r.log.Info("btrunner", "start")
	return nil
}

func (r *BtRunner) Run() error {
	if !r.started || r.stopped {
		return common.StateError("BtRunner", "run")
	}
	r.log.Info("btrunner", "run")
	started := time.Now()
	defer func() { r.stats.elapsed = time.Since(started) }()

	for {
		bbo, ok, err := r.reader.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		if err := r.runtime.Ingest(bbo); err != nil {
			return err
		}
		if r.stats.firstMS == 0 {
			r.stats.firstMS = bbo.TimestampMS
		}
		r.stats.lastMS = bbo.TimestampMS
		r.stats.ticksServed++

		due, err := r.clock.Advance(bbo.TimestampMS)
		if err != nil {
			return err
		}
		if due {
			r.stats.passesTriggered++
			stop, err := r.runtime.MainLoop(bbo.TimestampMS)
			if err != nil {
				return err
			}
			if stop {
				break
			}
		}
	}
	if err := r.runtime.Stop("end_date"); err != nil {
		return err
	}
	return r.verify()
}

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

func (r *BtRunner) Stop() error {
	if r.stopped {
		return nil
	}
	r.started = false
	r.stopped = true
	runtimeErr := r.runtime.Stop("parent_stop")
	readerErr := r.reader.Stop()
	r.clock.Stop()

	status := "failed"
	if r.stats.replayCompleted && runtimeErr == nil && readerErr == nil {
		status = "success"
	}
	var memory stdruntime.MemStats
	stdruntime.ReadMemStats(&memory)
	r.log.Info(
		"btrunner",
		"stop status=%s loader=parquet ticks=%d/%d passes=%d/%d first_ts_ms=%d last_ts_ms=%d replay_completed=%t replay_ms=%d heap_mb=%.1f total_alloc_mb=%.1f gc_runs=%d gc_pause_ms=%.3f",
		status, r.stats.ticksServed, r.stats.ticksExpected,
		r.stats.passesTriggered, r.stats.passesExpected,
		r.stats.firstMS, r.stats.lastMS, r.stats.replayCompleted,
		r.stats.elapsed.Milliseconds(), float64(memory.HeapAlloc)/(1<<20),
		float64(memory.TotalAlloc)/(1<<20), memory.NumGC,
		float64(memory.PauseTotalNs)/1e6,
	)
	if runtimeErr != nil {
		return runtimeErr
	}
	if readerErr != nil {
		return readerErr
	}
	if !r.stats.replayCompleted {
		return fmt.Errorf("BtRunner replay did not complete")
	}
	return nil
}
