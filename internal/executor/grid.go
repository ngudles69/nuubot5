package executor

import (
	"errors"
	"fmt"
	"sort"

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

const gridDeploymentPct = 95

type gridExecutor struct {
	log            *logging.Logger
	nuubot         *setup.Nuubot
	account        account.Account
	spec           botspec.ExecutorSpec
	cycleNumber    int
	executorNumber int
	signalMS       uint64
	side           string
	lastBBO        market.BBO
	levels         []GridLevel
	accountResult  account.Result
	cancellations  uint64
	closureOrders  uint64
	retries        uint64
	roundTrips     uint64
	hasResult      bool
	status         Status
	exitReason     string
}

var _ Executor = (*gridExecutor)(nil)
var _ StartHandler = (*gridExecutor)(nil)
var _ BBOHandler = (*gridExecutor)(nil)
var _ BBOIngestHandler = (*gridExecutor)(nil)
var _ AccountReconciler = (*gridExecutor)(nil)
var _ ReconHandler = (*gridExecutor)(nil)

// Section 1 - Program Flow

// OnInit initializes GridExecutor and logs its validated levels without submitting Orders.
func (e *gridExecutor) OnInit(ctx Context) error {
	// Step 1: bind GridExecutor inputs
	e.log = ctx.Nuubot.Log
	e.nuubot = ctx.Nuubot
	e.spec = ctx.Spec

	// Step 2: validate GridExecutor state
	if e.status != Configured {
		return fmt.Errorf("grid executor cannot initialize from current state")
	}
	e.status = Starting

	// Step 3: validate GridExecutor config
	var equity = ctx.StartingEquityUSDC
	if !equity.IsPositive() {
		equity = ctx.Spec.CapitalUSDC
	}
	if !equity.IsPositive() ||
		ctx.Spec.GridLevels < 3 || ctx.Spec.GridLevels > 1024 ||
		!ctx.Spec.RangePct.IsPositive() ||
		ctx.Spec.RangePct.GreaterThanOrEqual(decimal.NewFromInt(1)) ||
		ctx.Spec.MinExpectedPnL.IsNegative() ||
		ctx.Spec.FeePct.IsNegative() ||
		ctx.Spec.SlippagePct.IsNegative() ||
		ctx.Spec.Resource.Venue != "simulator" ||
		ctx.Spec.Resource.Network != "simnet" ||
		ctx.Spec.Resource.PhysicalAccountID == "" ||
		ctx.Spec.Resource.Symbol == "" ||
		(ctx.Spec.PersistMode != ledger.None && ctx.Spec.PersistMode != ledger.Max) {
		e.status = Error
		return fmt.Errorf("grid executor config is invalid")
	}

	// Step 4: validate fixed side
	e.side = ctx.Spec.Side
	if e.side != Long && e.side != Short {
		e.status = Error
		return fmt.Errorf("%w: grid executor requires one configured side", ErrRejected)
	}

	// Step 5: retain current BBO and identity
	if ctx.LatestBBO.TimestampMS == 0 || ctx.LatestBBO.Price <= 0 {
		e.status = Error
		return fmt.Errorf("%w: grid executor requires current BBO", ErrRejected)
	}
	e.lastBBO = ctx.LatestBBO
	e.cycleNumber = ctx.CycleNumber
	e.executorNumber = ctx.ExecutorNumber
	e.signalMS = ctx.Signal.TimestampMS()

	// Step 6: calculate Grid levels
	var err error
	e.levels, err = calculateGridLevels(
		ctx.Nuubot,
		ctx.Spec,
		ctx.LatestBBO,
		equity,
	)
	if err != nil {
		e.status = Error
		return fmt.Errorf("initialize grid executor: %w", err)
	}

	// Step 7: initialize Account
	var ledgerID = uint64(ctx.CycleNumber)<<32 | uint64(ctx.ExecutorNumber)
	err = e.account.Init(account.Config{
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
		return fmt.Errorf("initialize grid executor: %w", err)
	}
	err = e.account.IngestBBO(ctx.LatestBBO)
	if err != nil {
		e.account.Stop()
		e.status = Error
		return fmt.Errorf("initialize grid executor: %w", err)
	}
	var existing account.Result
	existing, err = e.account.Result()
	if err != nil {
		return e.stopError(fmt.Errorf("initialize grid executor: %w", err))
	}
	if existing.Ledger.Trades != 0 {
		return e.stopError(fmt.Errorf(
			"initialize grid executor: persisted Trade recovery is pending Runner",
		))
	}

	// Step 8: log validated Grid table
	e.logGridTable()

	return nil
}

// OnStart submits the initial Grid after every sibling Executor initializes.
func (e *gridExecutor) OnStart() error {
	// Step 1: validate start state
	if e.status != Starting {
		return fmt.Errorf("grid executor cannot start from current state")
	}

	// Step 2: submit initial Grid at cycle-start BBO
	for _, index := range e.activeLevelIndexes() {
		var err = e.submitLevel(index, true, e.lastBBO.TimestampMS)
		if err != nil {
			return e.failSubmission(err)
		}
	}
	// Step 3: mark GridExecutor running
	e.status = Running

	// Step 4: log start completed
	e.log.Info(fmt.Sprintf(
		"executor initialized cycle=%d executor=%d kind=grid side=%s signal_ts_ms=%d levels=%d active_levels=%d",
		e.cycleNumber,
		e.executorNumber,
		e.side,
		e.signalMS,
		len(e.levels),
		len(e.levels)-2,
	))
	return nil
}

// OnStop cancels Grid Orders, closes every open Trade, and stops GridExecutor.
func (e *gridExecutor) OnStop(reason string) error {
	// Step 1: validate stop state
	if e.status == Stopped || (e.status == Error && e.hasResult) {
		return nil
	}
	if e.status == Error {
		return fmt.Errorf("stop grid executor: executor is in error state")
	}

	// Step 2: mark GridExecutor stopping
	e.status = Stopping
	if e.exitReason == "" {
		e.exitReason = reason
	}

	// Step 3: reconcile current Account truth
	var snapshot, _, _, err = e.account.Reconcile(e.lastBBO.TimestampMS, true)
	if err != nil {
		return e.stopError(fmt.Errorf("stop grid executor: %w", err))
	}

	// Step 4: cancel active Orders
	var active = orderedCancellations(e.account.ActiveOrders())
	if len(active) > 0 {
		var cloids = make([]string, len(active))
		for index, current := range active {
			cloids[index] = current.CLOID
		}
		err = e.account.CancelOrders(cloids, e.lastBBO.TimestampMS)
		if err != nil {
			return e.stopError(fmt.Errorf("stop grid executor: %w", err))
		}
		e.cancellations += uint64(len(cloids))
		snapshot, _, _, err = e.account.Reconcile(e.lastBBO.TimestampMS, true)
		if err != nil {
			return e.stopError(fmt.Errorf("stop grid executor: %w", err))
		}
	}

	// Step 5: close open Trades
	for _, owned := range e.account.OpenTrades() {
		if owned.OpenQuantity.IsZero() {
			continue
		}
		var level, found = e.levelForTrade(owned.TradeID)
		if !found {
			return e.stopError(fmt.Errorf(
				"stop grid executor: open Trade %d lacks Grid level",
				owned.TradeID,
			))
		}
		var closeSide = order.Sell
		if owned.Side == trade.Short {
			closeSide = order.Buy
		}
		var price = decimal.NewFromFloat(e.lastBBO.Price)
		err = e.placeExistingWithRetries(account.OrderSpec{
			TradeID:     owned.TradeID,
			OrderLevel:  level,
			Role:        order.Stop,
			Side:        closeSide,
			Type:        order.Limit,
			TimeInForce: order.IOC,
			Quantity:    owned.OpenQuantity,
			Price:       &price,
			ReduceOnly:  true,
			TimestampMS: e.lastBBO.TimestampMS,
		})
		if err != nil {
			return e.stopError(fmt.Errorf("stop grid executor: %w", err))
		}
		e.closureOrders++
	}

	// Step 6: reconcile final Venue truth
	snapshot, _, _, err = e.account.Reconcile(e.lastBBO.TimestampMS, true)
	if err != nil {
		return e.stopError(fmt.Errorf("stop grid executor: %w", err))
	}
	var activeOrders = e.account.ActiveOrders()
	if !snapshot.PositionQuantity.IsZero() || len(activeOrders) != 0 {
		return e.stopError(fmt.Errorf(
			"stop grid executor: Account is not flat and inactive position=%s active_orders=%d",
			snapshot.PositionQuantity,
			len(activeOrders),
		))
	}
	err = e.refreshLevelStates(e.lastBBO.TimestampMS)
	if err != nil {
		return e.stopError(fmt.Errorf("stop grid executor: %w", err))
	}

	// Step 7: capture terminal Account result
	var result account.Result
	result, err = e.account.Result()
	if err != nil {
		return e.stopError(fmt.Errorf("stop grid executor: %w", err))
	}
	e.roundTrips = e.account.CountOrders(order.TakeProfit, order.Filled)

	// Step 8: stop Account
	err = e.account.Stop()
	if err != nil {
		e.status = Error
		return fmt.Errorf("stop grid executor: %w", err)
	}

	// Step 9: cache terminal Account result
	e.accountResult = result.Clone()
	e.hasResult = true
	e.status = Stopped

	// Step 10: log stop completed
	e.log.Info(fmt.Sprintf(
		"executor stopped cycle=%d executor=%d kind=grid trades=%d orders=%d fills=%d cancellations=%d closure_orders=%d retries=%d round_trips=%d stop_reason=%s",
		e.cycleNumber,
		e.executorNumber,
		result.Ledger.Trades,
		result.Ledger.Orders,
		result.Ledger.Fills,
		e.cancellations,
		e.closureOrders,
		e.retries,
		e.roundTrips,
		e.exitReason,
	))
	return nil
}

// Section 2 - Domain Helpers

// Section 2.1 - Market Data

// IngestBBO advances the owned Simulator before Grid policy handling.
func (e *gridExecutor) IngestBBO(bbo market.BBO) error {
	if bbo.Symbol != "" && bbo.Symbol != e.spec.Resource.Symbol {
		return nil
	}
	e.lastBBO = bbo
	return e.account.IngestBBO(bbo)
}

// OnBBO records the latest Grid BBO and requests boundary exit.
func (e *gridExecutor) OnBBO(bbo market.BBO) {
	if bbo.Symbol != "" && bbo.Symbol != e.spec.Resource.Symbol {
		return
	}
	e.lastBBO = bbo
	if e.status != Running || len(e.levels) < 3 {
		return
	}
	var price = decimal.NewFromFloat(bbo.Price)
	var lower = e.levels[0].GridPrice
	var upper = e.levels[len(e.levels)-1].GridPrice
	switch {
	case price.LessThanOrEqual(lower):
		e.status = Stopping
		if e.side == Long {
			e.exitReason = "stop_loss"
		} else {
			e.exitReason = "take_profit"
		}
	case price.GreaterThanOrEqual(upper):
		e.status = Stopping
		if e.side == Long {
			e.exitReason = "take_profit"
		} else {
			e.exitReason = "stop_loss"
		}
	}
}

// Section 2.2 - Reconciliation and Policy

// Reconcile refreshes the owned Account when dirty or forced.
func (e *gridExecutor) Reconcile(
	nowMS uint64,
	forced bool,
) (account.Snapshot, bool, uint64, error) {
	return e.account.Reconcile(nowMS, forced)
}

// OnRecon re-enters completed levels.
func (e *gridExecutor) OnRecon() error {
	if e.status != Running {
		return nil
	}

	// Step 1: read current Nuubot time
	var nowMS = e.nuubot.Clock.NowMS()

	// Step 2: re-enter completed levels
	for _, index := range e.activeLevelIndexes() {
		var level = &e.levels[index]
		if level.CurrentTradeID == 0 {
			continue
		}
		var owned, err = e.account.Trade(level.CurrentTradeID)
		if err != nil {
			return fmt.Errorf("run grid executor recon: %w", err)
		}
		level.CurrentTradeNo = owned.TradeNo
		level.CurrentTradeStatus = string(owned.Status)
		if owned.Status != trade.Closed && owned.Status != trade.Canceled {
			continue
		}
		level.LastCompletedMS = nowMS
		level.Status = "completed"
		err = e.submitLevel(index, false, nowMS)
		if err != nil {
			return e.failSubmission(err)
		}
	}
	return nil
}

// Section 2.3 - Observation

// Status returns GridExecutor's canonical lifecycle state.
func (e *gridExecutor) Status() Status {
	return e.status
}

// ExitReason returns GridExecutor's terminal reason.
func (e *gridExecutor) ExitReason() string {
	return e.exitReason
}

// Telemetry returns one immutable current GridExecutor observation.
func (e *gridExecutor) Telemetry() Telemetry {
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

// Result returns one terminal GridExecutor result.
func (e *gridExecutor) Result() (Result, error) {
	if !e.hasResult {
		return Result{}, errors.New("grid executor result is unavailable")
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
		Cancellations: e.cancellations,
		ClosureOrders: e.closureOrders,
		Retries:       e.retries,
		RoundTrips:    e.roundTrips,
		Levels:        append([]GridLevel(nil), e.levels...),
		Account:       &accountResult,
	}, nil
}

// Section 2.4 - Submission

func (e *gridExecutor) submitLevel(index int, initial bool, nowMS uint64) error {
	var level = &e.levels[index]
	level.Status = "submitting"
	var entry = level.ReentryPrice
	if initial {
		entry = level.InitialEntryPrice
	}
	var entrySide = order.Buy
	var exitSide = order.Sell
	if e.side == Short {
		entrySide = order.Sell
		exitSide = order.Buy
	}
	var lastErr error
	var attempts int
	for attempt := 0; attempt < 3; attempt++ {
		attempts++
		level.SubmissionAttempts++
		if attempt > 0 {
			e.retries++
		}
		var placed account.PlaceResult
		placed, lastErr = e.account.PlaceOrders([]account.OrderSpec{
			{
				OrderLevel:  level.Level,
				Role:        order.Entry,
				Side:        entrySide,
				Type:        order.Limit,
				TimeInForce: order.GTC,
				Quantity:    level.Quantity,
				Price:       &entry,
				TimestampMS: nowMS,
			},
			{
				OrderLevel:   level.Level,
				Role:         order.TakeProfit,
				Side:         exitSide,
				Type:         order.Trigger,
				TimeInForce:  order.GTC,
				Quantity:     level.Quantity,
				Price:        &level.ExitPrice,
				TriggerPrice: &level.ExitPrice,
				ReduceOnly:   true,
				TimestampMS:  nowMS,
			},
		})
		if placed.TradeID != 0 {
			level.CurrentTradeID = placed.TradeID
			var owned, tradeErr = e.account.Trade(placed.TradeID)
			if tradeErr == nil {
				level.CurrentTradeNo = owned.TradeNo
				level.CurrentTradeStatus = string(owned.Status)
			}
		}
		if lastErr != nil {
			if errors.Is(lastErr, account.ErrNotSubmitted) && attempt < 2 {
				continue
			}
			break
		}
		var owned trade.ReconState
		owned, lastErr = e.account.Trade(placed.TradeID)
		if lastErr != nil {
			break
		}
		level.CurrentTradeID = placed.TradeID
		level.CurrentTradeNo = owned.TradeNo
		level.CurrentTradeStatus = string(owned.Status)
		level.Status = "active"
		level.LastSubmittedMS = nowMS
		if initial {
			level.InitialSubmissionCompleted = true
		}
		return nil
	}
	level.Status = "error"
	return fmt.Errorf(
		"submit grid level %d after %d attempts: %w",
		level.Level,
		attempts,
		lastErr,
	)
}

func (e *gridExecutor) placeExistingWithRetries(spec account.OrderSpec) error {
	var err error
	var attempts int
	for attempt := 0; attempt < 3; attempt++ {
		attempts++
		if attempt > 0 {
			e.retries++
		}
		_, err = e.account.PlaceOrders([]account.OrderSpec{spec})
		if err == nil {
			return nil
		}
		if !errors.Is(err, account.ErrNotSubmitted) {
			break
		}
	}
	return fmt.Errorf("submit closure after %d attempts: %w", attempts, err)
}

func (e *gridExecutor) failSubmission(err error) error {
	e.exitReason = "submission_error"
	var stopErr = e.OnStop(e.exitReason)
	e.status = Error
	return errors.Join(err, stopErr)
}

// Section 2.5 - Level State

func (e *gridExecutor) levelForTrade(tradeID uint64) (uint16, bool) {
	for _, level := range e.levels {
		if level.CurrentTradeID == tradeID {
			return level.Level, true
		}
	}
	return 0, false
}

func (e *gridExecutor) refreshLevelStates(nowMS uint64) error {
	for index := 1; index < len(e.levels)-1; index++ {
		var level = &e.levels[index]
		if level.CurrentTradeID == 0 {
			continue
		}
		var owned, err = e.account.Trade(level.CurrentTradeID)
		if err != nil {
			return err
		}
		level.CurrentTradeNo = owned.TradeNo
		level.CurrentTradeStatus = string(owned.Status)
		if owned.Status == trade.Closed || owned.Status == trade.Canceled {
			level.Status = "completed"
			level.LastCompletedMS = nowMS
		}
	}
	return nil
}

func (e *gridExecutor) activeLevelIndexes() []int {
	var indexes = make([]int, 0, len(e.levels)-2)
	if e.side == Long {
		for index := 1; index < len(e.levels)-1; index++ {
			indexes = append(indexes, index)
		}
		return indexes
	}
	for index := len(e.levels) - 2; index > 0; index-- {
		indexes = append(indexes, index)
	}
	return indexes
}

func (e *gridExecutor) logGridTable() {
	e.log.Info(fmt.Sprintf(
		"grid table cycle=%d executor=%d side=%s levels=%d",
		e.cycleNumber,
		e.executorNumber,
		e.side,
		len(e.levels),
	))
	for _, level := range e.levels {
		e.log.Info(fmt.Sprintf(
			"grid level cycle=%d executor=%d level=%d grid_price=%s entry_price=%s exit_price=%s side=%s size=%s notional=%s intended_action=%s",
			e.cycleNumber,
			e.executorNumber,
			level.Level,
			level.GridPrice,
			level.InitialEntryPrice,
			level.ExitPrice,
			e.side,
			level.Quantity,
			level.InitialNotional,
			level.IntendedAction,
		))
	}
}

// Section 2.6 - Grid Calculation

func calculateGridLevels(
	nuubot *setup.Nuubot,
	spec botspec.ExecutorSpec,
	bbo market.BBO,
	equity decimal.Decimal,
) ([]GridLevel, error) {
	var start = decimal.NewFromFloat(bbo.Price)
	var one = decimal.NewFromInt(1)
	var minNotionalUSDC = decimal.NewFromInt(
		int64(nuubot.App.Hyperliquid.MinOrderNotionalUSDC),
	)
	var lower = nuubot.Meta.RoundPrice(start.Mul(one.Sub(spec.RangePct)))
	var upper = nuubot.Meta.RoundPrice(start.Mul(one.Add(spec.RangePct)))
	var step = upper.Sub(lower).Div(decimal.NewFromInt(int64(spec.GridLevels - 1)))
	var prices = make([]decimal.Decimal, spec.GridLevels)
	for index := range prices {
		prices[index] = nuubot.Meta.RoundPrice(
			lower.Add(step.Mul(decimal.NewFromInt(int64(index)))),
		)
		if index > 0 && !prices[index].GreaterThan(prices[index-1]) {
			return nil, fmt.Errorf("rounded Grid prices are not strictly increasing")
		}
	}
	var levels = make([]GridLevel, spec.GridLevels)
	for index, price := range prices {
		levels[index] = GridLevel{
			Level:     uint16(index),
			Boundary:  index == 0 || index == len(prices)-1,
			GridPrice: price,
			Status:    "ready",
		}
	}
	levels[0].Status = "boundary"
	levels[len(levels)-1].Status = "boundary"
	if spec.Side == Long {
		levels[0].IntendedAction = "stop_loss_bound"
		levels[len(levels)-1].IntendedAction = "take_profit_bound"
	} else {
		levels[0].IntendedAction = "take_profit_bound"
		levels[len(levels)-1].IntendedAction = "stop_loss_bound"
	}

	var deployed = equity.Mul(decimal.NewFromInt(gridDeploymentPct)).
		Div(decimal.NewFromInt(100))
	var slice = deployed.Div(decimal.NewFromInt(int64(spec.GridLevels - 2)))
	var feeRate = spec.FeePct.Div(decimal.NewFromInt(100))
	for index := 1; index < len(levels)-1; index++ {
		var level = &levels[index]
		level.ReentryPrice = level.GridPrice
		level.InitialEntryPrice = level.GridPrice
		var exitIndex = index + 1
		level.IntendedAction = "enter_long"
		if spec.Side == Long {
			if start.LessThan(level.InitialEntryPrice) {
				level.InitialEntryPrice = nuubot.Meta.RoundPrice(start)
			}
		} else {
			exitIndex = index - 1
			level.IntendedAction = "enter_short"
			if start.GreaterThan(level.InitialEntryPrice) {
				level.InitialEntryPrice = nuubot.Meta.RoundPrice(start)
			}
		}
		level.ExitPrice = levels[exitIndex].GridPrice
		var sizingPrice = decimal.Max(level.InitialEntryPrice, level.ReentryPrice)
		level.Quantity = nuubot.Meta.RoundSize(
			slice.Div(one.Add(feeRate)).Div(sizingPrice),
		)
		if !level.Quantity.IsPositive() {
			return nil, fmt.Errorf("Grid level %d quantity rounds to zero", level.Level)
		}
		level.InitialNotional = level.Quantity.Mul(level.InitialEntryPrice)
		level.ReentryNotional = level.Quantity.Mul(level.ReentryPrice)
		var exitNotional = level.Quantity.Mul(level.ExitPrice)
		if level.InitialNotional.LessThan(minNotionalUSDC) ||
			level.ReentryNotional.LessThan(minNotionalUSDC) ||
			exitNotional.LessThan(minNotionalUSDC) {
			return nil, fmt.Errorf("Grid level %d is below minimum notional", level.Level)
		}
		if level.InitialNotional.Mul(one.Add(feeRate)).GreaterThan(slice) ||
			level.ReentryNotional.Mul(one.Add(feeRate)).GreaterThan(slice) {
			return nil, fmt.Errorf("Grid level %d exceeds its capital slice", level.Level)
		}
		level.InitialEntryCommission = level.InitialNotional.Mul(feeRate)
		level.ReentryCommission = level.ReentryNotional.Mul(feeRate)
		level.ExitCommission = exitNotional.Mul(feeRate)
		var initialGross = gridGross(
			spec.Side,
			level.InitialEntryPrice,
			level.ExitPrice,
			level.Quantity,
		)
		var reentryGross = gridGross(
			spec.Side,
			level.ReentryPrice,
			level.ExitPrice,
			level.Quantity,
		)
		level.InitialExpectedPnL = initialGross.
			Sub(level.InitialEntryCommission).
			Sub(level.ExitCommission)
		level.ReentryExpectedPnL = reentryGross.
			Sub(level.ReentryCommission).
			Sub(level.ExitCommission)
		if !level.InitialExpectedPnL.GreaterThan(spec.MinExpectedPnL) ||
			!level.ReentryExpectedPnL.GreaterThan(spec.MinExpectedPnL) {
			return nil, fmt.Errorf(
				"Grid level %d expected PnL does not exceed %s",
				level.Level,
				spec.MinExpectedPnL,
			)
		}
	}
	return levels, nil
}

func orderedCancellations(active []order.ActiveState) []order.ActiveState {
	var priority = func(role string) int {
		switch role {
		case order.TakeProfit:
			return 0
		case order.StopLoss:
			return 1
		case order.Entry:
			return 2
		default:
			return 3
		}
	}
	sort.SliceStable(active, func(left int, right int) bool {
		return priority(active[left].Role) < priority(active[right].Role)
	})
	return active
}

func gridGross(
	side string,
	entry decimal.Decimal,
	exit decimal.Decimal,
	quantity decimal.Decimal,
) decimal.Decimal {
	if side == Short {
		return entry.Sub(exit).Mul(quantity)
	}
	return exit.Sub(entry).Mul(quantity)
}

// Section 3 - Generic Helpers

func (e *gridExecutor) stopError(err error) error {
	e.status = Error
	return errors.Join(err, e.account.Stop())
}
