package executor

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/botspec"
	"nuubot/internal/market"
	"nuubot/internal/setup"
	"nuubot/internal/signaler"
)

// ErrRejected identifies an Executor admission rejection.
var ErrRejected = errors.New("executor rejected bot cycle")

const (
	Long  = "long"
	Short = "short"
)

// Result contains one immutable terminal Executor result.
type Result struct {
	ID            string
	Role          string
	Kind          string
	Side          string
	Resource      botspec.Resource
	Status        Status
	ExitReason    string
	CapitalUSDC   decimal.Decimal
	OrderSizeUSDC decimal.Decimal
	Cancellations uint64
	ClosureOrders uint64
	Retries       uint64
	RoundTrips    uint64
	Levels        []GridLevel
	Account       *account.Result
}

// Telemetry contains one immutable current Executor observation.
type Telemetry struct {
	ID       string
	Kind     string
	Resource botspec.Resource
	Status   Status
	Account  *account.Snapshot
}

// Clone returns one independently owned Executor result.
func (r Result) Clone() Result {
	var copied = r
	copied.Levels = append([]GridLevel(nil), r.Levels...)
	if r.Account != nil {
		var accountResult = r.Account.Clone()
		copied.Account = &accountResult
	}
	return copied
}

// GridLevel contains one Grid Executor's calculated level and runtime state.
type GridLevel struct {
	Level                      uint16
	Boundary                   bool
	GridPrice                  decimal.Decimal
	InitialEntryPrice          decimal.Decimal
	ReentryPrice               decimal.Decimal
	ExitPrice                  decimal.Decimal
	Quantity                   decimal.Decimal
	InitialNotional            decimal.Decimal
	ReentryNotional            decimal.Decimal
	InitialEntryCommission     decimal.Decimal
	ReentryCommission          decimal.Decimal
	ExitCommission             decimal.Decimal
	InitialExpectedPnL         decimal.Decimal
	ReentryExpectedPnL         decimal.Decimal
	IntendedAction             string
	CurrentTradeID             uint64
	CurrentTradeNo             uint32
	CurrentTradeStatus         string
	Status                     string
	InitialSubmissionCompleted bool
	SubmissionAttempts         uint32
	LastSubmittedMS            uint64
	LastCompletedMS            uint64
}

// Status identifies one Executor lifecycle state.
type Status string

const (
	Configured Status = "configured"
	Starting   Status = "starting"
	Running    Status = "running"
	Paused     Status = "paused"
	Stopping   Status = "stopping"
	Stopped    Status = "stopped"
	Error      Status = "error"
)

// BotCycleContext contains one Executor's BotCycle initialization inputs.
type BotCycleContext struct {
	Nuubot             *setup.Nuubot
	CycleNumber        int
	ExecutorNumber     int
	Signal             signaler.Package
	Spec               botspec.ExecutorSpec
	LatestBBO          market.BBO
	StartingEquityUSDC decimal.Decimal
	Status             Status
}

// Executor defines the required lifecycle for one execution policy.
type Executor interface {
	OnInit(BotCycleContext) error
	OnStop(string) error
	Status() Status
	ExitReason() string
	Telemetry() Telemetry
	Result() (Result, error)
}

// StartHandler starts after every sibling Executor initializes.
type StartHandler interface {
	OnStart() error
}

// BBOHandler consumes normal Executor BBO events.
type BBOHandler interface {
	OnBBO(market.BBO)
}

// BBOIngestHandler consumes Simulator-only BBO ingestion.
type BBOIngestHandler interface {
	IngestBBO(market.BBO) error
}

// AccountReconciler refreshes one Executor's Account truth.
type AccountReconciler interface {
	Reconcile(uint64, bool) (account.Snapshot, bool, uint64, error)
}

// ReconHandler consumes one accepted reconciliation barrier.
type ReconHandler interface {
	OnRecon() error
}

// SignalHandler consumes one immutable Signal package.
type SignalHandler interface {
	OnSignal(signaler.Package) error
}

// Section 1 - Program Flow

// Create selects and initializes the configured Executor.
func Create(ctx BotCycleContext) (Executor, error) {
	// Step 1: validate Executor status
	switch ctx.Status {
	case Configured, Starting, Running, Paused, Stopping:
	case Stopped, Error:
		return nil, fmt.Errorf("cannot initialize terminal executor status: %s", ctx.Status)
	default:
		return nil, fmt.Errorf("unknown executor status: %s", ctx.Status)
	}

	// Step 2: select Executor
	var selected Executor
	switch ctx.Spec.Kind {
	case "observer":
		selected = &observer{status: ctx.Status}
	case "trade":
		selected = &tradeExecutor{status: ctx.Status}
	case "grid":
		selected = &gridExecutor{status: ctx.Status}
	default:
		return nil, fmt.Errorf("unknown executor: %s", ctx.Spec.Kind)
	}

	// Step 3: initialize Executor
	var err = selected.OnInit(ctx)
	if err != nil {
		return nil, err
	}
	return selected, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
