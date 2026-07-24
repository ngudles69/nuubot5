package ledger

import (
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/fill"
	"nuubot/internal/order"
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
	err = actual.RecordSubmit([]SubmitOutcome{{
		OrderID:      plan.OrderIDs[0],
		VenueOrderID: 11,
	}})
	if err != nil {
		t.Fatalf("record submit: %v", err)
	}
	var recon = ReconInput{
		Orders: []OrderEvidence{{
			CLOID:        created.Snapshot().CLOID,
			VenueOrderID: 11,
			Status:       order.Filled,
			TimestampMS:  12,
		}},
		Fills: []FillEvidence{{
			CLOID:        created.Snapshot().CLOID,
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
	err = actual.Recon(recon)
	if err != nil {
		t.Fatalf("repeat reconciliation: %v", err)
	}
	var result Result
	result, err = actual.Result(&requestedPrice)
	if err != nil {
		t.Fatalf("read ledger result: %v", err)
	}
	if result.FillsThroughMS != 12 || len(result.Trades) != 1 ||
		len(result.Trades[0].Orders) != 1 ||
		len(result.Trades[0].Orders[0].Fills) != 1 {
		t.Fatalf("unexpected ledger result: %+v", result)
	}

	recon.Fills[0].Price = decimal.RequireFromString("101")
	err = actual.Recon(recon)
	if err == nil {
		t.Fatal("actual error nil, expected contradictory fill rejection")
	}
	var after Result
	after, err = actual.Result(&requestedPrice)
	if err != nil {
		t.Fatalf("read ledger after rejection: %v", err)
	}
	if after.Trades[0].GrossPnL.String() != result.Trades[0].GrossPnL.String() ||
		after.Trades[0].Orders[0].Fills[0].Price.String() != "100" {
		t.Fatalf("failed recon changed ledger: %+v", after)
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
	if err = first.RecordSubmit([]SubmitOutcome{{
		OrderID:      plan.OrderIDs[0],
		VenueOrderID: 17,
	}}); err != nil {
		t.Fatalf("record submit: %v", err)
	}
	if err = first.Recon(ReconInput{
		Orders: []OrderEvidence{{
			CLOID:        created.Snapshot().CLOID,
			VenueOrderID: 17,
			Status:       order.Filled,
			TimestampMS:  12,
		}},
		Fills: []FillEvidence{{
			CLOID:        created.Snapshot().CLOID,
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
	if err = first.Stop(); err != nil {
		t.Fatalf("stop first ledger: %v", err)
	}

	var restored Ledger
	if err = restored.Init(cfg); err != nil {
		t.Fatalf("initialize restored ledger: %v", err)
	}
	var result Result
	result, err = restored.Result(&requestedPrice)
	if err != nil {
		t.Fatalf("read restored ledger: %v", err)
	}
	if result.FillsThroughMS != 12 || len(result.Trades) != 1 ||
		len(result.Trades[0].Orders) != 1 ||
		len(result.Trades[0].Orders[0].Fills) != 1 ||
		result.Trades[0].Orders[0].Fills[0].VenueTID != 18 {
		t.Fatalf("unexpected restored ledger: %+v", result)
	}
	if err = restored.Stop(); err != nil {
		t.Fatalf("stop restored ledger: %v", err)
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
		created.Snapshot().CLOID,
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

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
