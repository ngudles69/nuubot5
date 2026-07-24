package botcycle

import (
	"fmt"

	"nuubot/internal/config"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// Control owns one active BotCycle and its Executors.
type Control struct {
	log       *logging.Logger
	number    int
	signal    signaler.Signal
	executors []executor.Executor
	ticks     uint64
	runs      uint64
	startMS   uint64
	endMS     uint64
	running   bool
	completed bool
	stopped   bool
}

// Section 1 - Program Flow

// Init prepares one BotCycle and its configured Executors.
func (c *Control) Init(
	log *logging.Logger,
	number int,
	signal signaler.Signal,
	configs []config.Executor,
) error {
	c.log = log
	c.number = number
	c.signal = signal

	// create executors
	c.executors = make([]executor.Executor, 0, len(configs))
	for index, cfg := range configs {
		var created, err = executor.Create(log, number, index+1, signal, cfg)
		if err != nil {
			return fmt.Errorf("create executor %d: %w", index+1, err)
		}
		c.executors = append(c.executors, created)
	}

	// initialize botcycle
	log.Info(fmt.Sprintf(
		"bot cycle initialized cycle=%d side=%s signal_ts_ms=%d available_ts_ms=%d",
		number,
		signal.Side,
		signal.SignalMS,
		signal.AvailableMS,
	))
	return nil
}

// Start starts every configured Executor.
func (c *Control) Start() error {
	if c.running || c.stopped {
		return fmt.Errorf("bot cycle cannot start from current state")
	}
	// start executors
	for _, executor := range c.executors {
		var err = executor.Start()
		if err != nil {
			_, _ = c.Stop("start_error")
			return fmt.Errorf("start executor: %w", err)
		}
	}

	// start botcycle
	c.running = true
	c.log.Info(fmt.Sprintf("bot cycle started cycle=%d", c.number))
	return nil
}

// Run executes one timer-driven Executor operation.
func (c *Control) Run(nowMS uint64) (bool, error) {
	if !c.running {
		return false, fmt.Errorf("bot cycle cannot run from current state")
	}
	c.runs++

	// run executors
	for _, executor := range c.executors {
		if !executor.Terminal() {
			executor.Run(nowMS)
		}
	}

	// check completion
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

	// stop executors
	var firstErr error
	for index := len(c.executors) - 1; index >= 0; index-- {
		if err := c.executors[index].Stop(reason); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("stop executor %d: %w", index+1, err)
		}
	}
	c.stopped = true

	// resolve exit reason
	var exitReason = c.exitReason(reason)

	// calculate duration
	var durationMS uint64
	if c.endMS >= c.startMS {
		durationMS = c.endMS - c.startMS
	}

	// report proof
	c.log.Info(fmt.Sprintf(
		"bot cycle stopped cycle=%d side=%s start_ts_ms=%d end_ts_ms=%d "+
			"duration_ms=%d executors=%d ticks_received=%d runs=%d stop_reason=%s",
		c.number,
		c.signal.Side,
		c.startMS,
		c.endMS,
		durationMS,
		len(c.executors),
		c.ticks,
		c.runs,
		exitReason,
	))
	return exitReason, firstErr
}

// Section 2 - Domain Helpers

// OnBBO distributes one BBO to active Executors.
func (c *Control) OnBBO(bbo market.BBO) {
	// record cycle time
	if c.startMS == 0 {
		c.startMS = bbo.TimestampMS
	}
	c.endMS = bbo.TimestampMS
	c.ticks++

	// ingest executor bbo
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
