package executor

import (
	"fmt"
	"log/slog"

	"nuubot5/internal/common"
	"nuubot5/internal/config"
	"nuubot5/internal/market"
	"nuubot5/internal/signaler"
)

type observerStats struct {
	ticks         uint64
	passes        uint64
	startMS       uint64
	endMS         uint64
	startPrice    float64
	stopLossPrice float64
	exitPrice     float64
	lastMS        uint64
	lastPrice     float64
	reason        string
}

type observer struct {
	log            *slog.Logger
	cycleNumber    int
	executorNumber int
	signal         signaler.Signal
	stopLossPct    float64
	stats          observerStats
	started        bool
	terminal       bool
	stopped        bool
}

// Program Flow

func newObserver(
	logger *slog.Logger,
	cycleNumber int,
	executorNumber int,
	signal signaler.Signal,
	cfg config.Executor,
) (*observer, error) {
	if cfg.StopLossPct <= 0 || cfg.StopLossPct >= 1 {
		return nil, fmt.Errorf("observer stop_loss_pct must be between 0 and 1")
	}
	log := logger.With(
		"component", "executor",
		"cycle", cycleNumber,
		"executor", executorNumber,
	)
	log.Info(
		"executor initialized",
		"event", "init",
		"status", "success",
		"kind", "observer",
		"side", signal.Side,
		"signal_ts_ms", signal.SignalMS,
		"available_ts_ms", signal.AvailableMS,
		"stop_loss_pct", cfg.StopLossPct,
	)
	return &observer{
		log: log, cycleNumber: cycleNumber, executorNumber: executorNumber,
		signal: signal, stopLossPct: cfg.StopLossPct,
	}, nil
}

func (e *observer) Start() error {
	if e.started || e.stopped {
		return common.StateError("observer executor", "start")
	}
	e.started = true
	e.log.Info(
		"executor started",
		"event", "start",
		"status", "success",
		"kind", "observer",
	)
	return nil
}

func (e *observer) Pass(_ uint64) bool {
	e.stats.passes++
	return e.terminal
}

func (e *observer) Stop(reason string) error {
	if e.stopped {
		return nil
	}
	if e.stats.reason == "" {
		e.stats.reason = reason
	}
	if e.stats.endMS == 0 {
		e.stats.endMS = e.stats.lastMS
		if e.stats.endMS == 0 {
			e.stats.endMS = e.signal.AvailableMS
		}
	}
	e.started = false
	e.terminal = true
	e.stopped = true
	e.log.Info(
		"executor stopped",
		"event", "stop",
		"status", "success",
		"side", e.signal.Side,
		"signal_ts_ms", e.signal.SignalMS,
		"available_ts_ms", e.signal.AvailableMS,
		"signal_price", e.signal.Price,
		"stop_loss_pct", e.stopLossPct,
		"start_ts_ms", e.stats.startMS,
		"end_ts_ms", e.stats.endMS,
		"duration_ms", common.Duration(e.stats.startMS, e.stats.endMS),
		"start_price", e.stats.startPrice,
		"stop_loss_price", e.stats.stopLossPrice,
		"exit_price", e.stats.exitPrice,
		"final_price", e.stats.lastPrice,
		"ticks_received", e.stats.ticks,
		"passes", e.stats.passes,
		"stop_reason", e.stats.reason,
	)
	return nil
}

// Domain Helpers

func (e *observer) OnBBO(bbo market.BBO) {
	if !e.started || e.terminal {
		return
	}
	e.stats.lastMS = bbo.TimestampMS
	e.stats.lastPrice = bbo.Price
	if e.stats.startMS == 0 {
		e.stats.startMS = bbo.TimestampMS
		e.stats.startPrice = bbo.Price
		if e.signal.Side == signaler.Long {
			e.stats.stopLossPrice = bbo.Price * (1 - e.stopLossPct)
		} else {
			e.stats.stopLossPrice = bbo.Price * (1 + e.stopLossPct)
		}
	}
	e.stats.ticks++
	triggered := e.signal.Side == signaler.Long && bbo.Price <= e.stats.stopLossPrice ||
		e.signal.Side == signaler.Short && bbo.Price >= e.stats.stopLossPrice
	if triggered {
		e.stats.endMS = bbo.TimestampMS
		e.stats.exitPrice = bbo.Price
		e.stats.reason = "stop_loss"
		e.terminal = true
	}
}

func (e *observer) Terminal() bool {
	return e.terminal
}

func (e *observer) ExitReason() string {
	return e.stats.reason
}
