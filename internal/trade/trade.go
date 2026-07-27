// Package trade owns one trading intent and its derived PnL.
package trade

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"nuubot/internal/fill"
	"nuubot/internal/order"
)

// Status identifies one canonical Trade lifecycle state.
type Status string

const (
	Pending  Status = "pending"
	Open     Status = "open"
	Closing  Status = "closing"
	Closed   Status = "closed"
	Canceled Status = "canceled"
	Error    Status = "error"
)

const (
	Long  = "long"
	Short = "short"
	Flat  = "flat"
)

// Input contains one admitted immutable Trade identity.
type Input struct {
	LedgerID    uint64
	TradeID     uint64
	TradeNo     uint32
	Account     string
	CycleNumber int
	Symbol      string
}

// Record contains one flat Trade database value.
type Record struct {
	Input
	Status            Status
	Side              string
	OpenQuantity      decimal.Decimal
	AverageEntryPrice decimal.Decimal
	HasAveragePrice   bool
	RealizedPnL       decimal.Decimal
	UnrealizedPnL     decimal.Decimal
	GrossPnL          decimal.Decimal
	Fees              decimal.Decimal
	NetPnL            decimal.Decimal
	OpenedMS          uint64
	ClosedMS          uint64
	UpdatedMS         uint64
}

// ReconState is the flat Trade value used by reconciliation decisions.
type ReconState = Record

// Summary contains stored finance and allocation-free reconciliation counts.
type Summary struct {
	Status        Status
	RealizedPnL   decimal.Decimal
	UnrealizedPnL decimal.Decimal
	GrossPnL      decimal.Decimal
	Fees          decimal.Decimal
	NetPnL        decimal.Decimal
	ActiveOrders  int
	Fills         int
	PendingOrders int
	PendingFills  int
}

type metrics struct {
	status            Status
	side              string
	openQuantity      decimal.Decimal
	averageEntryPrice decimal.Decimal
	hasAveragePrice   bool
	realizedPnL       decimal.Decimal
	unrealizedPnL     decimal.Decimal
	grossPnL          decimal.Decimal
	fees              decimal.Decimal
	netPnL            decimal.Decimal
	openedMS          uint64
	closedMS          uint64
	updatedMS         uint64
}

type execution struct {
	side        string
	quantity    decimal.Decimal
	price       decimal.Decimal
	fee         decimal.Decimal
	timestampMS uint64
	venueTID    uint64
}

// Trade owns one coherent set of Orders.
type Trade struct {
	input     Input
	metrics   metrics
	orders    map[uint64]*order.Order
	finalized bool
}

// Section 1 - Program Flow

// New creates one pending Trade.
func New(input Input) (*Trade, error) {
	// validate Trade identity
	if input.LedgerID == 0 || input.TradeID == 0 ||
		input.TradeNo == 0 || input.TradeNo > 0x001fffff ||
		input.Account == "" || input.CycleNumber <= 0 || input.Symbol == "" {
		return nil, fmt.Errorf("create trade: complete identity is required")
	}

	// initialize pending Trade
	return &Trade{
		input:   input,
		metrics: metrics{status: Pending, side: Flat},
		orders:  make(map[uint64]*order.Order),
	}, nil
}

// AddOrder attaches one owned Order and refreshes the Trade.
func (t *Trade) AddOrder(created *order.Order) error {
	// validate Order ownership
	if created == nil {
		return fmt.Errorf("add trade order: order is required")
	}
	var state = created.ReconState()
	if state.LedgerID != t.input.LedgerID ||
		state.TradeID != t.input.TradeID ||
		state.Account != t.input.Account ||
		state.CycleNumber != t.input.CycleNumber ||
		state.Symbol != t.input.Symbol {
		return fmt.Errorf("add trade order: ownership mismatch")
	}
	if t.finalized {
		return fmt.Errorf("add trade order: terminal trade cannot change")
	}
	if _, exists := t.orders[state.OrderID]; exists {
		return fmt.Errorf("add trade order: duplicate order id %d", state.OrderID)
	}
	for _, existing := range t.orders {
		if existing.ReconState().CLOID == state.CLOID {
			return fmt.Errorf("add trade order: duplicate cloid %s", state.CLOID)
		}
	}

	// attach Order
	t.orders[state.OrderID] = created

	// refresh Trade
	var err = t.Refresh()
	if err != nil {
		delete(t.orders, state.OrderID)
		return err
	}
	return nil
}

// Refresh derives Trade state and economics from owned Orders and Fills.
func (t *Trade) Refresh() error {
	// order Fill evidence
	var executions, activeOrders, activeCloseOrders, pendingFills = t.executions()

	// calculate exposure
	var calculated, err = calculate(executions, activeOrders, activeCloseOrders, pendingFills)
	if err != nil {
		return fmt.Errorf("refresh trade: %w", err)
	}
	calculated, err = calculateFinance(calculated, nil)
	if err != nil {
		return fmt.Errorf("refresh trade: %w", err)
	}
	if t.finalized && !sameMetrics(calculated, t.metrics) {
		return fmt.Errorf("refresh trade: terminal trade values changed")
	}

	// derive status
	t.metrics = calculated
	t.finalized = isTerminal(calculated.status)
	return nil
}

// RefreshRecon derives and stores canonical reconciliation finance at one mark.
func (t *Trade) RefreshRecon(markPrice *decimal.Decimal) error {
	var executions, activeOrders, activeCloseOrders, pendingFills = t.executions()
	var calculated, err = calculate(executions, activeOrders, activeCloseOrders, pendingFills)
	if err != nil {
		return fmt.Errorf("refresh trade recon: %w", err)
	}
	calculated, err = calculateFinance(calculated, markPrice)
	if err != nil {
		return fmt.Errorf("refresh trade recon: %w", err)
	}
	if t.finalized && !sameMetrics(calculated, t.metrics) {
		return fmt.Errorf("refresh trade recon: terminal trade values changed")
	}
	t.metrics = calculated
	t.finalized = isTerminal(calculated.status)
	return nil
}

// RefreshMark updates stored finance from existing exposure without reading Orders or Fills.
func (t *Trade) RefreshMark(markPrice *decimal.Decimal) error {
	if t.finalized || !t.metrics.openQuantity.IsPositive() {
		return nil
	}
	var calculated, err = calculateFinance(t.metrics, markPrice)
	if err != nil {
		return fmt.Errorf("refresh trade recon: %w", err)
	}
	t.metrics = calculated
	return nil
}

// MarkedRecord returns one flat Trade value at the supplied mark.
func (t *Trade) MarkedRecord(markPrice *decimal.Decimal) (Record, error) {
	return t.record(markPrice, true)
}

// Record returns one flat stored Trade value.
func (t *Trade) Record() Record {
	var current, _ = t.record(nil, false)
	return current
}

// ReconState returns allocation-free Trade identity and mutable state.
func (t *Trade) ReconState() ReconState {
	return ReconState{
		Input:             t.input,
		Status:            t.metrics.status,
		Side:              t.metrics.side,
		OpenQuantity:      t.metrics.openQuantity,
		AverageEntryPrice: t.metrics.averageEntryPrice,
		HasAveragePrice:   t.metrics.hasAveragePrice,
		RealizedPnL:       t.metrics.realizedPnL,
		UnrealizedPnL:     t.metrics.unrealizedPnL,
		GrossPnL:          t.metrics.grossPnL,
		Fees:              t.metrics.fees,
		NetPnL:            t.metrics.netPnL,
		OpenedMS:          t.metrics.openedMS,
		ClosedMS:          t.metrics.closedMS,
		UpdatedMS:         t.metrics.updatedMS,
	}
}

// Summary returns stored finance and allocation-free reconciliation counts.
func (t *Trade) Summary() Summary {
	var result = Summary{
		Status:        t.metrics.status,
		RealizedPnL:   t.metrics.realizedPnL,
		UnrealizedPnL: t.metrics.unrealizedPnL,
		GrossPnL:      t.metrics.grossPnL,
		Fees:          t.metrics.fees,
		NetPnL:        t.metrics.netPnL,
	}
	for _, owned := range t.orders {
		var current = owned.Summary()
		if current.Active {
			result.ActiveOrders++
		}
		result.Fills += current.Fills
		result.PendingFills += current.PendingFills
		if current.ReconciliationPending {
			result.PendingOrders++
		}
	}
	return result
}

// Clone returns one independently owned Trade.
func (t *Trade) Clone() *Trade {
	var clone = *t
	clone.orders = make(map[uint64]*order.Order, len(t.orders))
	for id, ownedOrder := range t.orders {
		clone.orders[id] = ownedOrder.Clone()
	}
	return &clone
}

// Order returns one owned mutable Order to the owning Ledger.
func (t *Trade) Order(orderID uint64) (*order.Order, bool) {
	var owned, ok = t.orders[orderID]
	return owned, ok
}

// EachOrder visits every owned Order without allocating a collection.
func (t *Trade) EachOrder(visit func(*order.Order) error) error {
	for _, owned := range t.orders {
		var err = visit(owned)
		if err != nil {
			return err
		}
	}
	return nil
}

// Section 2 - Domain Helpers

func (t *Trade) record(
	markPrice *decimal.Decimal,
	requireMark bool,
) (Record, error) {
	var finance = t.metrics
	if requireMark {
		if finance.openQuantity.IsPositive() && (markPrice == nil || !markPrice.IsPositive()) {
			return Record{}, fmt.Errorf("read trade record: positive mark price is required")
		}
		var err error
		finance, err = calculateFinance(finance, markPrice)
		if err != nil {
			return Record{}, fmt.Errorf("read trade record: %w", err)
		}
	}
	return Record{
		Input:             t.input,
		Status:            t.metrics.status,
		Side:              t.metrics.side,
		OpenQuantity:      t.metrics.openQuantity,
		AverageEntryPrice: t.metrics.averageEntryPrice,
		HasAveragePrice:   t.metrics.hasAveragePrice,
		RealizedPnL:       finance.realizedPnL,
		UnrealizedPnL:     finance.unrealizedPnL,
		GrossPnL:          finance.grossPnL,
		Fees:              finance.fees,
		NetPnL:            finance.netPnL,
		OpenedMS:          t.metrics.openedMS,
		ClosedMS:          t.metrics.closedMS,
		UpdatedMS:         t.metrics.updatedMS,
	}, nil
}

func (t *Trade) executions() ([]execution, int, int, int) {
	var executions = make([]execution, 0)
	var activeOrders int
	var activeCloseOrders int
	var pendingFills int
	for _, ownedOrder := range t.orders {
		var state = ownedOrder.ReconState()
		if state.Active {
			activeOrders++
			if state.ReduceOnly || isCloseRole(state.Role) {
				activeCloseOrders++
			}
		}
		ownedOrder.EachFill(func(ownedFill *fill.Fill) error {
			var current = ownedFill.State()
			var fee = decimal.Zero
			if current.HasFee {
				fee = current.Fee
			} else {
				pendingFills++
			}
			executions = append(executions, execution{
				side:        current.Side,
				quantity:    current.Quantity,
				price:       current.Price,
				fee:         fee,
				timestampMS: current.TimestampMS,
				venueTID:    current.VenueTID,
			})
			return nil
		})
	}
	sort.Slice(executions, func(left int, right int) bool {
		if executions[left].timestampMS == executions[right].timestampMS {
			return executions[left].venueTID < executions[right].venueTID
		}
		return executions[left].timestampMS < executions[right].timestampMS
	})
	return executions, activeOrders, activeCloseOrders, pendingFills
}

func calculate(
	executions []execution,
	activeOrders int,
	activeCloseOrders int,
	pendingFills int,
) (metrics, error) {
	var result = metrics{status: Pending, side: Flat}
	var signed = decimal.Zero
	var average = decimal.Zero
	for _, current := range executions {
		var delta = current.quantity
		if current.side == order.Sell {
			delta = delta.Neg()
		}
		result.fees = result.fees.Add(current.fee)
		if result.openedMS == 0 || current.timestampMS < result.openedMS {
			result.openedMS = current.timestampMS
		}
		if current.timestampMS > result.updatedMS {
			result.updatedMS = current.timestampMS
		}
		if signed.IsZero() {
			signed = delta
			average = current.price
			continue
		}
		if sameSign(signed, delta) {
			var total = signed.Abs().Add(current.quantity)
			average = signed.Abs().Mul(average).
				Add(current.quantity.Mul(current.price)).
				Div(total)
			signed = signed.Add(delta)
			continue
		}
		if current.quantity.GreaterThan(signed.Abs()) {
			return metrics{}, fmt.Errorf("fill reverses trade")
		}
		if signed.IsPositive() {
			result.realizedPnL = result.realizedPnL.Add(
				current.price.Sub(average).Mul(current.quantity),
			)
		} else {
			result.realizedPnL = result.realizedPnL.Add(
				average.Sub(current.price).Mul(current.quantity),
			)
		}
		signed = signed.Add(delta)
		if signed.IsZero() {
			average = decimal.Zero
			result.closedMS = current.timestampMS
		}
	}
	result.openQuantity = signed.Abs()
	result.averageEntryPrice = average
	result.hasAveragePrice = result.openQuantity.IsPositive()
	if signed.IsPositive() {
		result.side = Long
	} else if signed.IsNegative() {
		result.side = Short
	}
	switch {
	case len(executions) == 0 && activeOrders > 0:
		result.status = Pending
	case len(executions) == 0:
		result.status = Canceled
	case result.openQuantity.IsZero() && (activeOrders > 0 || pendingFills > 0):
		result.status = Closing
	case result.openQuantity.IsZero():
		result.status = Closed
	case activeCloseOrders > 0:
		result.status = Closing
	default:
		result.status = Open
	}
	return result, nil
}

func calculateFinance(current metrics, markPrice *decimal.Decimal) (metrics, error) {
	current.unrealizedPnL = decimal.Zero
	if current.openQuantity.IsPositive() && markPrice != nil {
		if !markPrice.IsPositive() {
			return metrics{}, fmt.Errorf("positive mark price is required")
		}
		if current.side == Long {
			current.unrealizedPnL = markPrice.Sub(current.averageEntryPrice).
				Mul(current.openQuantity)
		} else {
			current.unrealizedPnL = current.averageEntryPrice.Sub(*markPrice).
				Mul(current.openQuantity)
		}
	}
	current.grossPnL = current.realizedPnL.Add(current.unrealizedPnL)
	current.netPnL = current.grossPnL.Sub(current.fees)
	return current, nil
}

func isCloseRole(role string) bool {
	switch role {
	case order.TakeProfit, order.StopLoss, order.Exit,
		order.Cleanup, order.Stop:
		return true
	default:
		return false
	}
}

func isTerminal(status Status) bool {
	return status == Closed || status == Canceled || status == Error
}

func sameMetrics(left metrics, right metrics) bool {
	return left.status == right.status &&
		left.side == right.side &&
		left.openQuantity.Equal(right.openQuantity) &&
		left.averageEntryPrice.Equal(right.averageEntryPrice) &&
		left.hasAveragePrice == right.hasAveragePrice &&
		left.realizedPnL.Equal(right.realizedPnL) &&
		left.unrealizedPnL.Equal(right.unrealizedPnL) &&
		left.grossPnL.Equal(right.grossPnL) &&
		left.fees.Equal(right.fees) &&
		left.netPnL.Equal(right.netPnL) &&
		left.openedMS == right.openedMS &&
		left.closedMS == right.closedMS &&
		left.updatedMS == right.updatedMS
}

// Section 3 - Generic Helpers

func sameSign(left decimal.Decimal, right decimal.Decimal) bool {
	return left.IsPositive() && right.IsPositive() ||
		left.IsNegative() && right.IsNegative()
}
