package clock

import (
	"fmt"
	"sync"

	"nuubot/internal/toolkit/logging"
)

// Kind selects one clock implementation.
type Kind string

const (
	// Tick selects replay time supplied by the clock owner.
	Tick Kind = "tick"
	// Wall selects current UTC wall time.
	Wall Kind = "wall"
)

// Clock defines the common TickClock and WallClock contract.
type Clock interface {
	Init(*logging.Logger, uint64) error
	Start() error
	Stop()
	Err() error
	NowMS() uint64
	RegisterTimer(Timer, func(uint64) error) error
	Advance(uint64) error
	NextFireMS(string) (uint64, bool)
	CancelTimer(string)
}

// clockState owns lifecycle and timer state shared by both Clocks.
type clockState struct {
	mu          sync.Mutex
	advanceMu   sync.Mutex
	log         *logging.Logger
	kind        Kind
	nowMS       uint64
	timers      map[string]*timerState
	advances    uint64
	timersFired uint64
	initialized bool
	started     bool
	stopped     bool
	err         error
}

// Section 1 - Program Flow

// Create constructs the selected Clock.
func Create(kind Kind) (Clock, error) {
	// select implementation
	switch kind {
	case Tick:
		return &TickClock{}, nil
	case Wall:
		return &WallClock{}, nil
	default:
		return nil, fmt.Errorf("unknown clock: %s", kind)
	}
}

func (c *clockState) init(log *logging.Logger, kind Kind, initialMS uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// validate state
	if log == nil {
		return fmt.Errorf("%s clock logger is required", kind)
	}
	if c.initialized {
		return fmt.Errorf("%s clock is already initialized", kind)
	}

	// initialize clock
	c.log = log
	c.kind = kind
	c.nowMS = initialMS
	c.timers = make(map[string]*timerState)
	c.initialized = true
	c.log.Info(fmt.Sprintf("%s clock initialized time_ms=%d", kind, initialMS))
	return nil
}

func (c *clockState) start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// validate state
	if !c.initialized || c.started || c.stopped {
		return fmt.Errorf("%s clock cannot start from current state", c.kind)
	}

	// start clock
	c.started = true
	c.log.Info(fmt.Sprintf("%s clock started", c.kind))
	return nil
}

func (c *clockState) stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized || c.stopped {
		return
	}

	// stop clock
	c.started = false
	c.stopped = true
	c.log.Info(fmt.Sprintf(
		"%s clock stopped advances=%d timers_triggered=%d active_timers=%d",
		c.kind,
		c.advances,
		c.timersFired,
		len(c.timers),
	))
}

// Section 2 - Domain Helpers

func (c *clockState) timeMS() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nowMS
}

func (c *clockState) clockErr() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *clockState) running() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started && !c.stopped
}

// Section 3 - Generic Helpers
