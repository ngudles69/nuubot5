package simulator

import (
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/ledger"
	"nuubot/internal/market"
	"nuubot/internal/order"
)

// Section 1 - Program Flow

func TestSimulatorMatchesBracketAndCancelsSibling(t *testing.T) {
	var actual Simulator
	var err = actual.Init(Config{
		LedgerID:    1,
		Name:        "sim",
		Account:     "sim",
		CycleNumber: 2,
		Symbol:      "BTC",
		Equity:      decimal.NewFromInt(1000),
		FeePct:      decimal.RequireFromString("0.035"),
		SlippagePct: decimal.Zero,
		PersistMode: "none",
	})
	if err != nil {
		t.Fatalf("initialize simulator: %v", err)
	}
	var first, _ = market.CreateBBO(1000, 100)
	var changed bool
	changed, err = actual.IngestBBO(first)
	if err != nil || changed {
		t.Fatalf("warm simulator changed=%t error=%v", changed, err)
	}
	var response SubmitResponse
	response, err = actual.PlaceOrders([]OrderRequest{
		request(1, order.Entry, order.Buy, "100", false),
		request(2, order.TakeProfit, order.Sell, "110", true),
		request(3, order.StopLoss, order.Sell, "90", true),
	})
	if err != nil {
		t.Fatalf("submit bracket: %v", err)
	}
	if len(response.Statuses) != 3 ||
		response.Statuses[0].Kind != "filled" ||
		response.Statuses[1].Kind != "waitingForTrigger" ||
		response.Statuses[2].Kind != "waitingForTrigger" {
		t.Fatalf("unexpected submit statuses: %+v", response.Statuses)
	}
	var second, _ = market.CreateBBO(2000, 111)
	changed, err = actual.IngestBBO(second)
	if err != nil || !changed {
		t.Fatalf("match take profit changed=%t error=%v", changed, err)
	}
	var state AccountState
	state, err = actual.AccountState()
	if err != nil {
		t.Fatalf("read account state: %v", err)
	}
	if !state.PositionSize.IsZero() || state.AccountValue.LessThan(decimal.NewFromInt(1000)) {
		t.Fatalf(
			"unexpected account state position=%s value=%s",
			state.PositionSize,
			state.AccountValue,
		)
	}
	var result = actual.Result()
	if len(result.Fills) != 2 {
		t.Fatalf("actual fills %d, expected 2", len(result.Fills))
	}
	var stopState, statusErr = actual.OrderStatus(request(3, order.StopLoss, order.Sell, "90", true).CLOID)
	if statusErr != nil || stopState.Status != order.Canceled {
		t.Fatalf("unexpected stop state=%+v error=%v", stopState, statusErr)
	}
}

func TestSimulatorDoesNotReuseBBOForRestingOrder(t *testing.T) {
	var actual Simulator
	var err = actual.Init(Config{
		LedgerID:    1,
		Name:        "sim",
		Account:     "sim",
		CycleNumber: 2,
		Symbol:      "BTC",
		Equity:      decimal.NewFromInt(1000),
		FeePct:      decimal.Zero,
		SlippagePct: decimal.Zero,
		PersistMode: "none",
	})
	if err != nil {
		t.Fatalf("initialize simulator: %v", err)
	}
	var first, _ = market.CreateBBO(1000, 100)
	if _, err = actual.IngestBBO(first); err != nil {
		t.Fatalf("warm simulator: %v", err)
	}
	var resting = request(1, order.Entry, order.Buy, "110", false)
	resting.TimeInForce = order.GTC
	var response SubmitResponse
	response, err = actual.PlaceOrders([]OrderRequest{resting})
	if err != nil || response.Statuses[0].Kind != "resting" {
		t.Fatalf("submit resting Order response=%+v error=%v", response, err)
	}
	var changed bool
	changed, err = actual.IngestBBO(first)
	if err != nil || changed {
		t.Fatalf("duplicate BBO changed resting Order changed=%t error=%v", changed, err)
	}
	var duplicateState OrderState
	duplicateState, err = actual.OrderStatus(resting.CLOID)
	if err != nil || duplicateState.Status != order.Open {
		t.Fatalf("duplicate BBO filled resting Order state=%+v error=%v", duplicateState, err)
	}
	var immediate = request(2, order.Entry, order.Buy, "100", false)
	immediate.TradeID = 2
	response, err = actual.PlaceOrders([]OrderRequest{immediate})
	if err != nil || response.Statuses[0].Kind != "filled" {
		t.Fatalf("submit immediate Order response=%+v error=%v", response, err)
	}
	var state OrderState
	state, err = actual.OrderStatus(resting.CLOID)
	if err != nil || state.Status != order.Open {
		t.Fatalf("resting Order reused old BBO state=%+v error=%v", state, err)
	}
}

func TestSimulatorMaxPersistenceRejectsUndurableMutations(t *testing.T) {
	t.Run("submit", func(t *testing.T) {
		var actual = persistentSimulator(t)
		if err := actual.store.db.Close(); err != nil {
			t.Fatalf("close Simulator store: %v", err)
		}
		var _, err = actual.PlaceOrders([]OrderRequest{
			request(1, order.Entry, order.Buy, "100", false),
		})
		if err == nil {
			t.Fatal("undurable submit was admitted")
		}
		var result = actual.Result()
		if result.NextVenueOrderID != 1 || result.NextVenueTID != 1 ||
			len(result.Orders) != 0 || len(result.Fills) != 0 {
			t.Fatalf("submit changed memory after persistence failure: %+v", result)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		var actual = persistentSimulator(t)
		var resting = request(1, order.Entry, order.Buy, "90", false)
		resting.TimeInForce = order.GTC
		if _, err := actual.PlaceOrders([]OrderRequest{resting}); err != nil {
			t.Fatalf("place resting Order: %v", err)
		}
		var before = actual.Result()
		if err := actual.store.db.Close(); err != nil {
			t.Fatalf("close Simulator store: %v", err)
		}
		var _, err = actual.CancelOrders([]string{resting.CLOID}, 1100)
		if err == nil {
			t.Fatal("undurable cancellation was admitted")
		}
		var state, statusErr = actual.OrderStatus(resting.CLOID)
		var after = actual.Result()
		if statusErr != nil || state.Status != order.Open ||
			len(after.Orders) != len(before.Orders) {
			t.Fatalf(
				"cancellation changed memory state=%+v before=%+v after=%+v error=%v",
				state,
				before,
				after,
				statusErr,
			)
		}
	})

	t.Run("fill", func(t *testing.T) {
		var actual = persistentSimulator(t)
		var resting = request(1, order.Entry, order.Buy, "90", false)
		resting.TimeInForce = order.GTC
		if _, err := actual.PlaceOrders([]OrderRequest{resting}); err != nil {
			t.Fatalf("place resting Order: %v", err)
		}
		if err := actual.store.db.Close(); err != nil {
			t.Fatalf("close Simulator store: %v", err)
		}
		var crossed, _ = market.CreateBBO(2000, 80)
		var _, err = actual.IngestBBO(crossed)
		if err == nil {
			t.Fatal("undurable Fill was admitted")
		}
		var state, statusErr = actual.OrderStatus(resting.CLOID)
		var accountState, accountErr = actual.AccountState()
		var result = actual.Result()
		if statusErr != nil || accountErr != nil || state.Status != order.Open ||
			len(result.Fills) != 0 || accountState.ObservedMS != 1000 {
			t.Fatalf(
				"Fill changed memory state=%+v account=%+v result=%+v errors=%v/%v",
				state,
				accountState,
				result,
				statusErr,
				accountErr,
			)
		}
	})
}

// Section 2 - Domain Helpers

func persistentSimulator(t *testing.T) *Simulator {
	t.Helper()
	var path = filepath.Join(t.TempDir(), "result.db")
	var ownedLedger ledger.Ledger
	var err = ownedLedger.Init(ledger.Config{
		ID:             1,
		CycleNumber:    2,
		ExecutorNumber: 1,
		Account:        "sim",
		Network:        "simnet",
		Symbol:         "BTC",
		PersistMode:    ledger.Max,
		Path:           path,
	})
	if err != nil {
		t.Fatalf("initialize Ledger schema: %v", err)
	}
	if err = ownedLedger.Stop(); err != nil {
		t.Fatalf("stop Ledger schema: %v", err)
	}
	var actual Simulator
	err = actual.Init(Config{
		LedgerID:    1,
		Name:        "sim",
		Account:     "sim",
		CycleNumber: 2,
		Symbol:      "BTC",
		Equity:      decimal.NewFromInt(1000),
		FeePct:      decimal.Zero,
		SlippagePct: decimal.Zero,
		PersistMode: ledger.Max,
		Path:        path,
	})
	if err != nil {
		t.Fatalf("initialize persistent Simulator: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if _, err = actual.IngestBBO(warm); err != nil {
		t.Fatalf("warm persistent Simulator: %v", err)
	}
	return &actual
}

func request(
	position uint16,
	role string,
	side string,
	price string,
	reduceOnly bool,
) OrderRequest {
	var value = decimal.RequireFromString(price)
	var timeInForce = order.IOC
	if role == order.TakeProfit || role == order.StopLoss {
		timeInForce = order.GTC
	}
	return OrderRequest{
		CLOID:        cloidFor(position),
		Symbol:       "BTC",
		TradeID:      1,
		OrderID:      uint64(position),
		Role:         role,
		Side:         side,
		Type:         order.Limit,
		TimeInForce:  timeInForce,
		Quantity:     decimal.NewFromInt(1),
		Price:        &value,
		TriggerPrice: trigger(role, value),
		ReduceOnly:   reduceOnly,
		TimestampMS:  1000,
	}
}

func trigger(role string, value decimal.Decimal) *decimal.Decimal {
	if role == order.TakeProfit || role == order.StopLoss {
		return &value
	}
	return nil
}

func cloidFor(position uint16) string {
	switch position {
	case 1:
		return "0x000002000101010000080202800003e8"
	case 2:
		return "0x000002000100020000080205000003e8"
	default:
		return "0x000002000100030000080207800003e8"
	}
}

// Section 3 - Generic Helpers
