package runtime

import (
	"errors"
	"fmt"
	"time"

	"nuubot/internal/botcycle"
	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/risk"
	"nuubot/internal/setup"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

type stats struct {
	ticks               uint64
	runs                uint64
	signalPackagesRead  uint64
	entrySignalsSkipped uint64
	cyclesStarted       uint64
	cyclesRejected      uint64
	cyclesClosed        uint64
	stopLossExits       uint64
}

// Runtime owns synchronous Bot decisions and its direct children.
type Runtime struct {
	log          *logging.Logger
	config       config.Runtime
	symbol       string
	signaler     signaler.Signaler
	risks        []risk.Risk
	cycle        *botcycle.Control
	stats        stats
	lastSignalMS uint64
	stopReason   string
	started      bool
	stopped      bool
}

// Section 1 - Program Flow

// Init prepares the Runtime and its configured children.
func (r *Runtime) Init(log *logging.Logger, ctx setup.Context, end time.Time) error {
	r.log = log
	r.config = ctx.Config.Runtime
	r.symbol = ctx.Bot.Symbol

	// create signaler
	var err error
	r.signaler, err = signaler.Create(
		log,
		r.config.Signaler,
		r.symbol,
		ctx.Bot.TicksPath,
		ctx.Bot.ReplayStart,
		end,
	)
	if err != nil {
		return fmt.Errorf("create signaler: %w", err)
	}

	// create risks
	for index, riskConfig := range r.config.Risks {
		var created, riskErr = risk.Create(log, index+1, riskConfig)
		if riskErr != nil {
			r.signaler.Stop()
			return fmt.Errorf("create risk %d: %w", index+1, riskErr)
		}
		r.risks = append(r.risks, created)
	}

	// initialize runtime
	log.Info("runtime initialized")
	return nil
}

// Start starts Runtime admission.
func (r *Runtime) Start() error {
	if r.started || r.stopped {
		return fmt.Errorf("runtime cannot start from current state")
	}

	// start runtime
	r.started = true
	r.log.Info("runtime started")
	return nil
}

// Run executes one timer-driven control pass.
func (r *Runtime) Run(nowMS uint64) (bool, error) {
	if !r.started || r.stopped {
		return false, fmt.Errorf("runtime cannot run from current state")
	}
	r.stats.runs++

	// assess risk stops
	for _, activeRisk := range r.risks {
		if activeRisk.AssessStop() {
			r.requestStop("risk")
		}
	}

	// check stop request
	if r.stopReason != "" {
		return true, nil
	}

	// check botcycle
	if r.cycle != nil {
		var completed, err = r.cycle.Run(nowMS)
		if err != nil {
			return false, fmt.Errorf("run bot cycle: %w", err)
		}
		if completed {
			err = r.closeCycle("completed")
			if err != nil {
				return false, fmt.Errorf("close completed bot cycle: %w", err)
			}
		}
	}

	// check max cycles
	if r.stats.cyclesClosed >= r.config.MaxCycles {
		r.requestStop("max_cycles")
		return true, nil
	}

	// read signal
	var packages = r.signaler.Signals(r.symbol, nowMS, 1)
	if len(packages) == 0 {
		return false, nil
	}
	var signalPackage = packages[len(packages)-1]
	if signalPackage.TimestampMS() <= r.lastSignalMS {
		return false, nil
	}

	// consume signal
	r.lastSignalMS = signalPackage.TimestampMS()
	r.stats.signalPackagesRead++
	var enterLong = signalPackage.EnterLong()
	var enterShort = signalPackage.EnterShort()
	if enterLong && enterShort {
		return false, fmt.Errorf(
			"signal package %d contains conflicting entry triggers",
			signalPackage.TimestampMS(),
		)
	}
	if !enterLong && !enterShort {
		return false, nil
	}
	if r.cycle != nil {
		r.stats.entrySignalsSkipped++
		return false, nil
	}

	// open botcycle
	var err = r.openCycle(signalPackage)
	if errors.Is(err, botcycle.ErrRejected) {
		r.stats.cyclesRejected++
		r.log.Warning(fmt.Sprintf(
			"bot cycle rejected signal_ts_ms=%d reason=%v",
			signalPackage.TimestampMS(),
			err,
		))
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open bot cycle: %w", err)
	}
	return false, nil
}

// Stop closes the active BotCycle and stops children.
func (r *Runtime) Stop(reason string) error {
	if r.stopped {
		return nil
	}

	// request stop
	r.requestStop(reason)
	r.started = false

	// stop botcycle
	var firstErr = r.closeCycle(r.stopReason)

	// stop risks
	for index := len(r.risks) - 1; index >= 0; index-- {
		r.risks[index].Stop()
	}

	// stop signaler
	r.signaler.Stop()

	// stop runtime
	r.stopped = true
	r.log.Info(fmt.Sprintf(
		"runtime stopped ticks_accepted=%d runs=%d signal_packages_read=%d "+
			"entry_signals_skipped=%d cycles_started=%d cycles_rejected=%d "+
			"cycles_closed=%d stop_loss_exits=%d stop_reason=%s",
		r.stats.ticks,
		r.stats.runs,
		r.stats.signalPackagesRead,
		r.stats.entrySignalsSkipped,
		r.stats.cyclesStarted,
		r.stats.cyclesRejected,
		r.stats.cyclesClosed,
		r.stats.stopLossExits,
		r.stopReason,
	))
	return firstErr
}

// Section 2 - Domain Helpers

// IngestBBO accepts one validated BBO.
func (r *Runtime) IngestBBO(bbo market.BBO) error {
	if !r.started || r.stopped || r.stopReason != "" {
		return fmt.Errorf("runtime cannot ingest bbo from current state")
	}

	// ingest botcycle bbo
	if r.cycle != nil {
		var err = r.cycle.IngestBBO(bbo)
		if err != nil {
			return fmt.Errorf("ingest bot cycle bbo: %w", err)
		}

		// deliver botcycle bbo
		r.cycle.OnBBO(bbo)
	}
	r.stats.ticks++
	return nil
}

func (r *Runtime) openCycle(signal signaler.Package) error {
	// initialize botcycle
	var cycle botcycle.Control
	var err = cycle.Init(
		r.log,
		int(r.stats.cyclesStarted+r.stats.cyclesRejected+1),
		r.signaler,
		signal,
		r.config.Executors,
	)
	if err != nil {
		return err
	}
	r.cycle = &cycle
	r.stats.cyclesStarted++
	return nil
}

func (r *Runtime) closeCycle(reason string) error {
	if r.cycle == nil {
		return nil
	}
	var cycle = r.cycle
	r.cycle = nil
	var exitReason, err = cycle.Stop(reason)
	if err != nil {
		return fmt.Errorf("stop bot cycle: %w", err)
	}
	r.stats.cyclesClosed++
	if exitReason == "stop_loss" {
		r.stats.stopLossExits++
	}
	return nil
}

func (r *Runtime) requestStop(reason string) {
	if r.stopReason == "" {
		r.stopReason = reason
	}
}

// Section 3 - Generic Helpers
