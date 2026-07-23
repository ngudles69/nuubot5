package clock

import (
	"fmt"

	"nuubot/internal/toolkit/logging"
)

// TickClock runs one registered timer from replay timestamps.
type TickClock struct {
	log      *logging.Logger
	interval uint64
	callback func(uint64) error
	nextMS   uint64
	started  bool
	ticks    uint64
	timers   uint64
	stopped  bool
}

// Section 1 - Program Flow

// New constructs one TickClock.
func New(log *logging.Logger) *TickClock {
	log.Info("tick clock initialized")
	return &TickClock{log: log}
}

// RegisterTimer registers the one replay timer.
func (c *TickClock) RegisterTimer(intervalMS uint64, callback func(uint64) error) error {
	if intervalMS == 0 {
		return fmt.Errorf("tick clock timer interval must be positive")
	}
	if callback == nil {
		return fmt.Errorf("tick clock timer callback is required")
	}
	if c.callback != nil {
		return fmt.Errorf("tick clock timer already registered")
	}
	c.interval = intervalMS
	c.callback = callback
	c.log.Info(fmt.Sprintf("tick clock timer registered interval_ms=%d", intervalMS))
	return nil
}

// Advance accepts one timestamp and runs the registered timer when scheduled.
func (c *TickClock) Advance(nowMS uint64) error {
	c.ticks++
	if c.callback == nil {
		return fmt.Errorf("tick clock timer is not registered")
	}
	if !c.started {
		if nowMS > ^uint64(0)-c.interval {
			return fmt.Errorf("tick clock overflow")
		}
		c.started = true
		c.nextMS = nowMS + c.interval
		return c.runTimer(nowMS)
	}
	if nowMS < c.nextMS {
		return nil
	}
	var intervals = (nowMS-c.nextMS)/c.interval + 1
	if intervals > (^uint64(0)-c.nextMS)/c.interval {
		return fmt.Errorf("tick clock overflow")
	}
	c.nextMS += intervals * c.interval
	return c.runTimer(nowMS)
}

func (c *TickClock) runTimer(nowMS uint64) error {
	c.timers++
	var err = c.callback(nowMS)
	if err != nil {
		return fmt.Errorf("run tick clock timer: %w", err)
	}
	return nil
}

// Stop reports final clock statistics once.
func (c *TickClock) Stop() {
	if c.stopped {
		return
	}
	c.stopped = true
	c.log.Info(fmt.Sprintf(
		"tick clock stopped ticks_seen=%d timers_triggered=%d",
		c.ticks,
		c.timers,
	))
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers

// Duration returns the non-negative difference between two millisecond timestamps.
func Duration(start, end uint64) uint64 {
	if end < start {
		return 0
	}
	return end - start
}
