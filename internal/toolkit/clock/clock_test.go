package clock

import (
	"errors"
	"io"
	"testing"
	"time"

	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestClocksShareLifecycleAndTimeContract(t *testing.T) {
	var log = logging.Create(io.Discard)
	var tickClock, err = Create(Tick)
	if err != nil {
		t.Fatalf("create TickClock: %v", err)
	}
	err = tickClock.Init(log, 100)
	if err != nil {
		t.Fatalf("initialize TickClock: %v", err)
	}
	if tickClock.NowMS() != 100 {
		t.Fatalf("actual TickClock time %d, expected 100", tickClock.NowMS())
	}

	var wallClock Clock
	wallClock, err = Create(Wall)
	if err != nil {
		t.Fatalf("create WallClock: %v", err)
	}
	var wallMS = wallClock.NowMS()
	err = wallClock.Init(log, wallMS)
	if err != nil {
		t.Fatalf("initialize WallClock: %v", err)
	}
	if wallClock.NowMS() < wallMS {
		t.Fatalf(
			"actual WallClock time %d, expected at least %d",
			wallClock.NowMS(),
			wallMS,
		)
	}
}

func TestTickClockFiresMultipleTimersInScheduledOrder(t *testing.T) {
	var calls []timerCall
	var tickClock = createTickClock(t, 100)
	registerTimer(t, tickClock, Timer{Name: "fast", IntervalMS: 10}, &calls)
	registerTimer(t, tickClock, Timer{Name: "slow", IntervalMS: 20}, &calls)
	var err = tickClock.Start()
	if err != nil {
		t.Fatalf("start TickClock: %v", err)
	}

	err = tickClock.Advance(120)
	if err != nil {
		t.Fatalf("advance TickClock: %v", err)
	}

	var expected = []timerCall{
		{name: "fast", fireMS: 110},
		{name: "fast", fireMS: 120},
		{name: "slow", fireMS: 120},
	}
	assertTimerCalls(t, calls, expected)
}

func TestTickClockUsesOptionalTimerStartAndStop(t *testing.T) {
	var startMS = uint64(105)
	var stopMS = uint64(125)
	var calls []timerCall
	var tickClock = createTickClock(t, 100)
	registerTimer(t, tickClock, Timer{
		Name:       "bounded",
		IntervalMS: 10,
		StartMS:    &startMS,
		StopMS:     &stopMS,
	}, &calls)
	var err = tickClock.Start()
	if err != nil {
		t.Fatalf("start TickClock: %v", err)
	}

	err = tickClock.Advance(140)
	if err != nil {
		t.Fatalf("advance TickClock: %v", err)
	}

	var expected = []timerCall{
		{name: "bounded", fireMS: 115},
		{name: "bounded", fireMS: 125},
	}
	assertTimerCalls(t, calls, expected)
	if _, exists := tickClock.NextFireMS("bounded"); exists {
		t.Fatal("bounded timer remains after its stop time")
	}
}

func TestTickClockRejectsBackwardTime(t *testing.T) {
	var tickClock = createStartedTickClock(t, 100)
	var err = tickClock.Advance(110)
	if err != nil {
		t.Fatalf("advance TickClock: %v", err)
	}

	err = tickClock.Advance(109)
	if err == nil {
		t.Fatal("backward TickClock advance succeeded")
	}
}

func TestTickClockReturnsCallbackError(t *testing.T) {
	var callbackErr = errors.New("callback failed")
	var tickClock = createTickClock(t, 100)
	var err = tickClock.RegisterTimer(
		Timer{Name: "runtime", IntervalMS: 10},
		func(uint64) error {
			return callbackErr
		},
	)
	if err != nil {
		t.Fatalf("register timer: %v", err)
	}
	err = tickClock.Start()
	if err != nil {
		t.Fatalf("start TickClock: %v", err)
	}

	err = tickClock.Advance(110)
	if !errors.Is(err, callbackErr) {
		t.Fatalf("actual error %v, expected callback error", err)
	}
}

func TestWallClockAdvancesItself(t *testing.T) {
	var wallClock, err = Create(Wall)
	if err != nil {
		t.Fatalf("create WallClock: %v", err)
	}
	var initialMS = wallClock.NowMS()
	err = wallClock.Init(logging.Create(io.Discard), initialMS)
	if err != nil {
		t.Fatalf("initialize WallClock: %v", err)
	}

	var fired = make(chan uint64, 1)
	err = wallClock.RegisterTimer(
		Timer{Name: "runtime", IntervalMS: 10},
		func(fireMS uint64) error {
			select {
			case fired <- fireMS:
			default:
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("register WallClock timer: %v", err)
	}
	err = wallClock.Start()
	if err != nil {
		t.Fatalf("start WallClock: %v", err)
	}
	defer wallClock.Stop()

	select {
	case fireMS := <-fired:
		if fireMS != initialMS+10 {
			t.Fatalf("actual fire time %d, expected %d", fireMS, initialMS+10)
		}
	case <-time.After(time.Second):
		t.Fatal("WallClock did not advance itself")
	}
}

// Section 2 - Domain Helpers

type timerCall struct {
	name   string
	fireMS uint64
}

func createTickClock(t *testing.T, initialMS uint64) Clock {
	t.Helper()
	var tickClock, err = Create(Tick)
	if err != nil {
		t.Fatalf("create TickClock: %v", err)
	}
	err = tickClock.Init(logging.Create(io.Discard), initialMS)
	if err != nil {
		t.Fatalf("initialize TickClock: %v", err)
	}
	return tickClock
}

func createStartedTickClock(
	t *testing.T,
	initialMS uint64,
) Clock {
	t.Helper()
	var tickClock = createTickClock(t, initialMS)
	var err = tickClock.Start()
	if err != nil {
		t.Fatalf("start TickClock: %v", err)
	}
	return tickClock
}

func registerTimer(
	t *testing.T,
	tickClock Clock,
	timer Timer,
	calls *[]timerCall,
) {
	t.Helper()
	var err = tickClock.RegisterTimer(timer, func(fireMS uint64) error {
		*calls = append(*calls, timerCall{name: timer.Name, fireMS: fireMS})
		return nil
	})
	if err != nil {
		t.Fatalf("register timer: %v", err)
	}
}

func assertTimerCalls(t *testing.T, actual, expected []timerCall) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("actual calls %v, expected %v", actual, expected)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("actual calls %v, expected %v", actual, expected)
		}
	}
}

// Section 3 - Generic Helpers
