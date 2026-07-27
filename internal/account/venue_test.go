package account

import (
	"bytes"
	"fmt"

	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
	"nuubot/internal/order"
)

type venueRecorder struct {
	placed      hyperliquid.PlaceOrderAction
	placedAtMS  uint64
	cancelled   hyperliquid.CancelByCLOIDAction
	cancelledAt uint64
}

// Section 1 - Program Flow

func TestAccountSendsOnlyOfficialVenueOrderAction(t *testing.T) {
	var actual Account
	var err = actual.Init(accountTestConfig(40, "recon"))
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	if err = actual.venue.Stop(); err != nil {
		t.Fatalf("stop initialized Venue: %v", err)
	}
	var recorder = &venueRecorder{}
	actual.venue = recorder

	var quantity = decimal.RequireFromString("0.11")
	var entry = decimal.NewFromInt(100)
	var takeProfit = decimal.NewFromInt(110)
	var stopLoss = decimal.NewFromInt(90)
	var placed PlaceResult
	placed, err = actual.PlaceOrders([]OrderSpec{
		{
			Role: order.Entry, Side: order.Buy, Type: order.Limit,
			TimeInForce: order.IOC, Quantity: quantity, Price: &entry,
			TimestampMS: 1000,
		},
		{
			Role: order.TakeProfit, Side: order.Sell, Type: order.Trigger,
			TimeInForce: order.GTC, Quantity: quantity, Price: &takeProfit,
			TriggerPrice: &takeProfit, ReduceOnly: true, TimestampMS: 1000,
		},
		{
			Role: order.StopLoss, Side: order.Sell, Type: order.Trigger,
			TimeInForce: order.GTC, Quantity: quantity, Price: &stopLoss,
			TriggerPrice: &stopLoss, ReduceOnly: true, TimestampMS: 1000,
		},
	})
	if err != nil {
		t.Fatalf("place official bracket: %v", err)
	}
	if placed.TradeID == 0 || len(placed.Orders) != 3 {
		t.Fatalf("actual placed result %+v, expected one Trade and three Orders", placed)
	}
	assertOfficialBracketAction(t, recorder.placed, recorder.placedAtMS)
	err = actual.CancelOrders([]string{placed.Orders[0].CLOID}, 2000)
	if err != nil {
		t.Fatalf("cancel official Order: %v", err)
	}
	if recorder.cancelled.Type != "cancelByCloid" ||
		len(recorder.cancelled.Cancels) != 1 ||
		recorder.cancelled.Cancels[0].Asset != 0 ||
		recorder.cancelled.Cancels[0].CLOID != placed.Orders[0].CLOID ||
		recorder.cancelledAt != 2000 {
		t.Fatalf(
			"actual official cancellation %+v at %d",
			recorder.cancelled,
			recorder.cancelledAt,
		)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

// Section 2 - Domain Helpers

func (v *venueRecorder) PlaceOrders(
	action hyperliquid.PlaceOrderAction,
	timestampMS uint64,
) ([]byte, error) {
	v.placed = action
	v.placedAtMS = timestampMS
	return []byte(fmt.Sprintf(
		`{"status":"ok","response":{"type":"order","data":{"statuses":[{"resting":{"cloid":%q}},{"resting":{"cloid":%q}},{"resting":{"cloid":%q}}]}}}`,
		action.Orders[0].CLOID,
		action.Orders[1].CLOID,
		action.Orders[2].CLOID,
	)), nil
}

func (v *venueRecorder) CancelOrders(
	action hyperliquid.CancelByCLOIDAction,
	timestampMS uint64,
) ([]byte, error) {
	v.cancelled = action
	v.cancelledAt = timestampMS
	return []byte(`{"status":"ok","response":{"type":"cancel","data":{"statuses":["success"]}}}`), nil
}

func (v *venueRecorder) IngestBBO(market.BBO) (bool, error) {
	return false, nil
}

func (v *venueRecorder) OpenOrders(string) ([]byte, error) {
	return []byte(`[]`), nil
}

func (v *venueRecorder) Fills(string, uint64, uint64) ([]byte, error) {
	return []byte(`[]`), nil
}

func (v *venueRecorder) OrderStatus(string, string) ([]byte, error) {
	return []byte(`{"status":"unknownOid"}`), nil
}

func (v *venueRecorder) AccountState(string) ([]byte, error) {
	return nil, nil
}

func (v *venueRecorder) Stop() error {
	return nil
}

func assertOfficialBracketAction(
	t *testing.T,
	action hyperliquid.PlaceOrderAction,
	timestampMS uint64,
) {
	t.Helper()
	if action.Type != "order" || action.Grouping != "normalTpsl" ||
		len(action.Orders) != 3 || timestampMS != 1000 {
		t.Fatalf("actual official action %+v at %d", action, timestampMS)
	}
	var entry = action.Orders[0]
	if entry.Asset != 0 || !entry.IsBuy || entry.Price != "100" ||
		entry.Size != "0.12223" || entry.ReduceOnly || entry.CLOID == "" ||
		entry.Type.Limit == nil || entry.Type.Limit.TimeInForce != order.IOC ||
		entry.Type.Trigger != nil {
		t.Fatalf("actual official entry %+v", entry)
	}
	for index, tpsl := range []string{"tp", "sl"} {
		var child = action.Orders[index+1]
		if child.Asset != 0 || child.IsBuy || !child.ReduceOnly ||
			child.Type.Limit != nil || child.Type.Trigger == nil ||
			child.Type.Trigger.IsMarket || child.Type.Trigger.TPSL != tpsl {
			t.Fatalf("actual official child %d %+v", index, child)
		}
	}
	var payload, err = hyperliquid.Encode(action)
	if err != nil {
		t.Fatalf("encode official action: %v", err)
	}
	for _, forbidden := range [][]byte{
		[]byte(`"account"`),
		[]byte(`"ledger"`),
		[]byte(`"order_id"`),
		[]byte(`"role"`),
		[]byte(`"symbol"`),
		[]byte(`"trade_id"`),
	} {
		if bytes.Contains(payload, forbidden) {
			t.Fatalf("official action contains private field %s: %s", forbidden, payload)
		}
	}
}

// Section 3 - Generic Helpers
