package botspec

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
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
	var templates = map[string]string{
		"macross_observer_v1.toml": MacrossObserver,
		"macross_trade_v1.toml":    MacrossTrade,
		"macross_grid_v1.toml":     MacrossGrid,
	}
	for name, botSpecID := range templates {
		var content, err = os.ReadFile(botTemplatePath(t, name))
		if err != nil {
			t.Fatalf("read %s template: %v", name, err)
		}
		var identity struct {
			BotSpec string `toml:"bot_spec"`
		}
		if _, err = toml.Decode(string(content), &identity); err != nil {
			t.Fatalf("decode %s template identity: %v", name, err)
		}
		if identity.BotSpec != botSpecID {
			t.Fatalf("%s BotSpecID = %q, want %q", name, identity.BotSpec, botSpecID)
		}
		if err = Validate(botSpecID, string(content)); err != nil {
			t.Fatalf("validate %s template: %v", name, err)
		}
	}
}

func TestReconSelectionValues(t *testing.T) {
	var content, err = os.ReadFile(botTemplatePath(t, "macross_grid_v1.toml"))
	if err != nil {
		t.Fatalf("read Grid template: %v", err)
	}
	for _, value := range []string{"recon", "recon2"} {
		var config = strings.Replace(string(content), `recon = "recon"`, `recon = "`+value+`"`, 1)
		var admitted, admitErr = admit(MacrossGrid, config)
		if admitErr != nil {
			t.Fatalf("admit %s: %v", value, admitErr)
		}
		if admitted.executors[0].Recon != value {
			t.Fatalf("admitted Recon = %q, want %q", admitted.executors[0].Recon, value)
		}
	}
	var invalid = strings.Replace(string(content), `recon = "recon"`, `recon = "invalid"`, 1)
	if err = Validate(MacrossGrid, invalid); err == nil {
		t.Fatal("invalid reconciliation selection was accepted")
	}
}

func TestGridLevelCountUsesCompleteTenBitRange(t *testing.T) {
	var content, err = os.ReadFile(botTemplatePath(t, "macross_grid_v1.toml"))
	if err != nil {
		t.Fatalf("read Grid template: %v", err)
	}
	var maximum = strings.Replace(string(content), "levels = 30", "levels = 1024", 1)
	if err = Validate(MacrossGrid, maximum); err != nil {
		t.Fatalf("validate 1024 Grid levels: %v", err)
	}
	var excessive = strings.Replace(string(content), "levels = 30", "levels = 1025", 1)
	if err = Validate(MacrossGrid, excessive); err == nil {
		t.Fatal("1025 Grid levels were accepted")
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

func botTemplatePath(t *testing.T, name string) string {
	t.Helper()
	var _, file, _, valid = runtime.Caller(0)
	if !valid {
		t.Fatal("locate test source")
	}
	var root = filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "templates", "bots", name)
}
