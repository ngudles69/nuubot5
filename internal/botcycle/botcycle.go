package botcycle

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// ErrRejected identifies a BotCycle admission rejection.
var ErrRejected = errors.New("bot cycle rejected")

// Inputs contains approved Executor inputs owned above BotCycle.
type Inputs struct {
	LatestBBOs     map[string]market.BBO
	ResourceEquity map[executor.Resource]decimal.Decimal
}

// Result contains one immutable terminal BotCycle result.
type Result struct {
	CycleNumber int
	Executors   []executor.Result
}

// Telemetry contains one immutable current BotCycle observation.
type Telemetry struct {
	CycleNumber int
	Status      string
	Executors   []executor.Telemetry
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
	signal signaler.Package,
	inputs Inputs,
	specs []executor.Spec,
) error {
	c.log = log
	c.number = number
	c.signal = signal

	// create executors
	c.executors = make([]executor.Executor, 0, len(specs))
	for index, spec := range specs {
		var created, err = executor.Create(executor.Context{
			Log:                log,
			CycleNumber:        number,
			ExecutorNumber:     index + 1,
			SignalTimestampMS:  signal.TimestampMS(),
			Spec:               spec,
			LatestBBO:          inputs.LatestBBOs[spec.Resource.Symbol],
			StartingEquityUSDC: inputs.ResourceEquity[spec.Resource],
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

	// start executors after every sibling initializes
	for index, activeExecutor := range c.executors {
		var starter, supported = activeExecutor.(executor.StartHandler)
		if !supported {
			continue
		}
		var err = starter.OnStart()
		if err != nil {
			return errors.Join(
				fmt.Errorf("start executor %d: %w", index+1, err),
				c.stopExecutors("start_error"),
			)
		}
	}

	// initialize botcycle
	c.running = true
	log.Info(fmt.Sprintf(
		"bot cycle initialized cycle=%d action=%s signal_ts_ms=%d",
		number,
		signal.Action(),
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

	// check coordinated completion
	for index, activeExecutor := range c.executors {
		switch activeExecutor.Status() {
		case executor.Error:
			return false, fmt.Errorf("executor %d entered error state", index+1)
		case executor.Stopping, executor.Stopped:
			c.completed = true
			return true, nil
		}
	}
	return false, nil
}

// Stop stops Executors in reverse ownership order.
func (c *Control) Stop(reason string) (string, error) {
	if c.stopped {
		return c.exitReason(reason), nil
	}
	c.running = false

	// stop executors
	var firstErr = c.stopExecutors(reason)

	// collect immutable Executor results
	for index, activeExecutor := range c.executors {
		var result, err = activeExecutor.Result()
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("collect Executor %d result: %w", index+1, err)
			continue
		}
		if err == nil {
			c.result.Executors = append(c.result.Executors, result.Clone())
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
	for _, current := range c.result.Executors {
		result.Executors = append(result.Executors, current.Clone())
	}
	return result
}

// Telemetry returns one immutable current BotCycle observation.
func (c *Control) Telemetry() Telemetry {
	var status = "configured"
	switch {
	case c.stopped:
		status = "stopped"
	case c.completed:
		status = "completed"
	case c.running:
		status = "running"
	}
	var result = Telemetry{
		CycleNumber: c.number,
		Status:      status,
	}
	for _, current := range c.executors {
		result.Executors = append(result.Executors, current.Telemetry())
	}
	return result
}

// IngestBBO routes one BBO through supported Simulator handlers.
func (c *Control) IngestBBO(bbo market.BBO) error {
	// ingest executor bbo
	for index, activeExecutor := range c.executors {
		var status = activeExecutor.Status()
		if status != executor.Running && status != executor.Stopping {
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
