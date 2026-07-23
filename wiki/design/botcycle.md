# BotCycle

## Purpose

Coordinate all Executors created for one accepted Signal.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/botcycle.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/executor.md`
- Nuubot5: `internal/botcycle/botcycle.go`

## Scope

BotCycle owns one Signal, configured Executors, bounded work statistics, and consolidated completion.

## Owner and Children

Runtime owns at most one active BotCycle.

BotCycle directly owns every configured Executor.

## Responsibilities

- Construct Executors through the factory.
- Pass the same accepted Signal to every Executor.
- Start Executors in configuration order.
- Deliver BBO values to non-terminal Executors.
- Run one timed pass on non-terminal Executors.
- Complete only when every Executor is terminal.
- Stop Executors in reverse ownership order.
- Consolidate one cycle exit reason.

## Does Not

- Select or release Signals.
- Assess Runtime Risks.
- Decode market data.
- Own Runtime stop policy.
- Create Account state in the current slice.

## Lifecycle

`New` constructs configured Executors.

`Start` starts every Executor.

Canonical target `Pass` advances one bounded control pass.

`Stop` is idempotent and stops children.

Current Go names `Pass` as `MainLoop`. This is documented pre-contract drift.

## Inputs and Outputs

Inputs are cycle number, accepted Signal, Executor configurations, BBO values, pass timestamps, and stop reason.

Outputs are completion state, consolidated exit reason, statistics, and errors.

## State and Invariants

All Executors receive the same Signal.

Completion requires every Executor terminal.

The cycle owns no Executor after Runtime closes it.

## Concurrency

BotCycle is synchronous.

## Persistence

None.

## Errors

Factory or start failures abort cycle creation.

Start failure triggers cleanup.

Stop attempts every Executor and returns the first error.

## Program Flow

```text
New
  create every configured Executor
  return stopped BotCycle

Start
  start Executors in order
  clean up on failure
  mark running

OnBBO
  record cycle range
  send BBO to non-terminal Executors

Pass
  run non-terminal Executors
  return whether all are terminal

Stop
  stop Executors in reverse order
  consolidate exit reason
  report BotCycle statistics
```

## Required Proof

- Every configured Executor starts once.
- Each admitted BBO reaches every non-terminal Executor.
- Completion waits for all Executors.
- Start failure cleans up started work.
- End-date stop closes an active cycle.

## Open Decisions

Future Account creation and registration belong to Executor design, not the current BotCycle.
