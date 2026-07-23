package runtime

import (
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

type Runtime struct {
	log        *common.Logger
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

func New(log *common.Logger, cfg config.Runtime, endAt *time.Time) (*Runtime, error) {
	signals, err := signaler.New(log, cfg.Signaler)
	if err != nil {
		return nil, err
	}
	risks := make([]risk.Risk, 0, len(cfg.Risks))
	for index, riskConfig := range cfg.Risks {
		created, err := risk.New(log, index+1, riskConfig)
		if err != nil {
			return nil, err
		}
		risks = append(risks, created)
	}
	var endMS uint64
	if endAt != nil {
		endMS = uint64(endAt.UnixMilli())
	}
	log.Info("runtime", "init end_ts_ms=%d", endMS)
	return &Runtime{log: log, config: cfg, signaler: signals, risks: risks, endMS: endMS}, nil
}

func (r *Runtime) BarsNeeded() []bars.Requirement {
	return r.signaler.BarsNeeded()
}

func (r *Runtime) PrepareBars(loaded []bars.Data) error {
	return r.signaler.Prepare(loaded)
}

func (r *Runtime) Start() error {
	if r.started || r.stopped {
		return common.StateError("Runtime", "start")
	}
	if err := r.signaler.Start(); err != nil {
		return err
	}
	r.started = true
	r.log.Info("runtime", "start")
	return nil
}

func (r *Runtime) Ingest(bbo market.BBO) error {
	if !r.started || r.stopped || r.stopReason != "" {
		return common.StateError("Runtime", "ingest BBO")
	}
	for {
		signal, available, err := r.signaler.Next(bbo.TimestampMS)
		if err != nil {
			return err
		}
		if !available {
			break
		}
		r.stats.signals++
		if r.cycle != nil {
			r.stats.signalsSkipped++
		} else if err := r.openCycle(signal); err != nil {
			return err
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

func (r *Runtime) MainLoop(nowMS uint64) (bool, error) {
	if !r.started || r.stopped {
		return false, common.StateError("Runtime", "run main loop")
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
	completed, err := r.cycle.MainLoop(nowMS)
	if err != nil {
		return false, err
	}
	if !completed {
		return false, nil
	}
	if err := r.closeCycle("completed"); err != nil {
		return false, err
	}
	if r.stats.cyclesClosed >= r.config.MaxCycles {
		return true, r.Stop("max_cycles")
	}
	return false, nil
}

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
		"runtime",
		"stop status=%s ticks_accepted=%d passes=%d signals_received=%d signals_skipped=%d cycles_started=%d cycles_closed=%d stop_loss_exits=%d end_date_exits=%d stop_reason=%s",
		status, r.stats.ticks, r.stats.passes, r.stats.signals, r.stats.signalsSkipped,
		r.stats.cyclesStarted, r.stats.cyclesClosed, r.stats.stopLossExits,
		r.stats.endDateExits, r.stopReason,
	)
	return firstErr
}

func (r *Runtime) openCycle(signal signaler.Signal) error {
	cycle, err := botcycle.New(r.log, int(r.stats.cyclesStarted+1), signal, r.config.Executors)
	if err != nil {
		return err
	}
	if err := cycle.Start(); err != nil {
		return err
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
		return err
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
