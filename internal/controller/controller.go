// Package controller owns synchronous Bot decisions.
package controller

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/bot"
	"nuubot/internal/botcycle"
	"nuubot/internal/botspec"

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
	AcctRecon(bool) (botcycle.ReconResult, error)
	OnRecon() error
	Run(signaler.Package) (bool, error)
	Stop(string) (string, error)
	OnBBO(market.BBO)
	Result() botcycle.Result
	Telemetry() botcycle.Telemetry
}

// Controller owns synchronous Bot decisions and its direct children.
type Controller struct {
	log             *logging.Logger
	nuubot          *setup.Nuubot
	signaler        signaler.Signaler
	risks           []risk.Risk
	cycle           cycleControl
	results         []botcycle.Result
	subscriptions   []*market.Subscription
	signalLog       []SignalDecision
	riskLog         []RiskDecision
	lastRisk        []risk.Decision
	resourceEquity  map[botspec.Resource]decimal.Decimal
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

// Init prepares one Controller from Setup and one typed BotSpec.
func (c *Controller) Init(nuubot *setup.Nuubot) error {
	// Step 1: retain Nuubot and log init
	c.nuubot = nuubot
	c.log = nuubot.Log
	c.log.Info("controller init")
	var botInput = nuubot.Bot
	var replayInput = botInput.Replay
	var spec = nuubot.BotSpec

	// Step 2: set Signaler replay range
	var end = replayInput.ReplayEnd
	if replayInput.EndAt != nil && replayInput.EndAt.Before(end) {
		end = *replayInput.EndAt
	}

	// Step 3: create Signaler
	var signals, err = signaler.Create(
		c.log,
		spec.Signaler,
		replayInput.Symbol,
		replayInput.TicksPath,
		replayInput.ReplayStart,
		end,
	)
	if err != nil {
		return fmt.Errorf("initialize Controller Signaler: %w", err)
	}

	// Step 4: create Risks
	var risks = make([]risk.Risk, 0, len(spec.Risks))
	for index, riskSpec := range spec.Risks {
		var created risk.Risk
		created, err = risk.Create(c.log, index+1, riskSpec.Kind)
		if err != nil {
			for stopIndex := len(risks) - 1; stopIndex >= 0; stopIndex-- {
				risks[stopIndex].Stop()
			}
			signals.Stop()
			return fmt.Errorf("initialize Controller Risk %d: %w", index+1, err)
		}
		risks = append(risks, created)
	}

	// Step 5: retain Controller components
	c.signaler = signals
	c.risks = risks

	// Step 6: initialize Controller state
	c.resourceEquity = make(map[botspec.Resource]decimal.Decimal)
	c.lastRisk = make([]risk.Decision, len(spec.Risks))

	// Step 7: subscribe to MarketData timing
	for _, key := range marketKeys(spec.Executors) {
		var current = key
		var subscription *market.Subscription
		subscription, err = nuubot.MarketData.SubscribeBBO(current, func() error {
			c.onBBO(current)
			return nil
		})
		if err != nil {
			c.stopSubscriptions()
			for index := len(c.risks) - 1; index >= 0; index-- {
				c.risks[index].Stop()
			}
			c.signaler.Stop()
			return fmt.Errorf("initialize Controller MarketData: %w", err)
		}
		c.subscriptions = append(c.subscriptions, subscription)
	}

	// Step 8: initialize resource capital
	for _, executorSpec := range spec.Executors {
		if !executorSpec.CapitalUSDC.IsPositive() {
			continue
		}
		if _, exists := c.resourceEquity[executorSpec.Resource]; exists {
			for index := len(c.risks) - 1; index >= 0; index-- {
				c.risks[index].Stop()
			}
			c.signaler.Stop()
			return fmt.Errorf(
				"controller specification repeats resource: %s",
				executorSpec.Resource.Key(),
			)
		}
		c.resourceEquity[executorSpec.Resource] = executorSpec.CapitalUSDC
		c.botCapital = c.botCapital.Add(executorSpec.CapitalUSDC)
	}
	c.peakEquity = c.botCapital

	// Step 9: log init completed
	c.log.Info(fmt.Sprintf(
		"controller init completed bot_spec=%s spec_hash=%s",
		botInput.BotSpecID,
		botInput.ConfigHash,
	))
	return nil
}

// Start starts Controller processing.
func (c *Controller) Start() error {
	// Step 1: log start
	c.log.Info("controller start")

	// Step 2: validate start state
	if c.started || c.stopped {
		return fmt.Errorf("controller cannot start from current state")
	}

	// Step 3: mark Controller started
	c.started = true

	// Step 4: log start completed
	c.log.Info("controller started")
	return nil
}

// Run executes one timer-driven control pass.
func (c *Controller) Run() (bool, error) {
	// Step 1: validate run state
	if !c.started || c.stopped {
		return false, fmt.Errorf("controller cannot run from current state")
	}

	// Step 2: read Clock, record run, and check stop request
	var nowMS = c.nuubot.Clock.NowMS()
	c.stats.runs++
	if c.stopReason != "" {
		return true, nil
	}

	// Step 3: reconcile active BotCycle
	var snapshots []account.Snapshot
	if c.cycle != nil {
		var result, err = c.cycle.AcctRecon(false)
		if result.MaxConsecutiveFailures >= 3 {
			if err != nil {
				return false, fmt.Errorf(
					"reconcile BotCycle failed %d consecutive times: %w",
					result.MaxConsecutiveFailures,
					err,
				)
			}
			return false, fmt.Errorf(
				"reconcile BotCycle failed %d consecutive times",
				result.MaxConsecutiveFailures,
			)
		}
		if err != nil {
			if result.Failed && result.MaxConsecutiveFailures > 0 {
				return false, nil
			}
			return false, fmt.Errorf("reconcile BotCycle: %w", err)
		}
		snapshots = result.Snapshots
	}

	// Step 4: assess Risks
	var blockStart bool
	var stopCycle bool
	var stopController bool
	var riskInput = c.riskInput(nowMS, snapshots)
	for index, policy := range c.risks {
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

	// Step 5: apply Risk decisions
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

	// Step 6: deliver accepted reconciliation
	if c.cycle != nil {
		if err := c.cycle.OnRecon(); err != nil {
			return false, fmt.Errorf("deliver BotCycle recon: %w", err)
		}
	}

	// Step 7: read current Signal
	var packages = c.signaler.Signals(
		c.nuubot.Bot.Replay.Symbol,
		nowMS,
		1,
	)
	if len(packages) == 0 {
		return false, nil
	}

	// Step 8: record new Signal
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

	// Step 9: run active BotCycle with current Signal
	if c.cycle != nil {
		var completed, err = c.cycle.Run(current)
		if err != nil {
			return false, fmt.Errorf("run BotCycle: %w", err)
		}
		if completed {
			if err = c.closeCycle("completed"); err != nil {
				return false, fmt.Errorf("close completed BotCycle: %w", err)
			}
			if c.stats.cyclesClosed >= c.nuubot.BotSpec.Controller.MaxCycles {
				c.requestStop("max_cycles")
				return true, nil
			}
			return false, nil
		}
	}

	// Step 10: apply Signal action
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
		if c.stats.cyclesStarted >= c.nuubot.BotSpec.Controller.MaxCycles {
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
	// Step 1: log stop
	c.log.Info("controller stop")

	// Step 2: ignore repeated stop request
	if c.stopped {
		c.log.Info("controller stopping - ignoring stop request")
		return nil
	}

	// Step 3: request Controller stop
	c.requestStop(reason)
	c.started = false

	// Step 4: stop MarketData subscriptions
	var firstErr = c.stopSubscriptions()

	// Step 5: close active BotCycle
	var cycleErr = c.closeCycle(c.stopReason)
	if firstErr == nil {
		firstErr = cycleErr
	}

	// Step 6: stop Risks
	for index := len(c.risks) - 1; index >= 0; index-- {
		c.risks[index].Stop()
	}

	// Step 7: stop Signaler
	c.signaler.Stop()

	// Step 8: mark Controller stopped
	c.stopped = true

	// Step 9: log stopped results and stats
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

	// Step 10: return stop error
	return firstErr
}

// Section 2 - Domain Helpers

// Result returns one independently owned terminal Controller result.
func (c *Controller) Result() Result {
	var botInput = c.nuubot.Bot
	var result = Result{
		Identity: bot.Identity{
			SweepID:    botInput.SweepID,
			BotID:      botInput.BotID,
			BotSpecID:  botInput.BotSpecID,
			ConfigTOML: botInput.ConfigTOML,
			ConfigHash: botInput.ConfigHash,
		},
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

func (c *Controller) onBBO(key market.Key) {
	// Step 1: read latest BBO
	var bbo, found = c.nuubot.MarketData.LatestBBO(key)
	if !found || !c.started || c.stopped || c.stopReason != "" {
		return
	}

	// Step 2: record Controller market time
	c.recordBBOGap(bbo.TimestampMS)
	// Step 3: record active BotCycle market time
	if c.cycle != nil {
		c.cycle.OnBBO(bbo)
	}

	// Step 4: record accepted tick
	c.stats.ticks++
}

func (c *Controller) recordBBOGap(nowMS uint64) {
	if c.firstMS == 0 {
		c.firstMS = nowMS
		c.lastMS = nowMS
		return
	}
	var bboGapMS = nowMS - c.lastMS
	if c.cycle == nil {
		c.timeOutCyclesMS += bboGapMS
	} else {
		c.timeInCyclesMS += bboGapMS
	}
	c.lastMS = nowMS
}

func (c *Controller) stopSubscriptions() error {
	var firstErr error
	for index := len(c.subscriptions) - 1; index >= 0; index-- {
		var err = c.subscriptions[index].Stop()
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	c.subscriptions = nil
	return firstErr
}

func marketKeys(specs []botspec.ExecutorSpec) []market.Key {
	var seen = make(map[market.Key]struct{})
	var keys = make([]market.Key, 0, len(specs))
	for _, spec := range specs {
		var key = market.Key{
			Venue:   spec.Resource.Venue,
			Network: spec.Resource.Network,
			Symbol:  spec.Resource.Symbol,
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}

func addReconStats(total *account.ReconStats, current account.ReconStats) {
	total.Calls += current.Calls
	total.SkippedClean += current.SkippedClean
	total.Executed += current.Executed
	total.Succeeded += current.Succeeded
	total.Failed += current.Failed
}

func (c *Controller) openCycle(signal signaler.Package) error {
	var cycle botcycle.BotCycle
	var err = cycle.Init(
		c.nuubot,
		int(c.stats.cyclesStarted+c.stats.cyclesRejected+1),
		signal,
		botcycle.Inputs{
			ResourceEquity: c.resourceEquity,
		},
	)
	if err != nil {
		return err
	}
	err = cycle.Start()
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
		for _, spec := range c.nuubot.BotSpec.Executors {
			var resource = botspec.Resource{
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
