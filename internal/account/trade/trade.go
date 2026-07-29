// Package trade owns one trading intent and its derived PnL.
package trade

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/fill"
	"nuubot/internal/account/order"
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

// Trade preserves one trading intent and its calculated finance.
type Trade struct {
	SweepID           uint64
	BotID             uint64
	Venue             string
	Network           string
	Account           string
	LedgerID          uint64
	TradeID           uint64
	CycleNumber       int
	Symbol            string
	Status            Status
	Side              string
	OpenQuantity      decimal.Decimal
	AverageEntryPrice decimal.Decimal
	RealizedPnL       decimal.Decimal
	UnrealizedPnL     decimal.Decimal
	GrossPnL          decimal.Decimal
	Fees              decimal.Decimal
	NetPnL            decimal.Decimal
	OpenedMS          uint64
	ClosedMS          uint64
	UpdatedMS         uint64
}

// Section 1 - Program Flow

// New creates one pending Trade.
func New(input Trade) (*Trade, error) {
	// Step 1: validate Trade identity
	var err = input.Validate()
	if err != nil {
		return nil, err
	}

	// Step 2: initialize pending Trade
	input.Status = Pending
	input.Side = Flat
	input.OpenQuantity = decimal.Zero
	input.AverageEntryPrice = decimal.Zero
	input.RealizedPnL = decimal.Zero
	input.UnrealizedPnL = decimal.Zero
	input.GrossPnL = decimal.Zero
	input.Fees = decimal.Zero
	input.NetPnL = decimal.Zero
	input.OpenedMS = 0
	input.ClosedMS = 0
	input.UpdatedMS = 0
	return &input, nil
}

// Update recalculates Trade state from Ledger-supplied Orders and Fills.
func (t *Trade) Update(orders []*order.Order, fills []*fill.Fill) (bool, error) {
	// Step 1: order Fill evidence
	var ordered = append([]*fill.Fill(nil), fills...)
	sort.Slice(ordered, func(left int, right int) bool {
		if ordered[left].TimestampMS == ordered[right].TimestampMS {
			return ordered[left].VenueTID < ordered[right].VenueTID
		}
		return ordered[left].TimestampMS < ordered[right].TimestampMS
	})

	// Step 2: calculate Trade state and finance
	var calculated, err = calculate(*t, orders, ordered)
	if err != nil {
		return false, fmt.Errorf("update trade: %w", err)
	}
	calculated, err = calculateFinance(calculated, nil)
	if err != nil {
		return false, fmt.Errorf("update trade: %w", err)
	}
	if sameState(*t, calculated) {
		return false, nil
	}

	// Step 3: publish calculated Trade values
	t.publish(calculated)
	return true, nil
}

// UpdateMark updates unrealized, gross, and net PnL at one mark.
func (t *Trade) UpdateMark(markPrice *decimal.Decimal) (bool, error) {
	// Step 1: skip closed or flat Trade
	if t.IsClosed() || !t.OpenQuantity.IsPositive() {
		return false, nil
	}

	// Step 2: calculate marked finance
	var calculated, err = calculateFinance(*t, markPrice)
	if err != nil {
		return false, fmt.Errorf("update trade mark: %w", err)
	}
	if sameState(*t, calculated) {
		return false, nil
	}

	// Step 3: publish marked finance
	t.publish(calculated)
	return true, nil
}

// IsClosed reports whether the current Trade snapshot is closed.
func (t *Trade) IsClosed() bool {
	return t.Status == Closed || t.Status == Canceled || t.Status == Error
}

// Section 2 - Domain Helpers

// Section 2.1 - Validation

// Validate validates one complete Trade identity.
func (t Trade) Validate() error {
	if t.SweepID == 0 || t.BotID == 0 || t.LedgerID == 0 || t.TradeID == 0 {
		return fmt.Errorf("validate trade: identity values must be positive")
	}
	if t.Venue == "" || t.Network == "" || t.Account == "" || t.Symbol == "" {
		return fmt.Errorf("validate trade: text identity must not be empty")
	}
	if t.CycleNumber <= 0 {
		return fmt.Errorf("validate trade: cycle must be positive")
	}
	return nil
}

// Section 2.2 - Trade Calculations

func calculate(
	base Trade,
	orders []*order.Order,
	fills []*fill.Fill,
) (Trade, error) {
	var result = base
	result.Status = Pending
	result.Side = Flat
	result.OpenQuantity = decimal.Zero
	result.AverageEntryPrice = decimal.Zero
	result.RealizedPnL = decimal.Zero
	result.UnrealizedPnL = decimal.Zero
	result.GrossPnL = decimal.Zero
	result.Fees = decimal.Zero
	result.NetPnL = decimal.Zero
	result.OpenedMS = 0
	result.ClosedMS = 0
	result.UpdatedMS = 0

	var activeOrders int
	var activeCloseOrders int
	for _, current := range orders {
		if current == nil || current.TradeID != base.TradeID {
			return Trade{}, fmt.Errorf("order ownership mismatch")
		}
		if !current.IsClosed() {
			activeOrders++
			if current.ReduceOnly || isCloseRole(current.Role) {
				activeCloseOrders++
			}
		}
	}

	var signed = decimal.Zero
	var average = decimal.Zero
	for _, current := range fills {
		if current == nil || current.TradeID != base.TradeID {
			return Trade{}, fmt.Errorf("fill ownership mismatch")
		}
		var delta = current.Quantity
		if current.Side == order.Sell {
			delta = delta.Neg()
		}
		if current.HasFee() {
			result.Fees = result.Fees.Add(*current.Fee)
		}
		if result.OpenedMS == 0 || current.TimestampMS < result.OpenedMS {
			result.OpenedMS = current.TimestampMS
		}
		if current.TimestampMS > result.UpdatedMS {
			result.UpdatedMS = current.TimestampMS
		}
		if signed.IsZero() {
			signed = delta
			average = current.Price
			continue
		}
		if sameSign(signed, delta) {
			var total = signed.Abs().Add(current.Quantity)
			average = signed.Abs().Mul(average).
				Add(current.Quantity.Mul(current.Price)).
				Div(total)
			signed = signed.Add(delta)
			continue
		}
		if current.Quantity.GreaterThan(signed.Abs()) {
			return Trade{}, fmt.Errorf("fill reverses trade")
		}
		if signed.IsPositive() {
			result.RealizedPnL = result.RealizedPnL.Add(
				current.Price.Sub(average).Mul(current.Quantity),
			)
		} else {
			result.RealizedPnL = result.RealizedPnL.Add(
				average.Sub(current.Price).Mul(current.Quantity),
			)
		}
		signed = signed.Add(delta)
		if signed.IsZero() {
			average = decimal.Zero
			result.ClosedMS = current.TimestampMS
		}
	}

	result.OpenQuantity = signed.Abs()
	result.AverageEntryPrice = average
	if signed.IsPositive() {
		result.Side = Long
	} else if signed.IsNegative() {
		result.Side = Short
	}
	switch {
	case len(fills) == 0 && activeOrders > 0:
		result.Status = Pending
	case len(fills) == 0:
		result.Status = Canceled
	case result.OpenQuantity.IsZero() && activeOrders > 0:
		result.Status = Closing
	case result.OpenQuantity.IsZero():
		result.Status = Closed
	case activeCloseOrders > 0:
		result.Status = Closing
	default:
		result.Status = Open
	}
	return result, nil
}

func calculateFinance(current Trade, markPrice *decimal.Decimal) (Trade, error) {
	current.UnrealizedPnL = decimal.Zero
	if current.OpenQuantity.IsPositive() && markPrice != nil {
		if !markPrice.IsPositive() {
			return Trade{}, fmt.Errorf("positive mark price is required")
		}
		if current.Side == Long {
			current.UnrealizedPnL = markPrice.Sub(current.AverageEntryPrice).
				Mul(current.OpenQuantity)
		} else {
			current.UnrealizedPnL = current.AverageEntryPrice.Sub(*markPrice).
				Mul(current.OpenQuantity)
		}
	}
	current.GrossPnL = current.RealizedPnL.Add(current.UnrealizedPnL)
	current.NetPnL = current.GrossPnL.Sub(current.Fees)
	return current, nil
}

func (t *Trade) publish(current Trade) {
	t.Status = current.Status
	t.Side = current.Side
	t.OpenQuantity = current.OpenQuantity
	t.AverageEntryPrice = current.AverageEntryPrice
	t.RealizedPnL = current.RealizedPnL
	t.UnrealizedPnL = current.UnrealizedPnL
	t.GrossPnL = current.GrossPnL
	t.Fees = current.Fees
	t.NetPnL = current.NetPnL
	t.OpenedMS = current.OpenedMS
	t.ClosedMS = current.ClosedMS
	t.UpdatedMS = current.UpdatedMS
}

func isCloseRole(role string) bool {
	switch role {
	case order.TakeProfit, order.StopLoss, order.Exit, order.Cleanup, order.Stop:
		return true
	default:
		return false
	}
}

func sameState(left Trade, right Trade) bool {
	return left.Status == right.Status &&
		left.Side == right.Side &&
		left.OpenQuantity.Equal(right.OpenQuantity) &&
		left.AverageEntryPrice.Equal(right.AverageEntryPrice) &&
		left.RealizedPnL.Equal(right.RealizedPnL) &&
		left.UnrealizedPnL.Equal(right.UnrealizedPnL) &&
		left.GrossPnL.Equal(right.GrossPnL) &&
		left.Fees.Equal(right.Fees) &&
		left.NetPnL.Equal(right.NetPnL) &&
		left.OpenedMS == right.OpenedMS &&
		left.ClosedMS == right.ClosedMS &&
		left.UpdatedMS == right.UpdatedMS
}

// Section 3 - Generic Helpers

func sameSign(left decimal.Decimal, right decimal.Decimal) bool {
	return left.IsPositive() && right.IsPositive() ||
		left.IsNegative() && right.IsNegative()
}
