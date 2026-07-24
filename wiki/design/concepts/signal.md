# Signal

Status: Implemented.
Covers: `internal/signaler/signaler.go`
Purpose: Carry one immutable, timestamped trading intent from Signaler to Runtime and BotCycle.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/signaler.rs`

## Scope

Signal contains event time, availability time, side, and observed close price.

## Owner and Children

Signaler creates Signals.

Runtime accepts them. BotCycle and Executors consume copies.

Signal owns no child.

## Responsibilities

- Identify long or short intent.
- Preserve the source Bar start timestamp.
- Preserve the earliest safe release timestamp.
- Preserve the source closing price.

## Does Not

- Decide release timing.
- Create BotCycles.
- Place orders.
- Track execution state.
- Own mutable lifecycle.

## Lifecycle

Calculate once, validate once, release once, and pass by value.

## Inputs and Outputs

Inputs are one closed indicator Bar and calculator decision.

Output is one `signaler.Signal`.

## State and Invariants

Side MUST be `long` or `short`.

Signal timestamp MUST precede availability timestamp.

Availability MUST represent the close of all data used by the decision.

Price MUST be the Signal Bar close.

## Concurrency

Signal is an immutable value after creation.

## Persistence

None.

## Errors

Signaler initialization rejects invalid timestamp ordering.

## Program Flow

```text
Init
  select calculator
  resolve requirements
  load ohlcv
  calculate signals
  validate signals
  initialize signaler

Run
  release signal
```

## Required Proof

- Equal-time BBO cannot observe a Signal early.
- Released Signals retain exact side, timestamps, and price.
- Each Signal releases once.

## Open Decisions

Future persistent Signal identity has no approved contract.
