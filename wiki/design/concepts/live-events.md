# Live Event Process

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Move admitted live BBO, bar, and user events into one Runner without moving trading policy into asynchronous work.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`

## Participants

- DataEngine acquires, validates, and multiplexes external events.
- Runner owns subscriptions and local feed state.
- Runtime consumes synchronous event and timer calls.
- Account and Ledger own dirty state.

## Ordered Flow

```text
BBO event
  validate and publish typed BBO through DataEngine
  update Runner-local BBO state
  ask Runtime to evaluate responsive exits on fast Clock timer

user event
  validate account identity through DataEngine
  mark matching Account and Ledger dirty
  ask Runtime to reconcile dirty Accounts on next recon timer

bar event
  validate completed Bar through DataEngine
  update Runner-local Bar state
  admit Bar through Runtime Signaler boundary
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
