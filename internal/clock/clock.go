package clock

import (
	"fmt"

	"nuubot5/internal/common"
)

type TickClock struct {
	log      *common.Logger
	interval uint64
	nextMS   uint64
	started  bool
	ticks    uint64
	passes   uint64
	stopped  bool
}

func New(logger *common.Logger, intervalMS uint64) *TickClock {
	logger.Info("tickclock", "init interval_ms=%d", intervalMS)
	return &TickClock{log: logger, interval: intervalMS}
}

func (c *TickClock) Advance(nowMS uint64) (bool, error) {
	c.ticks++
	if !c.started {
		if nowMS > ^uint64(0)-c.interval {
			return false, fmt.Errorf("TickClock overflow")
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
		return false, fmt.Errorf("TickClock overflow")
	}
	c.nextMS += intervals * c.interval
	c.passes++
	return true, nil
}

func (c *TickClock) Stop() {
	if c.stopped {
		return
	}
	c.stopped = true
	c.log.Info("tickclock", "stop status=success ticks_seen=%d passes_due=%d", c.ticks, c.passes)
}
