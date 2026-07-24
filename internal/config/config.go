package config

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
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
	SharedData    string `toml:"shared_data"`
	SweepDatabase string `toml:"sweep_database"`
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
	Kind        string  `toml:"kind"`
	StopLossPct float64 `toml:"stop_loss_pct"`
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
	if cfg.Paths.SharedData == "" || cfg.Paths.SweepDatabase == "" {
		return cfg, fmt.Errorf("configured paths must not be empty")
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
	for _, executor := range cfg.Executors {
		if executor.Kind != "observer" {
			return fmt.Errorf("unknown executor: %s", executor.Kind)
		}
		if math.IsNaN(executor.StopLossPct) || math.IsInf(executor.StopLossPct, 0) ||
			executor.StopLossPct <= 0 || executor.StopLossPct >= 1 {
			return fmt.Errorf("observer stop_loss_pct must be between 0 and 1")
		}
	}
	for _, risk := range cfg.Risks {
		if risk.Kind != "balanced" {
			return fmt.Errorf("unknown risk: %s", risk.Kind)
		}
	}
	return nil
}

// Section 3 - Generic Helpers

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
