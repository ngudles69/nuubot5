package executor

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/botspec"
	"nuubot/internal/market"
)

// Section 1 - Program Flow

func TestTradeExecutorRunsOneBracket(t *testing.T) {
	var actual = newTradeExecutor(t, true, 100)
	if actual.Telemetry().Account != nil {
		t.Fatal("unobserved Account telemetry was reported")
	}
	var err error
	if _, _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if actual.Telemetry().Account == nil {
		t.Fatal("observed Account telemetry is missing")
	}
	if err = actual.OnRecon(); err != nil {
		t.Fatalf("submit bracket: %v", err)
	}
	var entrySnapshot account.Snapshot
	entrySnapshot, _, _, err = actual.Reconcile(3_000, false)
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
	if _, _, _, err = actual.Reconcile(4_000, false); err != nil {
		t.Fatalf("reconcile take-profit: %v", err)
	}
	if err = actual.OnRecon(); err != nil {
		t.Fatalf("complete bracket: %v", err)
	}
	if actual.Status() != Stopping {
		t.Fatalf("actual status %q, expected %q", actual.Status(), Stopping)
	}
	if err = actual.OnStop("completed"); err != nil {
		t.Fatalf("stop TradeExecutor: %v", err)
	}
	var executorResult Result
	executorResult, err = actual.Result()
	if err != nil {
		t.Fatalf("read Executor result: %v", err)
	}
	var result = *executorResult.Account
	if result.Ledger.Trades != 1 || result.Ledger.Orders != 3 ||
		result.Ledger.Summary.OpenTrades != 0 || result.Ledger.Fills != 2 {
		t.Fatalf("unexpected TradeExecutor result: %+v", result)
	}
}

func TestTradeExecutorStopClosesOpenPosition(t *testing.T) {
	var actual = newTradeExecutor(t, true, 100)
	var err error
	if _, _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if err = actual.OnRecon(); err != nil {
		t.Fatalf("submit bracket: %v", err)
	}
	if _, _, _, err = actual.Reconcile(3_000, false); err != nil {
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
	var executorResult Result
	executorResult, err = actual.Result()
	if err != nil {
		t.Fatalf("read Executor result: %v", err)
	}
	var result = *executorResult.Account
	if result.Ledger.Fills != 2 || result.Ledger.Summary.OpenTrades != 0 {
		t.Fatalf("unexpected shutdown result: %+v", result)
	}
}

func TestTradeExecutorStopClosesRoundedShortPosition(t *testing.T) {
	var actual = newTradeExecutor(t, false, 104241)
	var err error
	if _, _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if err = actual.OnRecon(); err != nil {
		t.Fatalf("submit bracket: %v", err)
	}
	if _, _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile entry: %v", err)
	}
	if err = actual.OnStop("parent_stop"); err != nil {
		t.Fatalf("stop rounded short TradeExecutor: %v", err)
	}
	var executorResult Result
	executorResult, err = actual.Result()
	if err != nil {
		t.Fatalf("read Executor result: %v", err)
	}
	var result = *executorResult.Account
	if result.Ledger.Fills != 2 || result.Ledger.Summary.OpenTrades != 0 {
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
	if _, _, _, err = first.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile first Account: %v", err)
	}
	if err = first.OnRecon(); err != nil {
		t.Fatalf("submit persisted bracket: %v", err)
	}
	if _, _, _, err = first.Reconcile(3_000, false); err != nil {
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
		Nuubot:         executorNuubot(t, resultPath),
		CycleNumber:    1,
		ExecutorNumber: 1,
		Signal:         executorTestSignal(t),
		LatestBBO:      initial,
		Spec: botspec.ExecutorSpec{
			ID:   "trade",
			Kind: "trade",
			Side: map[bool]string{true: Long, false: Short}[enterLong],
			Resource: botspec.Resource{
				Venue:             "simulator",
				Network:           "simnet",
				PhysicalAccountID: "sim",
				Symbol:            "BTC",
			},
			CapitalUSDC:   decimal.NewFromInt(1000),
			OrderSizeUSDC: decimal.NewFromInt(11),
			TakeProfitPct: decimal.RequireFromString("0.10"),
			StopLossPct:   decimal.RequireFromString("0.10"),
			FeePct:        decimal.RequireFromString("0.035"),
			SlippagePct:   decimal.Zero,
			PersistMode:   persistMode,
		},
	}
}

// Section 3 - Generic Helpers
