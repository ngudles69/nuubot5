package account

import (
	"database/sql"
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"

	"nuubot/internal/account/ledger"
	"nuubot/internal/account/order"
	"nuubot/internal/account/trade"
	appconfig "nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/setup"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestAccountRunsOneReconciledBracket(t *testing.T) {
	var actual Account
	var err = actual.Init(Config{
		Nuubot: accountNuubot(meta.Instrument{
			Network:       "testnet",
			Kind:          "perp",
			Symbol:        "BTC",
			AssetID:       0,
			SizeDecimals:  5,
			PriceDecimals: 1,
		}, ""),
		LedgerID:       1,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Name:           "sim",
		Venue:          "simulator",
		Network:        "simnet",
		Symbol:         "BTC",
		EquityUSDC:     decimal.NewFromInt(1000),
		FeePct:         decimal.RequireFromString("0.035"),
		SlippagePct:    decimal.Zero,
		PersistMode:    "none",
	})
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var first, _ = market.CreateBBO(1000, 100)
	err = ingestAccountBBO(&actual, first)
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
	snapshot, _, _, err = actual.Reconcile(1000, false)
	if err != nil {
		t.Fatalf("reconcile entry: %v", err)
	}
	if snapshot.Generation != 1 || !snapshot.PositionQuantity.IsPositive() {
		t.Fatalf(
			"actual generation=%d entry position=%s, expected generation 1 and positive position",
			snapshot.Generation,
			snapshot.PositionQuantity,
		)
	}
	var skipped Snapshot
	var refreshed bool
	skipped, refreshed, _, err = actual.Reconcile(1000, false)
	if err != nil {
		t.Fatalf("skip clean reconciliation: %v", err)
	}
	if refreshed || skipped.Generation != snapshot.Generation ||
		actual.ReconciliationTelemetry().Outcome != ReconSkipped {
		t.Fatalf(
			"unexpected skip refreshed=%t generation=%d telemetry=%+v",
			refreshed,
			skipped.Generation,
			actual.ReconciliationTelemetry(),
		)
	}
	if actual.reconStats != (ReconStats{
		Calls: 2, SkippedClean: 1, Executed: 1, Succeeded: 1,
	}) {
		t.Fatalf("unexpected Recon counters: %+v", actual.reconStats)
	}
	var entryValue = snapshot.AccountValue
	var mark, _ = market.CreateBBO(1500, 105)
	if err = ingestAccountBBO(&actual, mark); err != nil {
		t.Fatalf("ingest mark BBO: %v", err)
	}
	snapshot, refreshed, _, err = actual.Reconcile(1500, false)
	if err != nil {
		t.Fatalf("reconcile mark: %v", err)
	}
	if !refreshed || !snapshot.AccountValue.GreaterThan(entryValue) {
		t.Fatalf(
			"actual refreshed=%t Account value=%s, expected above %s",
			refreshed,
			snapshot.AccountValue,
			entryValue,
		)
	}
	var second, _ = market.CreateBBO(2000, 111)
	err = ingestAccountBBO(&actual, second)
	if err != nil {
		t.Fatalf("ingest take-profit BBO: %v", err)
	}
	snapshot, _, _, err = actual.Reconcile(2000, false)
	if err != nil {
		t.Fatalf("reconcile exit: %v", err)
	}
	if snapshot.Generation != 3 || !snapshot.PositionQuantity.IsZero() || snapshot.Fills != 2 {
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
	if result.Ledger.Trades != 1 || result.Ledger.Fills != 2 {
		t.Fatalf("unexpected Account result: %+v", result)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestAccountSchedulesDirtyReconAndCleanSweep(t *testing.T) {
	var cfg = accountTestConfig(2, "recon1")
	cfg.Nuubot.Runtime = appconfig.Runtime{
		ControllerIntervalMS: 1,
		ReconIntervalMS:      10,
		ReconSweepIntervalMS: 60,
		TelemetryIntervalMS:  10,
	}
	var actual Account
	var err = actual.Init(cfg)
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var initial, _ = market.CreateBBO(100, 100)
	if err = ingestAccountBBO(&actual, initial); err != nil {
		t.Fatalf("ingest initial BBO: %v", err)
	}

	var _, executed, _, reconErr = actual.Reconcile(100, false)
	if reconErr != nil || !executed || actual.lastReconMS != 100 {
		t.Fatalf(
			"initial Recon executed=%t last_ms=%d error=%v",
			executed,
			actual.lastReconMS,
			reconErr,
		)
	}

	actual.dirty = true
	_, executed, _, reconErr = actual.Reconcile(105, false)
	if reconErr != nil || executed || actual.lastReconMS != 100 {
		t.Fatalf(
			"early dirty Recon executed=%t last_ms=%d error=%v",
			executed,
			actual.lastReconMS,
			reconErr,
		)
	}
	_, executed, _, reconErr = actual.Reconcile(110, false)
	if reconErr != nil || !executed || actual.lastReconMS != 110 {
		t.Fatalf(
			"due dirty Recon executed=%t last_ms=%d error=%v",
			executed,
			actual.lastReconMS,
			reconErr,
		)
	}

	_, executed, _, reconErr = actual.Reconcile(169, false)
	if reconErr != nil || executed || actual.lastReconMS != 110 {
		t.Fatalf(
			"early clean sweep executed=%t last_ms=%d error=%v",
			executed,
			actual.lastReconMS,
			reconErr,
		)
	}
	_, executed, _, reconErr = actual.Reconcile(170, false)
	if reconErr != nil || !executed || actual.lastReconMS != 170 {
		t.Fatalf(
			"due clean sweep executed=%t last_ms=%d error=%v",
			executed,
			actual.lastReconMS,
			reconErr,
		)
	}
}

func TestAccountReconFailurePublishesOnlyTelemetry(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var actual Account
	var err = actual.Init(simulatorFailureConfig(path, 6))
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&actual, warm); err != nil {
		t.Fatalf("warm Account: %v", err)
	}
	var price = decimal.NewFromInt(99)
	_, err = actual.PlaceOrders([]OrderSpec{{
		Role:        order.Entry,
		Side:        order.Buy,
		Type:        order.Limit,
		TimeInForce: order.GTC,
		Quantity:    decimal.RequireFromString("0.12"),
		Price:       &price,
		TimestampMS: 1000,
	}})
	if err != nil {
		t.Fatalf("place resting Order: %v", err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open result database: %v", err)
	}
	_, err = db.Exec(`
		CREATE TRIGGER fail_recon_cursor
		BEFORE UPDATE ON account_ledger
		BEGIN
			SELECT RAISE(FAIL, 'injected Recon failure');
		END;`)
	if err != nil {
		t.Fatalf("inject Recon failure: %v", err)
	}
	var snapshot Snapshot
	var refreshed bool
	var consecutiveFailures uint64
	snapshot, refreshed, consecutiveFailures, err = actual.Reconcile(1100, true)
	if err == nil {
		t.Fatal("failed Recon returned nil error")
	}
	var telemetry = actual.ReconciliationTelemetry()
	if refreshed || snapshot.ObservedMS != 0 || consecutiveFailures != 1 ||
		telemetry.Outcome != ReconFailed || telemetry.ConsecutiveFailures != 1 {
		t.Fatalf(
			"unexpected first failed Recon snapshot=%+v refreshed=%t failures=%d telemetry=%+v",
			snapshot,
			refreshed,
			consecutiveFailures,
			telemetry,
		)
	}
	snapshot, refreshed, consecutiveFailures, err = actual.Reconcile(1150, true)
	if err == nil {
		t.Fatal("second failed Recon returned nil error")
	}
	telemetry = actual.ReconciliationTelemetry()
	if refreshed || snapshot.ObservedMS != 0 || consecutiveFailures != 2 ||
		telemetry.Outcome != ReconFailed || telemetry.ConsecutiveFailures != 2 {
		t.Fatalf(
			"unexpected second failed Recon snapshot=%+v refreshed=%t failures=%d telemetry=%+v",
			snapshot,
			refreshed,
			consecutiveFailures,
			telemetry,
		)
	}
	if _, err = db.Exec(`DROP TRIGGER fail_recon_cursor`); err != nil {
		t.Fatalf("remove Recon failure: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close result database: %v", err)
	}
	snapshot, refreshed, consecutiveFailures, err = actual.Reconcile(1200, true)
	if err != nil {
		t.Fatalf("retry Recon: %v", err)
	}
	telemetry = actual.ReconciliationTelemetry()
	if !refreshed || snapshot.Generation != 1 || consecutiveFailures != 0 ||
		telemetry.Outcome != ReconSucceeded || telemetry.ConsecutiveFailures != 0 {
		t.Fatalf(
			"unexpected successful retry snapshot=%+v refreshed=%t telemetry=%+v",
			snapshot,
			refreshed,
			telemetry,
		)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestAccountMaxPersistenceRecoversDirtyVenueState(t *testing.T) {
	var cfg = Config{
		Nuubot: accountNuubot(meta.Instrument{
			Network:       "testnet",
			Kind:          "perp",
			Symbol:        "BTC",
			AssetID:       0,
			SizeDecimals:  5,
			PriceDecimals: 1,
		}, t.TempDir()+"/result.db"),
		LedgerID:       7,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Name:           "sim",
		Venue:          "simulator",
		Network:        "simnet",
		Symbol:         "BTC",
		EquityUSDC:     decimal.NewFromInt(1000),
		FeePct:         decimal.RequireFromString("0.035"),
		SlippagePct:    decimal.Zero,
		PersistMode:    "max",
	}
	var first Account
	var err = first.Init(cfg)
	if err != nil {
		t.Fatalf("initialize first Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&first, warm); err != nil {
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
	if err = restored.Init(cfg); err != nil {
		t.Fatalf("initialize restored Account: %v", err)
	}
	var restoredWarm, _ = market.CreateBBO(1100, 100)
	if err = ingestAccountBBO(&restored, restoredWarm); err != nil {
		t.Fatalf("warm restored Account: %v", err)
	}
	var snapshot Snapshot
	snapshot, _, _, err = restored.Reconcile(1100, false)
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
	if err = ingestAccountBBO(&restored, exitBBO); err != nil {
		t.Fatalf("ingest restored take-profit BBO: %v", err)
	}
	snapshot, _, _, err = restored.Reconcile(2000, false)
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
	err = mismatched.Init(cfg)
	if err == nil {
		t.Fatal("Simulator policy mismatch was admitted")
	}
}

func TestAccountMaxPersistenceRecoversSimulatorSubmitFailure(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var cfg = simulatorFailureConfig(path, 8)
	var first Account
	var err = first.Init(cfg)
	if err != nil {
		t.Fatalf("initialize first Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&first, warm); err != nil {
		t.Fatalf("warm first Account: %v", err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open result database: %v", err)
	}
	if _, err = db.Exec(`DROP TABLE simulator_venue_state`); err != nil {
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
	if !errors.Is(err, ErrNotSubmitted) {
		t.Fatalf("actual error %v, expected proven non-submission", err)
	}
	var result Result
	result, err = first.Result()
	if err != nil {
		t.Fatalf("read failed Account result: %v", err)
	}
	var failedOrder, orderErr = first.Order(1)
	if orderErr != nil || result.Ledger.Trades != 1 || result.Ledger.Orders != 1 ||
		failedOrder.Status != order.Error {
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
	if err = restored.Init(cfg); err != nil {
		t.Fatalf("initialize restored Account: %v", err)
	}
	var restoredWarm, _ = market.CreateBBO(1100, 100)
	if err = ingestAccountBBO(&restored, restoredWarm); err != nil {
		t.Fatalf("warm restored Account: %v", err)
	}
	var snapshot Snapshot
	snapshot, _, _, err = restored.Reconcile(1100, false)
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

func TestAccountDoesNotMarkAcceptedSimulatorOrderRetriable(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var cfg = simulatorFailureConfig(path, 9)
	var actual Account
	var err = actual.Init(cfg)
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&actual, warm); err != nil {
		t.Fatalf("warm Account: %v", err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open result database: %v", err)
	}
	if _, err = db.Exec(`
		CREATE TRIGGER fail_submitted_order
		BEFORE INSERT ON account_order
		WHEN NEW.status = 'submitted'
		BEGIN
			SELECT RAISE(FAIL, 'injected Ledger failure');
		END;`); err != nil {
		t.Fatalf("inject Ledger failure: %v", err)
	}
	var price = decimal.NewFromInt(99)
	var placed PlaceResult
	placed, err = actual.PlaceOrders([]OrderSpec{{
		Role:        order.Entry,
		Side:        order.Buy,
		Type:        order.Limit,
		TimeInForce: order.GTC,
		Quantity:    decimal.RequireFromString("0.12"),
		Price:       &price,
		TimestampMS: 1000,
	}})
	if err == nil {
		t.Fatal("Ledger persistence failure was admitted")
	}
	if placed.TradeID == 0 || errors.Is(err, ErrNotSubmitted) {
		t.Fatalf(
			"actual Trade ID=%d error=%v, expected uncertain accepted outcome",
			placed.TradeID,
			err,
		)
	}
	if _, err = db.Exec(`DROP TRIGGER fail_submitted_order`); err != nil {
		t.Fatalf("remove Ledger failure: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close result database: %v", err)
	}
	if _, _, _, err = actual.Reconcile(1100, true); err != nil {
		t.Fatalf("reconcile accepted Simulator Order: %v", err)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestAccountReconRetainsImmediateFillAfterSubmitPersistenceFailure(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var cfg = simulatorFailureConfig(path, 10)
	var actual Account
	var err = actual.Init(cfg)
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&actual, warm); err != nil {
		t.Fatalf("warm Account: %v", err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open result database: %v", err)
	}
	if _, err = db.Exec(`
		CREATE TRIGGER fail_submitted_order
		BEFORE INSERT ON account_order
		WHEN NEW.status = 'submitted'
		BEGIN
			SELECT RAISE(FAIL, 'injected Ledger failure');
		END;`); err != nil {
		t.Fatalf("inject Ledger failure: %v", err)
	}
	var price = decimal.NewFromInt(100)
	var placed PlaceResult
	placed, err = actual.PlaceOrders([]OrderSpec{{
		Role:        order.Entry,
		Side:        order.Buy,
		Type:        order.Limit,
		TimeInForce: order.IOC,
		Quantity:    decimal.RequireFromString("0.11"),
		Price:       &price,
		TimestampMS: 1000,
	}})
	if err == nil {
		t.Fatal("Ledger persistence failure was admitted")
	}
	if placed.TradeID == 0 || errors.Is(err, ErrNotSubmitted) {
		t.Fatalf(
			"actual Trade ID=%d error=%v, expected committed Venue Fill",
			placed.TradeID,
			err,
		)
	}
	if _, err = db.Exec(`DROP TRIGGER fail_submitted_order`); err != nil {
		t.Fatalf("remove Ledger failure: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close result database: %v", err)
	}
	var snapshot Snapshot
	snapshot, _, _, err = actual.Reconcile(1100, true)
	if err != nil {
		t.Fatalf("reconcile committed Venue Fill: %v", err)
	}
	var execution, found = actual.Fill(1)
	var filledOrder, orderErr = actual.Order(1)
	if !found || orderErr != nil || filledOrder.VenueOrderID != 1 ||
		filledOrder.Status != order.Filled || snapshot.Fills != 1 ||
		snapshot.PendingOrders != 0 || snapshot.PendingFills != 0 ||
		!snapshot.PositionQuantity.Equal(decimal.RequireFromString("0.11")) ||
		!snapshot.EntryPrice.Equal(price) || !snapshot.AccountValue.Equal(cfg.EquityUSDC) ||
		!execution.Quantity.Equal(decimal.RequireFromString("0.11")) ||
		!execution.Price.Equal(price) || !execution.HasFee ||
		!execution.Fee.IsZero() {
		t.Fatalf(
			"unexpected recovered Fill snapshot=%+v Order=%+v Fill=%+v",
			snapshot,
			filledOrder,
			execution,
		)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestEnrichFillCLOIDsRejectsAmbiguousVenueIdentity(t *testing.T) {
	var cases = []struct {
		name   string
		orders []ledger.OrderEvidence
		fills  []ledger.FillEvidence
	}{
		{
			name: "duplicate mapping",
			orders: []ledger.OrderEvidence{
				{CLOID: "first", VenueOrderID: 7},
				{CLOID: "first", VenueOrderID: 7},
			},
		},
		{
			name: "conflicting mapping",
			orders: []ledger.OrderEvidence{
				{CLOID: "first", VenueOrderID: 7},
				{CLOID: "second", VenueOrderID: 7},
			},
		},
		{
			name:   "conflicting Fill",
			orders: []ledger.OrderEvidence{{CLOID: "first", VenueOrderID: 7}},
			fills:  []ledger.FillEvidence{{CLOID: "second", VenueOrderID: 7}},
		},
	}
	for _, current := range cases {
		var err = enrichFillCLOIDs(current.orders, current.fills)
		if err == nil {
			t.Fatalf("%s ambiguity was admitted", current.name)
		}
	}
}

func TestAccountReconTelemetryReportsNoBulkOrderStatusQueries(t *testing.T) {
	var actual Account
	var err = actual.Init(accountTestConfig(34, "recon"))
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var price = decimal.NewFromInt(100)
	_, err = actual.PlaceOrders([]OrderSpec{{
		Role: order.Entry, Side: order.Buy, Type: order.Limit, TimeInForce: order.GTC,
		Quantity: decimal.RequireFromString("0.11"), Price: &price, TimestampMS: 1000,
	}})
	if err != nil {
		t.Fatalf("place open Order: %v", err)
	}
	if _, _, _, err = actual.Reconcile(1000, false); err != nil {
		t.Fatalf("reconcile open Order: %v", err)
	}
	var telemetry = actual.ReconciliationTelemetry()
	if telemetry.OrderStatusQueries != 0 {
		t.Fatalf("actual OrderStatus queries %d, expected 0", telemetry.OrderStatusQueries)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestAccountReconTelemetryPreservesFailedOrderStatusQuery(t *testing.T) {
	var actual Account
	var err = actual.Init(accountTestConfig(35, "recon"))
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var plan ledger.Plan
	plan, err = actual.ledger.PlanTrade(1)
	if err != nil {
		t.Fatalf("plan Trade: %v", err)
	}
	var price = decimal.NewFromInt(100)
	var created *order.Order
	created, err = order.New(order.Input{
		LedgerID: actual.config.LedgerID, TradeID: plan.Trade.TradeID,
		OrderID: plan.OrderIDs[0], Account: actual.config.Name,
		CycleNumber: actual.config.CycleNumber, Symbol: actual.config.Symbol,
		BatchNo: 1, OrderPos: 1, CLOID: "0x00000000000000000000000000000023",
		Role: order.Entry, Side: order.Buy, Type: order.Limit, TimeInForce: order.GTC,
		RequestedQuantity: decimal.RequireFromString("0.11"),
		RequestedPrice:    &price,
		TimestampMS:       1000,
	})
	if err != nil {
		t.Fatalf("create Order: %v", err)
	}
	if err = actual.ledger.CreateTrade(plan.Trade, []*order.Order{created}); err != nil {
		t.Fatalf("create Trade: %v", err)
	}
	if err = actual.ledger.RecordSubmit([]ledger.SubmitOutcome{{
		OrderID: plan.OrderIDs[0], VenueOrderID: 99,
	}}); err != nil {
		t.Fatalf("record submit: %v", err)
	}
	if _, _, _, err = actual.Reconcile(1000, false); err == nil {
		t.Fatal("missing submitted Venue Order returned nil error")
	}
	var telemetry = actual.ReconciliationTelemetry()
	if telemetry.Outcome != ReconFailed || telemetry.OrderStatusQueries != 1 {
		t.Fatalf("unexpected failed Recon telemetry: %+v", telemetry)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestAccountMaxPersistenceRepairsMissingSubmittingOrder(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var cfg = simulatorFailureConfig(path, 9)
	var first Account
	var err = first.Init(cfg)
	if err != nil {
		t.Fatalf("initialize first Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&first, warm); err != nil {
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
		DROP TABLE simulator_venue_state;`)
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
	var failedOrder, orderErr = first.Order(1)
	if orderErr != nil || failedOrder.Status != order.Error {
		t.Fatalf("unexpected untrusted in-memory Order state: %+v", failedOrder)
	}
	var storedStatus string
	if err = db.QueryRow(
		`SELECT status FROM account_order WHERE ledger_id = ? AND order_id = ?`,
		cfg.LedgerID,
		uint64(1),
	).Scan(&storedStatus); err != nil {
		t.Fatalf("read rolled-back Order state: %v", err)
	}
	if storedStatus != string(order.Created) {
		t.Fatalf("failed transaction changed durable Order state: %s", storedStatus)
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
	if err = restored.Init(cfg); err != nil {
		t.Fatalf("initialize restored Account: %v", err)
	}
	var restoredWarm, _ = market.CreateBBO(1100, 100)
	if err = ingestAccountBBO(&restored, restoredWarm); err != nil {
		t.Fatalf("warm restored Account: %v", err)
	}
	var snapshot Snapshot
	snapshot, _, _, err = restored.Reconcile(1100, false)
	if err != nil {
		t.Fatalf("repair missing submitting Order: %v", err)
	}
	result, err = restored.Result()
	if err != nil {
		t.Fatalf("read repaired Account result: %v", err)
	}
	var repairedOrder order.Record
	repairedOrder, orderErr = restored.Order(1)
	if orderErr != nil || repairedOrder.Status != order.Error ||
		snapshot.ActiveOrders != 0 || snapshot.OpenTrades != 0 {
		t.Fatalf("unexpected repaired Account state snapshot=%+v result=%+v", snapshot, result)
	}
	var telemetry = restored.ReconciliationTelemetry()
	if telemetry.OrderStatusQueries != 1 {
		t.Fatalf(
			"actual failed OrderStatus queries %d, expected 1",
			telemetry.OrderStatusQueries,
		)
	}
	if err = restored.Stop(); err != nil {
		t.Fatalf("stop restored Account: %v", err)
	}
}

func TestAccountRepairsCursorAdvancedMissingFee(t *testing.T) {
	var actual Account
	var err = actual.Init(accountTestConfig(20, "recon"))
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&actual, warm); err != nil {
		t.Fatalf("warm Account: %v", err)
	}
	var price = decimal.NewFromInt(100)
	_, err = actual.PlaceOrders([]OrderSpec{{
		Role: order.Entry, Side: order.Buy, Type: order.Limit, TimeInForce: order.IOC,
		Quantity: decimal.RequireFromString("0.11"), Price: &price, TimestampMS: 1000,
	}})
	if err != nil {
		t.Fatalf("place entry: %v", err)
	}
	setFillFeeAvailableForTest(t, &actual, 1, false)
	var snapshot Snapshot
	snapshot, _, _, err = actual.Reconcile(1000, false)
	if err != nil {
		t.Fatalf("reconcile missing fee: %v", err)
	}
	if snapshot.PendingOrders != 1 || snapshot.PendingFills != 1 {
		t.Fatalf("unexpected initial pending state: %+v", snapshot)
	}
	var later, _ = market.CreateBBO(3000, 101)
	if err = ingestAccountBBO(&actual, later); err != nil {
		t.Fatalf("advance mark: %v", err)
	}
	if _, _, _, err = actual.Reconcile(3000, false); err != nil {
		t.Fatalf("advance Fill cursor: %v", err)
	}
	if actual.ledger.FillsThroughMS() != 3000 {
		t.Fatalf("actual Fill cursor=%d, expected 3000", actual.ledger.FillsThroughMS())
	}
	setFillFeeAvailableForTest(t, &actual, 1, true)
	snapshot, _, _, err = actual.Reconcile(4000, false)
	if err != nil {
		t.Fatalf("repair delayed fee: %v", err)
	}
	var telemetry = actual.ReconciliationTelemetry()
	if snapshot.PendingOrders != 0 || snapshot.PendingFills != 0 ||
		len(telemetry.FillQueries) != 2 || telemetry.FillQueries[0].Kind != "discovery" ||
		telemetry.FillQueries[0].StartMS != 3000 || telemetry.FillQueries[1].Kind != "repair" ||
		telemetry.FillQueries[1].StartMS > 1000 || telemetry.FillQueries[1].EndMS < 1000 ||
		telemetry.FillQueries[1].FeesEnriched != 1 || telemetry.FillQueries[1].PendingMatched != 1 {
		t.Fatalf("unexpected repaired snapshot=%+v telemetry=%+v", snapshot, telemetry)
	}
	var execution, found = actual.Fill(1)
	if !found || !execution.HasFee || !execution.Fee.Equal(decimal.RequireFromString("0.00385")) {
		t.Fatalf("unexpected enriched Fill: %+v", execution)
	}
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

func TestAccountKeepsFeeIncompleteClosurePendingAndFinalFinanceStatic(t *testing.T) {
	var actual Account
	var err = actual.Init(accountTestConfig(21, "recon"))
	if err != nil {
		t.Fatalf("initialize Account: %v", err)
	}
	var warm, _ = market.CreateBBO(1000, 100)
	if err = ingestAccountBBO(&actual, warm); err != nil {
		t.Fatalf("warm Account: %v", err)
	}
	var quantity = decimal.RequireFromString("0.11")
	var entry = decimal.NewFromInt(100)
	var exit = decimal.NewFromInt(110)
	_, err = actual.PlaceOrders([]OrderSpec{
		{Role: order.Entry, Side: order.Buy, Type: order.Limit, TimeInForce: order.IOC,
			Quantity: quantity, Price: &entry, TimestampMS: 1000},
		{Role: order.Stop, Side: order.Sell, Type: order.Limit, TimeInForce: order.GTC,
			Quantity: quantity, Price: &exit, ReduceOnly: true, TimestampMS: 1000},
	})
	if err != nil {
		t.Fatalf("place entry and exit: %v", err)
	}
	if _, _, _, err = actual.Reconcile(1000, false); err != nil {
		t.Fatalf("reconcile entry: %v", err)
	}
	var closeBBO, _ = market.CreateBBO(2000, 111)
	if err = ingestAccountBBO(&actual, closeBBO); err != nil {
		t.Fatalf("fill exit: %v", err)
	}
	setFillFeeAvailableForTest(t, &actual, 2, false)
	var pending Snapshot
	pending, _, _, err = actual.Reconcile(2000, false)
	if err != nil {
		t.Fatalf("reconcile fee-incomplete exit: %v", err)
	}
	var pendingTrade, tradeErr = actual.Trade(1)
	var pendingOrder, orderErr = actual.Order(2)
	if tradeErr != nil || orderErr != nil || pending.PendingOrders != 1 || pending.PendingFills != 1 ||
		pendingTrade.Status != trade.Closing || !pendingTrade.UnrealizedPnL.IsZero() ||
		pendingOrder.Status == order.Filled {
		t.Fatalf("closure appeared complete snapshot=%+v Trade=%+v", pending, pendingTrade)
	}
	setFillFeeAvailableForTest(t, &actual, 2, true)
	var closed Snapshot
	closed, _, _, err = actual.Reconcile(3000, false)
	if err != nil {
		t.Fatalf("finalize closure: %v", err)
	}
	var closedTrade trade.ReconState
	closedTrade, tradeErr = actual.Trade(1)
	if tradeErr != nil || closedTrade.Status != trade.Closed || !closedTrade.UnrealizedPnL.IsZero() ||
		closed.PendingOrders != 0 || closed.PendingFills != 0 {
		t.Fatalf("unexpected closed finance snapshot=%+v Trade=%+v", closed, closedTrade)
	}
	var changedMark, _ = market.CreateBBO(4000, 150)
	if err = ingestAccountBBO(&actual, changedMark); err != nil {
		t.Fatalf("change closed mark: %v", err)
	}
	if _, _, _, err = actual.Reconcile(4000, true); err != nil {
		t.Fatalf("force closed Recon: %v", err)
	}
	var finalTrade trade.ReconState
	finalTrade, tradeErr = actual.Trade(1)
	if tradeErr != nil {
		t.Fatalf("read final Trade: %v", tradeErr)
	}
	assertTradeFinanceEqual(t, finalTrade, closedTrade)
	if err = actual.Stop(); err != nil {
		t.Fatalf("stop Account: %v", err)
	}
}

// Section 2 - Domain Helpers

func accountTestConfig(ledgerID uint64, recon string) Config {
	return Config{
		Nuubot: accountNuubot(meta.Instrument{
			Network: "testnet", Kind: "perp", Symbol: "BTC",
			AssetID: 0, SizeDecimals: 5, PriceDecimals: 1,
		}, ""),
		LedgerID: ledgerID, CycleNumber: 2, ExecutorNumber: 3, Name: "sim",
		Venue: "simulator", Network: "simnet", Symbol: "BTC",
		EquityUSDC: decimal.NewFromInt(1000),
		FeePct:     decimal.RequireFromString("0.035"), SlippagePct: decimal.Zero,
		PersistMode: "none", Recon: recon,
	}
}

func accountNuubot(
	instrument meta.Instrument,
	resultPath string,
) *setup.Nuubot {
	return &setup.Nuubot{
		Log: logging.Create(io.Discard),
		App: appconfig.App{
			Hyperliquid: appconfig.Hyperliquid{MinOrderNotionalUSDC: 11},
		},
		Runtime: appconfig.Runtime{
			ControllerIntervalMS: 1,
			ReconIntervalMS:      2,
			ReconSweepIntervalMS: 1 << 63,
			TelemetryIntervalMS:  10,
		},
		MarketData:  market.CreateMarketData(),
		Meta:        instrument,
		ResultPath:  resultPath,
		RuntimePath: resultPath,
	}
}

func ingestAccountBBO(actual *Account, bbo market.BBO) error {
	return actual.config.Nuubot.MarketData.IngestBBO(market.Key{
		Venue:   actual.config.Venue,
		Network: actual.config.Network,
		Symbol:  actual.config.Symbol,
	}, bbo)
}

func assertTradeFinanceEqual(t *testing.T, actual trade.ReconState, expected trade.ReconState) {
	t.Helper()
	if actual.Status != expected.Status || !actual.RealizedPnL.Equal(expected.RealizedPnL) ||
		!actual.UnrealizedPnL.Equal(expected.UnrealizedPnL) ||
		!actual.GrossPnL.Equal(expected.GrossPnL) || !actual.Fees.Equal(expected.Fees) ||
		!actual.NetPnL.Equal(expected.NetPnL) {
		t.Fatalf("Trade finance mismatch actual=%+v expected=%+v", actual, expected)
	}
}

func simulatorFailureConfig(path string, ledgerID uint64) Config {
	return Config{
		Nuubot: accountNuubot(meta.Instrument{
			Network:       "mainnet",
			Kind:          "perp",
			Symbol:        "BTC",
			AssetID:       0,
			SizeDecimals:  5,
			PriceDecimals: 1,
		}, path),
		LedgerID:       ledgerID,
		CycleNumber:    2,
		ExecutorNumber: 3,
		Name:           "sim",
		Venue:          "simulator",
		Network:        "simnet",
		Symbol:         "BTC",
		EquityUSDC:     decimal.NewFromInt(1000),
		FeePct:         decimal.Zero,
		SlippagePct:    decimal.Zero,
		PersistMode:    "max",
	}
}

func setFillFeeAvailableForTest(
	t *testing.T,
	actual *Account,
	venueTID uint64,
	available bool,
) {
	t.Helper()
	var controller, supported = actual.venue.(interface {
		SetFillFeeAvailableForTest(uint64, bool) error
	})
	if !supported {
		t.Fatal("Venue does not support delayed-fee test control")
	}
	var err = controller.SetFillFeeAvailableForTest(venueTID, available)
	if err != nil {
		t.Fatalf("set Fill fee availability: %v", err)
	}
}

func restoreSimulatorTable(db *sql.DB) error {
	var _, err = db.Exec(`
		CREATE TABLE simulator_venue_state (
			account_name    TEXT NOT NULL,
			symbol          TEXT NOT NULL,
			schema_version  INTEGER NOT NULL,
			payload_json    TEXT NOT NULL,
			updated_ms      INTEGER NOT NULL,
			PRIMARY KEY (account_name, symbol)
		)`)
	return err
}

// Section 3 - Generic Helpers
