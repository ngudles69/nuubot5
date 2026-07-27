package botcycle

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/botspec"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/setup"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// ErrRejected identifies a BotCycle admission rejection.
var ErrRejected = errors.New("bot cycle rejected")

// Inputs contains approved Executor inputs owned above BotCycle.
type Inputs struct {
	ResourceEquity map[botspec.Resource]decimal.Decimal
}

// Result contains one immutable terminal BotCycle result.
type Result struct {
	CycleNumber int
	StartMS     uint64
	EndMS       uint64
	DurationMS  uint64
	Recon       account.ReconStats
	Executors   []executor.Result
}

// ReconResult contains one complete Account reconciliation barrier result.
type ReconResult struct {
	Snapshots              []account.Snapshot
	Failed                 bool
	MaxConsecutiveFailures uint64
}

// Telemetry contains one immutable current BotCycle observation.
type Telemetry struct {
	CycleNumber int
	Status      string
	Executors   []executor.Telemetry
}

// BotCycle owns one coordinated Executor lifecycle.
type BotCycle struct {
	log       *logging.Logger
	nuubot    *setup.Nuubot
	number    int
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
func (c *BotCycle) Init(
	nuubot *setup.Nuubot,
	number int,
	signal signaler.Package,
	inputs Inputs,
) error {
	// Step 1: retain BotCycle inputs and log init
	c.log = nuubot.Log
	c.nuubot = nuubot
	c.number = number
	c.log.Info("bot cycle init")

	// Step 2: create Executors
	var specs = nuubot.BotSpec.Executors
	c.executors = make([]executor.Executor, 0, len(specs))
	for index, spec := range specs {
		var created, err = executor.Create(executor.BotCycleContext{
			Nuubot:             nuubot,
			CycleNumber:        number,
			ExecutorNumber:     index + 1,
			Signal:             signal,
			Spec:               spec,
			StartingEquityUSDC: inputs.ResourceEquity[spec.Resource],
			Status:             executor.Status(nuubot.Bot.Status),
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

	// Step 3: log init completed
	c.log.Info(fmt.Sprintf(
		"bot cycle init completed cycle=%d action=%s signal_ts_ms=%d",
		number,
		signal.Action(),
		signal.TimestampMS(),
	))
	return nil
}

// Start starts every initialized Executor as one coordinated unit.
func (c *BotCycle) Start() error {
	// Step 1: log start
	c.log.Info("bot cycle start")

	// Step 2: start Executors after every sibling initializes
	for index, activeExecutor := range c.executors {
		var startHandler, supported = activeExecutor.(executor.StartHandler)
		if !supported {
			continue
		}
		var err = startHandler.OnStart()
		if err != nil {
			return errors.Join(
				fmt.Errorf("start executor %d: %w", index+1, err),
				c.stopExecutors("start_error"),
			)
		}
	}

	// Step 3: mark BotCycle running
	c.running = true

	// Step 4: log start completed
	c.log.Info("bot cycle started")
	return nil
}

// Run delivers the current Signal and checks BotCycle completion.
func (c *BotCycle) Run(signal signaler.Package) (bool, error) {
	// Step 1: validate run state
	if !c.running {
		return false, fmt.Errorf("bot cycle cannot run from current state")
	}

	// Step 2: record run
	c.runs++

	// Step 3: deliver current Signal to running Executors
	for index, activeExecutor := range c.executors {
		if activeExecutor.Status() != executor.Running {
			continue
		}
		var handler, supported = activeExecutor.(executor.SignalHandler)
		if !supported {
			continue
		}
		var err = handler.OnSignal(signal)
		if err != nil {
			return false, fmt.Errorf("deliver executor %d Signal: %w", index+1, err)
		}
	}

	// Step 4: check coordinated completion
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
func (c *BotCycle) Stop(reason string) (string, error) {
	// Step 1: log stop
	c.log.Info("bot cycle stop")

	// Step 2: ignore repeated stop request
	if c.stopped {
		c.log.Info("bot cycle stopping - ignoring stop request")
		return c.exitReason(reason), nil
	}

	// Step 3: mark BotCycle not running
	c.running = false

	// Step 4: stop Executors
	var firstErr = c.stopExecutors(reason)

	// Step 5: collect immutable Executor results
	for index, activeExecutor := range c.executors {
		var result, err = activeExecutor.Result()
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("collect Executor %d result: %w", index+1, err)
			continue
		}
		if err == nil {
			c.result.Executors = append(c.result.Executors, result.Clone())
			if result.Account != nil {
				addReconStats(&c.result.Recon, result.Account.Recon)
			}
		}
	}

	// Step 6: mark BotCycle completed and stopped
	c.completed = true
	c.stopped = true

	// Step 7: resolve exit reason
	var exitReason = c.exitReason(reason)

	// Step 8: calculate terminal result
	var durationMS uint64
	if c.endMS >= c.startMS {
		durationMS = c.endMS - c.startMS
	}
	c.result.CycleNumber = c.number
	c.result.StartMS = c.startMS
	c.result.EndMS = c.endMS
	c.result.DurationMS = durationMS

	// Step 9: report results and stats
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

// Section 2.1 - Account Reconciliation

// AcctRecon refreshes every capable Executor Account as one barrier.
func (c *BotCycle) AcctRecon(forced bool) (ReconResult, error) {
	// Step 1: read current Clock time
	var nowMS = c.nuubot.Clock.NowMS()

	// Step 2: reconcile running Executor Accounts
	var result ReconResult
	var failures []error
	for index, activeExecutor := range c.executors {
		if activeExecutor.Status() != executor.Running {
			continue
		}
		var reconciler, supported = activeExecutor.(executor.AccountReconciler)
		if !supported {
			continue
		}
		var snapshot, _, consecutiveFailures, err = reconciler.Reconcile(nowMS, forced)
		if consecutiveFailures > result.MaxConsecutiveFailures {
			result.MaxConsecutiveFailures = consecutiveFailures
		}
		if err != nil {
			failures = append(
				failures,
				fmt.Errorf("reconcile executor %d Account: %w", index+1, err),
			)
			continue
		}
		result.Snapshots = append(result.Snapshots, snapshot)
	}

	// Step 3: reject partial Account snapshots
	if len(failures) != 0 {
		result.Failed = true
		result.Snapshots = nil
		return result, errors.Join(failures...)
	}
	return result, nil
}

// OnRecon delivers one accepted reconciliation barrier.
func (c *BotCycle) OnRecon() error {
	// deliver accepted reconciliation to running Executors
	for index, activeExecutor := range c.executors {
		if activeExecutor.Status() != executor.Running {
			continue
		}
		var handler, supported = activeExecutor.(executor.ReconHandler)
		if !supported {
			continue
		}
		var err = handler.OnRecon()
		if err != nil {
			return fmt.Errorf("run executor %d recon handler: %w", index+1, err)
		}
	}
	return nil
}

// Section 2.2 - Results and Telemetry

// Result returns one independently owned terminal BotCycle result.
func (c *BotCycle) Result() Result {
	// Step 1: copy BotCycle result
	var result = Result{
		CycleNumber: c.result.CycleNumber,
		StartMS:     c.result.StartMS,
		EndMS:       c.result.EndMS,
		DurationMS:  c.result.DurationMS,
		Recon:       c.result.Recon,
	}

	// Step 2: copy Executor results
	for _, current := range c.result.Executors {
		result.Executors = append(result.Executors, current.Clone())
	}
	return result
}

// Telemetry returns one immutable current BotCycle observation.
func (c *BotCycle) Telemetry() Telemetry {
	// Step 1: resolve BotCycle status
	var status = "configured"
	switch {
	case c.stopped:
		status = "stopped"
	case c.completed:
		status = "completed"
	case c.running:
		status = "running"
	}

	// Step 2: build BotCycle telemetry
	var result = Telemetry{
		CycleNumber: c.number,
		Status:      status,
	}
	for _, current := range c.executors {
		result.Executors = append(result.Executors, current.Telemetry())
	}
	return result
}

// Section 2.3 - Market Data

// OnBBO records BotCycle market time.
func (c *BotCycle) OnBBO(bbo market.BBO) {
	if c.startMS == 0 {
		c.startMS = bbo.TimestampMS
	}
	c.endMS = bbo.TimestampMS
	c.ticks++
}

// Section 2.4 - Lifecycle Helpers

func (c *BotCycle) stopExecutors(reason string) error {
	var firstErr error
	for index := len(c.executors) - 1; index >= 0; index-- {
		var err = c.executors[index].OnStop(reason)
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("stop executor %d: %w", index+1, err)
		}
	}
	return firstErr
}

func (c *BotCycle) exitReason(fallback string) string {
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

func addReconStats(total *account.ReconStats, current account.ReconStats) {
	total.Calls += current.Calls
	total.SkippedClean += current.SkippedClean
	total.Executed += current.Executed
	total.Succeeded += current.Succeeded
	total.Failed += current.Failed
}

// Section 3 - Generic Helpers
