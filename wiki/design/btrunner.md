# BtRunner

## Purpose

Run one bounded historical Bot replay and prove exact completion.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/bin/nuubot-btrunner.rs`
- Nuubot4: `D:/rust/nuubot4/src/btrunner.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/btrunner.md`
- Nuubot5 command: `cmd/nuubot-btrunner/main.go`
- Nuubot5: `internal/btrunner/btrunner.go`

## Scope

BtRunner composes and supervises one TickClock, replay Reader, and Runtime.

## Owner and Children

The command owns BtRunner.

BtRunner directly owns:

- one replay Reader;
- one TickClock; and
- one Runtime.

## Responsibilities

- Prepare one admitted Bot replay.
- Resolve the effective replay end.
- Start replay at `ReplayStart`.
- Load Signaler bars before starting.
- Serve every validated BBO to Runtime.
- Trigger Runtime passes through TickClock.
- Stop Runtime at the effective end.
- Verify exact replay counts and range.
- Stop direct children in reverse ownership order.

## Does Not

- Calculate signals.
- Manage BotCycle internals.
- Execute trading policy.
- Collect child statistics for re-logging.
- Persist replay results.
- Own Account, Ledger, Trade, Order, Fill, or Simulator.

## Lifecycle

`New` prepares the complete stopped object.

`Start` starts Runtime.

`Run` performs one finite replay.

`Stop` is idempotent and releases direct children.

## Inputs and Outputs

Inputs are logger, repository root, configuration, Sweep ID, and Bot ID.

Output is successful exact replay or one terminal error.

## State and Invariants

The effective end is the earlier Bot end or replay end.

Replay start MUST precede effective end.

Current BtRunner intentionally ignores `StartAt`.

Reader, Bars, expected counts, and expected boundaries all start at `ReplayStart`.

Every accepted tick reaches Runtime before TickClock advances.

Successful completion requires exact ticks, passes, first timestamp, and last timestamp.

## Concurrency

BtRunner is synchronous.

## Persistence

BtRunner reads configuration, SQLite Bot specification, and Parquet market data.

BtRunner writes no domain state.

## Errors

Preparation, child, replay, verification, and shutdown errors propagate to the command boundary.

The command MUST attempt `Stop` after a started `Run`, including failed runs.

Current Go verifies after Runtime end-date stop. Final command teardown still calls BtRunner `Stop`.

## Program Flow

```text
New
  run Setup
  use ReplayStart as replay start
  ignore StartAt
  choose effective end
  create TickClock
  create replay Reader
  create Runtime
  load required Bars
  prepare Signaler
  calculate expected proof

Start
  start Runtime
  mark started

Run
  read next BBO
  send BBO to Runtime
  record replay evidence
  advance TickClock
  run Runtime Pass when due
  stop at Runtime request or reader completion
  stop Runtime with end_date
  verify exact replay

Stop
  stop Runtime
  stop replay Reader
  stop TickClock
  report BtRunner statistics
```

## Required Proof

- 7,948,800 ticks and 794,880 passes complete for Sweep 6 Bot 9.
- The first and last timestamps match the selected range.
- Runtime reports 55 signals and 18 closed cycles.
- Exit zero and `replay_completed=true` both exist.
- Canonical `-tags noasm` fresh-process stability remains proven.

## Open Decisions

Result persistence has no approved owner or schema.
