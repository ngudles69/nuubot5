package clock

import "nuubot/internal/toolkit/logging"

// TickClock advances from admitted replay timestamps.
type TickClock struct {
	state clockState
}

// Section 1 - Program Flow

// Init prepares TickClock at its initial replay timestamp.
func (c *TickClock) Init(log *logging.Logger, initialMS uint64) error {
	// initialize clock
	return c.state.init(log, Tick, initialMS)
}

// Start starts TickClock timer admission.
func (c *TickClock) Start() error {
	// start clock
	return c.state.start()
}

// Stop stops TickClock timer admission.
func (c *TickClock) Stop() {
	// stop clock
	c.state.stop()
}

// Section 2 - Domain Helpers

// Err returns TickClock's terminal callback error.
func (c *TickClock) Err() error {
	// read error
	return c.state.clockErr()
}

// NowMS returns TickClock's admitted replay time.
func (c *TickClock) NowMS() uint64 {
	// read time
	return c.state.timeMS()
}

// RegisterTimer registers one named TickClock timer.
func (c *TickClock) RegisterTimer(timer Timer, callback func(uint64) error) error {
	// register timer
	return c.state.registerTimer(timer, callback)
}

// Advance advances TickClock and fires every due timer.
func (c *TickClock) Advance(nowMS uint64) error {
	// check timers
	return c.state.timerCheck(nowMS)
}

// NextFireMS returns one TickClock timer's next scheduled timestamp.
func (c *TickClock) NextFireMS(name string) (uint64, bool) {
	// read next fire
	return c.state.nextFireMS(name)
}

// CancelTimer cancels one TickClock timer.
func (c *TickClock) CancelTimer(name string) {
	// cancel timer
	c.state.cancelTimer(name)
}

// Section 3 - Generic Helpers
