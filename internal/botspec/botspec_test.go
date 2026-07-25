package botspec

import (
	"os"
	"path/filepath"
	"testing"
)

// Section 1 - Program Flow

func TestValidateAllowsExtraFields(t *testing.T) {
	var configTOML = validTradeConfig() + "\nfuture_field = \"stored but ignored\"\n"
	var err = Validate("macross_trade_bot", configTOML)
	if err != nil {
		t.Fatalf("validate Config with extra field: %v", err)
	}
}

func TestValidateRejectsMissingRequiredField(t *testing.T) {
	var err = Validate(
		"macross_trade_bot",
		`bot_spec = "macross_trade_bot"`,
	)
	if err == nil {
		t.Fatal("Config without required fields was accepted")
	}
}

func TestValidateRejectsDuplicateResource(t *testing.T) {
	var configTOML = validTradeConfig() + `

[[executors]]
id = "duplicate"
role = "hedge"
kind = "trade"
side = "short"
venue = "simulator"
network = "simnet"
physical_account_id = "sbacktest"
symbol = "BTC"
capital_usdc = "1000"
order_size_usdc = "11"
take_profit_pct = "0.01"
stop_loss_pct = "0.01"
fee_pct = "0.05"
slippage_pct = "0.02"
persist_mode = "none"
`
	var err = Validate("macross_trade_bot", configTOML)
	if err == nil {
		t.Fatal("duplicate Account-symbol resource was accepted")
	}
}

func TestCanonicalTemplatesValidate(t *testing.T) {
	for _, botSpecID := range []string{MacrossObserver, MacrossTrade} {
		var content, err = os.ReadFile(filepath.Join("templates", botSpecID+".toml"))
		if err != nil {
			t.Fatalf("read %s template: %v", botSpecID, err)
		}
		if err = Validate(botSpecID, string(content)); err != nil {
			t.Fatalf("validate %s template: %v", botSpecID, err)
		}
	}
}

// Section 2 - Domain Helpers

func validTradeConfig() string {
	return `
bot_spec = "macross_trade_bot"

[controller]
max_cycles = 999

[signaler]
signal_timeframe = "1h"
regime_timeframe = "4h"
fast_ma = 9
slow_ma = 21
regime_ema = 200

[[executors]]
id = "trade"
role = "trade"
kind = "trade"
side = "long"
venue = "simulator"
network = "simnet"
physical_account_id = "sbacktest"
symbol = "BTC"
capital_usdc = "1000"
order_size_usdc = "11"
take_profit_pct = "0.01"
stop_loss_pct = "0.01"
fee_pct = "0.05"
slippage_pct = "0.02"
persist_mode = "none"

[[risks]]
kind = "balanced"
`
}

// Section 3 - Generic Helpers
