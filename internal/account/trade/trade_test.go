package trade

import (
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/fill"
	"nuubot/internal/account/order"
)

// Section 1 - Program Flow

func TestTradeCalculatesWeightedEntryAndClosePnL(t *testing.T) {
	var actual, err = New(Input{
		LedgerID:    1,
		TradeID:     2,
		TradeNo:     1,
		Account:     "sim",
		CycleNumber: 3,
		Symbol:      "BTC",
	})
	if err != nil {
		t.Fatalf("create trade: %v", err)
	}
	var first = filledOrder(t, 1, order.Buy, order.Entry, "1", "100")
	var second = filledOrder(t, 2, order.Buy, order.Entry, "1", "120")
	var closeOrder = filledOrder(t, 3, order.Sell, order.Stop, "2", "130")
	for _, created := range []*order.Order{first, second, closeOrder} {
		err = actual.AddOrder(created)
		if err != nil {
			t.Fatalf("add order: %v", err)
		}
	}
	var record = actual.Record()
	if record.Status != Closed || record.Side != Flat {
		t.Fatalf("actual state status=%s side=%s", record.Status, record.Side)
	}
	if record.RealizedPnL.String() != "40" ||
		record.GrossPnL.String() != "40" ||
		record.NetPnL.String() != "40" {
		t.Fatalf(
			"unexpected pnl realized=%s gross=%s net=%s",
			record.RealizedPnL,
			record.GrossPnL,
			record.NetPnL,
		)
	}
}

func TestTradeRejectsReversal(t *testing.T) {
	var actual, err = New(Input{
		LedgerID:    1,
		TradeID:     2,
		TradeNo:     1,
		Account:     "sim",
		CycleNumber: 3,
		Symbol:      "BTC",
	})
	if err != nil {
		t.Fatalf("create trade: %v", err)
	}
	err = actual.AddOrder(filledOrder(t, 1, order.Buy, order.Entry, "1", "100"))
	if err != nil {
		t.Fatalf("add entry: %v", err)
	}
	err = actual.AddOrder(filledOrder(t, 2, order.Sell, order.Stop, "2", "90"))
	if err == nil {
		t.Fatal("actual error nil, expected reversal rejection")
	}
}

func TestTradeRefreshMarkPreservesStructureAndTerminalValues(t *testing.T) {
	var actual, err = New(Input{
		LedgerID:    1,
		TradeID:     2,
		TradeNo:     1,
		Account:     "sim",
		CycleNumber: 3,
		Symbol:      "BTC",
	})
	if err != nil {
		t.Fatalf("create trade: %v", err)
	}
	err = actual.AddOrder(filledOrder(t, 1, order.Buy, order.Entry, "1", "100"))
	if err != nil {
		t.Fatalf("add entry: %v", err)
	}
	var firstMark = decimal.NewFromInt(110)
	err = actual.RefreshRecon(&firstMark)
	if err != nil {
		t.Fatalf("refresh initial mark: %v", err)
	}
	var before = actual.ReconState()
	var nextMark = decimal.NewFromInt(120)
	err = actual.RefreshMark(&nextMark)
	if err != nil {
		t.Fatalf("refresh next mark: %v", err)
	}
	var marked = actual.ReconState()
	if marked.Status != before.Status || marked.Side != before.Side ||
		!marked.OpenQuantity.Equal(before.OpenQuantity) ||
		!marked.AverageEntryPrice.Equal(before.AverageEntryPrice) ||
		marked.HasAveragePrice != before.HasAveragePrice ||
		!marked.RealizedPnL.Equal(before.RealizedPnL) ||
		!marked.Fees.Equal(before.Fees) || marked.OpenedMS != before.OpenedMS ||
		marked.ClosedMS != before.ClosedMS || marked.UpdatedMS != before.UpdatedMS ||
		marked.UnrealizedPnL.String() != "20" || marked.GrossPnL.String() != "20" ||
		marked.NetPnL.String() != "20" {
		t.Fatalf("mark refresh changed structure: before=%+v after=%+v", before, marked)
	}

	err = actual.AddOrder(filledOrder(t, 2, order.Sell, order.Stop, "1", "130"))
	if err != nil {
		t.Fatalf("add close: %v", err)
	}
	var closed = actual.metrics
	var terminalMark = decimal.NewFromInt(150)
	err = actual.RefreshMark(&terminalMark)
	if err != nil {
		t.Fatalf("refresh terminal mark: %v", err)
	}
	if !sameMetrics(actual.metrics, closed) {
		t.Fatalf("terminal mark changed values: before=%+v after=%+v", closed, actual.metrics)
	}
}

// Section 2 - Domain Helpers

func filledOrder(
	t *testing.T,
	orderID uint64,
	side string,
	role string,
	quantity string,
	price string,
) *order.Order {
	t.Helper()
	var requestedPrice = decimal.RequireFromString(price)
	var created, err = order.New(order.Input{
		LedgerID:          1,
		TradeID:           2,
		OrderID:           orderID,
		Account:           "sim",
		CycleNumber:       3,
		Symbol:            "BTC",
		BatchNo:           1,
		OrderPos:          uint16(orderID),
		CLOID:             "0x0000000000000000000000000000000" + string(rune('0'+orderID)),
		Role:              role,
		Side:              side,
		Type:              order.Limit,
		TimeInForce:       order.IOC,
		RequestedQuantity: decimal.RequireFromString(quantity),
		RequestedPrice:    &requestedPrice,
		TimestampMS:       orderID,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	var fee = decimal.Zero
	err = created.ApplyFill(fill.Input{
		LedgerID:     1,
		TradeID:      2,
		OrderID:      orderID,
		Account:      "sim",
		CycleNumber:  3,
		Symbol:       "BTC",
		CLOID:        created.Record().CLOID,
		VenueOrderID: orderID,
		VenueTID:     orderID,
		Side:         side,
		Quantity:     decimal.RequireFromString(quantity),
		Price:        decimal.RequireFromString(price),
		TimestampMS:  10 + orderID,
		Fee:          &fee,
	})
	if err != nil {
		t.Fatalf("fill order: %v", err)
	}
	return created
}

// Section 3 - Generic Helpers
