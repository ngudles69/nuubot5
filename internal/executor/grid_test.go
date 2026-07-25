package executor

import (
	"bytes"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/cloid"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/order"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestGridExecutorRunsAllLevelsAndFlattensAtBound(t *testing.T) {
	var ctx = gridExecutorContext()
	var created, err = Create(ctx)
	if err != nil {
		t.Fatalf("create GridExecutor: %v", err)
	}
	var actual = created.(*gridExecutor)
	if err = actual.OnStart(); err != nil {
		t.Fatalf("start GridExecutor: %v", err)
	}
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if err = actual.OnRecon(3_000); err != nil {
		t.Fatalf("submit initial Grid: %v", err)
	}
	var initial, resultErr = actual.account.Result()
	if resultErr != nil {
		t.Fatalf("read initial Grid result: %v", resultErr)
	}
	if len(initial.Ledger.Trades) != 8 || countOrders(initial) != 16 {
		t.Fatalf(
			"actual trades=%d orders=%d, expected trades=8 orders=16",
			len(initial.Ledger.Trades),
			countOrders(initial),
		)
	}
	for _, owned := range initial.Ledger.Trades {
		for _, submitted := range owned.Orders {
			var identity cloid.Fields
			identity, err = cloid.Decode(submitted.CLOID)
			if err != nil {
				t.Fatalf("decode Grid CLOID: %v", err)
			}
			if identity.OrderLevel == 0 || identity.OrderLevel >= 9 {
				t.Fatalf("actual Order level %d, expected active level", identity.OrderLevel)
			}
			if !submitted.RequestedQuantity.Equal(
				actual.levels[identity.OrderLevel].Quantity,
			) {
				t.Fatalf(
					"actual normalized quantity %s, expected stored quantity %s",
					submitted.RequestedQuantity,
					actual.levels[identity.OrderLevel].Quantity,
				)
			}
		}
	}

	var middle = market.BBO{TimestampMS: 4_000, Price: 100}
	if err = actual.IngestBBO(middle); err != nil {
		t.Fatalf("ingest middle BBO: %v", err)
	}
	actual.OnBBO(middle)
	var snapshot, _, reconErr = actual.Reconcile(4_000, false)
	if reconErr != nil {
		t.Fatalf("reconcile middle BBO: %v", reconErr)
	}
	if snapshot.Fills != 4 {
		t.Fatalf("actual initial fills %d, expected 4", snapshot.Fills)
	}

	var adverse = market.BBO{TimestampMS: 5_000, Price: 94}
	if err = actual.IngestBBO(adverse); err != nil {
		t.Fatalf("ingest adverse BBO: %v", err)
	}
	actual.OnBBO(adverse)
	if actual.Status() != Stopping || actual.ExitReason() != "stop_loss" {
		t.Fatalf(
			"actual status=%s reason=%s, expected stopping stop_loss",
			actual.Status(),
			actual.ExitReason(),
		)
	}
	if err = actual.OnStop("completed"); err != nil {
		t.Fatalf("stop GridExecutor: %v", err)
	}
	var executorResult Result
	executorResult, err = actual.Result()
	if err != nil {
		t.Fatalf("read GridExecutor result: %v", err)
	}
	var result = *executorResult.Account
	var stopOrders int
	for _, owned := range result.Ledger.Trades {
		for _, submitted := range owned.Orders {
			if submitted.Role == order.Stop {
				stopOrders++
			}
		}
	}
	if !result.Snapshot.PositionQuantity.IsZero() ||
		result.Snapshot.ActiveOrders != 0 ||
		executorResult.ClosureOrders != 8 ||
		stopOrders != 8 ||
		len(executorResult.Levels) != 10 {
		t.Fatalf("unexpected terminal Grid result: %+v", executorResult)
	}
}

func TestGridCalculationKeepsInitialEntryInsideCapitalSlice(t *testing.T) {
	var spec = gridExecutorContext().Spec
	var levels, err = calculateGridLevels(
		spec,
		market.BBO{TimestampMS: 3_000, Price: 100},
		decimal.NewFromInt(100),
	)
	if err != nil {
		t.Fatalf("calculate Grid levels: %v", err)
	}
	var slice = decimal.RequireFromString("11.875")
	if len(levels) != 10 ||
		!levels[0].GridPrice.Equal(decimal.NewFromInt(95)) ||
		!levels[9].GridPrice.Equal(decimal.NewFromInt(105)) {
		t.Fatalf("unexpected Grid geometry: %+v", levels)
	}
	for index := 1; index < len(levels)-1; index++ {
		var level = levels[index]
		var deployed = level.InitialNotional.Add(level.InitialEntryCommission)
		var redeployed = level.ReentryNotional.Add(level.ReentryCommission)
		if deployed.GreaterThan(slice) ||
			redeployed.GreaterThan(slice) ||
			level.InitialNotional.LessThan(decimal.NewFromInt(11)) ||
			!level.InitialExpectedPnL.IsPositive() ||
			!level.ReentryExpectedPnL.IsPositive() {
			t.Fatalf("invalid calculated level: %+v", level)
		}
	}
}

func TestShortGridSubmitsTopDown(t *testing.T) {
	var ctx = gridExecutorContext()
	ctx.Spec.Side = Short
	var created, err = Create(ctx)
	if err != nil {
		t.Fatalf("create short GridExecutor: %v", err)
	}
	var actual = created.(*gridExecutor)
	if err = actual.OnStart(); err != nil {
		t.Fatalf("start short GridExecutor: %v", err)
	}
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile short Account: %v", err)
	}
	if err = actual.OnRecon(3_000); err != nil {
		t.Fatalf("submit short Grid: %v", err)
	}
	var result account.Result
	result, err = actual.account.Result()
	if err != nil {
		t.Fatalf("read short Grid result: %v", err)
	}
	var expected = uint16(8)
	for _, owned := range result.Ledger.Trades {
		var identity cloid.Fields
		identity, err = cloid.Decode(owned.Orders[0].CLOID)
		if err != nil {
			t.Fatalf("decode short Grid CLOID: %v", err)
		}
		if identity.OrderLevel != expected {
			t.Fatalf(
				"actual short level %d, expected %d",
				identity.OrderLevel,
				expected,
			)
		}
		expected--
	}
}

func TestGridCountsBoundaryTickRoundTrips(t *testing.T) {
	var created, err = Create(gridExecutorContext())
	if err != nil {
		t.Fatalf("create GridExecutor: %v", err)
	}
	var actual = created.(*gridExecutor)
	if err = actual.OnStart(); err != nil {
		t.Fatalf("start GridExecutor: %v", err)
	}
	if _, _, err = actual.Reconcile(3_000, false); err != nil {
		t.Fatalf("reconcile initial Account: %v", err)
	}
	if err = actual.OnRecon(3_000); err != nil {
		t.Fatalf("submit initial Grid: %v", err)
	}
	var entry = market.BBO{TimestampMS: 4_000, Price: 99}
	if err = actual.IngestBBO(entry); err != nil {
		t.Fatalf("ingest entry BBO: %v", err)
	}
	actual.OnBBO(entry)
	var boundary = market.BBO{TimestampMS: 5_000, Price: 105}
	if err = actual.IngestBBO(boundary); err != nil {
		t.Fatalf("ingest boundary BBO: %v", err)
	}
	actual.OnBBO(boundary)
	if err = actual.OnStop("completed"); err != nil {
		t.Fatalf("stop GridExecutor: %v", err)
	}
	var result Result
	result, err = actual.Result()
	if err != nil {
		t.Fatalf("read GridExecutor result: %v", err)
	}
	if result.RoundTrips != 5 {
		t.Fatalf(
			"actual boundary round trips %d, expected 5",
			result.RoundTrips,
		)
	}
}

func TestGridCalculationRejectsUnprofitableLevel(t *testing.T) {
	var spec = gridExecutorContext().Spec
	spec.MinExpectedPnL = decimal.NewFromInt(1)
	var _, err = calculateGridLevels(
		spec,
		market.BBO{TimestampMS: 3_000, Price: 100},
		spec.CapitalUSDC,
	)
	if err == nil {
		t.Fatal("unprofitable Grid was accepted")
	}
}

// Section 2 - Domain Helpers

func gridExecutorContext() Context {
	return Context{
		Log:                logging.Create(&bytes.Buffer{}),
		CycleNumber:        1,
		ExecutorNumber:     1,
		SignalTimestampMS:  2_000,
		LatestBBO:          market.BBO{TimestampMS: 3_000, Price: 100},
		StartingEquityUSDC: decimal.NewFromInt(100),
		Spec: Spec{
			ID:   "grid",
			Role: "grid",
			Kind: "grid",
			Side: Long,
			Resource: Resource{
				Venue:             "simulator",
				Network:           "simnet",
				PhysicalAccountID: "sim",
				Symbol:            "BTC",
			},
			CapitalUSDC:     decimal.NewFromInt(100),
			GridLevels:      10,
			RangePct:        decimal.RequireFromString("0.05"),
			MinExpectedPnL:  decimal.Zero,
			FeePct:          decimal.RequireFromString("0.05"),
			SlippagePct:     decimal.Zero,
			PersistMode:     "none",
			MinNotionalUSDC: decimal.NewFromInt(11),
			Meta: meta.Instrument{
				Network:       "testnet",
				Kind:          "perp",
				Symbol:        "BTC",
				AssetID:       0,
				SizeDecimals:  5,
				PriceDecimals: 1,
			},
		},
	}
}

// Section 3 - Generic Helpers
