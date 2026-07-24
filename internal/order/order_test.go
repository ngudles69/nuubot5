package order

import (
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/fill"
)

// Section 1 - Program Flow

func TestOrderAggregatesIdempotentFills(t *testing.T) {
	var created, err = New(validInput())
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	var first = validFill(1, "0.4", "100")
	err = created.ApplyFill(first)
	if err != nil {
		t.Fatalf("apply first fill: %v", err)
	}
	err = created.ApplyFill(first)
	if err != nil {
		t.Fatalf("repeat first fill: %v", err)
	}
	err = created.ApplyFill(validFill(2, "0.6", "110"))
	if err != nil {
		t.Fatalf("apply second fill: %v", err)
	}
	var snapshot = created.Snapshot()
	if snapshot.Status != Filled || snapshot.Active {
		t.Fatalf("actual state status=%s active=%t", snapshot.Status, snapshot.Active)
	}
	if snapshot.FilledQuantity.String() != "1" ||
		snapshot.RemainingQuantity.String() != "0" ||
		snapshot.AverageFillPrice.String() != "106" {
		t.Fatalf(
			"unexpected totals filled=%s remaining=%s average=%s",
			snapshot.FilledQuantity,
			snapshot.RemainingQuantity,
			snapshot.AverageFillPrice,
		)
	}
}

func TestOrderRejectsTerminalReopen(t *testing.T) {
	var created, err = New(validInput())
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	err = created.ApplyVenueState(VenueState{
		VenueOrderID: 1,
		Status:       Canceled,
		TimestampMS:  10,
	})
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	err = created.ApplyVenueState(VenueState{
		VenueOrderID: 1,
		Status:       Open,
		TimestampMS:  11,
	})
	if err == nil {
		t.Fatal("actual error nil, expected terminal reopen rejection")
	}
}

// Section 2 - Domain Helpers

func validInput() Input {
	var price = decimal.RequireFromString("100")
	return Input{
		LedgerID:          1,
		TradeID:           2,
		OrderID:           3,
		Account:           "sim",
		CycleNumber:       4,
		Symbol:            "BTC",
		BatchNo:           1,
		OrderPos:          1,
		CLOID:             "0x00000000000000000000000000000001",
		Role:              Entry,
		Side:              Buy,
		Type:              Limit,
		TimeInForce:       IOC,
		RequestedQuantity: decimal.NewFromInt(1),
		RequestedPrice:    &price,
		TimestampMS:       5,
	}
}

func validFill(tid uint64, quantity string, price string) fill.Input {
	return fill.Input{
		LedgerID:     1,
		TradeID:      2,
		OrderID:      3,
		Account:      "sim",
		CycleNumber:  4,
		Symbol:       "BTC",
		CLOID:        "0x00000000000000000000000000000001",
		VenueOrderID: 1,
		VenueTID:     tid,
		Side:         fill.Buy,
		Quantity:     decimal.RequireFromString(quantity),
		Price:        decimal.RequireFromString(price),
		TimestampMS:  6 + tid,
	}
}

// Section 3 - Generic Helpers
