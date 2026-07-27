package fill

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	Buy  = "B"
	Sell = "A"
)

// Input contains one normalized execution observation.
type Input struct {
	LedgerID     uint64
	TradeID      uint64
	OrderID      uint64
	Account      string
	CycleNumber  int
	Symbol       string
	CLOID        string
	VenueOrderID uint64
	VenueTID     uint64
	Side         string
	Quantity     decimal.Decimal
	Price        decimal.Decimal
	TimestampMS  uint64
	Fee          *decimal.Decimal
	Liquidity    string
	Raw          string
}

// Fill preserves one admitted execution and optional later metadata.
type Fill struct {
	input     Input
	fee       decimal.Decimal
	hasFee    bool
	liquidity string
	raw       string
}

// Record contains one immutable Fill value.
type Record struct {
	LedgerID     uint64
	TradeID      uint64
	OrderID      uint64
	Account      string
	CycleNumber  int
	Symbol       string
	CLOID        string
	VenueOrderID uint64
	VenueTID     uint64
	Side         string
	Quantity     decimal.Decimal
	Price        decimal.Decimal
	TimestampMS  uint64
	Fee          decimal.Decimal
	HasFee       bool
	Liquidity    string
	Raw          string
}

// Section 1 - Program Flow

// New creates one Fill from normalized execution evidence.
func New(input Input) (*Fill, error) {
	// validate complete execution identity
	var err = validateInput(input)
	if err != nil {
		return nil, err
	}

	// keep immutable execution
	var created = &Fill{input: input}
	created.input.Fee = nil

	// keep available metadata
	if input.Fee != nil {
		created.fee = *input.Fee
		created.hasFee = true
	}
	created.liquidity = input.Liquidity
	created.raw = input.Raw
	return created, nil
}

// Enrich applies later metadata without changing execution identity.
func (f *Fill) Enrich(input Input) error {
	// reject changed execution
	var err = validateInput(input)
	if err != nil {
		return err
	}
	if !f.sameExecution(input) {
		return fmt.Errorf("enrich fill: changed execution for venue tid %d", input.VenueTID)
	}
	if input.Fee != nil && f.hasFee && !f.fee.Equal(*input.Fee) {
		return fmt.Errorf("enrich fill: changed fee for venue tid %d", input.VenueTID)
	}
	if input.Liquidity != "" && f.liquidity != "" && f.liquidity != input.Liquidity {
		return fmt.Errorf("enrich fill: changed liquidity for venue tid %d", input.VenueTID)
	}

	// accept later metadata
	if input.Fee != nil {
		f.fee = *input.Fee
		f.hasFee = true
	}
	if input.Liquidity != "" {
		f.liquidity = input.Liquidity
	}
	if input.Raw != "" {
		f.raw = input.Raw
	}
	return nil
}

// State returns one allocation-free Fill value.
func (f *Fill) State() Record {
	return Record{
		LedgerID:     f.input.LedgerID,
		TradeID:      f.input.TradeID,
		OrderID:      f.input.OrderID,
		Account:      f.input.Account,
		CycleNumber:  f.input.CycleNumber,
		Symbol:       f.input.Symbol,
		CLOID:        f.input.CLOID,
		VenueOrderID: f.input.VenueOrderID,
		VenueTID:     f.input.VenueTID,
		Side:         f.input.Side,
		Quantity:     f.input.Quantity,
		Price:        f.input.Price,
		TimestampMS:  f.input.TimestampMS,
		Fee:          f.fee,
		HasFee:       f.hasFee,
		Liquidity:    f.liquidity,
		Raw:          f.raw,
	}
}

// HasFee reports whether Venue fee evidence is complete.
func (f *Fill) HasFee() bool {
	return f.hasFee
}

// Clone returns one independently owned Fill.
func (f *Fill) Clone() *Fill {
	var clone = *f
	return &clone
}

// Section 2 - Domain Helpers

func (f *Fill) sameExecution(input Input) bool {
	return f.input.LedgerID == input.LedgerID &&
		f.input.TradeID == input.TradeID &&
		f.input.OrderID == input.OrderID &&
		f.input.Account == input.Account &&
		f.input.CycleNumber == input.CycleNumber &&
		f.input.Symbol == input.Symbol &&
		f.input.CLOID == input.CLOID &&
		f.input.VenueOrderID == input.VenueOrderID &&
		f.input.VenueTID == input.VenueTID &&
		f.input.Side == input.Side &&
		f.input.Quantity.Equal(input.Quantity) &&
		f.input.Price.Equal(input.Price) &&
		f.input.TimestampMS == input.TimestampMS
}

func validateInput(input Input) error {
	if input.LedgerID == 0 || input.TradeID == 0 || input.OrderID == 0 ||
		input.VenueOrderID == 0 || input.VenueTID == 0 {
		return fmt.Errorf("create fill: identity values must be positive")
	}
	if input.Account == "" || input.Symbol == "" || input.CLOID == "" {
		return fmt.Errorf("create fill: text identity must not be empty")
	}
	if input.CycleNumber <= 0 || input.TimestampMS == 0 {
		return fmt.Errorf("create fill: cycle and timestamp must be positive")
	}
	if input.Side != Buy && input.Side != Sell {
		return fmt.Errorf("create fill: unknown side %q", input.Side)
	}
	if !input.Quantity.IsPositive() || !input.Price.IsPositive() {
		return fmt.Errorf("create fill: quantity and price must be positive")
	}
	return nil
}

// Section 3 - Generic Helpers
