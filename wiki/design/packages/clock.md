# Clock Package

Status: Implemented.
Covers: `internal/toolkit/clock/*.go`
Purpose: Drive named timers from replay or UTC wall time through one contract.

## Canonical Source

- NautilusTrader: `D:/rust/nuutrader-references/nautilus_trader/crates/common/src/clock.rs`
- NautilusTrader: `D:/rust/nuutrader-references/nautilus_trader/crates/common/src/timer.rs`
- NautilusTrader: `D:/rust/nuutrader-references/nautilus_trader/crates/common/src/live/clock.rs`

## Contract

TickClock and WallClock implement:

```text
Init
Start
Err
NowMS
RegisterTimer
Advance
NextFireMS
CancelTimer
Stop
```

`Create` selects TickClock or WallClock.

TickClock `NowMS` returns its last admitted replay timestamp.

WallClock `NowMS` returns current UTC wall time.

## File Ownership

- `clock.go` owns the contract, factory, and shared Clock state.
- `timer.go` owns timer definitions, validation, ordering, scheduling, and dispatch.
- `tickclock.go` owns replay-driven advancement.
- `wallclock.go` owns wall-time reads and its advancement loop.

## Program Flow

```text
create
  select implementation

init
  validate state
  initialize clock

NowMS
  read time

RegisterTimer
  validate state
  validate timer
  schedule timer
  register timer

start
  validate state
  start clock

Err
  read error

Advance
  validate state
  validate time
  check timers
    select next timer
    schedule timer
    run timer callback
  advance time

NextFireMS
  read next fire

CancelTimer
  cancel timer

stop
  stop clock
```

## Timer Contract

- Timer names are unique per Clock.
- Interval must be positive.
- Missing start uses the Clock initialization timestamp.
- First fire occurs at start plus interval.
- Optional stop is inclusive when a scheduled fire equals it.
- Every due interval fires.
- Events order by scheduled timestamp, then timer name.
- Callbacks receive scheduled fire time.
- Callback errors return through `Advance`.
- WallClock callback errors also remain available through `Err`.
- Backward advancement fails.

## Trigger Ownership

BtRunner advances TickClock from admitted replay ticks.

WallClock `Start` launches its wall-time loop. The loop waits for the next
timer and calls `Advance(NowMS())`.

Clock owns timer checking. Runtime policy stays inside registered callbacks.
