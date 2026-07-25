package executor

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/toolkit/logging"
)

// ErrRejected identifies an Executor admission rejection.
var ErrRejected = errors.New("executor rejected bot cycle")

const (
	Long  = "long"
	Short = "short"
)

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

// Spec contains one immutable Executor definition.
type Spec struct {
	ID              string
	Role            string
	Kind            string
	Side            string
	Resource        Resource
	CapitalUSDC     decimal.Decimal
	OrderSizeUSDC   decimal.Decimal
	TakeProfitPct   decimal.Decimal
	StopLossPct     decimal.Decimal
	FeePct          decimal.Decimal
	SlippagePct     decimal.Decimal
	PersistMode     string
	Meta            meta.Instrument
	MinNotionalUSDC decimal.Decimal
	ResultPath      string
}

// Result contains one immutable terminal Executor result.
type Result struct {
	ID            string
	Role          string
	Kind          string
	Side          string
	Resource      Resource
	Status        Status
	ExitReason    string
	CapitalUSDC   decimal.Decimal
	OrderSizeUSDC decimal.Decimal
	Account       *account.Result
}

// Clone returns one independently owned Executor result.
func (r Result) Clone() Result {
	var copied = r
	if r.Account != nil {
		var accountResult = r.Account.Clone()
		copied.Account = &accountResult
	}
	return copied
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

// Context contains one Executor's initialization inputs.
type Context struct {
	Log                *logging.Logger
	CycleNumber        int
	ExecutorNumber     int
	SignalTimestampMS  uint64
	Spec               Spec
	LatestBBO          market.BBO
	StartingEquityUSDC decimal.Decimal
}

// Executor defines the required lifecycle for one execution policy.
type Executor interface {
	OnInit(Context) error
	OnStop(string) error
	Status() Status
	ExitReason() string
	Result() (Result, error)
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
	Reconcile(uint64, bool) (account.Snapshot, bool, error)
}

// ReconHandler consumes one accepted reconciliation barrier.
type ReconHandler interface {
	OnRecon(uint64) error
}

// Section 1 - Program Flow

// Create selects and initializes the configured Executor.
func Create(ctx Context) (Executor, error) {
	// select executor
	var selected Executor
	switch ctx.Spec.Kind {
	case "observer":
		selected = &observer{status: Configured}
	case "trade":
		selected = &tradeExecutor{status: Configured}
	default:
		return nil, fmt.Errorf("unknown executor: %s", ctx.Spec.Kind)
	}

	// initialize executor
	var err = selected.OnInit(ctx)
	if err != nil {
		return nil, err
	}
	return selected, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
