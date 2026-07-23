package runtime

import (
	"fmt"
	"log/slog"
	"time"

	"nuubot/internal/botcycle"
	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/ohlcv"
	"nuubot/internal/risk"
	"nuubot/internal/setup"
	"nuubot/internal/signaler"
	nuuerrors "nuubot/internal/toolkit/errors"
)

type stats struct {
	ticks          uint64
	passes         uint64
	signals        uint64
	signalsSkipped uint64
	cyclesStarted  uint64
	cyclesClosed   uint64
	stopLossExits  uint64
	endDateExits   uint64
}

// Runtime owns synchronous trading decisions and its direct children.
type Runtime struct {
	logger     *slog.Logger
	log        *slog.Logger
	config     config.Runtime
	signaler   *signaler.Signaler
	risks      []risk.Risk
	cycle      *botcycle.Control
	endMS      uint64
	stats      stats
	stopReason string
	started    bool
	stopped    bool
}

// Section 1 - Program Flow

// Init constructs one Runtime and its configured children.
func Init(logger *slog.Logger, ctx setup.Context, end time.Time) (*Runtime, error) {
	var cfg = ctx.Config.Runtime
	var signals, err = signaler.Create(logger, cfg.Signaler)
	if err != nil {
		return nil, fmt.Errorf("create signaler: %w", err)
	}
	var requirements = signals.Requirements()
	var loaded = make([]signaler.Series, 0, len(requirements))
	for _, requirement := range requirements {
		var duration, durationErr = requirement.Interval.Duration()
		if durationErr != nil {
			return nil, fmt.Errorf("resolve signaler interval: %w", durationErr)
		}
		var start = ctx.Bot.ReplayStart.Add(-duration * time.Duration(requirement.PriorRows))
		var rows, loadErr = ohlcv.Load(ctx.Bot.TicksPath, requirement.Interval, start, end)
		if loadErr != nil {
			return nil, fmt.Errorf("load signaler OHLCV: %w", loadErr)
		}
		loaded = append(loaded, signaler.Series{Data: rows, PriorRows: requirement.PriorRows})
	}
	err = signals.Prepare(loaded)
	if err != nil {
		return nil, fmt.Errorf("prepare signaler: %w", err)
	}
	var risks = make([]risk.Risk, 0, len(cfg.Risks))
	for index, riskConfig := range cfg.Risks {
		var created, riskErr = risk.Create(logger, index+1, riskConfig)
		if riskErr != nil {
			return nil, fmt.Errorf("create risk %d: %w", index+1, riskErr)
		}
		risks = append(risks, created)
	}
	var endMS = uint64(end.UnixMilli())
	var log = logger.With("component", "runtime")
	log.Info(
		"runtime initialized",
		"event", "init",
		"status", "success",
		"end_ts_ms", endMS,
	)
	return &Runtime{
		logger: logger, log: log, config: cfg, signaler: signals, risks: risks, endMS: endMS,
	}, nil
}

// Start starts Runtime children and admission.
func (r *Runtime) Start() error {
	if r.started || r.stopped {
		return nuuerrors.StateError("runtime", "start")
	}
	var err = r.signaler.Start()
	if err != nil {
		return fmt.Errorf("start signaler: %w", err)
	}
	r.started = true
	r.log.Info("runtime started", "event", "start", "status", "success")
	return nil
}

// Pass executes one timer-driven control pass.
func (r *Runtime) Pass(nowMS uint64) (bool, error) {
	if !r.started || r.stopped {
		return false, nuuerrors.StateError("runtime", "pass")
	}
	r.stats.passes++

	for _, assessed := range r.risks {
		if assessed.Assess() {
			r.requestStop("risk")
		}
	}
	if r.stopReason != "" {
		return true, nil
	}
	if r.cycle == nil {
		return false, nil
	}
	var completed, err = r.cycle.Pass(nowMS)
	if err != nil {
		return false, fmt.Errorf("pass bot cycle: %w", err)
	}
	if !completed {
		return false, nil
	}
	err = r.closeCycle("completed")
	if err != nil {
		return false, fmt.Errorf("close completed bot cycle: %w", err)
	}
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
	r.requestStop(reason)
	r.started = false
	var firstErr = r.closeCycle(r.stopReason)
	for index := len(r.risks) - 1; index >= 0; index-- {
		r.risks[index].Stop()
	}
	r.signaler.Stop()
	r.stopped = true
	var status = "success"
	if firstErr != nil {
		status = "failed"
	}
	r.log.Info(
		"runtime stopped",
		"event", "stop",
		"status", status,
		"ticks_accepted", r.stats.ticks,
		"passes", r.stats.passes,
		"signals_received", r.stats.signals,
		"signals_skipped", r.stats.signalsSkipped,
		"cycles_started", r.stats.cyclesStarted,
		"cycles_closed", r.stats.cyclesClosed,
		"stop_loss_exits", r.stats.stopLossExits,
		"end_date_exits", r.stats.endDateExits,
		"stop_reason", r.stopReason,
	)
	return firstErr
}

// Section 2 - Domain Helpers

// Ingest accepts one validated BBO.
func (r *Runtime) Ingest(bbo market.BBO) error {
	if !r.started || r.stopped || r.stopReason != "" {
		return nuuerrors.StateError("runtime", "ingest bbo")
	}
	for {
		var signal, available, err = r.signaler.Next(bbo.TimestampMS)
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
	if r.cycle != nil {
		r.cycle.OnBBO(bbo)
	}
	r.stats.ticks++
	if r.endMS != 0 && bbo.TimestampMS >= r.endMS {
		r.requestStop("end_date")
	}
	return nil
}

func (r *Runtime) openCycle(signal signaler.Signal) error {
	var cycle, err = botcycle.New(
		r.logger,
		int(r.stats.cyclesStarted+1),
		signal,
		r.config.Executors,
	)
	if err != nil {
		return err
	}
	err = cycle.Start()
	if err != nil {
		return fmt.Errorf("start bot cycle: %w", err)
	}
	r.cycle = cycle
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
	switch exitReason {
	case "stop_loss":
		r.stats.stopLossExits++
	case "end_date":
		r.stats.endDateExits++
	}
	return nil
}

func (r *Runtime) requestStop(reason string) {
	if r.stopReason == "" {
		r.stopReason = reason
	}
}

// Section 3 - Generic Helpers
