package order

import (
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/fill"
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
	var record = created.Record()
	if record.Status != Filled || record.Active {
		t.Fatalf("actual state status=%s active=%t", record.Status, record.Active)
	}
	if record.FilledQuantity.String() != "1" ||
		record.RemainingQuantity.String() != "0" ||
		record.AverageFillPrice.String() != "106" {
		t.Fatalf(
			"unexpected totals filled=%s remaining=%s average=%s",
			record.FilledQuantity,
			record.RemainingQuantity,
			record.AverageFillPrice,
		)
	}
}

func TestOrderKeepsFilledAcknowledgementReconciliationPending(t *testing.T) {
	var created, err = New(validInput())
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err = created.ApplyVenueState(VenueState{
		VenueOrderID: 1, Status: Filled, TimestampMS: 10,
	}); err != nil {
		t.Fatalf("apply filled acknowledgement: %v", err)
	}
	created.RefreshRecon()
	var pending = created.Record()
	if pending.Status != Filled || !pending.Active || !pending.ReconciliationPending {
		t.Fatalf("filled acknowledgement appeared locally complete: %+v", pending)
	}
}

func TestOrderWaitsForCompleteFillFees(t *testing.T) {
	var created, err = New(validInput())
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	var input = validFill(1, "1", "100")
	input.Fee = nil
	if err = created.ApplyFill(input); err != nil {
		t.Fatalf("apply fee-incomplete Fill: %v", err)
	}
	var incomplete = created.ComparisonState()
	var pending = created.Record()
	if pending.Status != PartiallyFilled || !pending.Active || !pending.ReconciliationPending {
		t.Fatalf("fee-incomplete Order appeared complete: %+v", pending)
	}
	var fee = decimal.Zero
	input.Fee = &fee
	if err = created.ApplyFill(input); err != nil {
		t.Fatalf("enrich Fill fee: %v", err)
	}
	if created.ComparisonState() != incomplete+1 {
		t.Fatal("zero-fee enrichment did not change comparison state once")
	}
	var completed = created.Record()
	if completed.Status != Filled || completed.Active || completed.ReconciliationPending {
		t.Fatalf("fee-complete Order remained pending: %+v", completed)
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

func TestOrderComparisonStateTracksSemanticChanges(t *testing.T) {
	var created, err = New(validInput())
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if created.ComparisonState() != 0 {
		t.Fatalf("actual comparison state %d, expected 0", created.ComparisonState())
	}
	if err = created.RecordSubmit(1, "", `{"submitted":true}`); err != nil {
		t.Fatalf("record submit: %v", err)
	}
	var submitted = created.ComparisonState()
	if submitted != 1 {
		t.Fatalf("actual submit state %d, expected 1", submitted)
	}
	if err = created.RecordSubmit(1, "", `{"submitted":true}`); err != nil {
		t.Fatalf("repeat submit: %v", err)
	}
	if created.ComparisonState() != submitted {
		t.Fatal("identical submit changed comparison state")
	}
	var venue = VenueState{
		VenueOrderID: 1,
		Status:       Open,
		TimestampMS:  10,
		Raw:          `{"open":true}`,
	}
	if err = created.ApplyVenueState(venue); err != nil {
		t.Fatalf("apply venue state: %v", err)
	}
	var opened = created.ComparisonState()
	if opened != submitted+1 {
		t.Fatalf("actual venue state %d, expected %d", opened, submitted+1)
	}
	if err = created.ApplyVenueState(venue); err != nil {
		t.Fatalf("repeat venue state: %v", err)
	}
	if created.ComparisonState() != opened {
		t.Fatal("identical venue state changed comparison state")
	}
	var input = validFill(1, "0.5", "100")
	input.Fee = nil
	if err = created.ApplyFill(input); err != nil {
		t.Fatalf("apply new Fill: %v", err)
	}
	var added = created.ComparisonState()
	if added != opened+1 {
		t.Fatalf("actual new Fill state %d, expected %d", added, opened+1)
	}
	if err = created.ApplyFill(input); err != nil {
		t.Fatalf("repeat Fill: %v", err)
	}
	if created.ComparisonState() != added {
		t.Fatal("identical Fill changed comparison state")
	}
	var rebate = decimal.RequireFromString("-0.25")
	input.Fee = &rebate
	if err = created.ApplyFill(input); err != nil {
		t.Fatalf("enrich Fill fee: %v", err)
	}
	var enriched = created.ComparisonState()
	if enriched != added+1 {
		t.Fatalf("actual fee state %d, expected %d", enriched, added+1)
	}
	if err = created.ApplyFill(input); err != nil {
		t.Fatalf("repeat enriched Fill: %v", err)
	}
	if created.ComparisonState() != enriched {
		t.Fatal("identical fee evidence changed comparison state")
	}
	if err = created.ApplyFill(validFill(2, "0.5", "100")); err != nil {
		t.Fatalf("apply terminal Fill: %v", err)
	}
	if created.ComparisonState() != enriched+1 || created.Record().Status != Filled {
		t.Fatalf(
			"terminal Fill state=%d status=%s",
			created.ComparisonState(),
			created.Record().Status,
		)
	}

	var pending, pendingErr = New(validInput())
	if pendingErr != nil {
		t.Fatalf("create pending order: %v", pendingErr)
	}
	var filled = VenueState{VenueOrderID: 1, Status: Filled, TimestampMS: 10}
	if pendingErr = pending.ApplyVenueState(filled); pendingErr != nil {
		t.Fatalf("apply filled acknowledgement: %v", pendingErr)
	}
	var acknowledged = pending.ComparisonState()
	pending.RefreshRecon()
	if pending.ComparisonState() != acknowledged+1 {
		t.Fatal("pending RefreshRecon did not change comparison state once")
	}
	var refreshed = pending.ComparisonState()
	if pendingErr = pending.ApplyVenueState(filled); pendingErr != nil {
		t.Fatalf("repeat filled acknowledgement: %v", pendingErr)
	}
	pending.RefreshRecon()
	if pending.ComparisonState() != refreshed {
		t.Fatal("identical filled evidence changed comparison state")
	}
}

func TestOrderComparisonReadsAllocateNothing(t *testing.T) {
	var created, err = New(validInput())
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	var allocations = testing.AllocsPerRun(1000, func() {
		_ = created.ComparisonState()
		_ = created.FillIdentity()
	})
	if allocations != 0 {
		t.Fatalf("comparison reads allocated %.2f objects", allocations)
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
	var fee = decimal.Zero
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
		Fee:          &fee,
	}
}

// Section 3 - Generic Helpers
