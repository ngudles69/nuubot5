package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// App contains process-wide, non-Bot configuration.
type App struct {
	Server      Server      `toml:"server"`
	Network     Network     `toml:"network"`
	Hyperliquid Hyperliquid `toml:"hyperliquid"`
	Process     Process     `toml:"process"`
	Paths       Paths       `toml:"paths"`
	BtRunner    BtRunner    `toml:"btrunner"`
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

// Section 1 - Program Flow

// LoadApp decodes and validates one App Config.
func LoadApp(path string) (App, error) {
	// decode toml
	var cfg App
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
	return cfg, nil
}

// Section 2 - Domain Helpers

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
