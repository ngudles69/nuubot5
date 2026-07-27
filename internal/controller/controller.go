// Package controller owns synchronous Bot decisions.
package controller

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/bot"
	"nuubot/internal/botcycle"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/risk"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

type stats struct {
	ticks               uint64
	runs                uint64
	signalPackagesRead  uint64
	startActionsSkipped uint64
	cyclesStarted       uint64
	cyclesRejected      uint64
	cyclesClosed        uint64
	stopLossExits       uint64
}

// SignalDecision records one newly observed Signaler value.
type SignalDecision struct {
	TimestampMS uint64
	Action      signaler.Action
}

// RiskDecision records one Risk result.
type RiskDecision struct {
	TimestampMS uint64
	Policy      int
	Decision    risk.Decision
}

// Result contains one immutable terminal Controller result.
type Result struct {
	Identity        bot.Identity
	Cycles          []botcycle.Result
	Signals         []SignalDecision
	Risks           []RiskDecision
	ExitReason      string
	FirstMS         uint64
	LastMS          uint64
	TimeInCyclesMS  uint64
	TimeOutCyclesMS uint64
	Recon           account.ReconStats
	BotCapital      decimal.Decimal
	NetPnL          decimal.Decimal
	BotEquity       decimal.Decimal
	PeakEquity      decimal.Decimal
	Drawdown        decimal.Decimal
	MaxDrawdown     decimal.Decimal
}

// Telemetry contains one immutable current Controller observation.
type Telemetry struct {
	Ticks               uint64
	Runs                uint64
	SignalPackagesRead  uint64
	StartActionsSkipped uint64
	CyclesStarted       uint64
	CyclesRejected      uint64
	CyclesClosed        uint64
	ActiveCycle         int
	BotCapital          decimal.Decimal
	BotBalance          decimal.Decimal
	BotEquity           decimal.Decimal
	NetPnL              decimal.Decimal
	PeakEquity          decimal.Decimal
	Drawdown            decimal.Decimal
	MaxDrawdown         decimal.Decimal
}

type cycleControl interface {
	Reconcile(uint64, bool) (botcycle.ReconResult, error)
	OnRecon(uint64) error
	Run(uint64) (bool, error)
	Stop(string) (string, error)
	IngestBBO(market.BBO) error
	OnBBO(market.BBO)
	Result() botcycle.Result
	Telemetry() botcycle.Telemetry
}

// Controller owns synchronous Bot decisions and its direct children.
type Controller struct {
	log             *logging.Logger
	definition      bot.Definition
	cycle           cycleControl
	results         []botcycle.Result
	latestBBOs      map[string]market.BBO
	signalLog       []SignalDecision
	riskLog         []RiskDecision
	lastRisk        []risk.Decision
	resourceEquity  map[executor.Resource]decimal.Decimal
	botCapital      decimal.Decimal
	peakEquity      decimal.Decimal
	maxDrawdown     decimal.Decimal
	stats           stats
	firstMS         uint64
	lastMS          uint64
	timeInCyclesMS  uint64
	timeOutCyclesMS uint64
	lastSignalMS    uint64
	stopReason      string
	started         bool
	stopped         bool
}

// Section 1 - Program Flow

// Init prepares one Controller from one admitted Bot definition.
func (c *Controller) Init(log *logging.Logger, definition bot.Definition) error {
	if log == nil ||
		definition.Identity.BotSpecID == "" ||
		definition.Identity.ConfigTOML == "" ||
		definition.SignalSymbol == "" ||
		definition.MaxCycles == 0 ||
		definition.Signaler == nil ||
		len(definition.Executors) == 0 {
		return fmt.Errorf("controller definition is incomplete")
	}
	c.log = log
	c.definition = definition
	c.latestBBOs = make(map[string]market.BBO)
	c.resourceEquity = make(map[executor.Resource]decimal.Decimal)
	c.lastRisk = make([]risk.Decision, len(definition.Risks))
	for _, spec := range definition.Executors {
		if !spec.CapitalUSDC.IsPositive() {
			continue
		}
		if _, exists := c.resourceEquity[spec.Resource]; exists {
			return fmt.Errorf(
				"controller definition repeats resource: %s",
				spec.Resource.Key(),
			)
		}
		c.resourceEquity[spec.Resource] = spec.CapitalUSDC
		c.botCapital = c.botCapital.Add(spec.CapitalUSDC)
	}
	c.peakEquity = c.botCapital
	log.Info(fmt.Sprintf(
		"controller initialized bot_spec=%s config_hash=%s",
		definition.Identity.BotSpecID,
		definition.Identity.ConfigHash,
	))
	return nil
}

// Start starts Controller admission.
func (c *Controller) Start() error {
	if c.started || c.stopped {
		return fmt.Errorf("controller cannot start from current state")
	}
	c.started = true
	c.log.Info("controller started")
	return nil
}

// Run executes one timer-driven control pass.
func (c *Controller) Run(nowMS uint64) (bool, error) {
	if !c.started || c.stopped {
		return false, fmt.Errorf("controller cannot run from current state")
	}
	c.stats.runs++
	if c.stopReason != "" {
		return true, nil
	}

	var snapshots []account.Snapshot
	if c.cycle != nil {
		var barrier, err = c.cycle.Reconcile(nowMS, false)
		if barrier.MaxConsecutiveFailures >= 3 {
			if err != nil {
				return false, fmt.Errorf(
					"reconcile BotCycle failed %d consecutive times: %w",
					barrier.MaxConsecutiveFailures,
					err,
				)
			}
			return false, fmt.Errorf(
				"reconcile BotCycle failed %d consecutive times",
				barrier.MaxConsecutiveFailures,
			)
		}
		if err != nil {
			if barrier.Failed && barrier.MaxConsecutiveFailures > 0 {
				return false, nil
			}
			return false, fmt.Errorf("reconcile BotCycle: %w", err)
		}
		snapshots = barrier.Snapshots
	}

	var blockStart bool
	var stopCycle bool
	var stopController bool
	var riskInput = c.riskInput(nowMS, snapshots)
	for index, policy := range c.definition.Risks {
		var decision = policy.Assess(riskInput)
		if decision != c.lastRisk[index] {
			c.lastRisk[index] = decision
			c.riskLog = append(c.riskLog, RiskDecision{
				TimestampMS: nowMS,
				Policy:      index + 1,
				Decision:    decision,
			})
		}
		switch decision {
		case risk.Allow:
		case risk.BlockCycleStart:
			blockStart = true
		case risk.StopCycle:
			stopCycle = true
		case risk.StopController:
			stopController = true
		default:
			return false, fmt.Errorf("Risk returned invalid decision: %s", decision)
		}
	}
	if stopController {
		c.requestStop("risk")
		return true, nil
	}
	if stopCycle {
		if c.cycle != nil {
			if err := c.closeCycle("risk"); err != nil {
				return false, err
			}
		}
		return false, nil
	}

	if c.cycle != nil {
		if err := c.cycle.OnRecon(nowMS); err != nil {
			return false, fmt.Errorf("deliver BotCycle recon: %w", err)
		}
		var completed, err = c.cycle.Run(nowMS)
		if err != nil {
			return false, fmt.Errorf("run BotCycle: %w", err)
		}
		if completed {
			if err = c.closeCycle("completed"); err != nil {
				return false, fmt.Errorf("close completed BotCycle: %w", err)
			}
			if c.stats.cyclesClosed >= c.definition.MaxCycles {
				c.requestStop("max_cycles")
				return true, nil
			}
			return false, nil
		}
	}

	var packages = c.definition.Signaler.Signals(
		c.definition.SignalSymbol,
		nowMS,
		1,
	)
	if len(packages) == 0 {
		return false, nil
	}
	var current = packages[len(packages)-1]
	var newSignal = current.TimestampMS() > c.lastSignalMS
	if newSignal {
		c.lastSignalMS = current.TimestampMS()
		c.stats.signalPackagesRead++
		c.signalLog = append(c.signalLog, SignalDecision{
			TimestampMS: current.TimestampMS(),
			Action:      current.Action(),
		})
	}

	switch current.Action() {
	case signaler.NoAction:
		return false, nil
	case signaler.StopCycle:
		if c.cycle != nil {
			if err := c.closeCycle("signal"); err != nil {
				return false, err
			}
		}
		return false, nil
	case signaler.StartCycle:
		if c.cycle != nil {
			if newSignal {
				c.stats.startActionsSkipped++
			}
			return false, nil
		}
		if blockStart {
			return false, nil
		}
		if c.stats.cyclesStarted >= c.definition.MaxCycles {
			c.requestStop("max_cycles")
			return true, nil
		}
		var err = c.openCycle(current)
		if errors.Is(err, botcycle.ErrRejected) {
			c.stats.cyclesRejected++
			c.log.Warning(fmt.Sprintf(
				"bot cycle rejected signal_ts_ms=%d reason=%v",
				current.TimestampMS(),
				err,
			))
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("open BotCycle: %w", err)
		}
		return false, nil
	default:
		return false, fmt.Errorf("Signaler returned invalid action: %s", current.Action())
	}
}

// Stop closes the active BotCycle and stops direct children.
func (c *Controller) Stop(reason string) error {
	if c.stopped {
		return nil
	}
	c.requestStop(reason)
	c.started = false
	var firstErr = c.closeCycle(c.stopReason)
	for index := len(c.definition.Risks) - 1; index >= 0; index-- {
		c.definition.Risks[index].Stop()
	}
	c.definition.Signaler.Stop()
	c.stopped = true
	c.log.Info(fmt.Sprintf(
		"controller stopped ticks_accepted=%d runs=%d signal_packages_read=%d "+
			"start_actions_skipped=%d cycles_started=%d cycles_rejected=%d "+
			"cycles_closed=%d time_in_cycles_ms=%d time_out_cycles_ms=%d "+
			"stop_loss_exits=%d stop_reason=%s",
		c.stats.ticks,
		c.stats.runs,
		c.stats.signalPackagesRead,
		c.stats.startActionsSkipped,
		c.stats.cyclesStarted,
		c.stats.cyclesRejected,
		c.stats.cyclesClosed,
		c.timeInCyclesMS,
		c.timeOutCyclesMS,
		c.stats.stopLossExits,
		c.stopReason,
	))
	return firstErr
}

// Section 2 - Domain Helpers

// IngestBBO accepts one validated symbol-qualified BBO.
func (c *Controller) IngestBBO(bbo market.BBO) error {
	if !c.started || c.stopped || c.stopReason != "" {
		return fmt.Errorf("controller cannot ingest BBO from current state")
	}
	if bbo.Symbol == "" {
		return fmt.Errorf("controller requires symbol-qualified BBO")
	}
	c.recordCycleTime(bbo.TimestampMS)
	c.latestBBOs[bbo.Symbol] = bbo
	if c.cycle != nil {
		if err := c.cycle.IngestBBO(bbo); err != nil {
			return fmt.Errorf("ingest BotCycle BBO: %w", err)
		}
		c.cycle.OnBBO(bbo)
	}
	c.stats.ticks++
	return nil
}

// Result returns one independently owned terminal Controller result.
func (c *Controller) Result() Result {
	var result = Result{
		Identity:        c.definition.Identity,
		Signals:         append([]SignalDecision(nil), c.signalLog...),
		Risks:           append([]RiskDecision(nil), c.riskLog...),
		ExitReason:      c.stopReason,
		FirstMS:         c.firstMS,
		LastMS:          c.lastMS,
		TimeInCyclesMS:  c.timeInCyclesMS,
		TimeOutCyclesMS: c.timeOutCyclesMS,
	}
	var metrics = c.riskInput(0, nil)
	result.BotCapital = metrics.BotCapital
	result.NetPnL = metrics.NetPnL
	result.BotEquity = metrics.BotEquity
	result.PeakEquity = metrics.PeakEquity
	result.Drawdown = metrics.CurrentDrawdown
	result.MaxDrawdown = metrics.MaximumDrawdown
	for _, cycle := range c.results {
		var copied = botcycle.Result{
			CycleNumber: cycle.CycleNumber,
			StartMS:     cycle.StartMS,
			EndMS:       cycle.EndMS,
			DurationMS:  cycle.DurationMS,
			Recon:       cycle.Recon,
		}
		addReconStats(&result.Recon, cycle.Recon)
		for _, current := range cycle.Executors {
			copied.Executors = append(copied.Executors, current.Clone())
		}
		result.Cycles = append(result.Cycles, copied)
	}
	return result
}

// Telemetry returns one immutable current Controller observation.
func (c *Controller) Telemetry() Telemetry {
	var equity = decimal.Zero
	var unrealizedPnL = decimal.Zero
	for _, value := range c.resourceEquity {
		equity = equity.Add(value)
	}
	var activeCycle int
	if c.cycle != nil {
		var cycle = c.cycle.Telemetry()
		activeCycle = cycle.CycleNumber
		for _, current := range cycle.Executors {
			if current.Account == nil {
				continue
			}
			var resourceEquity = c.resourceEquity[current.Resource]
			equity = equity.Sub(resourceEquity).Add(current.Account.AccountValue)
			unrealizedPnL = unrealizedPnL.Add(current.Account.UnrealizedPnL)
		}
	}
	var peakEquity = c.peakEquity
	if equity.GreaterThan(peakEquity) {
		peakEquity = equity
	}
	var drawdown = peakEquity.Sub(equity)
	if drawdown.IsNegative() {
		drawdown = decimal.Zero
	}
	var maxDrawdown = c.maxDrawdown
	if drawdown.GreaterThan(maxDrawdown) {
		maxDrawdown = drawdown
	}
	return Telemetry{
		Ticks:               c.stats.ticks,
		Runs:                c.stats.runs,
		SignalPackagesRead:  c.stats.signalPackagesRead,
		StartActionsSkipped: c.stats.startActionsSkipped,
		CyclesStarted:       c.stats.cyclesStarted,
		CyclesRejected:      c.stats.cyclesRejected,
		CyclesClosed:        c.stats.cyclesClosed,
		ActiveCycle:         activeCycle,
		BotCapital:          c.botCapital,
		BotBalance:          equity.Sub(unrealizedPnL),
		BotEquity:           equity,
		NetPnL:              equity.Sub(c.botCapital),
		PeakEquity:          peakEquity,
		Drawdown:            drawdown,
		MaxDrawdown:         maxDrawdown,
	}
}

func (c *Controller) recordCycleTime(nowMS uint64) {
	if c.firstMS == 0 {
		c.firstMS = nowMS
		c.lastMS = nowMS
		return
	}
	var elapsed = nowMS - c.lastMS
	if c.cycle == nil {
		c.timeOutCyclesMS += elapsed
	} else {
		c.timeInCyclesMS += elapsed
	}
	c.lastMS = nowMS
}

func addReconStats(total *account.ReconStats, current account.ReconStats) {
	total.Calls += current.Calls
	total.SkippedClean += current.SkippedClean
	total.Executed += current.Executed
	total.Succeeded += current.Succeeded
	total.Failed += current.Failed
}

func (c *Controller) openCycle(signal signaler.Package) error {
	var cycle botcycle.Control
	var err = cycle.Init(
		c.log,
		int(c.stats.cyclesStarted+c.stats.cyclesRejected+1),
		signal,
		botcycle.Inputs{
			LatestBBOs:     c.latestBBOs,
			ResourceEquity: c.resourceEquity,
		},
		c.definition.Executors,
	)
	if err != nil {
		return err
	}
	c.cycle = &cycle
	c.stats.cyclesStarted++
	return nil
}

func (c *Controller) closeCycle(reason string) error {
	if c.cycle == nil {
		return nil
	}
	var cycle = c.cycle
	c.cycle = nil
	var exitReason, err = cycle.Stop(reason)
	if err != nil {
		return fmt.Errorf("stop BotCycle: %w", err)
	}
	var result = cycle.Result()
	for _, executorResult := range result.Executors {
		if executorResult.Account == nil {
			continue
		}
		c.resourceEquity[executorResult.Resource] =
			executorResult.Account.Snapshot.AccountValue
	}
	c.results = append(c.results, result)
	c.stats.cyclesClosed++
	if exitReason == "stop_loss" {
		c.stats.stopLossExits++
	}
	return nil
}

func (c *Controller) requestStop(reason string) {
	if c.stopReason == "" {
		c.stopReason = reason
	}
}

func (c *Controller) riskInput(
	nowMS uint64,
	snapshots []account.Snapshot,
) risk.Input {
	var equity = decimal.Zero
	for _, value := range c.resourceEquity {
		equity = equity.Add(value)
	}
	for _, snapshot := range snapshots {
		for _, spec := range c.definition.Executors {
			var resource = executor.Resource{
				Venue:             snapshot.Venue,
				Network:           snapshot.Network,
				PhysicalAccountID: snapshot.Account,
				Symbol:            snapshot.Symbol,
			}
			if spec.Resource == resource {
				equity = equity.Sub(c.resourceEquity[resource]).Add(snapshot.AccountValue)
				break
			}
		}
	}
	if equity.GreaterThan(c.peakEquity) {
		c.peakEquity = equity
	}
	var drawdown = c.peakEquity.Sub(equity)
	if drawdown.IsNegative() {
		drawdown = decimal.Zero
	}
	if drawdown.GreaterThan(c.maxDrawdown) {
		c.maxDrawdown = drawdown
	}
	return risk.Input{
		TimestampMS:     nowMS,
		ActiveCycle:     c.cycle != nil,
		CompletedCycles: c.stats.cyclesClosed,
		Accounts:        append([]account.Snapshot(nil), snapshots...),
		BotCapital:      c.botCapital,
		NetPnL:          equity.Sub(c.botCapital),
		BotEquity:       equity,
		PeakEquity:      c.peakEquity,
		CurrentDrawdown: drawdown,
		MaximumDrawdown: c.maxDrawdown,
	}
}

// Section 3 - Generic Helpers
