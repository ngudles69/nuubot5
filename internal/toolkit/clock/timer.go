package clock

import (
	"fmt"
	"math"
)

// Timer defines one named repeating schedule.
type Timer struct {
	Name       string
	IntervalMS uint64
	StartMS    *uint64
	StopMS     *uint64
}

type timerState struct {
	name       string
	intervalMS uint64
	stopMS     uint64
	hasStop    bool
	callback   func(uint64) error
	nextFireMS uint64
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

func (c *clockState) registerTimer(
	timer Timer,
	callback func(uint64) error,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// validate state
	if !c.initialized || c.started || c.stopped {
		return fmt.Errorf("%s clock cannot register timer from current state", c.kind)
	}

	// validate timer
	if timer.Name == "" {
		return fmt.Errorf("%s clock timer name is required", c.kind)
	}
	if timer.IntervalMS == 0 {
		return fmt.Errorf("%s clock timer interval must be positive", c.kind)
	}
	if callback == nil {
		return fmt.Errorf("%s clock timer callback is required", c.kind)
	}
	if _, exists := c.timers[timer.Name]; exists {
		return fmt.Errorf("%s clock timer already registered: %s", c.kind, timer.Name)
	}

	var startMS = c.nowMS
	if timer.StartMS != nil {
		startMS = *timer.StartMS
	}
	if startMS > math.MaxUint64-timer.IntervalMS {
		return fmt.Errorf("%s clock timer next fire overflows: %s", c.kind, timer.Name)
	}
	if timer.StopMS != nil && *timer.StopMS <= startMS {
		return fmt.Errorf(
			"%s clock timer stop must follow start: %s",
			c.kind,
			timer.Name,
		)
	}

	// schedule timer
	var nextFireMS = startMS + timer.IntervalMS
	var stopMS uint64
	if timer.StopMS != nil {
		stopMS = *timer.StopMS
	}

	// register timer
	c.timers[timer.Name] = &timerState{
		name:       timer.Name,
		intervalMS: timer.IntervalMS,
		stopMS:     stopMS,
		hasStop:    timer.StopMS != nil,
		callback:   callback,
		nextFireMS: nextFireMS,
	}
	c.log.Info(fmt.Sprintf(
		"%s clock timer registered name=%s interval_ms=%d next_fire_ms=%d",
		c.kind,
		timer.Name,
		timer.IntervalMS,
		nextFireMS,
	))
	return nil
}

func (c *clockState) timerCheck(nowMS uint64) error {
	c.advanceMu.Lock()
	defer c.advanceMu.Unlock()

	c.mu.Lock()

	// validate state
	if !c.started || c.stopped {
		c.mu.Unlock()
		return fmt.Errorf("%s clock cannot advance from current state", c.kind)
	}

	// validate time
	if nowMS < c.nowMS {
		var kind = c.kind
		var priorMS = c.nowMS
		c.mu.Unlock()
		return fmt.Errorf(
			"%s clock moved backward: %d -> %d",
			kind,
			priorMS,
			nowMS,
		)
	}

	// check timers
	c.advances++
	c.mu.Unlock()
	return c.checkTimers(nowMS)
}

func (c *clockState) checkTimers(nowMS uint64) error {
	for {
		c.mu.Lock()
		if c.stopped {
			c.mu.Unlock()
			return nil
		}

		// select next timer
		var timer = c.nextDueTimer(nowMS)
		if timer == nil {
			// advance time
			c.nowMS = nowMS
			c.mu.Unlock()
			return nil
		}

		// schedule timer
		var fireMS = timer.nextFireMS
		var callback = timer.callback
		var name = timer.name
		c.nowMS = fireMS
		c.scheduleNext(timer)
		c.timersFired++
		c.mu.Unlock()

		// run timer callback
		var err = callback(fireMS)
		if err != nil {
			var wrapped = fmt.Errorf(
				"run %s clock timer %s: %w",
				c.kind,
				name,
				err,
			)
			c.mu.Lock()
			c.err = wrapped
			c.mu.Unlock()
			return wrapped
		}
	}
}

func (c *clockState) nextFireMS(name string) (uint64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var timer, exists = c.timers[name]
	if !exists {
		return 0, false
	}
	return timer.nextFireMS, true
}

func (c *clockState) cancelTimer(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.timers, name)
}

func (c *clockState) nextTimerMS() (uint64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var nextMS uint64
	var exists bool
	for name, timer := range c.timers {
		if timer.hasStop && timer.nextFireMS > timer.stopMS {
			delete(c.timers, name)
			continue
		}
		if !exists || timer.nextFireMS < nextMS {
			nextMS = timer.nextFireMS
			exists = true
		}
	}
	return nextMS, exists
}

func (c *clockState) nextDueTimer(nowMS uint64) *timerState {
	// ponytail: linear scan suits a few Bot timers; use container/heap if timer counts become measurable.
	var due *timerState
	for name, timer := range c.timers {
		if timer.hasStop && timer.nextFireMS > timer.stopMS {
			delete(c.timers, name)
			continue
		}
		if timer.nextFireMS > nowMS {
			continue
		}
		if due == nil ||
			timer.nextFireMS < due.nextFireMS ||
			(timer.nextFireMS == due.nextFireMS && timer.name < due.name) {
			due = timer
		}
	}
	return due
}

func (c *clockState) scheduleNext(timer *timerState) {
	if timer.nextFireMS > math.MaxUint64-timer.intervalMS {
		delete(c.timers, timer.name)
		return
	}
	timer.nextFireMS += timer.intervalMS
	if timer.hasStop && timer.nextFireMS > timer.stopMS {
		delete(c.timers, timer.name)
	}
}

// Section 3 - Generic Helpers
