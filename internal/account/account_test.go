package account

import (
	"database/sql"
	"io"
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"

	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/order"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestAccountRunsOneReconciledBracket(t *testing.T) {
	var actual Account
	var err = actual.Init(logging.Create(io.Discard), Config{
		LedgerID:       1,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Name:           "sim",
		Venue:          "simulator",
		Network:        "simnet",
		Symbol:         "BTC",
		Meta: meta.Instrument{
			Network:       "testnet",
			Kind:          "perp",
			Symbol:        "BTC",
			AssetID:       0,
			SizeDecimals:  5,
			PriceDecimals: 1,
		},
		MinNotionalUSDC: decimal.NewFromInt(11),
		EquityUSDC:      decimal.NewFromInt(1000),
		FeePct:          decimal.RequireFromString("0.035"),
		SlippagePct:     decimal.Zero,
		PersistMode:     "none",
	})
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var first, _ = market.CreateBBO(1000, 100)
	err = actual.IngestBBO(first)
	if err != nil {
		t.Fatalf("ingest first BBO: %v", err)
	}
	var quantity = decimal.RequireFromString("0.11")
	var entry = decimal.NewFromInt(100)
	var takeProfit = decimal.NewFromInt(110)
	var stopLoss = decimal.NewFromInt(90)
	var placed PlaceResult
	placed, err = actual.PlaceOrders([]OrderSpec{
		{
			Role:        order.Entry,
			Side:        order.Buy,
			Type:        order.Limit,
			TimeInForce: order.IOC,
			Quantity:    quantity,
			Price:       &entry,
			TimestampMS: 1000,
		},
		{
			Role:         order.TakeProfit,
			Side:         order.Sell,
			Type:         order.Trigger,
			TimeInForce:  order.GTC,
			Quantity:     quantity,
			Price:        &takeProfit,
			TriggerPrice: &takeProfit,
			ReduceOnly:   true,
			TimestampMS:  1000,
		},
		{
			Role:         order.StopLoss,
			Side:         order.Sell,
			Type:         order.Trigger,
			TimeInForce:  order.GTC,
			Quantity:     quantity,
			Price:        &stopLoss,
			TriggerPrice: &stopLoss,
			ReduceOnly:   true,
			TimestampMS:  1000,
		},
	})
	if err != nil {
		t.Fatalf("place bracket: %v", err)
	}
	if len(placed.Orders) != 3 {
		t.Fatalf("actual Orders %d, expected 3", len(placed.Orders))
	}
	var snapshot Snapshot
	snapshot, _, err = actual.Reconcile(1000, false)
	if err != nil {
		t.Fatalf("reconcile entry: %v", err)
	}
	if !snapshot.PositionQuantity.IsPositive() {
		t.Fatalf("actual entry position %s, expected positive", snapshot.PositionQuantity)
	}
	var second, _ = market.CreateBBO(2000, 111)
	err = actual.IngestBBO(second)
	if err != nil {
		t.Fatalf("ingest take-profit BBO: %v", err)
	}
	snapshot, _, err = actual.Reconcile(2000, false)
	if err != nil {
		t.Fatalf("reconcile exit: %v", err)
	}
	if !snapshot.PositionQuantity.IsZero() || snapshot.Fills != 2 {
		t.Fatalf(
			"actual final position=%s fills=%d",
			snapshot.PositionQuantity,
			snapshot.Fills,
		)
	}
	var result Result
	result, err = actual.Result()
	if err != nil {
		t.Fatalf("read Account result: %v", err)
	}
	if len(result.Ledger.Trades) != 1 || result.Simulator == nil ||
		len(result.Simulator.Fills) != 2 {
		t.Fatalf("unexpected Account result: %+v", result)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestAccountMaxPersistenceRecoversDirtyVenueState(t *testing.T) {
	var cfg = Config{
		LedgerID:       7,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Name:           "sim",
		Venue:          "simulator",
		Network:        "simnet",
		Symbol:         "BTC",
		Meta: meta.Instrument{
			Network:       "testnet",
			Kind:          "perp",
			Symbol:        "BTC",
			AssetID:       0,
			SizeDecimals:  5,
			PriceDecimals: 1,
		},
		MinNotionalUSDC: decimal.NewFromInt(11),
		EquityUSDC:      decimal.NewFromInt(1000),
		FeePct:          decimal.RequireFromString("0.035"),
		SlippagePct:     decimal.Zero,
		PersistMode:     "max",
		ResultPath:      t.TempDir() + "/result.db",
	}
	var first Account
	var err = first.Init(logging.Create(io.Discard), cfg)
	if err != nil {
		t.Fatalf("initialize first Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = first.IngestBBO(warm); err != nil {
		t.Fatalf("ingest first BBO: %v", err)
	}
	var quantity = decimal.RequireFromString("0.11")
	var entry = decimal.NewFromInt(100)
	var takeProfit = decimal.NewFromInt(110)
	var stopLoss = decimal.NewFromInt(90)
	_, err = first.PlaceOrders([]OrderSpec{
		{
			Role:        order.Entry,
			Side:        order.Buy,
			Type:        order.Limit,
			TimeInForce: order.IOC,
			Quantity:    quantity,
			Price:       &entry,
			TimestampMS: 1000,
		},
		{
			Role:         order.TakeProfit,
			Side:         order.Sell,
			Type:         order.Trigger,
			TimeInForce:  order.GTC,
			Quantity:     quantity,
			Price:        &takeProfit,
			TriggerPrice: &takeProfit,
			ReduceOnly:   true,
			TimestampMS:  1000,
		},
		{
			Role:         order.StopLoss,
			Side:         order.Sell,
			Type:         order.Trigger,
			TimeInForce:  order.GTC,
			Quantity:     quantity,
			Price:        &stopLoss,
			TriggerPrice: &stopLoss,
			ReduceOnly:   true,
			TimestampMS:  1000,
		},
	})
	if err != nil {
		t.Fatalf("place first bracket: %v", err)
	}
	if err = first.Stop(); err != nil {
		t.Fatalf("stop first Account: %v", err)
	}

	var restored Account
	if err = restored.Init(logging.Create(io.Discard), cfg); err != nil {
		t.Fatalf("initialize restored Account: %v", err)
	}
	var restoredWarm, _ = market.CreateBBO(1100, 100)
	if err = restored.IngestBBO(restoredWarm); err != nil {
		t.Fatalf("warm restored Account: %v", err)
	}
	var snapshot Snapshot
	snapshot, _, err = restored.Reconcile(1100, false)
	if err != nil {
		t.Fatalf("reconcile restored entry: %v", err)
	}
	if !snapshot.PositionQuantity.IsPositive() || snapshot.Fills != 1 {
		t.Fatalf(
			"actual restored position=%s fills=%d",
			snapshot.PositionQuantity,
			snapshot.Fills,
		)
	}
	var exitBBO, _ = market.CreateBBO(2000, 111)
	if err = restored.IngestBBO(exitBBO); err != nil {
		t.Fatalf("ingest restored take-profit BBO: %v", err)
	}
	snapshot, _, err = restored.Reconcile(2000, false)
	if err != nil {
		t.Fatalf("reconcile restored exit: %v", err)
	}
	if !snapshot.PositionQuantity.IsZero() || snapshot.Fills != 2 {
		t.Fatalf(
			"actual final position=%s fills=%d",
			snapshot.PositionQuantity,
			snapshot.Fills,
		)
	}
	if err = restored.Stop(); err != nil {
		t.Fatalf("stop restored Account: %v", err)
	}

	var mismatched Account
	cfg.FeePct = decimal.RequireFromString("0.04")
	err = mismatched.Init(logging.Create(io.Discard), cfg)
	if err == nil {
		t.Fatal("Simulator policy mismatch was admitted")
	}
}

func TestAccountMaxPersistenceRecoversSimulatorSubmitFailure(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var cfg = simulatorFailureConfig(path, 8)
	var first Account
	var err = first.Init(logging.Create(io.Discard), cfg)
	if err != nil {
		t.Fatalf("initialize first Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = first.IngestBBO(warm); err != nil {
		t.Fatalf("warm first Account: %v", err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open result database: %v", err)
	}
	if _, err = db.Exec(`DROP TABLE simulator_state`); err != nil {
		t.Fatalf("remove Simulator store: %v", err)
	}
	var price = decimal.NewFromInt(100)
	_, err = first.PlaceOrders([]OrderSpec{{
		Role:        order.Entry,
		Side:        order.Buy,
		Type:        order.Limit,
		TimeInForce: order.IOC,
		Quantity:    decimal.RequireFromString("0.11"),
		Price:       &price,
		TimestampMS: 1000,
	}})
	if err == nil {
		t.Fatal("undurable Simulator submission was admitted")
	}
	var result Result
	result, err = first.Result()
	if err != nil {
		t.Fatalf("read failed Account result: %v", err)
	}
	if len(result.Ledger.Trades) != 1 ||
		len(result.Ledger.Trades[0].Orders) != 1 ||
		result.Ledger.Trades[0].Orders[0].Status != order.Error ||
		len(result.Simulator.Orders) != 0 {
		t.Fatalf("unexpected failed submission result: %+v", result)
	}
	if err = restoreSimulatorTable(db); err != nil {
		t.Fatalf("restore Simulator store: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close result database: %v", err)
	}
	if err = first.Stop(); err != nil {
		t.Fatalf("stop first Account: %v", err)
	}

	var restored Account
	if err = restored.Init(logging.Create(io.Discard), cfg); err != nil {
		t.Fatalf("initialize restored Account: %v", err)
	}
	var restoredWarm, _ = market.CreateBBO(1100, 100)
	if err = restored.IngestBBO(restoredWarm); err != nil {
		t.Fatalf("warm restored Account: %v", err)
	}
	var snapshot Snapshot
	snapshot, _, err = restored.Reconcile(1100, false)
	if err != nil {
		t.Fatalf("reconcile restored Account: %v", err)
	}
	if !snapshot.PositionQuantity.IsZero() || snapshot.ActiveOrders != 0 ||
		snapshot.OpenTrades != 0 {
		t.Fatalf("unexpected restored snapshot: %+v", snapshot)
	}
	if err = restored.Stop(); err != nil {
		t.Fatalf("stop restored Account: %v", err)
	}
}

func TestAccountMaxPersistenceRepairsMissingSubmittingOrder(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var cfg = simulatorFailureConfig(path, 9)
	var first Account
	var err = first.Init(logging.Create(io.Discard), cfg)
	if err != nil {
		t.Fatalf("initialize first Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = first.IngestBBO(warm); err != nil {
		t.Fatalf("warm first Account: %v", err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open result database: %v", err)
	}
	_, err = db.Exec(`
		CREATE TRIGGER fail_error_order
		BEFORE INSERT ON account_order
		WHEN NEW.status = 'error'
		BEGIN
			SELECT RAISE(FAIL, 'injected Ledger failure');
		END;
		DROP TABLE simulator_state;`)
	if err != nil {
		t.Fatalf("inject persistence failures: %v", err)
	}
	var price = decimal.NewFromInt(100)
	_, err = first.PlaceOrders([]OrderSpec{{
		Role:        order.Entry,
		Side:        order.Buy,
		Type:        order.Limit,
		TimeInForce: order.IOC,
		Quantity:    decimal.RequireFromString("0.11"),
		Price:       &price,
		TimestampMS: 1000,
	}})
	if err == nil {
		t.Fatal("double persistence failure was admitted")
	}
	var result Result
	result, err = first.Result()
	if err != nil {
		t.Fatalf("read failed Account result: %v", err)
	}
	if result.Ledger.Trades[0].Orders[0].Status != order.Created {
		t.Fatalf("unexpected in-memory Order state: %+v", result)
	}
	if _, err = db.Exec(`DROP TRIGGER fail_error_order`); err != nil {
		t.Fatalf("remove Ledger failure: %v", err)
	}
	if err = restoreSimulatorTable(db); err != nil {
		t.Fatalf("restore Simulator store: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close result database: %v", err)
	}
	if err = first.Stop(); err != nil {
		t.Fatalf("stop first Account: %v", err)
	}

	var restored Account
	if err = restored.Init(logging.Create(io.Discard), cfg); err != nil {
		t.Fatalf("initialize restored Account: %v", err)
	}
	var restoredWarm, _ = market.CreateBBO(1100, 100)
	if err = restored.IngestBBO(restoredWarm); err != nil {
		t.Fatalf("warm restored Account: %v", err)
	}
	var snapshot Snapshot
	snapshot, _, err = restored.Reconcile(1100, false)
	if err != nil {
		t.Fatalf("repair missing submitting Order: %v", err)
	}
	result, err = restored.Result()
	if err != nil {
		t.Fatalf("read repaired Account result: %v", err)
	}
	if result.Ledger.Trades[0].Orders[0].Status != order.Error ||
		snapshot.ActiveOrders != 0 || snapshot.OpenTrades != 0 {
		t.Fatalf("unexpected repaired Account state snapshot=%+v result=%+v", snapshot, result)
	}
	if err = restored.Stop(); err != nil {
		t.Fatalf("stop restored Account: %v", err)
	}
}

// Section 2 - Domain Helpers

func simulatorFailureConfig(path string, ledgerID uint64) Config {
	return Config{
		LedgerID:       ledgerID,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Name:           "sim",
		Venue:          "simulator",
		Network:        "simnet",
		Symbol:         "BTC",
		Meta: meta.Instrument{
			Network:       "mainnet",
			Kind:          "perp",
			Symbol:        "BTC",
			AssetID:       0,
			SizeDecimals:  5,
			PriceDecimals: 1,
		},
		MinNotionalUSDC: decimal.NewFromInt(11),
		EquityUSDC:      decimal.NewFromInt(1000),
		FeePct:          decimal.Zero,
		SlippagePct:     decimal.Zero,
		PersistMode:     "max",
		ResultPath:      path,
	}
}

func restoreSimulatorTable(db *sql.DB) error {
	var _, err = db.Exec(`
		CREATE TABLE simulator_state (
			ledger_id       INTEGER PRIMARY KEY REFERENCES account_ledger(ledger_id),
			schema_version  INTEGER NOT NULL,
			payload_json    TEXT NOT NULL,
			updated_ms      INTEGER NOT NULL
		)`)
	return err
}

// Section 3 - Generic Helpers
