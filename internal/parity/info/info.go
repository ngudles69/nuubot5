// Package info owns parity probes for Hyperliquid's info endpoint.
package info

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"nuubot/internal/hyperliquid"
)

type clearinghouseClient interface {
	ClearinghouseStatePayload(
		context.Context,
		string,
	) (hyperliquid.Response, error)
}

// Target identifies one admitted parity target.
type Target struct {
	Network     string
	Account     string
	Address     string
	EvidenceDir string
}

// Result contains one completed info probe summary.
type Result struct {
	EvidenceDir       string
	Duration          time.Duration
	Positions         int
	Equity            decimal.Decimal
	MarginUsed        decimal.Decimal
	MaintenanceMargin decimal.Decimal
	Withdrawable      decimal.Decimal
}

type report struct {
	Network    string `json:"network"`
	Account    string `json:"account"`
	Operation  string `json:"operation"`
	CapturedAt string `json:"captured_at"`
	StatusCode int    `json:"status_code"`
	DurationMS int64  `json:"duration_ms"`
	RawPayload string `json:"raw_payload,omitempty"`
	Normalized string `json:"normalized,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
