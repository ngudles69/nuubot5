package executor

import (
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/botspec"
	appconfig "nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/setup"
)

// Section 1 - Program Flow

func TestGridCalculationKeepsInitialEntryInsideCapitalSlice(t *testing.T) {
	var nuubot, spec = gridCalculationInputs()
	var levels, err = calculateGridLevels(
		nuubot,
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

func TestGridCalculationRejectsUnprofitableLevel(t *testing.T) {
	var nuubot, spec = gridCalculationInputs()
	spec.MinExpectedPnL = decimal.NewFromInt(1)
	var _, err = calculateGridLevels(
		nuubot,
		spec,
		market.BBO{TimestampMS: 3_000, Price: 100},
		spec.CapitalUSDC,
	)
	if err == nil {
		t.Fatal("unprofitable Grid was accepted")
	}
}

// Section 2 - Domain Helpers

func gridCalculationInputs() (*setup.Nuubot, botspec.ExecutorSpec) {
	return &setup.Nuubot{
			App: appconfig.App{
				Hyperliquid: appconfig.Hyperliquid{MinOrderNotionalUSDC: 11},
			},
			Meta: meta.Instrument{
				Network:       "testnet",
				Kind:          "perp",
				Symbol:        "BTC",
				AssetID:       0,
				SizeDecimals:  5,
				PriceDecimals: 1,
			},
		}, botspec.ExecutorSpec{
			ID:   "grid",
			Role: "grid",
			Kind: "grid",
			Side: Long,
			Resource: botspec.Resource{
				Venue:             "simulator",
				Network:           "simnet",
				PhysicalAccountID: "sim",
				Symbol:            "BTC",
			},
			CapitalUSDC:    decimal.NewFromInt(100),
			GridLevels:     10,
			RangePct:       decimal.RequireFromString("0.05"),
			MinExpectedPnL: decimal.Zero,
			FeePct:         decimal.RequireFromString("0.05"),
			SlippagePct:    decimal.Zero,
			PersistMode:    "none",
		}
}

// Section 3 - Generic Helpers
