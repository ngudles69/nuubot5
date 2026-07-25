// Package telemetry carries immutable periodic run observations.
package telemetry

import "github.com/shopspring/decimal"

// Sample contains one compact BtRunner telemetry observation.
type Sample struct {
	Sequence            uint64          `json:"sequence"`
	TimestampMS         uint64          `json:"timestamp_ms"`
	Terminal            bool            `json:"terminal"`
	TicksServed         uint64          `json:"ticks_served"`
	ControllerRuns      uint64          `json:"controller_runs"`
	SignalPackages      uint64          `json:"signal_packages"`
	StartActionsSkipped uint64          `json:"start_actions_skipped"`
	CyclesStarted       uint64          `json:"cycles_started"`
	CyclesRejected      uint64          `json:"cycles_rejected"`
	CyclesClosed        uint64          `json:"cycles_closed"`
	ActiveCycle         int             `json:"active_cycle"`
	BotCapital          decimal.Decimal `json:"bot_capital"`
	BotBalance          decimal.Decimal `json:"bot_balance"`
	BotEquity           decimal.Decimal `json:"bot_equity"`
	NetPnL              decimal.Decimal `json:"net_pnl"`
	PeakEquity          decimal.Decimal `json:"peak_equity"`
	Drawdown            decimal.Decimal `json:"drawdown"`
	MaxDrawdown         decimal.Decimal `json:"max_drawdown"`
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
