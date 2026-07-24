package executor

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/order"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestTradeExecutorRunsOneBracket(t *testing.T) {
	var actual = newTradeExecutor(t, true, 100)
	var err error
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if err = actual.OnRecon(3_000); err != nil {
		t.Fatalf("submit bracket: %v", err)
	}
	var entrySnapshot account.Snapshot
	entrySnapshot, _, err = actual.Reconcile(3_000, false)
	if err != nil {
		t.Fatalf("reconcile entry: %v", err)
	}
	if !entrySnapshot.PositionQuantity.IsPositive() {
		t.Fatalf("entry position was not reconciled: %+v", entrySnapshot)
	}
	var takeProfit = market.BBO{TimestampMS: 4_000, Price: 111}
	if err = actual.IngestBBO(takeProfit); err != nil {
		t.Fatalf("ingest take-profit BBO: %v", err)
	}
	actual.OnBBO(takeProfit)
	if _, _, err = actual.Reconcile(4_000, false); err != nil {
		t.Fatalf("reconcile take-profit: %v", err)
	}
	if err = actual.OnRecon(4_000); err != nil {
		t.Fatalf("complete bracket: %v", err)
	}
	if actual.Status() != Stopping {
		t.Fatalf("actual status %q, expected %q", actual.Status(), Stopping)
	}
	if err = actual.OnStop("completed"); err != nil {
		t.Fatalf("stop TradeExecutor: %v", err)
	}
	var result account.Result
	result, err = actual.AccountResult()
	if err != nil {
		t.Fatalf("read Account result: %v", err)
	}
	if len(result.Ledger.Trades) != 1 ||
		len(result.Ledger.Trades[0].Orders) != 3 ||
		result.Ledger.Trades[0].Status != "closed" ||
		result.Ledger.Trades[0].Orders[0].Role != order.Entry ||
		len(result.Simulator.Fills) != 2 {
		t.Fatalf("unexpected TradeExecutor result: %+v", result)
	}
}

func TestTradeExecutorStopClosesOpenPosition(t *testing.T) {
	var actual = newTradeExecutor(t, true, 100)
	var err error
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if err = actual.OnRecon(3_000); err != nil {
		t.Fatalf("submit bracket: %v", err)
	}
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile entry: %v", err)
	}
	var later = market.BBO{TimestampMS: 4_000, Price: 100}
	if err = actual.IngestBBO(later); err != nil {
		t.Fatalf("ingest later BBO: %v", err)
	}
	actual.OnBBO(later)
	if err = actual.OnStop("parent_stop"); err != nil {
		t.Fatalf("stop open TradeExecutor: %v", err)
	}
	var result account.Result
	result, err = actual.AccountResult()
	if err != nil {
		t.Fatalf("read Account result: %v", err)
	}
	if len(result.Simulator.Fills) != 2 ||
		!result.Ledger.Trades[0].OpenQuantity.IsZero() {
		t.Fatalf("unexpected shutdown result: %+v", result)
	}
}

func TestTradeExecutorStopClosesRoundedShortPosition(t *testing.T) {
	var actual = newTradeExecutor(t, false, 104241)
	var err error
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if err = actual.OnRecon(3_000); err != nil {
		t.Fatalf("submit bracket: %v", err)
	}
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile entry: %v", err)
	}
	if err = actual.OnStop("parent_stop"); err != nil {
		t.Fatalf("stop rounded short TradeExecutor: %v", err)
	}
	var result account.Result
	result, err = actual.AccountResult()
	if err != nil {
		t.Fatalf("read Account result: %v", err)
	}
	if len(result.Simulator.Fills) != 2 ||
		!result.Ledger.Trades[0].OpenQuantity.IsZero() {
		t.Fatalf("unexpected rounded short shutdown result: %+v", result)
	}
}

func TestTradeExecutorRejectsPersistedTradeUntilRunnerRecoveryExists(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "result.db")
	var ctx = tradeExecutorContext(t, true, 100, "max", path)
	var created, err = Create(ctx)
	if err != nil {
		t.Fatalf("create first TradeExecutor: %v", err)
	}
	var first = created.(*tradeExecutor)
	if _, _, err = first.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile first Account: %v", err)
	}
	if err = first.OnRecon(3_000); err != nil {
		t.Fatalf("submit persisted bracket: %v", err)
	}
	if _, _, err = first.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile persisted entry: %v", err)
	}
	if err = first.account.Stop(); err != nil {
		t.Fatalf("stop first Account: %v", err)
	}

	_, err = Create(ctx)
	if err == nil || !strings.Contains(err.Error(), "recovery is pending Runner") {
		t.Fatalf("unexpected recovered Trade result: %v", err)
	}
}

// Section 2 - Domain Helpers

func newTradeExecutor(t *testing.T, enterLong bool, price float64) *tradeExecutor {
	t.Helper()
	var ctx = tradeExecutorContext(t, enterLong, price, "none", "")
	var created, err = Create(ctx)
	if err != nil {
		t.Fatalf("create TradeExecutor: %v", err)
	}
	return created.(*tradeExecutor)
}

func tradeExecutorContext(
	t *testing.T,
	enterLong bool,
	price float64,
	persistMode string,
	resultPath string,
) Context {
	t.Helper()
	var initial = market.BBO{TimestampMS: 3_000, Price: price}
	return Context{
		Log:            logging.Create(&bytes.Buffer{}),
		CycleNumber:    1,
		ExecutorNumber: 1,
		Signal:         testSignal(t, enterLong, !enterLong),
		LatestBBO:      initial,
		Meta: meta.Instrument{
			Network:       "testnet",
			Kind:          "perp",
			Symbol:        "BTC",
			AssetID:       0,
			SizeDecimals:  5,
			PriceDecimals: 1,
		},
		MinNotional: decimal.NewFromInt(11),
		Config: config.Executor{
			Kind:                 "trade",
			AccountName:          "sim",
			Network:              "simnet",
			OrderNotionalUSDC:    "11",
			TakeProfitPct:        "0.10",
			StopLossPct:          "0.10",
			SimulatorEquityUSDC:  "1000",
			SimulatorFeePct:      "0.035",
			SimulatorSlippagePct: "0",
			PersistMode:          persistMode,
		},
		ResultPath: resultPath,
	}
}

// Section 3 - Generic Helpers
