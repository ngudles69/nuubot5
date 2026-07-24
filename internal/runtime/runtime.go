package runtime

import (
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
	ticks          uint64
	runs           uint64
	signals        uint64
	signalsSkipped uint64
	cyclesStarted  uint64
	cyclesClosed   uint64
	stopLossExits  uint64
}

// Runtime owns synchronous trading decisions and its direct children.
type Runtime struct {
	log        *logging.Logger
	config     config.Runtime
	signaler   signaler.Signaler
	risks      []risk.Risk
	cycle      *botcycle.Control
	stats      stats
	stopReason string
	started    bool
	stopped    bool
}

// Section 1 - Program Flow

// Init prepares the Runtime and its configured children.
func (r *Runtime) Init(log *logging.Logger, ctx setup.Context, end time.Time) error {
	r.log = log
	r.config = ctx.Config.Runtime

	// initialize signaler
	var err = r.signaler.Init(
		log,
		r.config.Signaler,
		ctx.Bot.TicksPath,
		ctx.Bot.ReplayStart,
		end,
	)
	if err != nil {
		return fmt.Errorf("initialize signaler: %w", err)
	}

	// create risks
	for index, riskConfig := range r.config.Risks {
		var created, riskErr = risk.Create(log, index+1, riskConfig)
		if riskErr != nil {
			return fmt.Errorf("create risk %d: %w", index+1, riskErr)
		}
		r.risks = append(r.risks, created)
	}

	// initialize runtime
	log.Info("runtime initialized")
	return nil
}

// Start starts Runtime children and admission.
func (r *Runtime) Start() error {
	if r.started || r.stopped {
		return fmt.Errorf("runtime cannot start from current state")
	}
	// start signaler
	var err = r.signaler.Start()
	if err != nil {
		return fmt.Errorf("start signaler: %w", err)
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
	if r.cycle == nil {
		return false, nil
	}

	// run botcycle
	var completed, err = r.cycle.Run(nowMS)
	if err != nil {
		return false, fmt.Errorf("run bot cycle: %w", err)
	}
	if !completed {
		return false, nil
	}

	// close botcycle
	err = r.closeCycle("completed")
	if err != nil {
		return false, fmt.Errorf("close completed bot cycle: %w", err)
	}

	// check max cycles
	if r.stats.cyclesClosed >= r.config.MaxCycles {
		r.requestStop("max_cycles")
		return true, nil
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
		"runtime stopped ticks_accepted=%d runs=%d signals_received=%d "+
			"signals_skipped=%d cycles_started=%d cycles_closed=%d "+
			"stop_loss_exits=%d stop_reason=%s",
		r.stats.ticks,
		r.stats.runs,
		r.stats.signals,
		r.stats.signalsSkipped,
		r.stats.cyclesStarted,
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

	// run signaler
	for {
		var signal, available, err = r.signaler.Run(bbo.TimestampMS)
		if err != nil {
			return fmt.Errorf("release signal: %w", err)
		}
		if !available {
			break
		}
		r.stats.signals++
		if r.cycle != nil {
			r.stats.signalsSkipped++
		} else {
			err = r.openCycle(signal)
			if err != nil {
				return fmt.Errorf("open bot cycle: %w", err)
			}
		}
	}

	// ingest botcycle bbo
	if r.cycle != nil {
		var ingestErr = r.cycle.IngestBBO(bbo)
		if ingestErr != nil {
			return fmt.Errorf("ingest bot cycle bbo: %w", ingestErr)
		}
		// deliver botcycle bbo
		r.cycle.OnBBO(bbo)
	}
	r.stats.ticks++
	return nil
}

func (r *Runtime) openCycle(signal signaler.Signal) error {
	// initialize botcycle
	var cycle botcycle.Control
	var err = cycle.Init(
		r.log,
		int(r.stats.cyclesStarted+1),
		signal,
		r.config.Executors,
	)
	if err != nil {
		return err
	}

	// start botcycle
	err = cycle.Start()
	if err != nil {
		return fmt.Errorf("start bot cycle: %w", err)
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
