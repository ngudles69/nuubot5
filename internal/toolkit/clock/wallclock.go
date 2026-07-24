package clock

import (
	"math"
	"sync"
	"time"

	"nuubot/internal/toolkit/logging"
)

// WallClock advances from current UTC wall time.
type WallClock struct {
	state    clockState
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

// Section 1 - Program Flow

// Init prepares WallClock at its initial wall timestamp.
func (c *WallClock) Init(log *logging.Logger, initialMS uint64) error {
	// initialize clock
	var err = c.state.init(log, Wall, initialMS)
	if err != nil {
		return err
	}

	// initialize loop
	c.stop = make(chan struct{})
	c.done = make(chan struct{})
	return nil
}

// Start starts WallClock timer admission and advancement.
func (c *WallClock) Start() error {
	// start clock
	var err = c.state.start()
	if err != nil {
		return err
	}

	// start loop
	go c.loop()
	return nil
}

// Stop stops WallClock advancement and timer admission.
func (c *WallClock) Stop() {
	// stop loop
	if c.state.running() {
		c.stopOnce.Do(func() {
			close(c.stop)
		})
		<-c.done
	}

	// stop clock
	c.state.stop()
}

// Section 2 - Domain Helpers

// Err returns WallClock's terminal callback error.
func (c *WallClock) Err() error {
	// read error
	return c.state.clockErr()
}

// NowMS returns current UTC wall time.
func (c *WallClock) NowMS() uint64 {
	// read time
	return uint64(time.Now().UnixMilli())
}

// RegisterTimer registers one named WallClock timer.
func (c *WallClock) RegisterTimer(timer Timer, callback func(uint64) error) error {
	// register timer
	return c.state.registerTimer(timer, callback)
}

// Advance advances WallClock and fires every due timer.
func (c *WallClock) Advance(nowMS uint64) error {
	// check timers
	return c.state.timerCheck(nowMS)
}

// NextFireMS returns one WallClock timer's next scheduled timestamp.
func (c *WallClock) NextFireMS(name string) (uint64, bool) {
	// read next fire
	return c.state.nextFireMS(name)
}

// CancelTimer cancels one WallClock timer.
func (c *WallClock) CancelTimer(name string) {
	// cancel timer
	c.state.cancelTimer(name)
}

func (c *WallClock) loop() {
	defer close(c.done)

	for {
		// read next timer
		var nextMS, exists = c.state.nextTimerMS()
		if !exists {
			<-c.stop
			return
		}

		// wait for timer
		var wait = wallWait(c.NowMS(), nextMS)
		var timer = time.NewTimer(wait)
		select {
		case <-c.stop:
			stopTimer(timer)
			return
		case <-timer.C:
		}

		// advance clock
		var err = c.Advance(c.NowMS())
		if err != nil {
			return
		}
	}
}

// Section 3 - Generic Helpers

func wallWait(nowMS, nextMS uint64) time.Duration {
	if nextMS <= nowMS {
		return 0
	}
	var waitMS = nextMS - nowMS
	var maxMS = uint64(math.MaxInt64 / int64(time.Millisecond))
	if waitMS > maxMS {
		waitMS = maxMS
	}
	return time.Duration(waitMS) * time.Millisecond
}

func stopTimer(timer *time.Timer) {
	if timer.Stop() {
		return
	}
	select {
	case <-timer.C:
	default:
	}
}
