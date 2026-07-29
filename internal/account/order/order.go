// Package order owns one submitted request and its Venue lifecycle.
package order

import (
	"fmt"

	"github.com/shopspring/decimal"
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

// Update contains one trusted Order acknowledgement or Venue observation.
type Update struct {
	VenueOrderID uint64
	VenueStatus  string
	Status       Status
	RejectReason string
	UpdatedMS    uint64
	RawJSON      string
}

// Order preserves one submitted request and its latest Venue state.
type Order struct {
	SweepID           uint64
	BotID             uint64
	Venue             string
	Network           string
	Account           string
	LedgerID          uint64
	TradeID           uint64
	OrderID           uint64
	CLOID             string
	VenueOrderID      uint64
	VenueStatus       string
	CycleNumber       int
	Symbol            string
	Level             uint16
	Role              string
	Side              string
	Type              string
	TimeInForce       string
	SubmittedQuantity decimal.Decimal
	SubmittedPrice    *decimal.Decimal
	TriggerPrice      *decimal.Decimal
	ReduceOnly        bool
	SubmittedMS       uint64
	Status            Status
	RejectReason      string
	UpdatedMS         uint64
	FilledQuantity    decimal.Decimal
	FilledNotional    decimal.Decimal
	AverageFillPrice  decimal.Decimal
	RemainingQuantity decimal.Decimal
	Fees              decimal.Decimal
	FillCount         int
	PendingFeeCount   int
	LastFillMS        uint64
	RawJSON           string
}

// Section 1 - Program Flow

// New creates one Order in created state.
func New(input Order) (*Order, error) {
	// Step 1: validate Order request
	var err = input.Validate()
	if err != nil {
		return nil, err
	}

	// Step 2: retain independently owned prices
	if input.SubmittedPrice != nil {
		var submittedPrice = *input.SubmittedPrice
		input.SubmittedPrice = &submittedPrice
	}
	if input.TriggerPrice != nil {
		var triggerPrice = *input.TriggerPrice
		input.TriggerPrice = &triggerPrice
	}

	// Step 3: initialize created state
	input.VenueOrderID = 0
	input.VenueStatus = ""
	input.Status = Created
	input.RejectReason = ""
	input.UpdatedMS = 0
	input.FilledQuantity = decimal.Zero
	input.FilledNotional = decimal.Zero
	input.AverageFillPrice = decimal.Zero
	input.RemainingQuantity = input.SubmittedQuantity
	input.Fees = decimal.Zero
	input.FillCount = 0
	input.PendingFeeCount = 0
	input.LastFillMS = 0
	input.RawJSON = ""
	return &input, nil
}

// Update applies one trusted acknowledgement or Venue observation.
func (o *Order) Update(update Update) (bool, error) {
	// Step 1: validate Order update
	if !validStatus(update.Status) || update.UpdatedMS == 0 {
		return false, fmt.Errorf("update order: valid status and timestamp are required")
	}
	if o.VenueOrderID != 0 && update.VenueOrderID != 0 &&
		o.VenueOrderID != update.VenueOrderID {
		return false, fmt.Errorf("update order: changed venue order identity")
	}

	// Step 2: detect changed Order evidence
	var venueOrderID = o.VenueOrderID
	if venueOrderID == 0 {
		venueOrderID = update.VenueOrderID
	}
	var changed = o.VenueOrderID != venueOrderID ||
		o.VenueStatus != update.VenueStatus ||
		o.Status != update.Status ||
		o.RejectReason != update.RejectReason ||
		o.UpdatedMS != update.UpdatedMS ||
		o.RawJSON != update.RawJSON
	if !changed {
		return false, nil
	}

	// Step 3: publish changed Order evidence
	o.VenueOrderID = venueOrderID
	o.VenueStatus = update.VenueStatus
	o.Status = update.Status
	o.RejectReason = update.RejectReason
	o.UpdatedMS = update.UpdatedMS
	o.RawJSON = update.RawJSON
	return true, nil
}

// IsClosed reports whether Order execution and fee evidence is complete.
func (o *Order) IsClosed() bool {
	if o.PendingFeeCount > 0 {
		return false
	}
	switch o.Status {
	case Canceled, Rejected, Expired, Error:
		return true
	case Filled:
		return o.FilledQuantity.Equal(o.SubmittedQuantity)
	default:
		return false
	}
}

// Slippage returns absolute and percentage price slippage when both prices exist.
func (o *Order) Slippage() (decimal.Decimal, decimal.Decimal, bool) {
	if o.SubmittedPrice == nil || !o.SubmittedPrice.IsPositive() ||
		o.FillCount == 0 || !o.AverageFillPrice.IsPositive() {
		return decimal.Zero, decimal.Zero, false
	}
	var amount = o.AverageFillPrice.Sub(*o.SubmittedPrice)
	if o.Side == Sell {
		amount = o.SubmittedPrice.Sub(o.AverageFillPrice)
	}
	var percent = amount.Div(*o.SubmittedPrice).Mul(decimal.NewFromInt(100))
	return amount, percent, true
}

// Section 2 - Domain Helpers

// Section 2.1 - Validation

// Validate validates one complete submitted Order request.
func (o Order) Validate() error {
	if o.SweepID == 0 || o.BotID == 0 || o.LedgerID == 0 ||
		o.TradeID == 0 || o.OrderID == 0 {
		return fmt.Errorf("validate order: identity values must be positive")
	}
	if o.Venue == "" || o.Network == "" || o.Account == "" ||
		o.Symbol == "" || o.CLOID == "" {
		return fmt.Errorf("validate order: text identity must not be empty")
	}
	if o.CycleNumber <= 0 || o.SubmittedMS == 0 {
		return fmt.Errorf("validate order: cycle and timestamp must be positive")
	}
	if !validRole(o.Role) || !validSide(o.Side) ||
		!validType(o.Type) || !validTimeInForce(o.TimeInForce) {
		return fmt.Errorf("validate order: invalid role, side, type, or time in force")
	}
	if !o.SubmittedQuantity.IsPositive() {
		return fmt.Errorf("validate order: submitted quantity must be positive")
	}
	if o.SubmittedPrice != nil && !o.SubmittedPrice.IsPositive() {
		return fmt.Errorf("validate order: submitted price must be positive")
	}
	if o.TriggerPrice != nil && !o.TriggerPrice.IsPositive() {
		return fmt.Errorf("validate order: trigger price must be positive")
	}
	if o.Type == Trigger && o.TriggerPrice == nil {
		return fmt.Errorf("validate order: trigger order requires trigger price")
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

// Section 3 - Generic Helpers
