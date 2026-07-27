package simulator

import (
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
)

func TestSimulatorOwnsCanonicalOrdersAndDetachedOfficialJSON(t *testing.T) {
	var actual = newSimulator(t, "none", "")
	var warm, _ = market.CreateBBO(1000, 100)
	if changed, err := ingestSimulatorBBO(actual, warm); err != nil || changed {
		t.Fatalf("warm simulator changed=%t error=%v", changed, err)
	}

	var payload, err = actual.PlaceOrders(bracketAction(), 1000)
	if err != nil {
		t.Fatalf("place bracket: %v", err)
	}
	var submit hyperliquid.SubmitResponse
	submit, err = hyperliquid.DecodeSubmitResponse(payload)
	if err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if len(submit.Statuses) != 3 ||
		submit.Statuses[0].Kind != "filled" ||
		submit.Statuses[1].Kind != "resting" ||
		submit.Statuses[2].Kind != "resting" ||
		submit.Statuses[0].VenueOrderID != 1 ||
		submit.Statuses[1].VenueOrderID != 2 ||
		submit.Statuses[2].VenueOrderID != 3 ||
		submit.Statuses[0].CLOID != cloidFor(1) ||
		submit.Statuses[1].CLOID != cloidFor(2) ||
		submit.Statuses[2].CLOID != cloidFor(3) {
		t.Fatalf("unexpected submit response: %+v", submit)
	}
	if len(actual.orders) != 3 || len(actual.ordersByCLOID) != 3 ||
		len(actual.activeOrders) != 2 || len(actual.fills) != 1 {
		t.Fatalf(
			"unexpected canonical state orders=%d index=%d active=%d fills=%d",
			len(actual.orders),
			len(actual.ordersByCLOID),
			len(actual.activeOrders),
			len(actual.fills),
		)
	}

	var openPayload []byte
	openPayload, err = actual.OpenOrders("sim")
	if err != nil {
		t.Fatalf("read open Orders: %v", err)
	}
	openPayload[0] = ' '
	var fresh []byte
	fresh, err = actual.OpenOrders("sim")
	if err != nil {
		t.Fatalf("read fresh open Orders: %v", err)
	}
	var open []hyperliquid.OpenOrder
	open, err = hyperliquid.DecodeOpenOrders(fresh)
	if err != nil || len(open) != 2 {
		t.Fatalf("decode open Orders rows=%d error=%v", len(open), err)
	}

	var takeProfit, _ = market.CreateBBO(2000, 111)
	if changed, matchErr := ingestSimulatorBBO(actual, takeProfit); matchErr != nil || !changed {
		t.Fatalf("match take profit changed=%t error=%v", changed, matchErr)
	}
	if len(actual.orders) != 3 || len(actual.activeOrders) != 0 || len(actual.fills) != 2 {
		t.Fatalf(
			"unexpected terminal state orders=%d active=%d fills=%d",
			len(actual.orders),
			len(actual.activeOrders),
			len(actual.fills),
		)
	}
	var fillsPayload []byte
	fillsPayload, err = actual.Fills("sim", 1, 3000)
	if err != nil {
		t.Fatalf("read Fills: %v", err)
	}
	var fills []hyperliquid.Fill
	fills, err = hyperliquid.DecodeFills(fillsPayload)
	if err != nil || len(fills) != 2 {
		t.Fatalf("decode Fills rows=%d error=%v", len(fills), err)
	}
	var stopPayload []byte
	stopPayload, err = actual.OrderStatus("sim", cloidFor(3))
	if err != nil {
		t.Fatalf("read stop status: %v", err)
	}
	var stop hyperliquid.OrderStatus
	stop, err = hyperliquid.DecodeOrderStatus(stopPayload)
	if err != nil || stop.Order == nil || stop.OrderStatus != orderCanceled {
		t.Fatalf("unexpected stop status=%+v error=%v", stop, err)
	}

	var later, _ = market.CreateBBO(3000, 89)
	if changed, matchErr := ingestSimulatorBBO(actual, later); matchErr != nil || changed {
		t.Fatalf("terminal Order rematched changed=%t error=%v", changed, matchErr)
	}
	if len(actual.fills) != 2 {
		t.Fatalf("terminal Order created duplicate Fill count=%d", len(actual.fills))
	}
}

func TestSimulatorPersistsEachCanonicalRecordOnce(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var actual = newSimulator(t, "max", path)
	var warm, _ = market.CreateBBO(1000, 100)
	if _, err := ingestSimulatorBBO(actual, warm); err != nil {
		t.Fatalf("warm simulator: %v", err)
	}
	if _, err := actual.PlaceOrders(bracketAction(), 1000); err != nil {
		t.Fatalf("place bracket: %v", err)
	}
	if err := actual.Stop(); err != nil {
		t.Fatalf("stop simulator: %v", err)
	}

	var restored = newSimulator(t, "max", path)
	if len(restored.orders) != 3 || len(restored.ordersByCLOID) != 3 ||
		len(restored.activeOrders) != 2 || len(restored.fills) != 1 {
		t.Fatalf(
			"unexpected restored state orders=%d index=%d active=%d fills=%d",
			len(restored.orders),
			len(restored.ordersByCLOID),
			len(restored.activeOrders),
			len(restored.fills),
		)
	}
	for _, row := range restored.orders {
		var expected = newComparisonKey(orderPrice(row))
		if compareComparisonKeys(row.comparisonKey, expected) != 0 {
			t.Fatalf("restored Order %d comparison key changed", row.venueOrderID)
		}
	}
	var stored, found, err = restored.store.load(restored.config)
	if err != nil || !found || len(stored.Orders) != 3 || len(stored.Fills) != 1 {
		t.Fatalf(
			"unexpected stored state found=%t orders=%d fills=%d error=%v",
			found,
			len(stored.Orders),
			len(stored.Fills),
			err,
		)
	}
	if err = restored.Stop(); err != nil {
		t.Fatalf("stop restored simulator: %v", err)
	}
}

func TestSimulatorTreatsOfficialCLOIDAsOpaque(t *testing.T) {
	var actual = newSimulator(t, "none", "")
	var warm, _ = market.CreateBBO(1000, 100)
	if _, err := ingestSimulatorBBO(actual, warm); err != nil {
		t.Fatalf("warm simulator: %v", err)
	}
	var action = hyperliquid.PlaceOrderAction{
		Type:     "order",
		Grouping: "na",
		Orders: []hyperliquid.OrderRequest{{
			Asset: 0,
			IsBuy: true,
			Price: "90",
			Size:  "1",
			Type: hyperliquid.OrderType{
				Limit: &hyperliquid.LimitOrderType{TimeInForce: "Gtc"},
			},
			CLOID: "0x00000000000000000000000000000000",
		}},
	}
	if _, err := actual.PlaceOrders(action, 1000); err != nil {
		t.Fatalf("opaque official CLOID was rejected: %v", err)
	}
	var statePayload, stateErr = actual.AccountState("sim")
	if stateErr != nil {
		t.Fatalf("read clearinghouse state: %v", stateErr)
	}
	var state hyperliquid.AccountState
	state, stateErr = hyperliquid.DecodeClearinghouseState(statePayload)
	if stateErr != nil || state.TimeMS != 1000 {
		t.Fatalf("unexpected clearinghouse time=%d error=%v", state.TimeMS, stateErr)
	}
	action.Orders[0].CLOID = "local-order-1"
	if _, err := actual.PlaceOrders(action, 1000); err == nil {
		t.Fatal("invalid official CLOID shape was admitted")
	}
}

func TestSimulatorPersistenceFailureDoesNotAdmitMutation(t *testing.T) {
	var actual = newSimulator(t, "max", filepath.Join(t.TempDir(), "result.db"))
	var warm, _ = market.CreateBBO(1000, 100)
	if _, err := ingestSimulatorBBO(actual, warm); err != nil {
		t.Fatalf("warm simulator: %v", err)
	}
	if err := actual.store.db.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	if _, err := actual.PlaceOrders(bracketAction(), 1000); err == nil {
		t.Fatal("undurable mutation was admitted")
	}
	if len(actual.orders) != 0 || len(actual.fills) != 0 ||
		actual.nextVenueOrderID != 1 || actual.nextVenueTID != 1 {
		t.Fatalf(
			"failed mutation changed truth orders=%d fills=%d oid=%d tid=%d",
			len(actual.orders),
			len(actual.fills),
			actual.nextVenueOrderID,
			actual.nextVenueTID,
		)
	}
}

func newSimulator(t *testing.T, persistMode string, path string) *Simulator {
	t.Helper()
	var actual Simulator
	var marketData = market.CreateMarketData()
	var key = market.Key{Venue: "simulator", Network: "simnet", Symbol: "BTC"}
	var err = actual.Init(Config{
		MarketData:  marketData,
		MarketKey:   key,
		OnChange:    func() {},
		Account:     "sim",
		Asset:       0,
		Symbol:      "BTC",
		Equity:      decimal.NewFromInt(1000),
		FeePct:      decimal.RequireFromString("0.035"),
		SlippagePct: decimal.Zero,
		PersistMode: persistMode,
		Path:        path,
	})
	if err != nil {
		t.Fatalf("initialize Simulator: %v", err)
	}
	return &actual
}

func ingestSimulatorBBO(actual *Simulator, bbo market.BBO) (bool, error) {
	var fills = len(actual.fills)
	var active = len(actual.activeOrders)
	var position = actual.currentPosition
	var err = actual.config.MarketData.IngestBBO(actual.config.MarketKey, bbo)
	var changed = fills != len(actual.fills) || active != len(actual.activeOrders) ||
		!position.size.Equal(actual.currentPosition.size) ||
		!position.entryPrice.Equal(actual.currentPosition.entryPrice) ||
		!position.realized.Equal(actual.currentPosition.realized) ||
		!position.fees.Equal(actual.currentPosition.fees)
	return changed, err
}

func bracketAction() hyperliquid.PlaceOrderAction {
	return hyperliquid.PlaceOrderAction{
		Type:     "order",
		Grouping: "normalTpsl",
		Orders: []hyperliquid.OrderRequest{
			{
				Asset: 0,
				IsBuy: true,
				Price: "100",
				Size:  "1",
				Type: hyperliquid.OrderType{
					Limit: &hyperliquid.LimitOrderType{TimeInForce: "Ioc"},
				},
				CLOID: cloidFor(1),
			},
			{
				Asset:      0,
				IsBuy:      false,
				Price:      "110",
				Size:       "1",
				ReduceOnly: true,
				Type: hyperliquid.OrderType{
					Trigger: &hyperliquid.TriggerOrderType{
						IsMarket:     false,
						TriggerPrice: "110",
						TPSL:         "tp",
					},
				},
				CLOID: cloidFor(2),
			},
			{
				Asset:      0,
				IsBuy:      false,
				Price:      "90",
				Size:       "1",
				ReduceOnly: true,
				Type: hyperliquid.OrderType{
					Trigger: &hyperliquid.TriggerOrderType{
						IsMarket:     false,
						TriggerPrice: "90",
						TPSL:         "sl",
					},
				},
				CLOID: cloidFor(3),
			},
		},
	}
}

func cloidFor(position int) string {
	switch position {
	case 1:
		return "0x000002000101010000080202800003e8"
	case 2:
		return "0x000002000100020000080205000003e8"
	default:
		return "0x000002000100030000080207800003e8"
	}
}

func TestComparisonKeyMatchesExactDecimalOrdering(t *testing.T) {
	var values = []string{
		"0.0000000000000000000000001",
		"0.0000000000000000000000002",
		"0.001",
		"0.01",
		"0.1",
		"0.9999999999999999999999999",
		"1",
		"1.0",
		"1.0000000000000000000000001",
		"9.999999999999999999999999",
		"10",
		"999999999999999999999999999.9",
		"1000000000000000000000000000",
		"1e100",
	}
	for _, leftText := range values {
		var left = decimal.RequireFromString(leftText)
		var leftKey = newComparisonKey(left)
		for _, rightText := range values {
			var right = decimal.RequireFromString(rightText)
			var rightKey = newComparisonKey(right)
			var actual = compareComparisonKeys(leftKey, rightKey)
			var expected = left.Cmp(right)
			if sign(actual) != sign(expected) {
				t.Fatalf(
					"compare %s to %s: key=%d decimal=%d",
					leftText,
					rightText,
					actual,
					expected,
				)
			}
		}
	}
}

func TestKeyCrossingMatchesDecimalCrossing(t *testing.T) {
	var values = []decimal.Decimal{
		decimal.RequireFromString("0.00000001"),
		decimal.RequireFromString("89.999999999999999999"),
		decimal.NewFromInt(90),
		decimal.RequireFromString("90.000000000000000001"),
		decimal.RequireFromString("109.999999999999999999"),
		decimal.NewFromInt(110),
		decimal.RequireFromString("110.000000000000000001"),
		decimal.RequireFromString("1000000000000000000000000"),
	}
	for _, kind := range []string{kindLimit, kindTP, kindSL} {
		for _, isBuy := range []bool{false, true} {
			for _, threshold := range values {
				var row = &simOrder{
					kind:          kind,
					isBuy:         isBuy,
					price:         threshold,
					comparisonKey: newComparisonKey(threshold),
				}
				for _, price := range values {
					var actual = crosses(row, newComparisonKey(price))
					var expected = decimalCrosses(row, price)
					if actual != expected {
						t.Fatalf(
							"kind=%s buy=%t threshold=%s price=%s actual=%t expected=%t",
							kind,
							isBuy,
							threshold,
							price,
							actual,
							expected,
						)
					}
				}
			}
		}
	}
}

func TestComparisonKeyCrossingAllocatesNothing(t *testing.T) {
	var threshold = decimal.RequireFromString("12345678901234567890.123456789")
	var row = simOrder{
		kind:          kindLimit,
		isBuy:         true,
		price:         threshold,
		comparisonKey: newComparisonKey(threshold),
	}
	var priceKey = newComparisonKey(
		decimal.RequireFromString("12345678901234567890.123456788"),
	)
	var crossed bool
	var allocations = testing.AllocsPerRun(1000, func() {
		crossed = crosses(&row, priceKey)
	})
	if !crossed || allocations != 0 {
		t.Fatalf("crossed=%t allocations=%f, expected true and zero", crossed, allocations)
	}
}

func FuzzComparisonKeyMatchesDecimal(f *testing.F) {
	f.Add("0.00000001", "1000000000000000000000000")
	f.Add("1.2300", "1.23")
	f.Add("999999999999.999999999999", "1000000000000")
	f.Fuzz(func(t *testing.T, leftText string, rightText string) {
		if len(leftText) > 128 || len(rightText) > 128 {
			t.Skip()
		}
		var left, leftErr = decimal.NewFromString(leftText)
		var right, rightErr = decimal.NewFromString(rightText)
		if leftErr != nil || rightErr != nil || !left.IsPositive() || !right.IsPositive() {
			t.Skip()
		}
		var actual = compareComparisonKeys(
			newComparisonKey(left),
			newComparisonKey(right),
		)
		if sign(actual) != sign(left.Cmp(right)) {
			t.Fatalf("compare %s to %s: key=%d decimal=%d", left, right, actual, left.Cmp(right))
		}
	})
}

func BenchmarkComparisonKeyCrossing(b *testing.B) {
	var threshold = decimal.RequireFromString("12345678901234567890.123456789")
	var row = simOrder{
		kind:          kindLimit,
		isBuy:         true,
		price:         threshold,
		comparisonKey: newComparisonKey(threshold),
	}
	var priceKey = newComparisonKey(
		decimal.RequireFromString("12345678901234567890.123456788"),
	)
	b.ReportAllocs()
	for b.Loop() {
		_ = crosses(&row, priceKey)
	}
}

func decimalCrosses(row *simOrder, price decimal.Decimal) bool {
	var threshold = orderPrice(row)
	switch row.kind {
	case kindTP:
		if !row.isBuy {
			return price.GreaterThanOrEqual(threshold)
		}
		return price.LessThanOrEqual(threshold)
	case kindSL:
		if !row.isBuy {
			return price.LessThanOrEqual(threshold)
		}
		return price.GreaterThanOrEqual(threshold)
	default:
		if row.isBuy {
			return price.LessThanOrEqual(threshold)
		}
		return price.GreaterThanOrEqual(threshold)
	}
}

func sign(value int) int {
	switch {
	case value < 0:
		return -1
	case value > 0:
		return 1
	default:
		return 0
	}
}
