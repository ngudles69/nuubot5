package fill

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	Buy  = "B"
	Sell = "A"
)

// Fill preserves one admitted execution and optional later fee evidence.
type Fill struct {
	FillID       uint64
	SweepID      uint64
	BotID        uint64
	Venue        string
	Network      string
	Account      string
	LedgerID     uint64
	TradeID      uint64
	OrderID      uint64
	CLOID        string
	VenueOrderID uint64
	VenueTID     uint64
	CycleNumber  int
	Symbol       string
	Side         string
	Quantity     decimal.Decimal
	Price        decimal.Decimal
	TimestampMS  uint64
	Fee          *decimal.Decimal
	Liquidity    string
	RawJSON      string
}

// Section 1 - Program Flow

// New creates one Fill from normalized execution evidence.
func New(input Fill) (*Fill, error) {
	// Step 1: validate Fill evidence
	var err = input.Validate()
	if err != nil {
		return nil, err
	}

	// Step 2: retain independently owned fee evidence
	if input.Fee != nil {
		var fee = *input.Fee
		input.Fee = &fee
	}
	return &input, nil
}

// Update applies later fee evidence and the latest raw Fill payload.
func (f *Fill) Update(fee *decimal.Decimal, rawJSON string) (bool, error) {
	// Step 1: reject conflicting fee evidence
	if fee != nil && f.Fee != nil && !f.Fee.Equal(*fee) {
		return false, fmt.Errorf("update fill: changed fee for venue tid %d", f.VenueTID)
	}

	// Step 2: apply later evidence
	var changed bool
	if fee != nil && f.Fee == nil {
		var copied = *fee
		f.Fee = &copied
		changed = true
	}
	if rawJSON != "" && rawJSON != f.RawJSON {
		f.RawJSON = rawJSON
		changed = true
	}
	return changed, nil
}

// HasFee reports whether Venue fee evidence is complete.
func (f *Fill) HasFee() bool {
	return f.Fee != nil
}

// Section 2 - Domain Helpers

// Section 2.1 - Validation and Identity

// Validate validates one complete Fill.
func (f Fill) Validate() error {
	if f.FillID == 0 || f.SweepID == 0 || f.BotID == 0 || f.LedgerID == 0 ||
		f.TradeID == 0 || f.OrderID == 0 || f.VenueTID == 0 {
		return fmt.Errorf("validate fill: identity values must be positive")
	}
	if f.Venue == "" || f.Network == "" || f.Account == "" ||
		f.Symbol == "" {
		return fmt.Errorf("validate fill: text identity must not be empty")
	}
	if f.CLOID == "" && f.VenueOrderID == 0 {
		return fmt.Errorf("validate fill: Exchange Order identity is required")
	}
	if f.CycleNumber <= 0 || f.TimestampMS == 0 {
		return fmt.Errorf("validate fill: cycle and timestamp must be positive")
	}
	if f.Side != Buy && f.Side != Sell {
		return fmt.Errorf("validate fill: unknown side %q", f.Side)
	}
	if !f.Quantity.IsPositive() || !f.Price.IsPositive() {
		return fmt.Errorf("validate fill: quantity and price must be positive")
	}
	return nil
}

// Section 3 - Generic Helpers
