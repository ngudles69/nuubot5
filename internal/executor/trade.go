package executor

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/ledger"
	"nuubot/internal/market"
	"nuubot/internal/order"
	"nuubot/internal/toolkit/logging"
	"nuubot/internal/trade"
)

type tradeExecutor struct {
	log            *logging.Logger
	account        account.Account
	spec           Spec
	cycleNumber    int
	executorNumber int
	signalMS       uint64
	side           string
	lastBBO        market.BBO
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
var _ BBOHandler = (*tradeExecutor)(nil)
var _ BBOIngestHandler = (*tradeExecutor)(nil)
var _ AccountReconciler = (*tradeExecutor)(nil)
var _ ReconHandler = (*tradeExecutor)(nil)

// Section 1 - Program Flow

// OnInit initializes TradeExecutor without submitting Orders.
func (e *tradeExecutor) OnInit(ctx Context) error {
	e.log = ctx.Log
	e.spec = ctx.Spec
	if e.status != Configured {
		return fmt.Errorf("trade executor cannot initialize from current state")
	}
	e.status = Starting

	// validate trade config
	e.notional = ctx.Spec.OrderSizeUSDC
	e.takeProfitPct = ctx.Spec.TakeProfitPct
	e.stopLossPct = ctx.Spec.StopLossPct
	var equity = ctx.StartingEquityUSDC
	if !equity.IsPositive() {
		equity = ctx.Spec.CapitalUSDC
	}
	if !e.notional.IsPositive() || !equity.IsPositive() ||
		!e.takeProfitPct.IsPositive() || e.takeProfitPct.GreaterThanOrEqual(decimal.NewFromInt(1)) ||
		!e.stopLossPct.IsPositive() || e.stopLossPct.GreaterThanOrEqual(decimal.NewFromInt(1)) ||
		ctx.Spec.FeePct.IsNegative() || ctx.Spec.SlippagePct.IsNegative() ||
		ctx.Spec.Resource.Venue != "simulator" ||
		ctx.Spec.Resource.Network != "simnet" ||
		ctx.Spec.Resource.PhysicalAccountID == "" ||
		ctx.Spec.Resource.Symbol == "" ||
		(ctx.Spec.PersistMode != ledger.None && ctx.Spec.PersistMode != ledger.Max) {
		e.status = Error
		return fmt.Errorf("trade executor config is invalid")
	}

	// admit fixed side
	e.side = ctx.Spec.Side
	if e.side != Long && e.side != Short {
		e.status = Error
		return fmt.Errorf("%w: trade executor requires one configured side", ErrRejected)
	}

	// admit current BBO
	if ctx.LatestBBO.TimestampMS == 0 || ctx.LatestBBO.Price <= 0 {
		e.status = Error
		return fmt.Errorf("%w: trade executor requires current BBO", ErrRejected)
	}
	e.lastBBO = ctx.LatestBBO
	e.cycleNumber = ctx.CycleNumber
	e.executorNumber = ctx.ExecutorNumber
	e.signalMS = ctx.SignalTimestampMS

	// initialize Account
	var ledgerID = uint64(ctx.CycleNumber)<<32 | uint64(ctx.ExecutorNumber)
	var err = e.account.Init(ctx.Log, account.Config{
		LedgerID:        ledgerID,
		CycleNumber:     ctx.CycleNumber,
		ExecutorNumber:  ctx.ExecutorNumber,
		Name:            ctx.Spec.Resource.PhysicalAccountID,
		Venue:           ctx.Spec.Resource.Venue,
		Network:         ctx.Spec.Resource.Network,
		Symbol:          ctx.Spec.Resource.Symbol,
		Meta:            ctx.Spec.Meta,
		MinNotionalUSDC: ctx.Spec.MinNotionalUSDC,
		EquityUSDC:      equity,
		FeePct:          ctx.Spec.FeePct,
		SlippagePct:     ctx.Spec.SlippagePct,
		PersistMode:     ctx.Spec.PersistMode,
		Recon:           ctx.Spec.Recon,
		ResultPath:      ctx.Spec.ResultPath,
	})
	if err != nil {
		e.status = Error
		return fmt.Errorf("initialize trade executor: %w", err)
	}
	err = e.account.IngestBBO(ctx.LatestBBO)
	if err != nil {
		e.account.Stop()
		e.status = Error
		return fmt.Errorf("initialize trade executor: %w", err)
	}
	var existing account.Result
	existing, err = e.account.Result()
	if err != nil {
		return e.stopError(fmt.Errorf("initialize trade executor: %w", err))
	}
	if len(existing.Ledger.Trades) != 0 {
		return e.stopError(fmt.Errorf(
			"initialize trade executor: persisted Trade recovery is pending Runner",
		))
	}

	// initialize TradeExecutor
	e.status = Running
	e.log.Info(fmt.Sprintf(
		"executor initialized cycle=%d executor=%d kind=trade side=%s signal_ts_ms=%d",
		e.cycleNumber,
		e.executorNumber,
		e.side,
		e.signalMS,
	))
	return nil
}

// OnStop closes exposure and stops TradeExecutor.
func (e *tradeExecutor) OnStop(reason string) error {
	if e.status == Stopped {
		return nil
	}
	if e.status == Error {
		return fmt.Errorf("stop trade executor: executor is in error state")
	}
	e.status = Stopping
	if e.exitReason == "" {
		e.exitReason = reason
	}

	// reconcile current Account truth
	var snapshot, _, err = e.account.Reconcile(e.lastBBO.TimestampMS, true)
	if err != nil {
		return e.stopError(fmt.Errorf("stop trade executor: %w", err))
	}

	// cancel active Orders
	var active = e.account.ActiveOrders()
	if len(active) > 0 {
		var cloids = make([]string, len(active))
		for index, current := range active {
			cloids[index] = current.CLOID
		}
		err = e.account.CancelOrders(cloids, e.lastBBO.TimestampMS)
		if err != nil {
			return e.stopError(fmt.Errorf("stop trade executor: %w", err))
		}
		snapshot, _, err = e.account.Reconcile(e.lastBBO.TimestampMS, true)
		if err != nil {
			return e.stopError(fmt.Errorf("stop trade executor: %w", err))
		}
	}

	// close open exposure
	if !snapshot.PositionQuantity.IsZero() {
		var side = order.Sell
		if snapshot.PositionQuantity.IsNegative() {
			side = order.Buy
		}
		var price = decimal.NewFromFloat(e.lastBBO.Price)
		_, err = e.account.PlaceOrders([]account.OrderSpec{{
			TradeID:     e.tradeID,
			Role:        order.Stop,
			Side:        side,
			Type:        order.Limit,
			TimeInForce: order.IOC,
			Quantity:    snapshot.PositionQuantity.Abs(),
			Price:       &price,
			ReduceOnly:  true,
			TimestampMS: e.lastBBO.TimestampMS,
		}})
		if err != nil {
			return e.stopError(fmt.Errorf("stop trade executor: %w", err))
		}
	}

	// reconcile final Venue truth
	snapshot, _, err = e.account.Reconcile(e.lastBBO.TimestampMS, true)
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

	// capture terminal Account result
	var result account.Result
	result, err = e.account.Result()
	if err != nil {
		return e.stopError(fmt.Errorf("stop trade executor: %w", err))
	}

	// stop Account
	err = e.account.Stop()
	if err != nil {
		e.status = Error
		return fmt.Errorf("stop trade executor: %w", err)
	}

	// cache terminal Account result
	e.accountResult = result.Clone()
	e.hasResult = true
	e.status = Stopped
	e.log.Info(fmt.Sprintf(
		"executor stopped cycle=%d executor=%d kind=trade trades=%d fills=%d stop_reason=%s",
		e.cycleNumber,
		e.executorNumber,
		len(result.Ledger.Trades),
		countFills(result),
		e.exitReason,
	))
	return nil
}

// Section 2 - Domain Helpers

// IngestBBO advances the owned Simulator before policy handling.
func (e *tradeExecutor) IngestBBO(bbo market.BBO) error {
	if bbo.Symbol != "" && bbo.Symbol != e.spec.Resource.Symbol {
		return nil
	}
	return e.account.IngestBBO(bbo)
}

// OnBBO records the latest normal Executor BBO.
func (e *tradeExecutor) OnBBO(bbo market.BBO) {
	if bbo.Symbol != "" && bbo.Symbol != e.spec.Resource.Symbol {
		return
	}
	e.lastBBO = bbo
}

// Reconcile refreshes the owned Account when dirty or forced.
func (e *tradeExecutor) Reconcile(
	nowMS uint64,
	forced bool,
) (account.Snapshot, bool, error) {
	return e.account.Reconcile(nowMS, forced)
}

// OnRecon runs TradeExecutor policy after one accepted reconciliation barrier.
func (e *tradeExecutor) OnRecon(_ uint64) error {
	if e.status != Running {
		return nil
	}

	// submit bracket when no Trade exists
	if e.tradeID == 0 {
		var entry = decimal.NewFromFloat(e.lastBBO.Price)
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
		var placed, err = e.account.PlaceOrders([]account.OrderSpec{
			{
				Role:        order.Entry,
				Side:        entrySide,
				Type:        order.Limit,
				TimeInForce: order.IOC,
				Quantity:    quantity,
				Price:       &entry,
				TimestampMS: e.lastBBO.TimestampMS,
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
				TimestampMS:  e.lastBBO.TimestampMS,
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
				TimestampMS:  e.lastBBO.TimestampMS,
			},
		})
		if err != nil {
			return fmt.Errorf("run trade executor recon: %w", err)
		}
		e.tradeID = placed.TradeID
		return nil
	}

	// check owned Trade completion
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

func countFills(result account.Result) int {
	var count int
	for _, ownedTrade := range result.Ledger.Trades {
		for _, ownedOrder := range ownedTrade.Orders {
			count += len(ownedOrder.Fills)
		}
	}
	return count
}

func (e *tradeExecutor) stopError(err error) error {
	e.status = Error
	return errors.Join(err, e.account.Stop())
}
