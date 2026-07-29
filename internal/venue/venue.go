// Package venue owns Account execution against one configured network.
package venue

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
	"nuubot/internal/simulator"
)

// Config contains one Venue's network, identity, and simnet policy.
type Config struct {
	MarketData  *market.MarketData
	MarketKey   market.Key
	Account     string
	Asset       int
	Symbol      string
	Equity      decimal.Decimal
	FeePct      decimal.Decimal
	SlippagePct decimal.Decimal
	PersistMode string
	Path        string
}

// Venue owns execution for one configured network.
type Venue struct {
	simulator   simulator.Simulator
	initialized bool
	stopped     bool
}

// Section 1 - Program Flow

// Init initializes Venue for its configured network.
func (v *Venue) Init(cfg Config) error {
	// Step 1: validate Venue lifecycle and network
	if v.initialized || v.stopped {
		return fmt.Errorf("initialize Venue: invalid lifecycle state")
	}
	if cfg.MarketKey.Network != "simnet" {
		return fmt.Errorf(
			"initialize Venue: network %q is not implemented",
			cfg.MarketKey.Network,
		)
	}

	// Step 2: initialize simulated Venue and Exchange
	var err = v.simulator.Init(simulator.Config{
		MarketData:  cfg.MarketData,
		MarketKey:   cfg.MarketKey,
		Account:     cfg.Account,
		Asset:       cfg.Asset,
		Symbol:      cfg.Symbol,
		Equity:      cfg.Equity,
		FeePct:      cfg.FeePct,
		SlippagePct: cfg.SlippagePct,
		PersistMode: cfg.PersistMode,
		Path:        cfg.Path,
	})
	if err != nil {
		return fmt.Errorf("initialize Venue: %w", err)
	}

	// Step 3: mark Venue initialized
	v.initialized = true
	return nil
}

// PlaceOrders submits one Order action to the configured network.
func (v *Venue) PlaceOrders(
	action hyperliquid.PlaceOrderAction,
	timestampMS uint64,
) ([]byte, error) {
	return v.simulator.PlaceOrders(action, timestampMS)
}

// CancelOrders submits one cancellation action to the configured network.
func (v *Venue) CancelOrders(
	action hyperliquid.CancelByCLOIDAction,
	timestampMS uint64,
) ([]byte, error) {
	return v.simulator.CancelOrders(action, timestampMS)
}

// OpenOrders returns current open Orders from the configured network.
func (v *Venue) OpenOrders(account string) ([]byte, error) {
	return v.simulator.OpenOrders(account)
}

// Fills returns current Fill history from the configured network.
func (v *Venue) Fills(account string, startMS uint64, endMS uint64) ([]byte, error) {
	return v.simulator.Fills(account, startMS, endMS)
}

// OrderStatus returns one current Order status from the configured network.
func (v *Venue) OrderStatus(account string, value string) ([]byte, error) {
	return v.simulator.OrderStatus(account, value)
}

// AccountState returns current Account state from the configured network.
func (v *Venue) AccountState(account string) ([]byte, error) {
	return v.simulator.AccountState(account)
}

// Stop releases Venue-owned resources.
func (v *Venue) Stop() error {
	// Step 1: ignore repeated stop
	if v.stopped {
		return nil
	}
	if !v.initialized {
		return fmt.Errorf("stop Venue: invalid lifecycle state")
	}

	// Step 2: stop simulated Venue and Exchange
	var err = v.simulator.Stop()
	if err != nil {
		return fmt.Errorf("stop Venue: %w", err)
	}

	// Step 3: mark Venue stopped
	v.initialized = false
	v.stopped = true
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
