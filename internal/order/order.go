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

// Record contains one flat Order database value.
type Record struct {
	Input
	VenueOrderID          uint64
	Status                Status
	Active                bool
	ReconciliationPending bool
	RejectReason          string
	UpdatedMS             uint64
	LastFillMS            uint64
	FilledQuantity        decimal.Decimal
	RemainingQuantity     decimal.Decimal
	AverageFillPrice      decimal.Decimal
	HasAveragePrice       bool
	Fees                  decimal.Decimal
	Raw                   string
	FillCount             int
}

// ReconState is the flat Order value used by reconciliation decisions.
type ReconState = Record

// ActiveState contains the fields required to reconcile or cancel an active Order.
type ActiveState struct {
	OrderID      uint64
	CLOID        string
	Role         string
	Status       Status
	VenueOrderID uint64
}

// Summary contains allocation-free reconciliation counts.
type Summary struct {
	Active                bool
	ReconciliationPending bool
	Fills                 int
	PendingFills          int
}

// Order preserves one immutable request and its advancing Venue evidence.
type Order struct {
	input                 Input
	venueOrderID          uint64
	status                Status
	active                bool
	reconciliationPending bool
	rejectReason          string
	updatedMS             uint64
	lastFillMS            uint64
	filledQuantity        decimal.Decimal
	remainingQuantity     decimal.Decimal
	averageFillPrice      decimal.Decimal
	hasAveragePrice       bool
	fees                  decimal.Decimal
	raw                   string
	fills                 map[uint64]*fill.Fill
	comparisonState       uint64
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
	if o.status != Created && o.status != Submitted {
		return fmt.Errorf("record order submit: order is already %s", o.status)
	}
	var changed = o.status != Submitted || o.rejectReason != rejectReason

	// preserve venue identity
	if venueOrderID != 0 {
		if o.venueOrderID != 0 && o.venueOrderID != venueOrderID {
			return fmt.Errorf("record order submit: changed venue order identity")
		}
		changed = changed || o.venueOrderID != venueOrderID
		o.venueOrderID = venueOrderID
	}

	// record acknowledgement
	o.status = Submitted
	o.rejectReason = rejectReason
	if raw != "" {
		changed = changed || o.raw != raw
		o.raw = raw
	}
	if changed {
		o.comparisonState++
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
	if (state.VenueOrderID == 0 || o.venueOrderID == state.VenueOrderID) &&
		o.status == state.Status &&
		o.rejectReason == state.RejectReason &&
		o.updatedMS == state.TimestampMS &&
		(state.Raw == "" || o.raw == state.Raw) {
		return nil
	}
	var changed = o.status != state.Status ||
		o.active != !isTerminal(state.Status) ||
		o.reconciliationPending ||
		o.rejectReason != state.RejectReason ||
		o.updatedMS != state.TimestampMS

	// preserve Venue identity
	if state.VenueOrderID != 0 {
		if o.venueOrderID != 0 && o.venueOrderID != state.VenueOrderID {
			return fmt.Errorf("apply order state: changed venue order identity")
		}
		changed = changed || o.venueOrderID != state.VenueOrderID
		o.venueOrderID = state.VenueOrderID
	}

	// advance lifecycle
	o.status = state.Status
	o.active = !isTerminal(state.Status)
	o.reconciliationPending = false
	o.rejectReason = state.RejectReason
	o.updatedMS = state.TimestampMS

	// preserve raw evidence
	if state.Raw != "" {
		changed = changed || o.raw != state.Raw
		o.raw = state.Raw
	}
	if changed {
		o.comparisonState++
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
	var changed = existing == nil
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
		var previous = existing.State()
		changed = (!previous.HasFee && input.Fee != nil) ||
			(previous.Liquidity == "" && input.Liquidity != "") ||
			(input.Raw != "" && previous.Raw != input.Raw)
		var err = existing.Enrich(input)
		if err != nil {
			return fmt.Errorf("apply order fill: %w", err)
		}
	}

	// refresh Fill totals
	o.refreshFills()
	if changed {
		o.comparisonState++
	}
	return nil
}

// RefreshRecon prevents Venue-filled evidence from becoming locally complete before its Fills and fees.
func (o *Order) RefreshRecon() {
	if o.status != Filled {
		return
	}
	if len(o.fills) == 0 {
		var changed = !o.active || !o.reconciliationPending
		o.active = true
		o.reconciliationPending = true
		if changed {
			o.comparisonState++
		}
		return
	}
	var status = o.status
	var active = o.active
	var pending = o.reconciliationPending
	o.refreshFills()
	if o.status != status || o.active != active || o.reconciliationPending != pending {
		o.comparisonState++
	}
}

// Record returns one flat Order database value.
func (o *Order) Record() Record {
	return o.ReconState()
}

// ReconState returns allocation-free Order identity and mutable state.
func (o *Order) ReconState() ReconState {
	return ReconState{
		Input:                 copyInput(o.input),
		VenueOrderID:          o.venueOrderID,
		Status:                o.status,
		Active:                o.active,
		ReconciliationPending: o.reconciliationPending,
		RejectReason:          o.rejectReason,
		UpdatedMS:             o.updatedMS,
		LastFillMS:            o.lastFillMS,
		FilledQuantity:        o.filledQuantity,
		RemainingQuantity:     o.remainingQuantity,
		AverageFillPrice:      o.averageFillPrice,
		HasAveragePrice:       o.hasAveragePrice,
		Fees:                  o.fees,
		Raw:                   o.raw,
		FillCount:             len(o.fills),
	}
}

// ComparisonState returns the allocation-free Order mutation revision.
func (o *Order) ComparisonState() uint64 {
	return o.comparisonState
}

// FillIdentity returns allocation-free ownership for one Order Fill.
func (o *Order) FillIdentity() fill.Input {
	return fill.Input{
		LedgerID:     o.input.LedgerID,
		TradeID:      o.input.TradeID,
		OrderID:      o.input.OrderID,
		Account:      o.input.Account,
		CycleNumber:  o.input.CycleNumber,
		Symbol:       o.input.Symbol,
		CLOID:        o.input.CLOID,
		VenueOrderID: o.venueOrderID,
		Side:         o.input.Side,
	}
}

// ActiveState returns focused active Order evidence.
func (o *Order) ActiveState() ActiveState {
	return ActiveState{
		OrderID:      o.input.OrderID,
		CLOID:        o.input.CLOID,
		Role:         o.input.Role,
		Status:       o.status,
		VenueOrderID: o.venueOrderID,
	}
}

// Summary returns allocation-free reconciliation counts.
func (o *Order) Summary() Summary {
	var result = Summary{
		Active:                o.active,
		ReconciliationPending: o.reconciliationPending,
		Fills:                 len(o.fills),
	}
	for _, execution := range o.fills {
		if !execution.HasFee() {
			result.PendingFills++
		}
	}
	return result
}

// EachFill visits every owned Fill without allocating a collection.
func (o *Order) EachFill(visit func(*fill.Fill) error) error {
	for _, execution := range o.fills {
		var err = visit(execution)
		if err != nil {
			return err
		}
	}
	return nil
}

// Fill returns one owned Fill by Venue identity.
func (o *Order) Fill(venueTID uint64) (*fill.Fill, bool) {
	var execution, exists = o.fills[venueTID]
	return execution, exists
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
	var venueComplete = o.status == Filled
	var quantity = decimal.Zero
	var notional = decimal.Zero
	var fees = decimal.Zero
	var feesComplete = true
	var lastMS uint64
	for _, execution := range o.fills {
		var current = execution.State()
		quantity = quantity.Add(current.Quantity)
		notional = notional.Add(current.Quantity.Mul(current.Price))
		if current.HasFee {
			fees = fees.Add(current.Fee)
		} else {
			feesComplete = false
		}
		if current.TimestampMS > lastMS {
			lastMS = current.TimestampMS
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
	if quantity.Equal(o.input.RequestedQuantity) && feesComplete {
		o.status = Filled
		o.active = false
		o.reconciliationPending = false
	} else if quantity.IsPositive() {
		o.status = PartiallyFilled
		o.active = true
		o.reconciliationPending = !feesComplete || venueComplete
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
