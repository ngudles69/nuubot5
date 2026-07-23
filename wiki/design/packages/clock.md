# Clock Package

Status: Implemented.
Covers: `internal/toolkit/clock/clock.go`
Purpose: Convert replay timestamps into deterministic Runtime pass decisions.

## Canonical Source

- `D:/rust/nuubot4/src/clock.rs`

## Scope & Responsibilities

TickClock owns one replay interval and its next due timestamp.

## Program Flow

```text
TickClock(log, interval)

init
  next = unset

start
  accept replay time

run(now)
  first tick   -> due
  now >= next -> advance next and return due
  otherwise   -> return not due

stop
  stop once
```

## Notes

- TickClock uses admitted replay time, never wall time.
