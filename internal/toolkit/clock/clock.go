package clock

import (
	"fmt"
	"log/slog"
)

// TickClock converts replay timestamps into timed passes.
type TickClock struct {
	log      *slog.Logger
	interval uint64
	nextMS   uint64
	started  bool
	ticks    uint64
	passes   uint64
	stopped  bool
}

// Section 1 - Program Flow

// New constructs one TickClock.
func New(logger *slog.Logger, intervalMS uint64) *TickClock {
	log := logger.With("component", "tickclock")
	log.Info(
		"tick clock initialized",
		"event", "init",
		"status", "success",
		"interval_ms", intervalMS,
	)
	return &TickClock{log: log, interval: intervalMS}
}

// Advance accepts one timestamp and reports whether a pass is due.
func (c *TickClock) Advance(nowMS uint64) (bool, error) {
	c.ticks++
	if !c.started {
		if nowMS > ^uint64(0)-c.interval {
			return false, fmt.Errorf("tick clock overflow")
		}
		c.started = true
		c.nextMS = nowMS + c.interval
		c.passes++
		return true, nil
	}
	if nowMS < c.nextMS {
		return false, nil
	}
	intervals := (nowMS-c.nextMS)/c.interval + 1
	if intervals > (^uint64(0)-c.nextMS)/c.interval {
		return false, fmt.Errorf("tick clock overflow")
	}
	c.nextMS += intervals * c.interval
	c.passes++
	return true, nil
}

// Stop reports final clock statistics once.
func (c *TickClock) Stop() {
	if c.stopped {
		return
	}
	c.stopped = true
	c.log.Info(
		"tick clock stopped",
		"event", "stop",
		"status", "success",
		"ticks_seen", c.ticks,
		"passes_due", c.passes,
	)
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
