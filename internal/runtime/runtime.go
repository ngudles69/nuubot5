package runtime

import (
	"fmt"
	"log/slog"
	"time"

	"nuubot5/internal/bars"
	"nuubot5/internal/botcycle"
	"nuubot5/internal/common"
	"nuubot5/internal/config"
	"nuubot5/internal/market"
	"nuubot5/internal/risk"
	"nuubot5/internal/signaler"
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

// Program Flow

// New constructs one Runtime and its configured children.
func New(logger *slog.Logger, cfg config.Runtime, endAt *time.Time) (*Runtime, error) {
	signals, err := signaler.New(logger, cfg.Signaler)
	if err != nil {
		return nil, fmt.Errorf("create signaler: %w", err)
	}
	risks := make([]risk.Risk, 0, len(cfg.Risks))
	for index, riskConfig := range cfg.Risks {
		created, err := risk.New(logger, index+1, riskConfig)
		if err != nil {
			return nil, fmt.Errorf("create risk %d: %w", index+1, err)
		}
		risks = append(risks, created)
	}
	var endMS uint64
	if endAt != nil {
		endMS = uint64(endAt.UnixMilli())
	}
	log := logger.With("component", "runtime")
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
		return common.StateError("runtime", "start")
	}
	if err := r.signaler.Start(); err != nil {
		return fmt.Errorf("start signaler: %w", err)
	}
	r.started = true
	r.log.Info("runtime started", "event", "start", "status", "success")
	return nil
}

// Pass executes one timer-driven control pass.
func (r *Runtime) Pass(nowMS uint64) (bool, error) {
	if !r.started || r.stopped {
		return false, common.StateError("runtime", "pass")
	}
	r.stats.passes++

	for _, assessed := range r.risks {
		if assessed.Assess() {
			r.requestStop("risk")
		}
	}
	if r.stopReason != "" {
		reason := r.stopReason
		return true, r.Stop(reason)
	}
	if r.cycle == nil {
		return false, nil
	}
	completed, err := r.cycle.Pass(nowMS)
	if err != nil {
		return false, fmt.Errorf("pass bot cycle: %w", err)
	}
	if !completed {
		return false, nil
	}
	if err := r.closeCycle("completed"); err != nil {
		return false, fmt.Errorf("close completed bot cycle: %w", err)
	}
	if r.stats.cyclesClosed >= r.config.MaxCycles {
		return true, r.Stop("max_cycles")
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
	firstErr := r.closeCycle(r.stopReason)
	for index := len(r.risks) - 1; index >= 0; index-- {
		r.risks[index].Stop()
	}
	r.signaler.Stop()
	r.stopped = true
	status := "success"
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

// Domain Helpers

// BarsNeeded returns Signaler bar requirements.
func (r *Runtime) BarsNeeded() []bars.Requirement {
	return r.signaler.BarsNeeded()
}

// PrepareBars prepares the Signaler with validated bars.
func (r *Runtime) PrepareBars(loaded []bars.Data) error {
	return r.signaler.Prepare(loaded)
}

// Ingest accepts one validated BBO.
func (r *Runtime) Ingest(bbo market.BBO) error {
	if !r.started || r.stopped || r.stopReason != "" {
		return common.StateError("runtime", "ingest bbo")
	}
	for {
		signal, available, err := r.signaler.Next(bbo.TimestampMS)
		if err != nil {
			return fmt.Errorf("release signal: %w", err)
		}
		if !available {
			break
		}
		r.stats.signals++
		if r.cycle != nil {
			r.stats.signalsSkipped++
		} else if err := r.openCycle(signal); err != nil {
			return fmt.Errorf("open bot cycle: %w", err)
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
	cycle, err := botcycle.New(r.logger, int(r.stats.cyclesStarted+1), signal, r.config.Executors)
	if err != nil {
		return err
	}
	if err := cycle.Start(); err != nil {
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
	cycle := r.cycle
	r.cycle = nil
	exitReason, err := cycle.Stop(reason)
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
