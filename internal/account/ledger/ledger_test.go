package ledger

import (
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/fill"
	"nuubot/internal/account/order"
	"nuubot/internal/account/trade"
)

// Section 1 - Program Flow

func TestLedgerReconcilesAtomicallyAndIdempotently(t *testing.T) {
	var actual Ledger
	var err = actual.Init(Config{
		ID:             1,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Account:        "sim",
		Network:        "simnet",
		Symbol:         "BTC",
		PersistMode:    None,
	})
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	var plan Plan
	plan, err = actual.PlanTrade(1)
	if err != nil {
		t.Fatalf("plan trade: %v", err)
	}
	var requestedPrice = decimal.RequireFromString("100")
	var created *order.Order
	created, err = order.New(order.Input{
		LedgerID:          1,
		TradeID:           plan.Trade.TradeID,
		OrderID:           plan.OrderIDs[0],
		Account:           "sim",
		CycleNumber:       2,
		Symbol:            "BTC",
		BatchNo:           1,
		OrderPos:          1,
		CLOID:             "0x00000000000000000000000000000001",
		Role:              order.Entry,
		Side:              order.Buy,
		Type:              order.Limit,
		TimeInForce:       order.IOC,
		RequestedQuantity: decimal.NewFromInt(1),
		RequestedPrice:    &requestedPrice,
		TimestampMS:       10,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	err = actual.CreateTrade(plan.Trade, []*order.Order{created})
	if err != nil {
		t.Fatalf("create trade: %v", err)
	}
	assertCachedSummary(t, &actual)
	err = actual.RecordSubmit([]SubmitOutcome{{
		OrderID:      plan.OrderIDs[0],
		VenueOrderID: 11,
	}})
	if err != nil {
		t.Fatalf("record submit: %v", err)
	}
	assertCachedSummary(t, &actual)
	var recon = ReconInput{
		Orders: []OrderEvidence{{
			CLOID:        created.Record().CLOID,
			VenueOrderID: 11,
			Status:       order.Filled,
			TimestampMS:  12,
		}},
		Fills: []FillEvidence{{
			CLOID:        created.Record().CLOID,
			VenueOrderID: 11,
			VenueTID:     12,
			Side:         fill.Buy,
			Quantity:     decimal.NewFromInt(1),
			Price:        requestedPrice,
			TimestampMS:  12,
		}},
		FillsThroughMS: 12,
		ObservedMS:     12,
	}
	err = actual.Recon(recon)
	if err != nil {
		t.Fatalf("reconcile ledger: %v", err)
	}
	assertCachedSummary(t, &actual)
	if len(actual.orders) != 1 || len(actual.cloids) != 1 || len(actual.fills) != 1 ||
		len(actual.pendingOrders) != 1 || len(actual.pendingFills) != 1 {
		t.Fatalf(
			"unexpected indexes orders=%d cloids=%d fills=%d pending_orders=%d pending_fills=%d",
			len(actual.orders),
			len(actual.cloids),
			len(actual.fills),
			len(actual.pendingOrders),
			len(actual.pendingFills),
		)
	}
	var unchanged = actual.trades[plan.Trade.TradeID]
	err = actual.Recon(recon)
	if err != nil {
		t.Fatalf("repeat reconciliation: %v", err)
	}
	assertCachedSummary(t, &actual)
	if actual.trades[plan.Trade.TradeID] != unchanged {
		t.Fatal("unchanged reconciliation replaced the Trade")
	}
	var rebate = decimal.RequireFromString("-0.25")
	recon.Fills[0].Fee = &rebate
	err = actual.Recon(recon)
	if err != nil {
		t.Fatalf("enrich reconciliation fee: %v", err)
	}
	assertCachedSummary(t, &actual)
	if len(actual.pendingOrders) != 0 || len(actual.pendingFills) != 0 {
		t.Fatalf(
			"fee enrichment remained pending orders=%d fills=%d",
			len(actual.pendingOrders),
			len(actual.pendingFills),
		)
	}
	var result Result
	result, err = actual.Result()
	if err != nil {
		t.Fatalf("read ledger result: %v", err)
	}
	var execution, found = actual.Fill(12)
	if result.FillsThroughMS != 12 || result.Trades != 1 || result.Orders != 1 ||
		result.Fills != 1 || !found || !execution.HasFee || !execution.Fee.Equal(rebate) {
		t.Fatalf("unexpected ledger result: %+v Fill=%+v", result, execution)
	}

	recon.Fills[0].Price = decimal.RequireFromString("101")
	err = actual.Recon(recon)
	if err == nil {
		t.Fatal("actual error nil, expected contradictory fill rejection")
	}
	assertCachedSummary(t, &actual)
	var after Result
	after, err = actual.Result()
	if err != nil {
		t.Fatalf("read ledger after rejection: %v", err)
	}
	var afterFill, _ = actual.Fill(12)
	if !after.Summary.GrossPnL.Equal(result.Summary.GrossPnL) ||
		afterFill.Price.String() != "100" {
		t.Fatalf("failed recon changed ledger: %+v Fill=%+v", after, afterFill)
	}
}

func TestLedgerFailedReconMutationKeepsSummaryAligned(t *testing.T) {
	var actual Ledger
	var err = actual.Init(Config{
		ID: 2, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: None,
	})
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	var _, created = createLedgerTrade(
		t,
		&actual,
		"0x00000000000000000000000000000002",
	)
	var attempt *ReconAttempt
	attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 20})
	if err != nil {
		t.Fatalf("prepare reconciliation: %v", err)
	}
	var price = decimal.NewFromInt(100)
	err = actual.UpdateReconFills(attempt, []FillEvidence{
		{
			CLOID: created.Record().CLOID, VenueOrderID: 21, VenueTID: 21,
			Side: fill.Buy, Quantity: decimal.RequireFromString("0.5"),
			Price: price, TimestampMS: 20,
		},
		{
			CLOID: created.Record().CLOID, VenueOrderID: 21, VenueTID: 22,
			Side: fill.Sell, Quantity: decimal.RequireFromString("0.5"),
			Price: price, TimestampMS: 20,
		},
	})
	if err == nil {
		t.Fatal("invalid second Fill was admitted")
	}
	if _, found := actual.Fill(21); !found {
		t.Fatal("failed reconciliation did not retain direct first Fill mutation")
	}
	assertCachedSummary(t, &actual)
}

func TestLedgerReconRefreshesOnlyChangedTradeStructure(t *testing.T) {
	var actual Ledger
	var err = actual.Init(Config{
		ID: 14, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: None,
	})
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	var firstPlan, firstOrder = createLedgerTrade(
		t,
		&actual,
		"0x00000000000000000000000000000018",
	)
	var secondPlan, secondOrder = createLedgerTrade(
		t,
		&actual,
		"0x00000000000000000000000000000019",
	)
	var fee = decimal.Zero
	var secondState = secondOrder.ReconState()
	err = secondOrder.ApplyFill(fill.Input{
		LedgerID:     secondState.LedgerID,
		TradeID:      secondState.TradeID,
		OrderID:      secondState.OrderID,
		Account:      secondState.Account,
		CycleNumber:  secondState.CycleNumber,
		Symbol:       secondState.Symbol,
		CLOID:        secondState.CLOID,
		VenueOrderID: 22,
		VenueTID:     22,
		Side:         fill.Buy,
		Quantity:     decimal.NewFromInt(1),
		Price:        decimal.NewFromInt(100),
		TimestampMS:  20,
		Fee:          &fee,
	})
	if err != nil {
		t.Fatalf("seed untouched Order evidence: %v", err)
	}

	var attempt *ReconAttempt
	attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 20})
	if err != nil {
		t.Fatalf("prepare reconciliation: %v", err)
	}
	err = actual.UpdateReconFills(attempt, []FillEvidence{{
		CLOID:        firstOrder.ReconState().CLOID,
		VenueOrderID: 21,
		VenueTID:     21,
		Side:         fill.Buy,
		Quantity:     decimal.NewFromInt(1),
		Price:        decimal.NewFromInt(100),
		TimestampMS:  20,
		Fee:          &fee,
	}})
	if err != nil {
		t.Fatalf("update changed Fill: %v", err)
	}
	var mark = decimal.NewFromInt(110)
	err = actual.UpdateReconTrades(attempt, &mark)
	if err != nil {
		t.Fatalf("update Trades: %v", err)
	}
	var first = actual.trades[firstPlan.Trade.TradeID].ReconState()
	var second = actual.trades[secondPlan.Trade.TradeID].ReconState()
	if first.Status != trade.Open || first.UnrealizedPnL.String() != "10" {
		t.Fatalf("changed Trade was not refreshed: %+v", first)
	}
	if second.Status != trade.Pending || !second.OpenQuantity.IsZero() {
		t.Fatalf("untouched Trade structure changed: %+v", second)
	}
	if _, touched := attempt.touchedTrades[firstPlan.Trade.TradeID]; !touched {
		t.Fatal("changed Trade was not marked touched")
	}
	if _, touched := attempt.touchedTrades[secondPlan.Trade.TradeID]; touched {
		t.Fatal("untouched Trade was marked touched")
	}
}

func TestLedgerMaxPersistenceRestoresEvidence(t *testing.T) {
	var cfg = Config{
		ID:             7,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Account:        "sim",
		Network:        "simnet",
		Symbol:         "BTC",
		PersistMode:    Max,
		Path:           t.TempDir() + "/result.db",
	}
	var first Ledger
	var err = first.Init(cfg)
	if err != nil {
		t.Fatalf("initialize first ledger: %v", err)
	}
	var simulatorTables int
	err = first.store.db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type = 'table' AND name = 'simulator_state'
	`).Scan(&simulatorTables)
	if err != nil {
		t.Fatalf("inspect ledger schema: %v", err)
	}
	if simulatorTables != 0 {
		t.Fatal("Ledger created Simulator persistence")
	}
	var plan Plan
	plan, err = first.PlanTrade(1)
	if err != nil {
		t.Fatalf("plan trade: %v", err)
	}
	var requestedPrice = decimal.RequireFromString("100")
	var created *order.Order
	created, err = order.New(order.Input{
		LedgerID:          cfg.ID,
		TradeID:           plan.Trade.TradeID,
		OrderID:           plan.OrderIDs[0],
		Account:           cfg.Account,
		CycleNumber:       cfg.CycleNumber,
		Symbol:            cfg.Symbol,
		BatchNo:           1,
		OrderPos:          1,
		CLOID:             "0x00000000000000000000000000000007",
		Role:              order.Entry,
		Side:              order.Buy,
		Type:              order.Limit,
		TimeInForce:       order.IOC,
		RequestedQuantity: decimal.NewFromInt(1),
		RequestedPrice:    &requestedPrice,
		TimestampMS:       10,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err = first.CreateTrade(plan.Trade, []*order.Order{created}); err != nil {
		t.Fatalf("create trade: %v", err)
	}
	assertCachedSummary(t, &first)
	if err = first.RecordSubmit([]SubmitOutcome{{
		OrderID:      plan.OrderIDs[0],
		VenueOrderID: 17,
	}}); err != nil {
		t.Fatalf("record submit: %v", err)
	}
	assertCachedSummary(t, &first)
	if err = first.Recon(ReconInput{
		Orders: []OrderEvidence{{
			CLOID:        created.Record().CLOID,
			VenueOrderID: 17,
			Status:       order.Filled,
			TimestampMS:  12,
		}},
		Fills: []FillEvidence{{
			CLOID:        created.Record().CLOID,
			VenueOrderID: 17,
			VenueTID:     18,
			Side:         fill.Buy,
			Quantity:     decimal.NewFromInt(1),
			Price:        requestedPrice,
			TimestampMS:  12,
		}},
		FillsThroughMS: 12,
		ObservedMS:     12,
	}); err != nil {
		t.Fatalf("reconcile first ledger: %v", err)
	}
	assertCachedSummary(t, &first)
	if err = first.Stop(); err != nil {
		t.Fatalf("stop first ledger: %v", err)
	}

	var restored Ledger
	if err = restored.Init(cfg); err != nil {
		t.Fatalf("initialize restored ledger: %v", err)
	}
	var result Result
	result, err = restored.Result()
	if err != nil {
		t.Fatalf("read restored ledger: %v", err)
	}
	var restoredFill, found = restored.Fill(18)
	if result.FillsThroughMS != 12 || result.Trades != 1 || result.Orders != 1 ||
		result.Fills != 1 || !found || restoredFill.VenueTID != 18 {
		t.Fatalf("unexpected restored ledger: %+v Fill=%+v", result, restoredFill)
	}
	assertCachedSummary(t, &restored)
	if err = restored.Stop(); err != nil {
		t.Fatalf("stop restored ledger: %v", err)
	}
}

func TestLedgerRejectsCrossTradeDuplicateBeforePersistence(t *testing.T) {
	var cfg = Config{
		ID: 8, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: Max,
		Path: t.TempDir() + "/result.db",
	}
	var actual Ledger
	var err = actual.Init(cfg)
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	t.Cleanup(func() { _ = actual.Stop() })
	var cloIDs = []string{
		"0x00000000000000000000000000000008",
		"0x00000000000000000000000000000009",
	}
	var tradeIDs = make([]uint64, 2)
	var orderIDs = make([]uint64, 2)
	var venueOrderIDs = []uint64{81, 82}
	var requestedPrice = decimal.NewFromInt(100)
	for index := range cloIDs {
		var plan Plan
		plan, err = actual.PlanTrade(1)
		if err != nil {
			t.Fatalf("plan Trade %d: %v", index, err)
		}
		var created *order.Order
		created, err = order.New(order.Input{
			LedgerID: cfg.ID, TradeID: plan.Trade.TradeID, OrderID: plan.OrderIDs[0],
			Account: cfg.Account, CycleNumber: cfg.CycleNumber, Symbol: cfg.Symbol,
			BatchNo: 1, OrderPos: 1, CLOID: cloIDs[index], Role: order.Entry,
			Side: order.Buy, Type: order.Limit, TimeInForce: order.IOC,
			RequestedQuantity: decimal.NewFromInt(1), RequestedPrice: &requestedPrice,
			TimestampMS: uint64(10 + index),
		})
		if err != nil {
			t.Fatalf("create Order %d: %v", index, err)
		}
		if err = actual.CreateTrade(plan.Trade, []*order.Order{created}); err != nil {
			t.Fatalf("create Trade %d: %v", index, err)
		}
		if err = actual.RecordSubmit([]SubmitOutcome{{
			OrderID: plan.OrderIDs[0], VenueOrderID: venueOrderIDs[index],
		}}); err != nil {
			t.Fatalf("record submit %d: %v", index, err)
		}
		tradeIDs[index] = plan.Trade.TradeID
		orderIDs[index] = plan.OrderIDs[0]
	}
	var beforeTrades = []*trade.Trade{actual.trades[tradeIDs[0]], actual.trades[tradeIDs[1]]}
	var beforeSummary = actual.Summary()
	var fee = decimal.Zero
	var recon = ReconInput{
		FillsThroughMS: 20, ObservedMS: 20, AccountStateRaw: `{"attempt":"changed"}`,
	}
	for index := range cloIDs {
		recon.Fills = append(recon.Fills, FillEvidence{
			CLOID: cloIDs[index], VenueOrderID: venueOrderIDs[index], VenueTID: 99,
			Side: fill.Buy, Quantity: decimal.NewFromInt(1), Price: requestedPrice,
			TimestampMS: 20, Fee: &fee,
		})
	}
	err = actual.Recon(recon)
	if err == nil {
		t.Fatal("cross-Trade duplicate Venue TID was admitted")
	}
	var afterSummary = actual.Summary()
	if actual.trades[tradeIDs[0]] != beforeTrades[0] ||
		actual.trades[tradeIDs[1]] != beforeTrades[1] || actual.fillsThroughMS != 0 ||
		actual.lastReconMS != 0 || actual.accountStateRaw != "" || len(actual.fills) != 0 ||
		!sameLedgerSummary(beforeSummary, afterSummary) {
		t.Fatalf("failed duplicate changed memory before=%+v after=%+v", beforeSummary, afterSummary)
	}
	var storedCursor sql.NullInt64
	var storedRecon sql.NullInt64
	var storedAccountState string
	if err = actual.store.db.QueryRow(
		`SELECT fills_through_ms, last_recon_ms, account_state_json
		 FROM account_ledger WHERE ledger_id = ?`, cfg.ID,
	).Scan(&storedCursor, &storedRecon, &storedAccountState); err != nil {
		t.Fatalf("read stored cursor: %v", err)
	}
	var storedFills int
	if err = actual.store.db.QueryRow(
		`SELECT COUNT(*) FROM account_fill WHERE ledger_id = ?`, cfg.ID,
	).Scan(&storedFills); err != nil {
		t.Fatalf("count stored Fills: %v", err)
	}
	if storedCursor.Valid || storedRecon.Valid || storedAccountState != "{}" || storedFills != 0 {
		t.Fatalf(
			"failed duplicate changed database cursor=%+v recon=%+v state=%s fills=%d",
			storedCursor,
			storedRecon,
			storedAccountState,
			storedFills,
		)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop ledger: %v", err)
	}
}

func TestLedgerMutationsPreserveDirectTradeOwnership(t *testing.T) {
	var actual Ledger
	var err = actual.Init(Config{
		ID: 10, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: None,
	})
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	var firstPlan, _ = createLedgerTrade(
		t,
		&actual,
		"0x00000000000000000000000000000010",
	)
	var first = actual.trades[firstPlan.Trade.TradeID]
	var rejectedPlan Plan
	rejectedPlan, err = actual.PlanTrade(1)
	if err != nil {
		t.Fatalf("plan rejected Trade: %v", err)
	}
	var duplicate = createOrder(
		t,
		&actual,
		rejectedPlan.Trade.TradeID,
		rejectedPlan.OrderIDs[0],
		1,
		"0x00000000000000000000000000000010",
	)
	err = actual.CreateTrade(rejectedPlan.Trade, []*order.Order{duplicate})
	if err == nil {
		t.Fatal("cross-Trade duplicate CLOID was admitted")
	}
	if actual.TradeCount() != 1 || actual.nextTradeID != rejectedPlan.Trade.TradeID ||
		actual.nextOrderID != rejectedPlan.OrderIDs[0] || actual.trades[firstPlan.Trade.TradeID] != first {
		t.Fatal("rejected CreateTrade mutated Ledger ownership")
	}
	var secondPlan, _ = createLedgerTrade(
		t,
		&actual,
		"0x00000000000000000000000000000011",
	)
	var second = actual.trades[secondPlan.Trade.TradeID]
	if actual.trades[firstPlan.Trade.TradeID] != first {
		t.Fatal("CreateTrade replaced an existing Trade")
	}

	var added = createLedgerOrder(
		t,
		&actual,
		secondPlan.Trade.TradeID,
		2,
		"0x00000000000000000000000000000012",
	)
	if err = actual.AddOrders(secondPlan.Trade.TradeID, []*order.Order{added}); err != nil {
		t.Fatalf("add Orders: %v", err)
	}
	assertCachedSummary(t, &actual)
	if actual.trades[firstPlan.Trade.TradeID] != first ||
		actual.trades[secondPlan.Trade.TradeID] != second {
		t.Fatal("AddOrders replaced Ledger-owned Trades")
	}
	if err = actual.RecordSubmit([]SubmitOutcome{{
		OrderID: added.ReconState().OrderID, VenueOrderID: 101,
	}}); err != nil {
		t.Fatalf("record submit: %v", err)
	}
	assertCachedSummary(t, &actual)
	if actual.trades[firstPlan.Trade.TradeID] != first ||
		actual.trades[secondPlan.Trade.TradeID] != second {
		t.Fatal("RecordSubmit replaced Ledger-owned Trades")
	}
}

func TestLedgerMaxMutationsPersistOnlyTouchedRows(t *testing.T) {
	var cfg = Config{
		ID: 11, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: Max,
		Path: t.TempDir() + "/result.db",
	}
	var actual Ledger
	var err = actual.Init(cfg)
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	t.Cleanup(func() { _ = actual.Stop() })
	var firstPlan, _ = createLedgerTrade(
		t,
		&actual,
		"0x00000000000000000000000000000013",
	)
	var secondPlan, _ = createLedgerTrade(
		t,
		&actual,
		"0x00000000000000000000000000000014",
	)
	var sentinel = `{"untouched":true}`
	_, err = actual.store.db.Exec(
		`UPDATE account_order SET raw_json = ? WHERE ledger_id = ? AND order_id = ?`,
		sentinel,
		cfg.ID,
		firstPlan.OrderIDs[0],
	)
	if err != nil {
		t.Fatalf("mark untouched Order: %v", err)
	}

	var added = createLedgerOrder(
		t,
		&actual,
		secondPlan.Trade.TradeID,
		2,
		"0x00000000000000000000000000000015",
	)
	if err = actual.AddOrders(secondPlan.Trade.TradeID, []*order.Order{added}); err != nil {
		t.Fatalf("add Orders: %v", err)
	}
	assertCachedSummary(t, &actual)
	if err = actual.RecordSubmit([]SubmitOutcome{{
		OrderID: added.ReconState().OrderID, VenueOrderID: 111,
	}}); err != nil {
		t.Fatalf("record submit: %v", err)
	}
	assertCachedSummary(t, &actual)
	var storedRaw string
	if err = actual.store.db.QueryRow(
		`SELECT raw_json FROM account_order WHERE ledger_id = ? AND order_id = ?`,
		cfg.ID,
		firstPlan.OrderIDs[0],
	).Scan(&storedRaw); err != nil {
		t.Fatalf("read untouched Order: %v", err)
	}
	if storedRaw != sentinel {
		t.Fatalf("untouched Order was rewritten: %s", storedRaw)
	}

	if err = actual.store.db.Close(); err != nil {
		t.Fatalf("close persistence for fault proof: %v", err)
	}
	err = actual.RecordSubmit([]SubmitOutcome{{
		OrderID: firstPlan.OrderIDs[0], VenueOrderID: 112,
	}})
	if err == nil {
		t.Fatal("persistence failure returned nil")
	}
	assertCachedSummary(t, &actual)
}

func TestLedgerSummaryReadsAllocateNothing(t *testing.T) {
	var actual Ledger
	var err = actual.Init(Config{
		ID: 12, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: None,
	})
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	createLedgerTrade(t, &actual, "0x00000000000000000000000000000016")
	var attempt *ReconAttempt
	attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 1})
	if err != nil {
		t.Fatalf("prepare reconciliation: %v", err)
	}
	var summaryAllocations = testing.AllocsPerRun(1000, func() {
		_ = actual.Summary()
	})
	if summaryAllocations != 0 {
		t.Fatalf("Summary allocated %.2f objects per read", summaryAllocations)
	}
	var reconAllocations = testing.AllocsPerRun(1000, func() {
		var _, readErr = actual.ReconSummary(attempt)
		if readErr != nil {
			t.Fatalf("read reconciliation summary: %v", readErr)
		}
	})
	if reconAllocations != 0 {
		t.Fatalf("ReconSummary allocated %.2f objects per read", reconAllocations)
	}
}

func BenchmarkLedgerUnchangedOrderSummaryMaintenance(b *testing.B) {
	var actual Ledger
	var err = actual.Init(Config{
		ID: 13, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: None,
	})
	if err != nil {
		b.Fatalf("initialize ledger: %v", err)
	}
	var _, created = createLedgerTrade(
		b,
		&actual,
		"0x00000000000000000000000000000017",
	)
	var evidence = []OrderEvidence{{
		CLOID: created.Record().CLOID, Status: order.Created, TimestampMS: 10,
	}}
	var attempt *ReconAttempt
	attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 20})
	if err != nil {
		b.Fatalf("prepare warm reconciliation: %v", err)
	}
	if err = actual.UpdateReconOrders(attempt, evidence); err != nil {
		b.Fatalf("warm unchanged Order: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 20})
		if err != nil {
			b.Fatalf("prepare reconciliation: %v", err)
		}
		if err = actual.UpdateReconOrders(attempt, evidence); err != nil {
			b.Fatalf("update unchanged Order: %v", err)
		}
	}
}

func TestLedgerStoreRejectsCrossTradeFill(t *testing.T) {
	var cfg = Config{
		ID:             9,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Account:        "sim",
		Network:        "simnet",
		Symbol:         "BTC",
		PersistMode:    Max,
		Path:           t.TempDir() + "/result.db",
	}
	var actual Ledger
	var err = actual.Init(cfg)
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	var plan Plan
	plan, err = actual.PlanTrade(1)
	if err != nil {
		t.Fatalf("plan trade: %v", err)
	}
	var requestedPrice = decimal.NewFromInt(100)
	var created *order.Order
	created, err = order.New(order.Input{
		LedgerID:          cfg.ID,
		TradeID:           plan.Trade.TradeID,
		OrderID:           plan.OrderIDs[0],
		Account:           cfg.Account,
		CycleNumber:       cfg.CycleNumber,
		Symbol:            cfg.Symbol,
		BatchNo:           1,
		OrderPos:          1,
		CLOID:             "0x00000000000000000000000000000009",
		Role:              order.Entry,
		Side:              order.Buy,
		Type:              order.Limit,
		TimeInForce:       order.IOC,
		RequestedQuantity: decimal.NewFromInt(1),
		RequestedPrice:    &requestedPrice,
		TimestampMS:       10,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err = actual.CreateTrade(plan.Trade, []*order.Order{created}); err != nil {
		t.Fatalf("create trade: %v", err)
	}
	_, err = actual.store.db.Exec(`
		INSERT INTO account_fill (
			ledger_id, trade_id, order_id, venue_tid, cloid, venue_order_id,
			account_name, cycle_no, symbol, side, qty, price, event_ms,
			raw_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.ID,
		999,
		plan.OrderIDs[0],
		1,
		created.Record().CLOID,
		1,
		cfg.Account,
		cfg.CycleNumber,
		cfg.Symbol,
		order.Buy,
		"1",
		"100",
		10,
		"{}",
	)
	if err == nil {
		t.Fatal("cross-Trade Fill insert succeeded")
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop ledger: %v", err)
	}
}

func TestLedgerOrderComparisonRoutesOnlySemanticChanges(t *testing.T) {
	var actual Ledger
	var err = actual.Init(Config{
		ID: 15, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: None,
	})
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	var plan, created = createLedgerTrade(
		t,
		&actual,
		"0x0000000000000000000000000000001a",
	)
	var evidence = []OrderEvidence{{
		CLOID: created.FillIdentity().CLOID, VenueOrderID: 15,
		Status: order.Filled, TimestampMS: 20,
	}}
	var attempt *ReconAttempt
	attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 20})
	if err != nil {
		t.Fatalf("prepare changed reconciliation: %v", err)
	}
	if err = actual.UpdateReconOrders(attempt, evidence); err != nil {
		t.Fatalf("update changed Order: %v", err)
	}
	if _, touched := attempt.touchedTrades[plan.Trade.TradeID]; !touched {
		t.Fatal("changed Order did not touch its Trade")
	}
	if _, touched := attempt.touchedOrders[plan.OrderIDs[0]]; !touched {
		t.Fatal("changed Order was not marked touched")
	}

	attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 20})
	if err != nil {
		t.Fatalf("prepare duplicate reconciliation: %v", err)
	}
	if err = actual.UpdateReconOrders(attempt, evidence); err != nil {
		t.Fatalf("update duplicate Order: %v", err)
	}
	if _, touched := attempt.touchedTrades[plan.Trade.TradeID]; touched {
		t.Fatal("duplicate Order touched its Trade")
	}
	if _, touched := attempt.touchedOrders[plan.OrderIDs[0]]; touched {
		t.Fatal("duplicate Order was marked touched")
	}
}

func TestLedgerReconcilesOfficialEvidenceByVenueOrderID(t *testing.T) {
	var actual Ledger
	var err = actual.Init(Config{
		ID: 16, CycleNumber: 2, ExecutorNumber: 3, Account: "sim",
		Network: "simnet", Symbol: "BTC", PersistMode: None,
	})
	if err != nil {
		t.Fatalf("initialize ledger: %v", err)
	}
	var plan, _ = createLedgerTrade(
		t,
		&actual,
		"0x0000000000000000000000000000001b",
	)
	err = actual.RecordSubmit([]SubmitOutcome{{
		OrderID: plan.OrderIDs[0], VenueOrderID: 16,
	}})
	if err != nil {
		t.Fatalf("record submit: %v", err)
	}
	var attempt *ReconAttempt
	attempt, err = actual.PrepareRecon(ReconInput{ObservedMS: 20})
	if err != nil {
		t.Fatalf("prepare reconciliation: %v", err)
	}
	err = actual.UpdateReconOrders(attempt, []OrderEvidence{{
		VenueOrderID: 16,
		Status:       order.Filled,
		TimestampMS:  20,
	}})
	if err != nil {
		t.Fatalf("update Order by Venue Order ID: %v", err)
	}
	err = actual.UpdateReconFills(attempt, []FillEvidence{{
		VenueOrderID: 16,
		VenueTID:     16,
		Side:         fill.Buy,
		Quantity:     decimal.NewFromInt(1),
		Price:        decimal.NewFromInt(100),
		TimestampMS:  20,
	}})
	if err != nil {
		t.Fatalf("update Fill by Venue Order ID: %v", err)
	}
	if _, exists := actual.fills[16]; !exists {
		t.Fatal("Venue Order ID Fill was not indexed")
	}
}

// Section 2 - Domain Helpers

func createLedgerTrade(
	t testing.TB,
	actual *Ledger,
	cloid string,
) (Plan, *order.Order) {
	t.Helper()
	var plan, err = actual.PlanTrade(1)
	if err != nil {
		t.Fatalf("plan Trade: %v", err)
	}
	var created = createOrder(t, actual, plan.Trade.TradeID, plan.OrderIDs[0], 1, cloid)
	if err = actual.CreateTrade(plan.Trade, []*order.Order{created}); err != nil {
		t.Fatalf("create Trade: %v", err)
	}
	assertCachedSummary(t, actual)
	return plan, created
}

func createLedgerOrder(
	t testing.TB,
	actual *Ledger,
	tradeID uint64,
	batchNo uint16,
	cloid string,
) *order.Order {
	t.Helper()
	var orderIDs, err = actual.PlanOrders(1)
	if err != nil {
		t.Fatalf("plan Orders: %v", err)
	}
	return createOrder(t, actual, tradeID, orderIDs[0], batchNo, cloid)
}

func createOrder(
	t testing.TB,
	actual *Ledger,
	tradeID uint64,
	orderID uint64,
	batchNo uint16,
	cloid string,
) *order.Order {
	t.Helper()
	var requestedPrice = decimal.NewFromInt(100)
	var created, err = order.New(order.Input{
		LedgerID: actual.config.ID, TradeID: tradeID, OrderID: orderID,
		Account: actual.config.Account, CycleNumber: actual.config.CycleNumber,
		Symbol: actual.config.Symbol, BatchNo: batchNo, OrderPos: 1, CLOID: cloid,
		Role: order.Entry, Side: order.Buy, Type: order.Limit, TimeInForce: order.IOC,
		RequestedQuantity: decimal.NewFromInt(1), RequestedPrice: &requestedPrice,
		TimestampMS: 10,
	})
	if err != nil {
		t.Fatalf("create Order: %v", err)
	}
	return created
}

func assertCachedSummary(t testing.TB, actual *Ledger) {
	t.Helper()
	var expected = completeTraversalSummary(actual)
	var cached = actual.Summary()
	if !sameLedgerSummary(cached, expected) {
		t.Fatalf("cached Ledger Summary mismatch cached=%+v traversal=%+v", cached, expected)
	}
}

func completeTraversalSummary(actual *Ledger) Summary {
	var result Summary
	for _, owned := range actual.trades {
		var current = owned.Summary()
		result.RealizedPnL = result.RealizedPnL.Add(current.RealizedPnL)
		result.UnrealizedPnL = result.UnrealizedPnL.Add(current.UnrealizedPnL)
		result.GrossPnL = result.GrossPnL.Add(current.GrossPnL)
		result.Fees = result.Fees.Add(current.Fees)
		result.NetPnL = result.NetPnL.Add(current.NetPnL)
		if tradeIsActive(current.Status) {
			result.OpenTrades++
		}
		result.ActiveOrders += current.ActiveOrders
		result.Fills += current.Fills
		result.PendingOrders += current.PendingOrders
		result.PendingFills += current.PendingFills
	}
	return result
}

func sameLedgerSummary(left Summary, right Summary) bool {
	return left.RealizedPnL.Equal(right.RealizedPnL) &&
		left.UnrealizedPnL.Equal(right.UnrealizedPnL) &&
		left.GrossPnL.Equal(right.GrossPnL) && left.Fees.Equal(right.Fees) &&
		left.NetPnL.Equal(right.NetPnL) && left.OpenTrades == right.OpenTrades &&
		left.ActiveOrders == right.ActiveOrders && left.Fills == right.Fills &&
		left.PendingOrders == right.PendingOrders && left.PendingFills == right.PendingFills
}

// Section 3 - Generic Helpers
