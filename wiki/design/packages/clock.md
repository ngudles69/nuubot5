# Clock Package

Status: Implemented.
Covers: `internal/toolkit/clock/clock.go`
Purpose: Run one registered timer from deterministic replay timestamps.

## Canonical Source

- `D:/rust/nuubot4/src/clock.rs`

## Scope & Responsibilities

TickClock owns one registered replay timer, its interval, and its next timestamp.

## Program Flow

```text
TickClock(log)

init
  timer = unset

register(interval, callback)
  timer = interval and callback
  next  = unset

run(now)
  first tick   -> invoke callback
  now >= next -> advance next and invoke callback
  otherwise   -> return
  propagate callback error

stop
  stop once
```

## Notes

- TickClock uses admitted replay time, never wall time.
- TickClock owns timer mechanics, not Runtime policy.
