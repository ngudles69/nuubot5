# WallClock

Status: Implemented.
Covers: `internal/toolkit/clock/wallclock.go`, `internal/toolkit/clock/timer.go`
Purpose: Drive live Bot callbacks from UTC wall time.

## Canonical Source

- NautilusTrader: `D:/rust/nuutrader-references/nautilus_trader/crates/common/src/live/clock.rs`

## Scope

WallClock implements the same Clock contract as TickClock.

Runner owns WallClock. WallClock owns its wall-time advancement loop.

## Flow

```text
Clock.Create(Wall)
WallClock.Init
  initialize clock
  initialize loop

WallClock.RegisterTimer
  register timer

WallClock.Start
  start clock
  start loop

WallClock.Loop
  read next timer
  wait for timer
  advance clock

WallClock.Stop
  stop loop
  stop clock
```

## Invariants

- `NowMS` returns current UTC wall time.
- `Advance` uses the same timer mechanics as TickClock.
- Timer callbacks receive their scheduled fire timestamp.
- Callback failure stops the loop and remains available through `Err`.
- WallClock serializes its own advancement.
- WallClock owns no Runtime, Account, feed, or trading policy.
