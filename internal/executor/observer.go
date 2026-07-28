package executor

import (
	"fmt"

	"nuubot/internal/botspec"
	"nuubot/internal/market"
	"nuubot/internal/setup"
	"nuubot/internal/toolkit/logging"
)

type observerStats struct {
	onBBOCount    uint64
	startMS       uint64
	endMS         uint64
	startPrice    float64
	stopLossPrice float64
	exitPrice     float64
	lastMS        uint64
	lastPrice     float64
	reason        string
}

type observer struct {
	log            *logging.Logger
	nuubot         *setup.Nuubot
	spec           botspec.ExecutorSpec
	cycleNumber    int
	executorNumber int
	signalMS       uint64
	signalPrice    float64
	side           string
	stopLossPct    float64
	subscription   *market.Subscription
	stats          observerStats
	status         Status
}

var _ Executor = (*observer)(nil)
var _ StartHandler = (*observer)(nil)

// Section 1 - Program Flow

// OnInit initializes ObserverExecutor.
func (e *observer) OnInit(ctx BotCycleContext) error {
	// Step 1: bind ObserverExecutor base inputs and log init
	e.log = ctx.Nuubot.Log
	e.nuubot = ctx.Nuubot
	e.spec = ctx.Spec
	e.log.Info(fmt.Sprintf(
		"executor init cycle=%d executor=%d id=%s kind=%s side=%s",
		ctx.CycleNumber,
		ctx.ExecutorNumber,
		ctx.Spec.ID,
		ctx.Spec.Kind,
		ctx.Spec.Side,
	))

	// Step 2: retain resolved ObserverExecutor status
	e.status = ctx.Status

	// Step 3: bind ObserverExecutor inputs
	e.cycleNumber = ctx.CycleNumber
	e.executorNumber = ctx.ExecutorNumber
	e.signalMS = ctx.Signal.TimestampMS()
	e.side = ctx.Spec.Side
	e.stopLossPct, _ = ctx.Spec.StopLossPct.Float64()

	// Step 4: validate ObserverExecutor config
	if e.stopLossPct <= 0 || e.stopLossPct >= 1 {
		e.status = Error
		return fmt.Errorf("observer stop_loss_pct must be between 0 and 1")
	}
	if e.side != Long && e.side != Short {
		e.status = Error
		return fmt.Errorf("%w: observer requires one configured side", ErrRejected)
	}

	// Step 5: log init completed
	e.log.Info(fmt.Sprintf(
		"executor init completed cycle=%d executor=%d id=%s kind=observer side=%s "+
			"signal_ts_ms=%d stop_loss_pct=%f",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
		e.signalMS,
		e.stopLossPct,
	))
	return nil
}

// OnStart starts ObserverExecutor after every sibling initializes.
func (e *observer) OnStart() error {
	// Step 1: log start
	e.log.Info(fmt.Sprintf(
		"executor start cycle=%d executor=%d id=%s kind=observer side=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
	))

	// Step 2: advance ObserverExecutor status
	if e.status == Configured {
		e.status = Starting
	}

	// Step 3: read latest BBO
	var key = e.marketKey()
	var bbo, found = e.nuubot.MarketData.LatestBBO(key)
	if !found {
		return fmt.Errorf("%w: observer executor requires current BBO", ErrRejected)
	}
	e.recordBBO(bbo)

	// Step 4: subscribe to MarketData
	var err error
	e.subscription, err = e.nuubot.MarketData.SubscribeBBO(key, e.onBBO)
	if err != nil {
		return fmt.Errorf("start observer executor: %w", err)
	}

	// Step 5: continue loaded ObserverExecutor state
	if e.status == Starting {
		e.status = Running
	}

	// Step 6: log start completed
	e.log.Info(fmt.Sprintf(
		"executor started cycle=%d executor=%d id=%s kind=observer side=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
	))
	return nil
}

// OnStop stops ObserverExecutor and reports final statistics.
func (e *observer) OnStop(reason string) error {
	// Step 1: log stop
	e.log.Info(fmt.Sprintf(
		"executor stop cycle=%d executor=%d id=%s kind=observer side=%s reason=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
		reason,
	))

	// Step 2: advance ObserverExecutor status
	if e.status == Stopped || e.status == Error {
		return nil
	}
	e.status = Stopping

	// Step 3: stop MarketData subscription
	if err := e.subscription.Stop(); err != nil {
		return fmt.Errorf("stop observer executor: %w", err)
	}
	e.subscription = nil

	// Step 4: preserve stop reason
	if e.stats.reason == "" {
		e.stats.reason = reason
	}

	// Step 5: preserve end time
	if e.stats.endMS == 0 {
		e.stats.endMS = e.stats.lastMS
		if e.stats.endMS == 0 {
			e.stats.endMS = e.signalMS
		}
	}

	// Step 6: mark ObserverExecutor stopped
	e.status = Stopped

	// Step 7: calculate duration
	var durationMS uint64
	if e.stats.endMS >= e.stats.startMS {
		durationMS = e.stats.endMS - e.stats.startMS
	}
	// Step 8: log stop completed
	e.log.Info(fmt.Sprintf(
		"executor stopped cycle=%d executor=%d id=%s kind=observer side=%s signal_ts_ms=%d "+
			"signal_price=%f stop_loss_pct=%f start_ts_ms=%d end_ts_ms=%d "+
			"duration_ms=%d start_price=%f stop_loss_price=%f exit_price=%f "+
			"final_price=%f on_bbo_count=%d stop_reason=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
		e.signalMS,
		e.signalPrice,
		e.stopLossPct,
		e.stats.startMS,
		e.stats.endMS,
		durationMS,
		e.stats.startPrice,
		e.stats.stopLossPrice,
		e.stats.exitPrice,
		e.stats.lastPrice,
		e.stats.onBBOCount,
		e.stats.reason,
	))
	return nil
}

// Section 2 - Domain Helpers

func (e *observer) onBBO() error {
	// Step 1: read latest BBO
	var bbo, found = e.nuubot.MarketData.LatestBBO(e.marketKey())
	if !found {
		return fmt.Errorf("observe BBO: latest BBO is unavailable")
	}
	// Step 2: record received BBO
	e.stats.onBBOCount++
	if e.status != Running {
		return nil
	}

	// Step 3: record current BBO
	e.recordBBO(bbo)

	// Step 4: assess stop loss
	var triggered = e.side == Long && bbo.Price <= e.stats.stopLossPrice ||
		e.side == Short && bbo.Price >= e.stats.stopLossPrice
	if triggered {
		e.stats.endMS = bbo.TimestampMS
		e.stats.exitPrice = bbo.Price
		e.stats.reason = "stop_loss"
		e.status = Stopping
	}
	return nil
}

func (e *observer) recordBBO(bbo market.BBO) {
	// Step 1: record latest BBO
	e.stats.lastMS = bbo.TimestampMS
	e.stats.lastPrice = bbo.Price
	if e.stats.startMS != 0 {
		return
	}

	// Step 2: establish ObserverExecutor entry
	e.stats.startMS = bbo.TimestampMS
	e.stats.startPrice = bbo.Price
	e.signalPrice = bbo.Price
	if e.side == Long {
		e.stats.stopLossPrice = bbo.Price * (1 - e.stopLossPct)
	} else {
		e.stats.stopLossPrice = bbo.Price * (1 + e.stopLossPct)
	}
}

func (e *observer) marketKey() market.Key {
	return market.Key{
		Venue:   e.spec.Resource.Venue,
		Network: e.spec.Resource.Network,
		Symbol:  e.spec.Resource.Symbol,
	}
}

// Status returns ObserverExecutor's canonical lifecycle state.
func (e *observer) Status() Status {
	return e.status
}

// ExitReason returns ObserverExecutor's terminal reason.
func (e *observer) ExitReason() string {
	return e.stats.reason
}

// Telemetry returns one immutable current Observer observation.
func (e *observer) Telemetry() Telemetry {
	return Telemetry{
		ID:       e.spec.ID,
		Kind:     e.spec.Kind,
		Resource: e.spec.Resource,
		Status:   e.status,
	}
}

// Result returns one terminal ObserverExecutor result.
func (e *observer) Result() (Result, error) {
	if e.status != Stopped {
		return Result{}, fmt.Errorf("observer executor result is unavailable")
	}
	return Result{
		ID:            e.spec.ID,
		Role:          e.spec.Role,
		Kind:          e.spec.Kind,
		Side:          e.spec.Side,
		Resource:      e.spec.Resource,
		Status:        e.status,
		ExitReason:    e.stats.reason,
		CapitalUSDC:   e.spec.CapitalUSDC,
		OrderSizeUSDC: e.spec.OrderSizeUSDC,
	}, nil
}

// Section 3 - Generic Helpers
