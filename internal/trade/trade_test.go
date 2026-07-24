package trade

import (
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/fill"
	"nuubot/internal/order"
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
	var closeOrder = filledOrder(t, 3, order.Sell, order.Close, "2", "130")
	for _, created := range []*order.Order{first, second, closeOrder} {
		err = actual.AddOrder(created)
		if err != nil {
			t.Fatalf("add order: %v", err)
		}
	}
	var snapshot, snapshotErr = actual.Snapshot(nil)
	if snapshotErr != nil {
		t.Fatalf("snapshot trade: %v", snapshotErr)
	}
	if snapshot.Status != Closed || snapshot.Side != Flat {
		t.Fatalf("actual state status=%s side=%s", snapshot.Status, snapshot.Side)
	}
	if snapshot.RealizedPnL.String() != "40" ||
		snapshot.GrossPnL.String() != "40" ||
		snapshot.NetPnL.String() != "40" {
		t.Fatalf(
			"unexpected pnl realized=%s gross=%s net=%s",
			snapshot.RealizedPnL,
			snapshot.GrossPnL,
			snapshot.NetPnL,
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
	err = actual.AddOrder(filledOrder(t, 2, order.Sell, order.Close, "2", "90"))
	if err == nil {
		t.Fatal("actual error nil, expected reversal rejection")
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
	err = created.ApplyFill(fill.Input{
		LedgerID:     1,
		TradeID:      2,
		OrderID:      orderID,
		Account:      "sim",
		CycleNumber:  3,
		Symbol:       "BTC",
		CLOID:        created.Snapshot().CLOID,
		VenueOrderID: orderID,
		VenueTID:     orderID,
		Side:         side,
		Quantity:     decimal.RequireFromString(quantity),
		Price:        decimal.RequireFromString(price),
		TimestampMS:  10 + orderID,
	})
	if err != nil {
		t.Fatalf("fill order: %v", err)
	}
	return created
}

// Section 3 - Generic Helpers
