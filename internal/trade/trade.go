// Package trade owns one trading intent and its derived PnL.
package trade

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

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

// Snapshot contains one immutable-by-contract Trade value.
type Snapshot struct {
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
	Orders            []order.Snapshot
}

type metrics struct {
	status            Status
	side              string
	openQuantity      decimal.Decimal
	averageEntryPrice decimal.Decimal
	hasAveragePrice   bool
	realizedPnL       decimal.Decimal
	fees              decimal.Decimal
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
	var snapshot = created.Snapshot()
	if snapshot.LedgerID != t.input.LedgerID ||
		snapshot.TradeID != t.input.TradeID ||
		snapshot.Account != t.input.Account ||
		snapshot.CycleNumber != t.input.CycleNumber ||
		snapshot.Symbol != t.input.Symbol {
		return fmt.Errorf("add trade order: ownership mismatch")
	}
	if t.finalized {
		return fmt.Errorf("add trade order: terminal trade cannot change")
	}
	if _, exists := t.orders[snapshot.OrderID]; exists {
		return fmt.Errorf("add trade order: duplicate order id %d", snapshot.OrderID)
	}
	for _, existing := range t.orders {
		if existing.Snapshot().CLOID == snapshot.CLOID {
			return fmt.Errorf("add trade order: duplicate cloid %s", snapshot.CLOID)
		}
	}

	// attach Order
	t.orders[snapshot.OrderID] = created

	// refresh Trade
	var err = t.Refresh()
	if err != nil {
		delete(t.orders, snapshot.OrderID)
		return err
	}
	return nil
}

// Refresh derives Trade state and economics from owned Orders and Fills.
func (t *Trade) Refresh() error {
	// order Fill evidence
	var executions = make([]execution, 0)
	var activeOrders int
	var activeCloseOrders int
	for _, ownedOrder := range t.orders {
		var snapshot = ownedOrder.Snapshot()
		if snapshot.Active {
			activeOrders++
			if snapshot.ReduceOnly || isCloseRole(snapshot.Role) {
				activeCloseOrders++
			}
		}
		for _, ownedFill := range snapshot.Fills {
			var fee = decimal.Zero
			if ownedFill.HasFee {
				fee = ownedFill.Fee
			}
			executions = append(executions, execution{
				side:        ownedFill.Side,
				quantity:    ownedFill.Quantity,
				price:       ownedFill.Price,
				fee:         fee,
				timestampMS: ownedFill.TimestampMS,
				venueTID:    ownedFill.VenueTID,
			})
		}
	}
	sort.Slice(executions, func(left int, right int) bool {
		if executions[left].timestampMS == executions[right].timestampMS {
			return executions[left].venueTID < executions[right].venueTID
		}
		return executions[left].timestampMS < executions[right].timestampMS
	})

	// calculate exposure
	var calculated, err = calculate(executions, activeOrders, activeCloseOrders)
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

// Snapshot returns one immutable-by-contract marked Trade value.
func (t *Trade) Snapshot(markPrice *decimal.Decimal) (Snapshot, error) {
	return t.snapshot(markPrice, true)
}

// State returns one independently owned unmarked Trade state for persistence.
func (t *Trade) State() Snapshot {
	var snapshot, _ = t.snapshot(nil, false)
	return snapshot
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

// Orders returns independently owned Order snapshots.
func (t *Trade) Orders() []order.Snapshot {
	var snapshots = make([]order.Snapshot, 0, len(t.orders))
	for _, owned := range t.orders {
		snapshots = append(snapshots, owned.Snapshot())
	}
	return snapshots
}

// Section 2 - Domain Helpers

func (t *Trade) snapshot(
	markPrice *decimal.Decimal,
	requireMark bool,
) (Snapshot, error) {
	// mark open exposure
	var unrealized = decimal.Zero
	if t.metrics.openQuantity.IsPositive() {
		if requireMark && (markPrice == nil || !markPrice.IsPositive()) {
			return Snapshot{}, fmt.Errorf("snapshot trade: positive mark price is required")
		}
		if markPrice != nil && t.metrics.side == Long {
			unrealized = markPrice.Sub(t.metrics.averageEntryPrice).
				Mul(t.metrics.openQuantity)
		} else if markPrice != nil {
			unrealized = t.metrics.averageEntryPrice.Sub(*markPrice).
				Mul(t.metrics.openQuantity)
		}
	}
	var gross = t.metrics.realizedPnL.Add(unrealized)

	// return immutable Trade values
	var orders = make([]order.Snapshot, 0, len(t.orders))
	for _, ownedOrder := range t.orders {
		orders = append(orders, ownedOrder.Snapshot())
	}
	sort.Slice(orders, func(left int, right int) bool {
		return orders[left].OrderID < orders[right].OrderID
	})
	return Snapshot{
		Input:             t.input,
		Status:            t.metrics.status,
		Side:              t.metrics.side,
		OpenQuantity:      t.metrics.openQuantity,
		AverageEntryPrice: t.metrics.averageEntryPrice,
		HasAveragePrice:   t.metrics.hasAveragePrice,
		RealizedPnL:       t.metrics.realizedPnL,
		UnrealizedPnL:     unrealized,
		GrossPnL:          gross,
		Fees:              t.metrics.fees,
		NetPnL:            gross.Sub(t.metrics.fees),
		OpenedMS:          t.metrics.openedMS,
		ClosedMS:          t.metrics.closedMS,
		UpdatedMS:         t.metrics.updatedMS,
		Orders:            orders,
	}, nil
}

func calculate(
	executions []execution,
	activeOrders int,
	activeCloseOrders int,
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
	case result.openQuantity.IsZero() && activeOrders > 0:
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

func isCloseRole(role string) bool {
	switch role {
	case order.TakeProfit, order.StopLoss, order.Exit,
		order.Close, order.Cleanup, order.Stop:
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
		left.fees.Equal(right.fees) &&
		left.openedMS == right.openedMS &&
		left.closedMS == right.closedMS &&
		left.updatedMS == right.updatedMS
}

// Section 3 - Generic Helpers

func sameSign(left decimal.Decimal, right decimal.Decimal) bool {
	return left.IsPositive() && right.IsPositive() ||
		left.IsNegative() && right.IsNegative()
}
