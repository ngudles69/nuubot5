package botcycle

import (
	"fmt"

	"nuubot/internal/config"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
)

// Control owns one active BotCycle and its Executors.
type Control struct {
	log       *logging.Logger
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
func New(log *logging.Logger, number int, signal signaler.Signal, configs []config.Executor) (*Control, error) {
	var executors = make([]executor.Executor, 0, len(configs))
	for index, cfg := range configs {
		var created, err = executor.Create(log, number, index+1, signal, cfg)
		if err != nil {
			return nil, fmt.Errorf("create executor %d: %w", index+1, err)
		}
		executors = append(executors, created)
	}
	log.Info(fmt.Sprintf(
		"bot cycle initialized cycle=%d side=%s signal_ts_ms=%d available_ts_ms=%d",
		number,
		signal.Side,
		signal.SignalMS,
		signal.AvailableMS,
	))
	return &Control{log: log, number: number, signal: signal, executors: executors}, nil
}

// Start starts every configured Executor.
func (c *Control) Start() error {
	if c.running || c.stopped {
		return fmt.Errorf("bot cycle cannot start from current state")
	}
	for _, executor := range c.executors {
		var err = executor.Start()
		if err != nil {
			_, _ = c.Stop("start_error")
			return fmt.Errorf("start executor: %w", err)
		}
	}
	c.running = true
	c.log.Info(fmt.Sprintf("bot cycle started cycle=%d", c.number))
	return nil
}

// Pass runs one timer-driven Executor pass.
func (c *Control) Pass(nowMS uint64) (bool, error) {
	if !c.running {
		return false, fmt.Errorf("bot cycle cannot pass from current state")
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
	c.log.Info(fmt.Sprintf(
		"bot cycle stopped cycle=%d side=%s start_ts_ms=%d end_ts_ms=%d "+
			"duration_ms=%d executors=%d ticks_received=%d passes=%d stop_reason=%s",
		c.number,
		c.signal.Side,
		c.startMS,
		c.endMS,
		clock.Duration(c.startMS, c.endMS),
		len(c.executors),
		c.ticks,
		c.passes,
		exitReason,
	))
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
