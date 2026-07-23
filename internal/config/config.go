package config

import (
	"fmt"
	"math"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Paths    Paths    `toml:"paths"`
	BtRunner BtRunner `toml:"btrunner"`
	Runtime  Runtime  `toml:"runtime"`
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

func Load(path string) (Config, error) {
	var cfg Config
	metadata, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("load config %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) != 0 {
		return cfg, fmt.Errorf("unknown config fields: %v", undecoded)
	}
	if cfg.Paths.SharedData == "" || cfg.Paths.SweepDatabase == "" {
		return cfg, fmt.Errorf("configured paths must not be empty")
	}
	if cfg.BtRunner.TimerIntervalMS == 0 {
		return cfg, fmt.Errorf("btrunner.timer_interval_ms must be positive")
	}
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

func Rooted(root, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}
