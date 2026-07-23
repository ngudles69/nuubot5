package btrunner

import (
	"fmt"
	stdruntime "runtime"
	"time"

	"nuubot/internal/replay"
	"nuubot/internal/runtime"
	"nuubot/internal/setup"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
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
	log           *logging.Logger
	reader        *replay.Reader
	clock         *clock.TickClock
	runtime       *runtime.Runtime
	stats         stats
	stopRequested bool
	started       bool
	stopped       bool
}

// Section 1 - Program Flow

// Init prepares one bounded historical replay.
func (r *BtRunner) Init(log *logging.Logger, sweepID, botID uint64) error {
	r.log = log

	// setup
	var ctx, err = setup.Init(log, sweepID, botID)
	if err != nil {
		return fmt.Errorf("initialize setup: %w", err)
	}

	// validate and set StartMS and endMS
	var start = ctx.Bot.ReplayStart
	var end = ctx.Bot.ReplayEnd
	if ctx.Bot.EndAt != nil && ctx.Bot.EndAt.Before(end) {
		end = *ctx.Bot.EndAt
	}
	if !start.Before(end) {
		return fmt.Errorf("bot end must follow replay start")
	}
	var startMS = uint64(start.UnixMilli())
	var endMS = uint64(end.UnixMilli())
	var durationMS = endMS - startMS

	// create clock
	r.clock = clock.New(log)
	err = r.clock.RegisterTimer(ctx.Config.BtRunner.TimerIntervalMS, r.runtimeRun)
	if err != nil {
		return fmt.Errorf("register runtime timer: %w", err)
	}

	// create tickReader
	r.reader, err = replay.NewReader(
		log,
		ctx.Bot.TicksPath,
		start,
		end,
	)
	if err != nil {
		return fmt.Errorf("create replay reader: %w", err)
	}

	// create runtime
	r.runtime, err = runtime.Init(log, ctx, end)
	if err != nil {
		return fmt.Errorf("initialize runtime: %w", err)
	}

	// create stats
	r.stats = stats{
		ticksExpected: durationMS / 1000,
		passesExpected: (durationMS + ctx.Config.BtRunner.TimerIntervalMS - 1) /
			ctx.Config.BtRunner.TimerIntervalMS,
		expectedFirstMS: startMS + 1000,
		expectedLastMS:  endMS,
	}

	log.Info(fmt.Sprintf("btrunner initialized: symbol=%s", ctx.Bot.Symbol))
	return nil
}

// Start starts the owned Runtime.
func (r *BtRunner) Start() error {
	if r.started || r.stopped {
		return fmt.Errorf("btrunner cannot start from current state")
	}
	var err = r.runtime.Start()
	if err != nil {
		return fmt.Errorf("start runtime: %w", err)
	}
	r.started = true
	r.log.Info("btrunner started")
	return nil
}

// Loop executes the complete bounded replay loop.
func (r *BtRunner) Loop() error {
	if !r.started || r.stopped {
		return fmt.Errorf("btrunner cannot loop from current state")
	}
	r.log.Info("btrunner loop started")
	var started = time.Now()
	defer func() { r.stats.elapsed = time.Since(started) }()

	for {

		// get next bbo
		var bbo, ok, err = r.reader.Next()
		if err != nil {
			return fmt.Errorf("read replay: %w", err)
		}
		if !ok {
			break
		}

		// runtime ingest bbo
		err = r.runtime.Ingest(bbo)
		if err != nil {
			return fmt.Errorf("ingest runtime bbo: %w", err)
		}

		// update stats
		if r.stats.firstMS == 0 {
			r.stats.firstMS = bbo.TimestampMS
		}
		r.stats.lastMS = bbo.TimestampMS
		r.stats.ticksServed++

		// clock advance
		err = r.clock.Advance(bbo.TimestampMS)
		if err != nil {
			return fmt.Errorf("advance tick clock: %w", err)
		}
		if r.stopRequested {
			break
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

	// clock stop
	r.clock.Stop()

	// reader stop
	var readerErr = r.reader.Stop()
	if readerErr != nil {
		readerErr = fmt.Errorf("stop replay reader: %w", readerErr)
	}

	// trigger runtime stop
	var runtimeErr = r.runtime.Stop("parent_stop")
	if runtimeErr != nil {
		runtimeErr = fmt.Errorf("stop runtime: %w", runtimeErr)
	}

	// get run stats
	var memory stdruntime.MemStats
	stdruntime.ReadMemStats(&memory)
	r.log.Info(fmt.Sprintf(
		"btrunner stopped loader=parquet ticks_served=%d ticks_expected=%d "+
			"passes_triggered=%d passes_expected=%d first_ts_ms=%d last_ts_ms=%d "+
			"replay_completed=%t replay_ms=%d heap_mb=%f total_alloc_mb=%f "+
			"gc_runs=%d gc_pause_ms=%f result=complete",
		r.stats.ticksServed,
		r.stats.ticksExpected,
		r.stats.passesTriggered,
		r.stats.passesExpected,
		r.stats.firstMS,
		r.stats.lastMS,
		r.stats.replayCompleted,
		r.stats.elapsed.Milliseconds(),
		float64(memory.HeapAlloc)/(1<<20),
		float64(memory.TotalAlloc)/(1<<20),
		memory.NumGC,
		float64(memory.PauseTotalNs)/1e6,
	))

	// check for errors
	if runtimeErr != nil {
		return runtimeErr
	}
	if readerErr != nil {
		return readerErr
	}
	if !r.stats.replayCompleted {
		return fmt.Errorf("btrunner replay did not complete")
	}

	// ok
	return nil
}

// Section 2 - Domain Helpers

func (r *BtRunner) runtimeRun(nowMS uint64) error {
	r.stats.passesTriggered++
	var stop, err = r.runtime.Run(nowMS)
	if err != nil {
		return fmt.Errorf("run runtime: %w", err)
	}
	if stop {
		r.stopRequested = true
	}
	return nil
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

// Section 3 - Generic Helpers
