// Package order owns one submitted request and its Venue lifecycle.
package order

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/fill"
)

// Status identifies one canonical Order lifecycle state.
type Status string

const (
	Created         Status = "created"
	Submitted       Status = "submitted"
	Open            Status = "open"
	PartiallyFilled Status = "partially_filled"
	Filled          Status = "filled"
	Canceled        Status = "canceled"
	Rejected        Status = "rejected"
	Expired         Status = "expired"
	Error           Status = "error"
)

const (
	Entry      = "entry"
	TakeProfit = "tp"
	StopLoss   = "sl"
	Exit       = "exit"
	Cleanup    = "cleanup"
	Stop       = "stop"

	Buy  = "B"
	Sell = "A"

	Limit   = "limit"
	Trigger = "trigger"
	Market  = "market"

	GTC = "Gtc"
	IOC = "Ioc"
	ALO = "Alo"
)

// Input contains one admitted immutable Order request.
type Input struct {
	LedgerID          uint64
	TradeID           uint64
	OrderID           uint64
	Account           string
	CycleNumber       int
	Symbol            string
	BatchNo           uint16
	OrderPos          uint16
	CLOID             string
	Role              string
	Side              string
	Type              string
	TimeInForce       string
	RequestedQuantity decimal.Decimal
	RequestedPrice    *decimal.Decimal
	TriggerPrice      *decimal.Decimal
	ReduceOnly        bool
	TimestampMS       uint64
}

// VenueState contains one normalized Venue lifecycle observation.
type VenueState struct {
	VenueOrderID uint64
	Status       Status
	RejectReason string
	TimestampMS  uint64
	Raw          string
}

// Snapshot contains one immutable-by-contract Order value.
type Snapshot struct {
	Input
	VenueOrderID      uint64
	Status            Status
	Active            bool
	RejectReason      string
	UpdatedMS         uint64
	LastFillMS        uint64
	FilledQuantity    decimal.Decimal
	RemainingQuantity decimal.Decimal
	AverageFillPrice  decimal.Decimal
	HasAveragePrice   bool
	Fees              decimal.Decimal
	Raw               string
	Fills             []fill.Snapshot
}

// Order preserves one immutable request and its advancing Venue evidence.
type Order struct {
	input             Input
	venueOrderID      uint64
	status            Status
	active            bool
	rejectReason      string
	updatedMS         uint64
	lastFillMS        uint64
	filledQuantity    decimal.Decimal
	remainingQuantity decimal.Decimal
	averageFillPrice  decimal.Decimal
	hasAveragePrice   bool
	fees              decimal.Decimal
	raw               string
	fills             map[uint64]*fill.Fill
}

// Section 1 - Program Flow

// New creates one Order in created state.
func New(input Input) (*Order, error) {
	// validate order request
	var err = validateInput(input)
	if err != nil {
		return nil, err
	}

	// initialize created state
	var created = &Order{
		input:             copyInput(input),
		status:            Created,
		active:            true,
		remainingQuantity: input.RequestedQuantity,
		fills:             make(map[uint64]*fill.Fill),
	}
	return created, nil
}

// RecordSubmit preserves one complete submission acknowledgement.
func (o *Order) RecordSubmit(venueOrderID uint64, rejectReason string, raw string) error {
	// preserve venue identity
	if venueOrderID != 0 {
		if o.venueOrderID != 0 && o.venueOrderID != venueOrderID {
			return fmt.Errorf("record order submit: changed venue order identity")
		}
		o.venueOrderID = venueOrderID
	}

	// record acknowledgement
	if o.status != Created && o.status != Submitted {
		return fmt.Errorf("record order submit: order is already %s", o.status)
	}
	o.status = Submitted
	o.rejectReason = rejectReason
	if raw != "" {
		o.raw = raw
	}
	return nil
}

// ApplyVenueState advances one Order from canonical Venue evidence.
func (o *Order) ApplyVenueState(state VenueState) error {
	// reject invalid transition
	if state.TimestampMS == 0 || !validStatus(state.Status) {
		return fmt.Errorf("apply order state: invalid status or timestamp")
	}
	if state.TimestampMS < o.updatedMS {
		return nil
	}
	if isTerminal(o.status) && state.Status != o.status {
		return fmt.Errorf("apply order state: terminal order cannot become %s", state.Status)
	}
	if !transitionAllowed(o.status, state.Status) {
		return fmt.Errorf("apply order state: invalid transition %s to %s", o.status, state.Status)
	}

	// preserve Venue identity
	if state.VenueOrderID != 0 {
		if o.venueOrderID != 0 && o.venueOrderID != state.VenueOrderID {
			return fmt.Errorf("apply order state: changed venue order identity")
		}
		o.venueOrderID = state.VenueOrderID
	}

	// advance lifecycle
	o.status = state.Status
	o.active = !isTerminal(state.Status)
	o.rejectReason = state.RejectReason
	o.updatedMS = state.TimestampMS

	// preserve raw evidence
	if state.Raw != "" {
		o.raw = state.Raw
	}
	return nil
}

// ApplyFill admits one owned execution and refreshes Fill totals.
func (o *Order) ApplyFill(input fill.Input) error {
	// validate Fill ownership
	if input.LedgerID != o.input.LedgerID ||
		input.TradeID != o.input.TradeID ||
		input.OrderID != o.input.OrderID ||
		input.Account != o.input.Account ||
		input.CycleNumber != o.input.CycleNumber ||
		input.Symbol != o.input.Symbol ||
		input.CLOID != o.input.CLOID ||
		input.Side != o.input.Side {
		return fmt.Errorf("apply order fill: execution ownership mismatch")
	}
	if o.venueOrderID != 0 && input.VenueOrderID != o.venueOrderID {
		return fmt.Errorf("apply order fill: venue order identity mismatch")
	}

	// add or enrich Fill
	var existing = o.fills[input.VenueTID]
	if existing == nil {
		var created *fill.Fill
		var err error
		created, err = fill.New(input)
		if err != nil {
			return fmt.Errorf("apply order fill: %w", err)
		}
		var total = o.filledQuantity.Add(input.Quantity)
		if total.GreaterThan(o.input.RequestedQuantity) {
			return fmt.Errorf("apply order fill: quantity exceeds request")
		}
		o.fills[input.VenueTID] = created
	} else {
		var err = existing.Enrich(input)
		if err != nil {
			return fmt.Errorf("apply order fill: %w", err)
		}
	}

	// refresh Fill totals
	o.refreshFills()
	return nil
}

// Snapshot returns one immutable-by-contract Order value.
func (o *Order) Snapshot() Snapshot {
	// return immutable Order values
	var fills = make([]fill.Snapshot, 0, len(o.fills))
	for _, execution := range o.fills {
		fills = append(fills, execution.Snapshot())
	}
	return Snapshot{
		Input:             copyInput(o.input),
		VenueOrderID:      o.venueOrderID,
		Status:            o.status,
		Active:            o.active,
		RejectReason:      o.rejectReason,
		UpdatedMS:         o.updatedMS,
		LastFillMS:        o.lastFillMS,
		FilledQuantity:    o.filledQuantity,
		RemainingQuantity: o.remainingQuantity,
		AverageFillPrice:  o.averageFillPrice,
		HasAveragePrice:   o.hasAveragePrice,
		Fees:              o.fees,
		Raw:               o.raw,
		Fills:             fills,
	}
}

// Clone returns one independently owned Order.
func (o *Order) Clone() *Order {
	var clone = *o
	clone.input = copyInput(o.input)
	clone.fills = make(map[uint64]*fill.Fill, len(o.fills))
	for id, execution := range o.fills {
		clone.fills[id] = execution.Clone()
	}
	return &clone
}

// Section 2 - Domain Helpers

func (o *Order) refreshFills() {
	var quantity = decimal.Zero
	var notional = decimal.Zero
	var fees = decimal.Zero
	var lastMS uint64
	for _, execution := range o.fills {
		var snapshot = execution.Snapshot()
		quantity = quantity.Add(snapshot.Quantity)
		notional = notional.Add(snapshot.Quantity.Mul(snapshot.Price))
		if snapshot.HasFee {
			fees = fees.Add(snapshot.Fee)
		}
		if snapshot.TimestampMS > lastMS {
			lastMS = snapshot.TimestampMS
		}
	}
	o.filledQuantity = quantity
	o.remainingQuantity = o.input.RequestedQuantity.Sub(quantity)
	o.fees = fees
	o.lastFillMS = lastMS
	o.hasAveragePrice = quantity.IsPositive()
	if o.hasAveragePrice {
		o.averageFillPrice = notional.Div(quantity)
	}
	if quantity.Equal(o.input.RequestedQuantity) {
		o.status = Filled
		o.active = false
	} else if quantity.IsPositive() && !isTerminal(o.status) {
		o.status = PartiallyFilled
		o.active = true
	}
}

func validateInput(input Input) error {
	if input.LedgerID == 0 || input.TradeID == 0 || input.OrderID == 0 ||
		input.Account == "" || input.CycleNumber <= 0 ||
		input.Symbol == "" || input.CLOID == "" || input.TimestampMS == 0 {
		return fmt.Errorf("create order: complete identity is required")
	}
	if input.BatchNo == 0 || input.BatchNo > 1000 ||
		input.OrderPos == 0 || input.OrderPos > 1000 {
		return fmt.Errorf("create order: batch and position must be from 1 to 1000")
	}
	if !validRole(input.Role) || !validSide(input.Side) ||
		!validType(input.Type) || !validTimeInForce(input.TimeInForce) {
		return fmt.Errorf("create order: invalid role, side, type, or time in force")
	}
	if !input.RequestedQuantity.IsPositive() {
		return fmt.Errorf("create order: requested quantity must be positive")
	}
	if input.RequestedPrice != nil && !input.RequestedPrice.IsPositive() {
		return fmt.Errorf("create order: requested price must be positive")
	}
	if input.TriggerPrice != nil && !input.TriggerPrice.IsPositive() {
		return fmt.Errorf("create order: trigger price must be positive")
	}
	if input.Type == Trigger && input.TriggerPrice == nil {
		return fmt.Errorf("create order: trigger order requires trigger price")
	}
	return nil
}

func validRole(value string) bool {
	switch value {
	case Entry, TakeProfit, StopLoss, Exit, Cleanup, Stop:
		return true
	default:
		return false
	}
}

func validSide(value string) bool {
	return value == Buy || value == Sell
}

func validType(value string) bool {
	return value == Limit || value == Trigger || value == Market
}

func validTimeInForce(value string) bool {
	return value == GTC || value == IOC || value == ALO
}

func validStatus(value Status) bool {
	switch value {
	case Created, Submitted, Open, PartiallyFilled, Filled,
		Canceled, Rejected, Expired, Error:
		return true
	default:
		return false
	}
}

func isTerminal(value Status) bool {
	switch value {
	case Filled, Canceled, Rejected, Expired, Error:
		return true
	default:
		return false
	}
}

func transitionAllowed(from Status, to Status) bool {
	if from == to {
		return true
	}
	switch from {
	case Created:
		return to == Submitted || to == Open || to == PartiallyFilled ||
			to == Filled || to == Rejected || to == Canceled || to == Error
	case Submitted:
		return to == Open || to == PartiallyFilled || to == Filled ||
			to == Canceled || to == Rejected || to == Expired || to == Error
	case Open:
		return to == PartiallyFilled || to == Filled ||
			to == Canceled || to == Expired || to == Error
	case PartiallyFilled:
		return to == Filled || to == Canceled || to == Expired || to == Error
	default:
		return false
	}
}

// Section 3 - Generic Helpers

func copyInput(input Input) Input {
	var copied = input
	if input.RequestedPrice != nil {
		var value = *input.RequestedPrice
		copied.RequestedPrice = &value
	}
	if input.TriggerPrice != nil {
		var value = *input.TriggerPrice
		copied.TriggerPrice = &value
	}
	return copied
}
