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
	MaxLeverage uint32
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

// Connect connects Venue to its configured network.
func (v *Venue) Connect(cfg Config) error {
	// Step 1: validate Venue lifecycle and network
	if v.initialized || v.stopped {
		return fmt.Errorf("connect Venue: invalid lifecycle state")
	}
	if cfg.MarketKey.Network != "simnet" {
		return fmt.Errorf(
			"connect Venue: network %q is not implemented",
			cfg.MarketKey.Network,
		)
	}

	// Step 2: connect simulated Venue and Exchange
	var err = v.simulator.Connect(simulator.Config{
		MarketData:  cfg.MarketData,
		MarketKey:   cfg.MarketKey,
		Account:     cfg.Account,
		Asset:       cfg.Asset,
		Symbol:      cfg.Symbol,
		MaxLeverage: cfg.MaxLeverage,
		Equity:      cfg.Equity,
		FeePct:      cfg.FeePct,
		SlippagePct: cfg.SlippagePct,
		PersistMode: cfg.PersistMode,
		Path:        cfg.Path,
	})
	if err != nil {
		return fmt.Errorf("connect Venue: %w", err)
	}

	// Step 3: mark Venue connected
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

// SetLeverage submits one leverage action to the configured network.
func (v *Venue) SetLeverage(
	action hyperliquid.UpdateLeverageAction,
	timestampMS uint64,
) ([]byte, error) {
	return v.simulator.SetLeverage(action, timestampMS)
}

// GetOpenOrders returns current open Orders from the configured network.
func (v *Venue) GetOpenOrders(account string) ([]byte, error) {
	return v.simulator.GetOpenOrders(account)
}

// GetOrderHistory returns recent Order history from the configured network.
func (v *Venue) GetOrderHistory(account string) ([]byte, error) {
	return v.simulator.GetOrderHistory(account)
}

// GetFillHistory returns current Fill history from the configured network.
func (v *Venue) GetFillHistory(
	account string,
	startMS uint64,
	endMS uint64,
) ([]byte, error) {
	return v.simulator.GetFillHistory(account, startMS, endMS)
}

// GetOrderStatus returns one current Order status from the configured network.
func (v *Venue) GetOrderStatus(account string, value string) ([]byte, error) {
	return v.simulator.GetOrderStatus(account, value)
}

// GetAccountState returns current Account state from the configured network.
func (v *Venue) GetAccountState(account string) ([]byte, error) {
	return v.simulator.GetAccountState(account)
}

// Disconnect releases Venue-owned resources.
func (v *Venue) Disconnect() error {
	// Step 1: ignore repeated disconnect
	if v.stopped {
		return nil
	}
	if !v.initialized {
		return fmt.Errorf("disconnect Venue: invalid lifecycle state")
	}

	// Step 2: disconnect simulated Venue and Exchange
	var err = v.simulator.Disconnect()
	if err != nil {
		return fmt.Errorf("disconnect Venue: %w", err)
	}

	// Step 3: mark Venue disconnected
	v.initialized = false
	v.stopped = true
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
