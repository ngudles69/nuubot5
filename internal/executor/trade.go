package executor

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/botspec"
	"nuubot/internal/ledger"
	"nuubot/internal/market"
	"nuubot/internal/order"
	"nuubot/internal/setup"
	"nuubot/internal/toolkit/logging"
	"nuubot/internal/trade"
)

type tradeExecutor struct {
	log            *logging.Logger
	nuubot         *setup.Nuubot
	account        account.Account
	spec           botspec.ExecutorSpec
	cycleNumber    int
	executorNumber int
	signalMS       uint64
	side           string
	notional       decimal.Decimal
	takeProfitPct  decimal.Decimal
	stopLossPct    decimal.Decimal
	tradeID        uint64
	accountResult  account.Result
	hasResult      bool
	status         Status
	exitReason     string
}

var _ Executor = (*tradeExecutor)(nil)
var _ StartHandler = (*tradeExecutor)(nil)
var _ AccountReconciler = (*tradeExecutor)(nil)
var _ ReconHandler = (*tradeExecutor)(nil)

// Section 1 - Program Flow

// OnInit initializes TradeExecutor without submitting Orders.
func (e *tradeExecutor) OnInit(ctx BotCycleContext) error {
	// Step 1: bind TradeExecutor inputs and log init
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

	// Step 2: reject terminal TradeExecutor state
	if e.status == Error || e.status == Stopped {
		return fmt.Errorf("trade executor cannot initialize from terminal status %s", e.status)
	}

	// Step 3: validate TradeExecutor config

	// Step 3.1: retain financial inputs
	e.notional = ctx.Spec.OrderSizeUSDC
	e.takeProfitPct = ctx.Spec.TakeProfitPct
	e.stopLossPct = ctx.Spec.StopLossPct
	var equity = ctx.StartingEquityUSDC
	if !equity.IsPositive() {
		equity = ctx.Spec.CapitalUSDC
	}

	// Step 3.2: validate Executor ID
	if ctx.Spec.ID == "" {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.3: validate capital and order size
	if !equity.IsPositive() || !e.notional.IsPositive() {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.4: validate take-profit percentage
	if !e.takeProfitPct.IsPositive() || e.takeProfitPct.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.5: validate stop-loss percentage
	if !e.stopLossPct.IsPositive() || e.stopLossPct.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.6: validate fees and slippage
	if ctx.Spec.FeePct.IsNegative() || ctx.Spec.SlippagePct.IsNegative() {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.7: validate Venue
	if ctx.Spec.Resource.Venue != "simulator" {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.8: validate network
	if ctx.Spec.Resource.Network != "simnet" {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.9: validate physical Account ID
	if ctx.Spec.Resource.PhysicalAccountID == "" {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.10: validate symbol
	if ctx.Spec.Resource.Symbol == "" {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.11: validate persistence mode
	if ctx.Spec.PersistMode != ledger.None && ctx.Spec.PersistMode != ledger.Max {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// Step 3.12: validate fixed side
	e.side = ctx.Spec.Side
	if e.side != Long && e.side != Short {
		e.status = Error
		return fmt.Errorf("%w: trade executor requires one configured side", ErrRejected)
	}

	// Step 4: retain TradeExecutor identity
	e.cycleNumber = ctx.CycleNumber
	e.executorNumber = ctx.ExecutorNumber
	e.signalMS = ctx.Signal.TimestampMS()

	// Step 5: initialize Account
	var ledgerID = uint64(ctx.CycleNumber)<<32 | uint64(ctx.ExecutorNumber)
	var err = e.account.Init(account.Config{
		Nuubot:         ctx.Nuubot,
		LedgerID:       ledgerID,
		CycleNumber:    ctx.CycleNumber,
		ExecutorNumber: ctx.ExecutorNumber,
		Name:           ctx.Spec.Resource.PhysicalAccountID,
		Venue:          ctx.Spec.Resource.Venue,
		Network:        ctx.Spec.Resource.Network,
		Symbol:         ctx.Spec.Resource.Symbol,
		EquityUSDC:     equity,
		FeePct:         ctx.Spec.FeePct,
		SlippagePct:    ctx.Spec.SlippagePct,
		PersistMode:    ctx.Spec.PersistMode,
		Recon:          ctx.Spec.Recon,
	})
	if err != nil {
		e.status = Error
		return fmt.Errorf("initialize trade executor: %w", err)
	}
	var existing account.Result
	existing, err = e.account.Result()
	if err != nil {
		return e.stopError(fmt.Errorf("initialize trade executor: %w", err))
	}
	if existing.Ledger.Trades != 0 {
		return e.stopError(fmt.Errorf(
			"initialize trade executor: persisted Trade recovery is pending Runner",
		))
	}

	// Step 6: log init completed
	e.log.Info(fmt.Sprintf(
		"executor init completed cycle=%d executor=%d id=%s kind=trade side=%s signal_ts_ms=%d",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
		e.signalMS,
	))
	return nil
}

// OnStart starts TradeExecutor after every sibling initializes.
func (e *tradeExecutor) OnStart() error {
	// Step 1: log start
	e.log.Info(fmt.Sprintf(
		"executor start cycle=%d executor=%d id=%s kind=trade side=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
	))

	// Step 2: reject terminal TradeExecutor state
	if e.status == Error || e.status == Stopped {
		return fmt.Errorf("trade executor cannot start from terminal status %s", e.status)
	}

	// Step 3: read latest BBO
	if _, err := e.latestBBO(); err != nil {
		return err
	}

	// Step 4: continue loaded TradeExecutor state
	if e.status == Configured || e.status == Starting {
		e.status = Running
	}

	// Step 5: log start completed
	e.log.Info(fmt.Sprintf(
		"executor started cycle=%d executor=%d id=%s kind=trade side=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
	))
	return nil
}

// OnStop closes exposure and stops TradeExecutor.
func (e *tradeExecutor) OnStop(reason string) error {
	// Step 1: log stop
	e.log.Info(fmt.Sprintf(
		"executor stop cycle=%d executor=%d id=%s kind=trade side=%s reason=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
		reason,
	))

	// Step 2: validate stop state
	if e.status == Stopped {
		return nil
	}
	if e.status == Error {
		return fmt.Errorf("stop trade executor: executor is in error state")
	}

	// Step 3: mark TradeExecutor stopping
	e.status = Stopping
	if e.exitReason == "" {
		e.exitReason = reason
	}

	// Step 4: read current time and latest BBO
	var nowMS = e.nuubot.Clock.NowMS()
	var bbo, err = e.latestBBO()
	if err != nil {
		return e.stopError(err)
	}

	// Step 5: reconcile current Account truth
	var snapshot account.Snapshot
	snapshot, _, _, err = e.account.Reconcile(nowMS, true)
	if err != nil {
		return e.stopError(fmt.Errorf("stop trade executor: %w", err))
	}

	// Step 6: cancel active Orders
	var active = e.account.ActiveOrders()
	if len(active) > 0 {
		var cloids = make([]string, len(active))
		for index, current := range active {
			cloids[index] = current.CLOID
		}
		err = e.account.CancelOrders(cloids, nowMS)
		if err != nil {
			return e.stopError(fmt.Errorf("stop trade executor: %w", err))
		}
		snapshot, _, _, err = e.account.Reconcile(nowMS, true)
		if err != nil {
			return e.stopError(fmt.Errorf("stop trade executor: %w", err))
		}
	}

	// Step 7: close open exposure
	if !snapshot.PositionQuantity.IsZero() {
		var side = order.Sell
		if snapshot.PositionQuantity.IsNegative() {
			side = order.Buy
		}
		var price = decimal.NewFromFloat(bbo.Price)
		_, err = e.account.PlaceOrders([]account.OrderSpec{{
			TradeID:     e.tradeID,
			Role:        order.Stop,
			Side:        side,
			Type:        order.Limit,
			TimeInForce: order.IOC,
			Quantity:    snapshot.PositionQuantity.Abs(),
			Price:       &price,
			ReduceOnly:  true,
			TimestampMS: nowMS,
		}})
		if err != nil {
			return e.stopError(fmt.Errorf("stop trade executor: %w", err))
		}
	}

	// Step 8: reconcile final Venue truth
	snapshot, _, _, err = e.account.Reconcile(nowMS, true)
	if err != nil {
		return e.stopError(fmt.Errorf("stop trade executor: %w", err))
	}
	var activeOrders = e.account.ActiveOrders()
	if !snapshot.PositionQuantity.IsZero() || len(activeOrders) != 0 {
		return e.stopError(fmt.Errorf(
			"stop trade executor: Account is not flat and inactive position=%s active_orders=%d",
			snapshot.PositionQuantity,
			len(activeOrders),
		))
	}

	// Step 9: capture terminal Account result
	var result account.Result
	result, err = e.account.Result()
	if err != nil {
		return e.stopError(fmt.Errorf("stop trade executor: %w", err))
	}

	// Step 10: stop Account
	err = e.account.Stop()
	if err != nil {
		e.status = Error
		return fmt.Errorf("stop trade executor: %w", err)
	}

	// Step 11: cache terminal Account result
	e.accountResult = result.Clone()
	e.hasResult = true
	e.status = Stopped

	// Step 12: log stop completed
	e.log.Info(fmt.Sprintf(
		"executor stopped cycle=%d executor=%d id=%s kind=trade side=%s trades=%d fills=%d stop_reason=%s",
		e.cycleNumber,
		e.executorNumber,
		e.spec.ID,
		e.side,
		result.Ledger.Trades,
		result.Ledger.Fills,
		e.exitReason,
	))
	return nil
}

// Section 2 - Domain Helpers

// Section 2.1 - Market Data

func (e *tradeExecutor) latestBBO() (market.BBO, error) {
	var bbo, found = e.nuubot.MarketData.LatestBBO(market.Key{
		Venue:   e.spec.Resource.Venue,
		Network: e.spec.Resource.Network,
		Symbol:  e.spec.Resource.Symbol,
	})
	if !found {
		return market.BBO{}, fmt.Errorf("%w: trade executor requires current BBO", ErrRejected)
	}
	return bbo, nil
}

// Section 2.2 - Reconciliation and Policy

// Reconcile refreshes the owned Account when dirty or forced.
func (e *tradeExecutor) Reconcile(
	nowMS uint64,
	forced bool,
) (account.Snapshot, bool, uint64, error) {
	return e.account.Reconcile(nowMS, forced)
}

// OnRecon runs TradeExecutor policy after one accepted reconciliation barrier.
func (e *tradeExecutor) OnRecon() error {
	if e.status != Running {
		return nil
	}

	// Step 1: submit bracket when no Trade exists
	if e.tradeID == 0 {
		var bbo, err = e.latestBBO()
		if err != nil {
			return err
		}
		var nowMS = e.nuubot.Clock.NowMS()
		var entry = decimal.NewFromFloat(bbo.Price)
		var quantity = e.notional.Div(entry)
		var takeProfit = entry.Mul(decimal.NewFromInt(1).Add(e.takeProfitPct))
		var stopLoss = entry.Mul(decimal.NewFromInt(1).Sub(e.stopLossPct))
		var entrySide = order.Buy
		var exitSide = order.Sell
		if e.side == Short {
			entrySide = order.Sell
			exitSide = order.Buy
			takeProfit = entry.Mul(decimal.NewFromInt(1).Sub(e.takeProfitPct))
			stopLoss = entry.Mul(decimal.NewFromInt(1).Add(e.stopLossPct))
		}
		var placed account.PlaceResult
		placed, err = e.account.PlaceOrders([]account.OrderSpec{
			{
				Role:        order.Entry,
				Side:        entrySide,
				Type:        order.Limit,
				TimeInForce: order.IOC,
				Quantity:    quantity,
				Price:       &entry,
				TimestampMS: nowMS,
			},
			{
				Role:         order.TakeProfit,
				Side:         exitSide,
				Type:         order.Trigger,
				TimeInForce:  order.GTC,
				Quantity:     quantity,
				Price:        &takeProfit,
				TriggerPrice: &takeProfit,
				ReduceOnly:   true,
				TimestampMS:  nowMS,
			},
			{
				Role:         order.StopLoss,
				Side:         exitSide,
				Type:         order.Trigger,
				TimeInForce:  order.GTC,
				Quantity:     quantity,
				Price:        &stopLoss,
				TriggerPrice: &stopLoss,
				ReduceOnly:   true,
				TimestampMS:  nowMS,
			},
		})
		if err != nil {
			return fmt.Errorf("run trade executor recon: %w", err)
		}
		e.tradeID = placed.TradeID
		return nil
	}

	// Step 2: check owned Trade completion
	var owned, err = e.account.Trade(e.tradeID)
	if err != nil {
		return fmt.Errorf("run trade executor recon: %w", err)
	}
	if (owned.Status == trade.Closed || owned.Status == trade.Canceled) &&
		len(e.account.ActiveOrders()) == 0 {
		e.exitReason = "completed"
		e.status = Stopping
	}
	return nil
}

// Section 2.3 - Observation

// Status returns TradeExecutor's canonical lifecycle state.
func (e *tradeExecutor) Status() Status {
	return e.status
}

// ExitReason returns TradeExecutor's terminal reason.
func (e *tradeExecutor) ExitReason() string {
	return e.exitReason
}

// Telemetry returns one immutable current TradeExecutor observation.
func (e *tradeExecutor) Telemetry() Telemetry {
	var result = Telemetry{
		ID:       e.spec.ID,
		Kind:     e.spec.Kind,
		Resource: e.spec.Resource,
		Status:   e.status,
	}
	var snapshot, observed = e.account.Telemetry()
	if observed {
		result.Account = &snapshot
	}
	return result
}

// Result returns one terminal TradeExecutor result.
func (e *tradeExecutor) Result() (Result, error) {
	if !e.hasResult {
		return Result{}, errors.New("trade executor result is unavailable")
	}
	var accountResult = e.accountResult.Clone()
	return Result{
		ID:            e.spec.ID,
		Role:          e.spec.Role,
		Kind:          e.spec.Kind,
		Side:          e.spec.Side,
		Resource:      e.spec.Resource,
		Status:        e.status,
		ExitReason:    e.exitReason,
		CapitalUSDC:   e.spec.CapitalUSDC,
		OrderSizeUSDC: e.spec.OrderSizeUSDC,
		Account:       &accountResult,
	}, nil
}

// Section 3 - Generic Helpers

func (e *tradeExecutor) stopError(err error) error {
	e.status = Error
	return errors.Join(err, e.account.Stop())
}
