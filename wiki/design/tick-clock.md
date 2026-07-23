# TickClock

## Purpose

Convert replay timestamps into deterministic Runtime pass decisions.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/clock.rs`
- Nuubot5: `internal/clock/clock.go`

## Scope

TickClock owns interval scheduling derived only from admitted replay timestamps.

## Owner and Children

BtRunner owns one TickClock. TickClock owns no child.

## Responsibilities

- Mark the first tick due.
- Track the next due timestamp.
- Advance across one or more missed intervals.
- Count seen ticks and due passes.
- Reject arithmetic overflow.

## Does Not

- Read wall time.
- Sleep.
- Call Runtime.
- Own callbacks.
- Read market data.

## Lifecycle

`New` records the positive configured interval.

Repeated `Advance` returns whether one Runtime pass is due.

`Stop` is idempotent and reports statistics.

## Inputs and Outputs

Input is one admitted timestamp in milliseconds.

Output is a due flag or overflow error.

## State and Invariants

BtRunner configuration MUST supply a positive interval.

First `Advance` returns due and anchors the next deadline.

Later calls return at most one due decision per tick.

## Concurrency

TickClock is synchronous and has one owner.

## Persistence

None.

## Errors

Timestamp arithmetic overflow fails the replay.

## Program Flow

```text
Advance
  count tick
  if first tick
    set next deadline
    return due
  if before deadline
    return not due
  move deadline beyond current timestamp
  return due
```

## Required Proof

- First tick is due.
- Ticks before the deadline are not due.
- Deadline and multi-interval jumps produce one due result.
- Overflow fails.

## Open Decisions

None.
