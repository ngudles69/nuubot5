package fill

import (
	"testing"

	"github.com/shopspring/decimal"
)

// Section 1 - Program Flow

func TestFillEnrichesDuplicateEvidence(t *testing.T) {
	var input = Input{
		LedgerID:     1,
		TradeID:      2,
		OrderID:      3,
		Account:      "sim",
		CycleNumber:  4,
		Symbol:       "BTC",
		CLOID:        "0x00000000000000000000000000000001",
		VenueOrderID: 5,
		VenueTID:     6,
		Side:         Buy,
		Quantity:     decimal.RequireFromString("0.1"),
		Price:        decimal.RequireFromString("50000"),
		TimestampMS:  7,
	}
	var actual, err = New(input)
	if err != nil {
		t.Fatalf("create fill: %v", err)
	}
	var fee = decimal.RequireFromString("1.25")
	input.Fee = &fee
	input.Liquidity = "taker"
	input.Raw = `{"tid":6}`
	err = actual.Enrich(input)
	if err != nil {
		t.Fatalf("enrich fill: %v", err)
	}
	err = actual.Enrich(input)
	if err != nil {
		t.Fatalf("repeat fill evidence: %v", err)
	}
	var snapshot = actual.Snapshot()
	if !snapshot.HasFee || !snapshot.Fee.Equal(fee) {
		t.Fatalf("actual fee %s present=%t, expected %s", snapshot.Fee, snapshot.HasFee, fee)
	}
	if snapshot.Liquidity != "taker" || snapshot.Raw != input.Raw {
		t.Fatalf(
			"actual liquidity=%q raw=%q, expected liquidity=%q raw=%q",
			snapshot.Liquidity,
			snapshot.Raw,
			"taker",
			input.Raw,
		)
	}
}

func TestFillRejectsChangedExecution(t *testing.T) {
	var input = validInput()
	var actual, err = New(input)
	if err != nil {
		t.Fatalf("create fill: %v", err)
	}
	input.Price = decimal.RequireFromString("50001")
	err = actual.Enrich(input)
	if err == nil {
		t.Fatalf("actual error nil, expected changed execution rejection")
	}
}

// Section 2 - Domain Helpers

func validInput() Input {
	return Input{
		LedgerID:     1,
		TradeID:      2,
		OrderID:      3,
		Account:      "sim",
		CycleNumber:  4,
		Symbol:       "BTC",
		CLOID:        "0x00000000000000000000000000000001",
		VenueOrderID: 5,
		VenueTID:     6,
		Side:         Sell,
		Quantity:     decimal.RequireFromString("0.1"),
		Price:        decimal.RequireFromString("50000"),
		TimestampMS:  7,
	}
}

// Section 3 - Generic Helpers
