package executor

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// ErrRejected identifies an Executor admission rejection.
var ErrRejected = errors.New("executor rejected bot cycle")

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
	Log            *logging.Logger
	CycleNumber    int
	ExecutorNumber int
	Signaler       signaler.Signaler
	Signal         signaler.Package
	Config         config.Executor
	LatestBBO      market.BBO
	Meta           meta.Instrument
	MinNotional    decimal.Decimal
	ResultPath     string
}

// Executor defines the required lifecycle for one execution policy.
type Executor interface {
	OnInit(Context) error
	OnStop(string) error
	Status() Status
	ExitReason() string
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

// AccountResultProvider returns cached terminal Account evidence.
type AccountResultProvider interface {
	AccountResult() (account.Result, error)
}

// Section 1 - Program Flow

// Create selects and initializes the configured Executor.
func Create(ctx Context) (Executor, error) {
	// select executor
	var selected Executor
	switch ctx.Config.Kind {
	case "observer":
		selected = &observer{status: Configured}
	case "trade":
		selected = &tradeExecutor{status: Configured}
	default:
		return nil, fmt.Errorf("unknown executor: %s", ctx.Config.Kind)
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
