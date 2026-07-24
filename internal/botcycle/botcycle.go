package botcycle

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/config"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// ErrRejected identifies a BotCycle admission rejection.
var ErrRejected = errors.New("bot cycle rejected")

// Inputs contains approved Executor inputs owned above BotCycle.
type Inputs struct {
	LatestBBO   market.BBO
	Meta        meta.Instrument
	MinNotional decimal.Decimal
	ResultPath  string
}

// Result contains one immutable terminal BotCycle result.
type Result struct {
	CycleNumber int
	Accounts    []account.Result
}

// Control owns one active BotCycle and its Executors.
type Control struct {
	log       *logging.Logger
	number    int
	signal    signaler.Package
	executors []executor.Executor
	result    Result
	ticks     uint64
	runs      uint64
	startMS   uint64
	endMS     uint64
	running   bool
	completed bool
	stopped   bool
}

// Section 1 - Program Flow

// Init prepares one BotCycle and its configured Executors.
func (c *Control) Init(
	log *logging.Logger,
	number int,
	signals signaler.Signaler,
	signal signaler.Package,
	inputs Inputs,
	configs []config.Executor,
) error {
	c.log = log
	c.number = number
	c.signal = signal

	// create executors
	c.executors = make([]executor.Executor, 0, len(configs))
	for index, cfg := range configs {
		var created, err = executor.Create(executor.Context{
			Log:            log,
			CycleNumber:    number,
			ExecutorNumber: index + 1,
			Signaler:       signals,
			Signal:         signal,
			Config:         cfg,
			LatestBBO:      inputs.LatestBBO,
			Meta:           inputs.Meta,
			MinNotional:    inputs.MinNotional,
			ResultPath:     inputs.ResultPath,
		})
		if err != nil {
			var reason = "init_error"
			if errors.Is(err, executor.ErrRejected) {
				reason = "admission_rejected"
			}
			var stopErr = c.stopExecutors(reason)
			if errors.Is(err, executor.ErrRejected) {
				return errors.Join(
					fmt.Errorf("%w: executor %d: %v", ErrRejected, index+1, err),
					stopErr,
				)
			}
			return errors.Join(
				fmt.Errorf("create executor %d: %w", index+1, err),
				stopErr,
			)
		}
		c.executors = append(c.executors, created)
	}

	// initialize botcycle
	c.running = true
	var side = signaler.Short
	if signal.EnterLong() {
		side = signaler.Long
	}
	log.Info(fmt.Sprintf(
		"bot cycle initialized cycle=%d side=%s signal_ts_ms=%d",
		number,
		side,
		signal.TimestampMS(),
	))
	return nil
}

// Run checks one timer-driven BotCycle completion pass.
func (c *Control) Run(_ uint64) (bool, error) {
	if !c.running {
		return false, fmt.Errorf("bot cycle cannot run from current state")
	}
	c.runs++

	// check completion
	c.completed = true
	for _, activeExecutor := range c.executors {
		switch activeExecutor.Status() {
		case executor.Stopping, executor.Stopped, executor.Error:
		default:
			c.completed = false
		}
	}
	return c.completed, nil
}

// Stop stops Executors in reverse ownership order.
func (c *Control) Stop(reason string) (string, error) {
	if c.stopped {
		return c.exitReason(reason), nil
	}
	c.running = false

	// stop executors
	var firstErr = c.stopExecutors(reason)

	// collect cached Account results
	for index, activeExecutor := range c.executors {
		var provider, supported = activeExecutor.(executor.AccountResultProvider)
		if !supported {
			continue
		}
		var result, err = provider.AccountResult()
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("collect executor %d Account result: %w", index+1, err)
			continue
		}
		if err == nil {
			c.result.Accounts = append(c.result.Accounts, result.Clone())
		}
	}
	c.result.CycleNumber = c.number
	c.completed = true
	c.stopped = true

	// resolve exit reason
	var exitReason = c.exitReason(reason)

	// calculate duration
	var durationMS uint64
	if c.endMS >= c.startMS {
		durationMS = c.endMS - c.startMS
	}

	// report proof
	c.log.Info(fmt.Sprintf(
		"bot cycle stopped cycle=%d start_ts_ms=%d end_ts_ms=%d "+
			"duration_ms=%d executors=%d ticks_received=%d runs=%d stop_reason=%s",
		c.number,
		c.startMS,
		c.endMS,
		durationMS,
		len(c.executors),
		c.ticks,
		c.runs,
		exitReason,
	))
	return exitReason, firstErr
}

// Section 2 - Domain Helpers

// Reconcile refreshes every capable Executor Account as one barrier.
func (c *Control) Reconcile(
	nowMS uint64,
	forced bool,
) ([]account.Snapshot, error) {
	var snapshots []account.Snapshot
	for index, activeExecutor := range c.executors {
		if activeExecutor.Status() != executor.Running {
			continue
		}
		var reconciler, supported = activeExecutor.(executor.AccountReconciler)
		if !supported {
			continue
		}
		var snapshot, _, err = reconciler.Reconcile(nowMS, forced)
		if err != nil {
			return nil, fmt.Errorf("reconcile executor %d Account: %w", index+1, err)
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

// OnRecon delivers one accepted reconciliation barrier.
func (c *Control) OnRecon(nowMS uint64) error {
	for index, activeExecutor := range c.executors {
		if activeExecutor.Status() != executor.Running {
			continue
		}
		var handler, supported = activeExecutor.(executor.ReconHandler)
		if !supported {
			continue
		}
		var err = handler.OnRecon(nowMS)
		if err != nil {
			return fmt.Errorf("run executor %d recon handler: %w", index+1, err)
		}
	}
	return nil
}

// Result returns one independently owned terminal BotCycle result.
func (c *Control) Result() Result {
	var result = Result{CycleNumber: c.result.CycleNumber}
	for _, current := range c.result.Accounts {
		result.Accounts = append(result.Accounts, current.Clone())
	}
	return result
}

// IngestBBO routes one BBO through supported Simulator handlers.
func (c *Control) IngestBBO(bbo market.BBO) error {
	// ingest executor bbo
	for index, activeExecutor := range c.executors {
		if activeExecutor.Status() != executor.Running {
			continue
		}
		var handler, supported = activeExecutor.(executor.BBOIngestHandler)
		if !supported {
			continue
		}
		var err = handler.IngestBBO(bbo)
		if err != nil {
			return fmt.Errorf("ingest executor %d bbo: %w", index+1, err)
		}
	}
	return nil
}

// OnBBO distributes one BBO through supported normal handlers.
func (c *Control) OnBBO(bbo market.BBO) {
	// record cycle time
	if c.startMS == 0 {
		c.startMS = bbo.TimestampMS
	}
	c.endMS = bbo.TimestampMS
	c.ticks++

	// deliver executor bbo
	for _, activeExecutor := range c.executors {
		if activeExecutor.Status() != executor.Running {
			continue
		}
		var handler, supported = activeExecutor.(executor.BBOHandler)
		if supported {
			handler.OnBBO(bbo)
		}
	}
}

func (c *Control) stopExecutors(reason string) error {
	var firstErr error
	for index := len(c.executors) - 1; index >= 0; index-- {
		var err = c.executors[index].OnStop(reason)
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("stop executor %d: %w", index+1, err)
		}
	}
	return firstErr
}

func (c *Control) exitReason(fallback string) string {
	if len(c.executors) == 0 {
		return fallback
	}
	var reason = c.executors[0].ExitReason()
	if reason == "" {
		return fallback
	}
	for _, activeExecutor := range c.executors[1:] {
		if activeExecutor.ExitReason() != reason {
			return "completed"
		}
	}
	return reason
}

// Section 3 - Generic Helpers
