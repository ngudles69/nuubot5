package executor

import (
	"fmt"

	"nuubot/internal/botspec"
	"nuubot/internal/market"
	"nuubot/internal/toolkit/logging"
)

type observerStats struct {
	ingestBBOCount uint64
	onBBOCount     uint64
	startMS        uint64
	endMS          uint64
	startPrice     float64
	stopLossPrice  float64
	exitPrice      float64
	lastMS         uint64
	lastPrice      float64
	reason         string
}

type observer struct {
	log            *logging.Logger
	spec           botspec.ExecutorSpec
	cycleNumber    int
	executorNumber int
	signalMS       uint64
	signalPrice    float64
	side           string
	stopLossPct    float64
	stats          observerStats
	status         Status
}

var _ Executor = (*observer)(nil)
var _ BBOHandler = (*observer)(nil)
var _ BBOIngestHandler = (*observer)(nil)

// Section 1 - Program Flow

// OnInit initializes ObserverExecutor.
func (e *observer) OnInit(ctx BotCycleContext) error {
	e.log = ctx.Nuubot.Log
	e.spec = ctx.Spec
	if e.status != Configured {
		return fmt.Errorf("observer executor cannot initialize from current state")
	}
	e.cycleNumber = ctx.CycleNumber
	e.executorNumber = ctx.ExecutorNumber
	e.signalMS = ctx.Signal.TimestampMS()
	e.side = ctx.Spec.Side
	e.stopLossPct, _ = ctx.Spec.StopLossPct.Float64()
	e.status = Starting

	// validate config
	if e.stopLossPct <= 0 || e.stopLossPct >= 1 {
		e.status = Error
		return fmt.Errorf("observer stop_loss_pct must be between 0 and 1")
	}
	if e.side != Long && e.side != Short {
		e.status = Error
		return fmt.Errorf("%w: observer requires one configured side", ErrRejected)
	}

	// initialize observer
	e.status = Running
	e.log.Info(fmt.Sprintf(
		"executor initialized cycle=%d executor=%d kind=observer side=%s "+
			"signal_ts_ms=%d stop_loss_pct=%f",
		e.cycleNumber,
		e.executorNumber,
		e.side,
		e.signalMS,
		e.stopLossPct,
	))
	return nil
}

// OnStop stops ObserverExecutor and reports final statistics.
func (e *observer) OnStop(reason string) error {
	if e.status == Stopped || e.status == Error {
		return nil
	}
	e.status = Stopping

	// preserve stop reason
	if e.stats.reason == "" {
		e.stats.reason = reason
	}

	// preserve end time
	if e.stats.endMS == 0 {
		e.stats.endMS = e.stats.lastMS
		if e.stats.endMS == 0 {
			e.stats.endMS = e.signalMS
		}
	}

	// stop observer
	e.status = Stopped

	// calculate duration
	var durationMS uint64
	if e.stats.endMS >= e.stats.startMS {
		durationMS = e.stats.endMS - e.stats.startMS
	}
	// report proof
	e.log.Info(fmt.Sprintf(
		"executor stopped cycle=%d executor=%d side=%s signal_ts_ms=%d "+
			"signal_price=%f stop_loss_pct=%f start_ts_ms=%d end_ts_ms=%d "+
			"duration_ms=%d start_price=%f stop_loss_price=%f exit_price=%f "+
			"final_price=%f ingest_bbo_count=%d on_bbo_count=%d stop_reason=%s",
		e.cycleNumber,
		e.executorNumber,
		e.side,
		e.signalMS,
		e.signalPrice,
		e.stopLossPct,
		e.stats.startMS,
		e.stats.endMS,
		durationMS,
		e.stats.startPrice,
		e.stats.stopLossPrice,
		e.stats.exitPrice,
		e.stats.lastPrice,
		e.stats.ingestBBOCount,
		e.stats.onBBOCount,
		e.stats.reason,
	))
	return nil
}

// Section 2 - Domain Helpers

// IngestBBO records one Simulator-only BBO delivery.
func (e *observer) IngestBBO(bbo market.BBO) error {
	if bbo.Symbol != "" && bbo.Symbol != e.spec.Resource.Symbol {
		return nil
	}
	// count ingested bbo
	e.stats.ingestBBOCount++
	return nil
}

// OnBBO observes one normal Executor BBO event.
func (e *observer) OnBBO(bbo market.BBO) {
	if bbo.Symbol != "" && bbo.Symbol != e.spec.Resource.Symbol {
		return
	}
	// count received bbo
	e.stats.onBBOCount++
	if e.status != Running {
		return
	}

	// record last bbo
	e.stats.lastMS = bbo.TimestampMS
	e.stats.lastPrice = bbo.Price

	// record entry
	if e.stats.startMS == 0 {
		e.stats.startMS = bbo.TimestampMS
		e.stats.startPrice = bbo.Price
		e.signalPrice = bbo.Price
		if e.side == Long {
			e.stats.stopLossPrice = bbo.Price * (1 - e.stopLossPct)
		} else {
			e.stats.stopLossPrice = bbo.Price * (1 + e.stopLossPct)
		}
	}

	// assess stop loss
	var triggered = e.side == Long && bbo.Price <= e.stats.stopLossPrice ||
		e.side == Short && bbo.Price >= e.stats.stopLossPrice
	if triggered {
		e.stats.endMS = bbo.TimestampMS
		e.stats.exitPrice = bbo.Price
		e.stats.reason = "stop_loss"
		e.status = Stopping
	}
}

// Status returns ObserverExecutor's canonical lifecycle state.
func (e *observer) Status() Status {
	return e.status
}

// ExitReason returns ObserverExecutor's terminal reason.
func (e *observer) ExitReason() string {
	return e.stats.reason
}

// Telemetry returns one immutable current Observer observation.
func (e *observer) Telemetry() Telemetry {
	return Telemetry{
		ID:       e.spec.ID,
		Kind:     e.spec.Kind,
		Resource: e.spec.Resource,
		Status:   e.status,
	}
}

// Result returns one terminal ObserverExecutor result.
func (e *observer) Result() (Result, error) {
	if e.status != Stopped {
		return Result{}, fmt.Errorf("observer executor result is unavailable")
	}
	return Result{
		ID:            e.spec.ID,
		Role:          e.spec.Role,
		Kind:          e.spec.Kind,
		Side:          e.spec.Side,
		Resource:      e.spec.Resource,
		Status:        e.status,
		ExitReason:    e.stats.reason,
		CapitalUSDC:   e.spec.CapitalUSDC,
		OrderSizeUSDC: e.spec.OrderSizeUSDC,
	}, nil
}

// Section 3 - Generic Helpers
