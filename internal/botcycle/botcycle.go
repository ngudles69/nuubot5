package botcycle

import (
	"fmt"
	"log/slog"

	"nuubot/internal/config"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/clock"
	nuuerrors "nuubot/internal/toolkit/errors"
)

// Control owns one active BotCycle and its Executors.
type Control struct {
	log       *slog.Logger
	number    int
	signal    signaler.Signal
	executors []executor.Executor
	ticks     uint64
	passes    uint64
	startMS   uint64
	endMS     uint64
	running   bool
	completed bool
	stopped   bool
}

// Section 1 - Program Flow

// New constructs one BotCycle.
func New(logger *slog.Logger, number int, signal signaler.Signal, configs []config.Executor) (*Control, error) {
	var executors = make([]executor.Executor, 0, len(configs))
	for index, cfg := range configs {
		var created, err = executor.Create(logger, number, index+1, signal, cfg)
		if err != nil {
			return nil, fmt.Errorf("create executor %d: %w", index+1, err)
		}
		executors = append(executors, created)
	}
	var log = logger.With("component", "botcycle", "cycle", number)
	log.Info(
		"bot cycle initialized",
		"event", "init",
		"status", "success",
		"side", signal.Side,
		"signal_ts_ms", signal.SignalMS,
		"available_ts_ms", signal.AvailableMS,
	)
	return &Control{log: log, number: number, signal: signal, executors: executors}, nil
}

// Start starts every configured Executor.
func (c *Control) Start() error {
	if c.running || c.stopped {
		return nuuerrors.StateError("bot cycle", "start")
	}
	for _, executor := range c.executors {
		var err = executor.Start()
		if err != nil {
			_, _ = c.Stop("start_error")
			return fmt.Errorf("start executor: %w", err)
		}
	}
	c.running = true
	c.log.Info("bot cycle started", "event", "start", "status", "success")
	return nil
}

// Pass runs one timer-driven Executor pass.
func (c *Control) Pass(nowMS uint64) (bool, error) {
	if !c.running {
		return false, nuuerrors.StateError("bot cycle", "pass")
	}
	c.passes++
	for _, executor := range c.executors {
		if !executor.Terminal() {
			executor.Pass(nowMS)
		}
	}
	c.completed = true
	for _, executor := range c.executors {
		if !executor.Terminal() {
			c.completed = false
			break
		}
	}
	return c.completed, nil
}

// Stop stops Executors in reverse ownership order.
func (c *Control) Stop(reason string) (string, error) {
	if c.stopped {
		return c.exitReason(reason), nil
	}
	c.running = false
	var firstErr error
	for index := len(c.executors) - 1; index >= 0; index-- {
		if err := c.executors[index].Stop(reason); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("stop executor %d: %w", index+1, err)
		}
	}
	c.stopped = true
	var exitReason = c.exitReason(reason)
	var status = "success"
	if firstErr != nil {
		status = "failed"
	}
	c.log.Info(
		"bot cycle stopped",
		"event", "stop",
		"status", status,
		"side", c.signal.Side,
		"start_ts_ms", c.startMS,
		"end_ts_ms", c.endMS,
		"duration_ms", clock.Duration(c.startMS, c.endMS),
		"executors", len(c.executors),
		"ticks_received", c.ticks,
		"passes", c.passes,
		"stop_reason", exitReason,
	)
	return exitReason, firstErr
}

// Section 2 - Domain Helpers

// OnBBO distributes one BBO to active Executors.
func (c *Control) OnBBO(bbo market.BBO) {
	if c.startMS == 0 {
		c.startMS = bbo.TimestampMS
	}
	c.endMS = bbo.TimestampMS
	c.ticks++
	for _, executor := range c.executors {
		if !executor.Terminal() {
			executor.OnBBO(bbo)
		}
	}
}

func (c *Control) exitReason(fallback string) string {
	if len(c.executors) == 0 {
		return fallback
	}
	var reason = c.executors[0].ExitReason()
	if reason == "" {
		return fallback
	}
	for _, executor := range c.executors[1:] {
		if executor.ExitReason() != reason {
			return "completed"
		}
	}
	return reason
}

// Section 3 - Generic Helpers
