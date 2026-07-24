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
	runsExpected    uint64
	runsTriggered   uint64
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
	reader        replay.Reader
	clock         clock.Clock
	runtime       runtime.Runtime
	stats         stats
	stopRequested bool
	started       bool
	stopped       bool
}

// Section 1 - Program Flow

// Init prepares one bounded historical replay.
func (r *BtRunner) Init(log *logging.Logger, sweepID, botID uint64) error {
	r.log = log

	// initialize setup
	var ctx, err = setup.Init(log, sweepID, botID)
	if err != nil {
		return fmt.Errorf("initialize setup: %w", err)
	}

	// set replay range
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
	r.clock, err = clock.Create(clock.Tick)
	if err != nil {
		return fmt.Errorf("create clock: %w", err)
	}

	// initialize clock
	err = r.clock.Init(log, startMS)
	if err != nil {
		return fmt.Errorf("initialize clock: %w", err)
	}

	// register runtime timer
	err = r.clock.RegisterTimer(clock.Timer{
		Name:       "runtime",
		IntervalMS: ctx.Config.BtRunner.TimerIntervalMS,
	}, r.runtimeRun)
	if err != nil {
		return fmt.Errorf("register runtime timer: %w", err)
	}

	// initialize replay reader
	err = r.reader.Init(
		log,
		ctx.Bot.TicksPath,
		start,
		end,
	)
	if err != nil {
		return fmt.Errorf("initialize replay reader: %w", err)
	}

	// initialize runtime
	err = r.runtime.Init(log, ctx, end)
	if err != nil {
		return fmt.Errorf("initialize runtime: %w", err)
	}

	// create proof
	r.stats = stats{
		ticksExpected: durationMS / 1000,
		runsExpected: (durationMS + ctx.Config.BtRunner.TimerIntervalMS - 1) /
			ctx.Config.BtRunner.TimerIntervalMS,
		expectedFirstMS: startMS + 1000,
		expectedLastMS:  endMS,
	}

	log.Info(fmt.Sprintf("btrunner initialized: symbol=%s", ctx.Bot.Symbol))
	return nil
}

// Start starts the owned Clock and Runtime.
func (r *BtRunner) Start() error {
	if r.started || r.stopped {
		return fmt.Errorf("btrunner cannot start from current state")
	}

	// start clock
	var err = r.clock.Start()
	if err != nil {
		return fmt.Errorf("start clock: %w", err)
	}

	// start runtime
	err = r.runtime.Start()
	if err != nil {
		r.clock.Stop()
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

		// read replay
		var bbo, ok, err = r.reader.Next()
		if err != nil {
			return fmt.Errorf("read replay: %w", err)
		}
		if !ok {
			break
		}

		// ingest runtime bbo
		err = r.runtime.Ingest(bbo)
		if err != nil {
			return fmt.Errorf("ingest runtime bbo: %w", err)
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
func (r *BtRunner) Stop() error {
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

	// stop runtime
	var runtimeErr = r.runtime.Stop("parent_stop")
	if runtimeErr != nil {
		runtimeErr = fmt.Errorf("stop runtime: %w", runtimeErr)
	}

	// report proof
	var memory stdruntime.MemStats
	stdruntime.ReadMemStats(&memory)
	r.log.Info(fmt.Sprintf(
		"btrunner stopped loader=parquet ticks_served=%d ticks_expected=%d "+
			"runs_triggered=%d runs_expected=%d first_ts_ms=%d last_ts_ms=%d "+
			"replay_completed=%t replay_ms=%d heap_mb=%f total_alloc_mb=%f "+
			"gc_runs=%d gc_pause_ms=%f result=complete",
		r.stats.ticksServed,
		r.stats.ticksExpected,
		r.stats.runsTriggered,
		r.stats.runsExpected,
		r.stats.firstMS,
		r.stats.lastMS,
		r.stats.replayCompleted,
		r.stats.elapsed.Milliseconds(),
		float64(memory.HeapAlloc)/(1<<20),
		float64(memory.TotalAlloc)/(1<<20),
		memory.NumGC,
		float64(memory.PauseTotalNs)/1e6,
	))

	// return stop errors
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

func (r *BtRunner) runtimeRun(nowMS uint64) error {
	// run runtime
	r.stats.runsTriggered++
	var stop, err = r.runtime.Run(nowMS)
	if err != nil {
		return fmt.Errorf("run runtime: %w", err)
	}
	// remember stop request
	if stop {
		r.stopRequested = true
	}
	return nil
}

func (r *BtRunner) verify() error {
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
