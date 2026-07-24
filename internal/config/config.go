package config

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/shopspring/decimal"
)

type Config struct {
	Server      Server      `toml:"server"`
	Network     Network     `toml:"network"`
	Hyperliquid Hyperliquid `toml:"hyperliquid"`
	Process     Process     `toml:"process"`
	Paths       Paths       `toml:"paths"`
	BtRunner    BtRunner    `toml:"btrunner"`
	Runtime     Runtime     `toml:"runtime"`
}

// Server defines the shared server listener.
type Server struct {
	Host string `toml:"host"`
	Port uint16 `toml:"port"`
}

// Network defines the default admitted network.
type Network struct {
	Default      string `toml:"default"`
	AllowMainnet bool   `toml:"allow_mainnet"`
}

// Hyperliquid defines shared Hyperliquid policy.
type Hyperliquid struct {
	MinOrderNotionalUSDC uint64 `toml:"min_order_notional_usdc"`
}

// Process defines shared process supervision values.
type Process struct {
	PollSeconds            uint64 `toml:"poll_seconds"`
	RequestTimeoutSeconds  uint64 `toml:"request_timeout_seconds"`
	FailureThreshold       uint64 `toml:"failure_threshold"`
	UnresponsiveSeconds    uint64 `toml:"unresponsive_seconds"`
	RestartLimit           uint64 `toml:"restart_limit"`
	RestartIntervalSeconds uint64 `toml:"restart_interval_seconds"`
}

type Paths struct {
	SharedData string `toml:"shared_data"`
	Database   string `toml:"database"`
}

type BtRunner struct {
	TimerIntervalMS uint64 `toml:"timer_interval_ms"`
}

type Runtime struct {
	MaxCycles uint64     `toml:"max_cycles"`
	Signaler  Signaler   `toml:"signaler"`
	Executors []Executor `toml:"executors"`
	Risks     []Risk     `toml:"risks"`
}

type Signaler struct {
	Kind            string `toml:"kind"`
	SignalTimeframe string `toml:"signal_timeframe"`
	RegimeTimeframe string `toml:"regime_timeframe"`
	FastMA          int    `toml:"fast_ma"`
	SlowMA          int    `toml:"slow_ma"`
	RSIPeriod       int    `toml:"rsi_period"`
	RegimeEMA       int    `toml:"regime_ema"`
	VolumePeriod    int    `toml:"volume_period"`
}

type Executor struct {
	Kind                 string `toml:"kind"`
	StopLossPct          string `toml:"stop_loss_pct"`
	AccountName          string `toml:"account_name"`
	Network              string `toml:"network"`
	OrderNotionalUSDC    string `toml:"order_notional_usdc"`
	TakeProfitPct        string `toml:"take_profit_pct"`
	SimulatorEquityUSDC  string `toml:"simulator_equity_usdc"`
	SimulatorFeePct      string `toml:"simulator_fee_pct"`
	SimulatorSlippagePct string `toml:"simulator_slippage_pct"`
	PersistMode          string `toml:"persist_mode"`
}

type Risk struct {
	Kind string `toml:"kind"`
}

// Section 1 - Program Flow

// Load decodes and validates one Config.
func Load(path string) (Config, error) {
	// decode toml
	var cfg Config
	metadata, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("load config %s: %v", path, err)
	}
	// reject unknown fields
	if undecoded := metadata.Undecoded(); len(undecoded) != 0 {
		return cfg, fmt.Errorf("unknown config fields: %v", undecoded)
	}
	// validate paths
	if cfg.Paths.SharedData == "" || cfg.Paths.Database == "" {
		return cfg, fmt.Errorf("configured paths must not be empty")
	}
	// validate network
	if cfg.Network.Default != "mainnet" && cfg.Network.Default != "testnet" {
		return cfg, fmt.Errorf("network.default must be mainnet or testnet")
	}
	// validate cadence
	if cfg.BtRunner.TimerIntervalMS == 0 {
		return cfg, fmt.Errorf("btrunner.timer_interval_ms must be positive")
	}
	// validate runtime
	if err := validateRuntime(cfg.Runtime); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Section 2 - Domain Helpers

func validateRuntime(cfg Runtime) error {
	if cfg.MaxCycles == 0 || len(cfg.Executors) == 0 {
		return fmt.Errorf("runtime requires max_cycles and at least one executor")
	}
	if cfg.Signaler.Kind != "macross" && cfg.Signaler.Kind != "rsi" {
		return fmt.Errorf("unknown signaler: %s", cfg.Signaler.Kind)
	}
	if cfg.Signaler.FastMA <= 0 || cfg.Signaler.FastMA >= cfg.Signaler.SlowMA ||
		cfg.Signaler.RSIPeriod <= 0 || cfg.Signaler.RegimeEMA <= 0 || cfg.Signaler.VolumePeriod <= 0 {
		return fmt.Errorf("invalid signaler periods")
	}
	var tradeExecutors int
	for _, executor := range cfg.Executors {
		switch executor.Kind {
		case "observer":
			var stopLoss, err = strconv.ParseFloat(executor.StopLossPct, 64)
			if err != nil || stopLoss <= 0 || stopLoss >= 1 {
				return fmt.Errorf("observer stop_loss_pct must be between 0 and 1")
			}
		case "trade":
			tradeExecutors++
			if executor.AccountName == "" || executor.Network != "simnet" ||
				(executor.PersistMode != "none" && executor.PersistMode != "max") {
				return fmt.Errorf("invalid trade executor identity or persistence")
			}
			var notional, err = positiveDecimal(executor.OrderNotionalUSDC)
			if err != nil {
				return fmt.Errorf("invalid trade executor order_notional_usdc: %w", err)
			}
			var equity decimal.Decimal
			equity, err = positiveDecimal(executor.SimulatorEquityUSDC)
			if err != nil || equity.LessThan(notional) {
				return fmt.Errorf("invalid trade executor simulator_equity_usdc")
			}
			for name, value := range map[string]string{
				"take_profit_pct":        executor.TakeProfitPct,
				"stop_loss_pct":          executor.StopLossPct,
				"simulator_fee_pct":      executor.SimulatorFeePct,
				"simulator_slippage_pct": executor.SimulatorSlippagePct,
			} {
				var parsed decimal.Decimal
				parsed, err = decimal.NewFromString(value)
				if err != nil || parsed.IsNegative() {
					return fmt.Errorf("invalid trade executor %s", name)
				}
				if (name == "take_profit_pct" || name == "stop_loss_pct") &&
					(!parsed.IsPositive() || parsed.GreaterThanOrEqual(decimal.NewFromInt(1))) {
					return fmt.Errorf("trade executor %s must be between 0 and 1", name)
				}
			}
		default:
			return fmt.Errorf("unknown executor: %s", executor.Kind)
		}
	}
	if tradeExecutors > 1 {
		return fmt.Errorf("runtime supports one trade executor")
	}
	for _, risk := range cfg.Risks {
		if risk.Kind != "balanced" {
			return fmt.Errorf("unknown risk: %s", risk.Kind)
		}
	}
	return nil
}

// Section 3 - Generic Helpers

func positiveDecimal(value string) (decimal.Decimal, error) {
	var parsed, err = decimal.NewFromString(value)
	if err != nil || !parsed.IsPositive() {
		return decimal.Zero, fmt.Errorf("expected positive decimal")
	}
	return parsed, nil
}

// Rooted resolves one configured path beneath root.
func Rooted(root, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}

// ResolveDataPath resolves one path inside the configured shared-data root.
func ResolveDataPath(root, path string) (string, error) {
	var err error
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve shared_data %s: %w", root, err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve data path %s: %w", path, err)
	}
	var relative string
	relative, err = filepath.Rel(root, path)
	if err != nil || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("data path is outside shared_data: %s", path)
	}
	return path, nil
}
