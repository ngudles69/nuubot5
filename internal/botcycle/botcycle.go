package botcycle

import (
	"nuubot5/internal/common"
	"nuubot5/internal/config"
	"nuubot5/internal/executor"
	"nuubot5/internal/market"
	"nuubot5/internal/signaler"
)

type Control struct {
	log       *common.Logger
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

func New(log *common.Logger, number int, signal signaler.Signal, configs []config.Executor) (*Control, error) {
	executors := make([]executor.Executor, 0, len(configs))
	for index, cfg := range configs {
		created, err := executor.New(log, number, index+1, signal, cfg)
		if err != nil {
			return nil, err
		}
		executors = append(executors, created)
	}
	log.Info("botcycle", "init cycle=%d side=%s signal_ts_ms=%d available_ts_ms=%d", number, signal.Side, signal.SignalMS, signal.AvailableMS)
	return &Control{log: log, number: number, signal: signal, executors: executors}, nil
}

func (c *Control) Start() error {
	if c.running || c.stopped {
		return common.StateError("BotCycle", "start")
	}
	for _, executor := range c.executors {
		if err := executor.Start(); err != nil {
			_, _ = c.Stop("start_error")
			return err
		}
	}
	c.running = true
	c.log.Info("botcycle", "start cycle=%d", c.number)
	return nil
}

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

func (c *Control) MainLoop(nowMS uint64) (bool, error) {
	if !c.running {
		return false, common.StateError("BotCycle", "run main loop")
	}
	c.passes++
	for _, executor := range c.executors {
		if !executor.Terminal() {
			executor.MainLoop(nowMS)
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

func (c *Control) Stop(reason string) (string, error) {
	if c.stopped {
		return c.exitReason(reason), nil
	}
	c.running = false
	var firstErr error
	for index := len(c.executors) - 1; index >= 0; index-- {
		if err := c.executors[index].Stop(reason); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	c.stopped = true
	exitReason := c.exitReason(reason)
	status := "success"
	if firstErr != nil {
		status = "failed"
	}
	c.log.Info(
		"botcycle",
		"stop status=%s cycle=%d side=%s start_ts_ms=%d end_ts_ms=%d duration_ms=%d executors=%d ticks_received=%d passes=%d stop_reason=%s",
		status, c.number, c.signal.Side, c.startMS, c.endMS, common.Duration(c.startMS, c.endMS),
		len(c.executors), c.ticks, c.passes, exitReason,
	)
	return exitReason, firstErr
}

func (c *Control) exitReason(fallback string) string {
	if len(c.executors) == 0 {
		return fallback
	}
	reason := c.executors[0].ExitReason()
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
