package botspec

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/signaler"
)

// ControllerSpec contains one Bot's Controller specification.
type ControllerSpec struct {
	MaxCycles uint64
}

// RiskSpec contains one Bot's Risk specification.
type RiskSpec struct {
	Kind string
}

// Resource identifies one exclusive exchange resource.
type Resource struct {
	Venue             string
	Network           string
	PhysicalAccountID string
	Symbol            string
}

// Key returns the exact resource identity.
func (r Resource) Key() string {
	return fmt.Sprintf(
		"%s/%s/%s/%s",
		r.Venue,
		r.Network,
		r.PhysicalAccountID,
		r.Symbol,
	)
}

// ExecutorSpec contains one immutable Executor specification.
type ExecutorSpec struct {
	ID             string
	Role           string
	Kind           string
	Side           string
	Resource       Resource
	CapitalUSDC    decimal.Decimal
	OrderSizeUSDC  decimal.Decimal
	GridLevels     int
	RangePct       decimal.Decimal
	MinExpectedPnL decimal.Decimal
	TakeProfitPct  decimal.Decimal
	StopLossPct    decimal.Decimal
	FeePct         decimal.Decimal
	SlippagePct    decimal.Decimal
	PersistMode    string
	Recon          string
}

// Spec contains one validated and shaped Bot specification.
type Spec struct {
	ID         string
	Controller ControllerSpec
	Signaler   signaler.Config
	Risks      []RiskSpec
	Executors  []ExecutorSpec
}

// Section 1 - Program Flow

// Build validates and shapes exact BotConfig TOML into one BotSpec.
func Build(botSpecID, configTOML string) (Spec, error) {
	var spec, err = build(botSpecID, configTOML)
	if err != nil {
		return Spec{}, err
	}
	spec.ID = botSpecID
	return spec, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
