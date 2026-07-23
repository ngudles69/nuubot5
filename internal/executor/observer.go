package executor

import (
	"fmt"

	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
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
	log            *logging.Logger
	cycleNumber    int
	executorNumber int
	signal         signaler.Signal
	stopLossPct    float64
	stats          observerStats
	started        bool
	terminal       bool
	stopped        bool
}

// Section 1 - Program Flow

func newObserver(
	log *logging.Logger,
	cycleNumber int,
	executorNumber int,
	signal signaler.Signal,
	cfg config.Executor,
) (*observer, error) {
	if cfg.StopLossPct <= 0 || cfg.StopLossPct >= 1 {
		return nil, fmt.Errorf("observer stop_loss_pct must be between 0 and 1")
	}
	log.Info(fmt.Sprintf(
		"executor initialized cycle=%d executor=%d kind=observer side=%s "+
			"signal_ts_ms=%d available_ts_ms=%d stop_loss_pct=%f",
		cycleNumber,
		executorNumber,
		signal.Side,
		signal.SignalMS,
		signal.AvailableMS,
		cfg.StopLossPct,
	))
	return &observer{
		log: log, cycleNumber: cycleNumber, executorNumber: executorNumber,
		signal: signal, stopLossPct: cfg.StopLossPct,
	}, nil
}

func (e *observer) Start() error {
	if e.started || e.stopped {
		return fmt.Errorf("observer executor cannot start from current state")
	}
	e.started = true
	e.log.Info(fmt.Sprintf(
		"executor started cycle=%d executor=%d kind=observer",
		e.cycleNumber,
		e.executorNumber,
	))
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
	e.log.Info(fmt.Sprintf(
		"executor stopped cycle=%d executor=%d side=%s signal_ts_ms=%d "+
			"available_ts_ms=%d signal_price=%f stop_loss_pct=%f start_ts_ms=%d "+
			"end_ts_ms=%d duration_ms=%d start_price=%f stop_loss_price=%f "+
			"exit_price=%f final_price=%f ticks_received=%d passes=%d stop_reason=%s",
		e.cycleNumber,
		e.executorNumber,
		e.signal.Side,
		e.signal.SignalMS,
		e.signal.AvailableMS,
		e.signal.Price,
		e.stopLossPct,
		e.stats.startMS,
		e.stats.endMS,
		clock.Duration(e.stats.startMS, e.stats.endMS),
		e.stats.startPrice,
		e.stats.stopLossPrice,
		e.stats.exitPrice,
		e.stats.lastPrice,
		e.stats.ticks,
		e.stats.passes,
		e.stats.reason,
	))
	return nil
}

// Section 2 - Domain Helpers

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

// Section 3 - Generic Helpers
