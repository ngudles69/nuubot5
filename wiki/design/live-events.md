# Live Event Process

## Purpose

Move admitted live BBO, bar, and user events into one Runner without moving trading policy into asynchronous work.

## Status

Approved — unimplemented.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`
- Nuubot5 contract: `wiki/ARCHITECTURE.md`

## Participants

- DataEngine acquires, validates, and multiplexes external events.
- Runner owns subscriptions and local feed state.
- Runtime consumes synchronous event and timer calls.
- Account and Ledger own dirty state.

## Ordered Flow

```text
BBO event
  DataEngine validates and publishes typed BBO
  Runner updates local BBO state
  fast Clock pass asks Runtime to evaluate responsive exits

user event
  DataEngine validates account identity
  Runner marks matching Account and Ledger dirty
  next recon timer asks Runtime to reconcile dirty Accounts

bar event
  DataEngine validates completed Bar
  Runner updates local Bar state
  Runtime admits the Bar through Signaler boundary
```

## Decisions

DataEngine decides admission and subscriber routing.

Runner decides which owned local state receives an event.

Runtime decides stop-loss, Risk, recon, BotCycle, and execution actions.

## Failure Handling

- Invalid events are rejected before local mutation.
- Subscription failure reaches Runner.
- Runtime failure reaches Runner supervision.
- Dirty state clears only after successful reconciliation.

## Does Not

- Reconcile immediately inside a user-event reader.
- Place orders inside a BBO reader.
- Let feed goroutines call Executor policy.
- Treat dirty hints as authoritative exchange truth.

## Required Proof

- BBO exits meet the configured responsive cadence.
- User events mark only their Account dirty.
- Recon waits for the Bot timer.
- Failed recon preserves dirty state.
- Event ordering and loss evidence remain observable.

