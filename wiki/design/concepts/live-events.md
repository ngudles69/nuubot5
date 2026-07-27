# Live Event Process

Status: Runner ownership approved; event and subscription mechanics remain unimplemented.
Covers: No implemented source.
Purpose: Record the Runner-owned live event flow without moving trading policy
into asynchronous work.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`

## Participants

- Runner-owned WebSocket acquires and validates external events.
- Runner owns the shared endpoint lifecycle.
- Controller consumes synchronous event and timer calls.
- Account solely owns reconciliation-dirty state.

## Ordered Flow

```text
BBO event
  validate and publish typed BBO through the selected transport
  update Runner-local BBO state
  ask Controller to evaluate responsive exits on fast Clock timer

user event
  validate account identity through the selected transport
  mark matching Account recon-dirty
  ask Controller to reconcile dirty Accounts on next recon timer

bar event
  validate completed Bar through the selected transport
  update Runner-local Bar state
  admit Bar through Controller Signaler boundary
```

## Decisions

Runner-owned WebSocket decides subscription, message admission, and callback routing.

Requesting components provide their subscription requirements and callbacks.

Controller decides stop-loss, Risk, recon, BotCycle, and execution actions.

## Failure Handling

- Invalid events are rejected before local mutation.
- Subscription failure reaches Runner.
- Controller failure reaches Runner supervision.
- Dirty state clears only after successful reconciliation.

WebSocket delivery is never required for final correctness.

Missing, duplicated, delayed, or reordered events are repaired by forced reconciliation.

When no user event arrives, the normal recon pass skips the clean Account.

This reduces Venue queries without making WebSocket delivery authoritative.

## Does Not

- Reconcile immediately inside a user-event reader.
- Place orders inside a BBO reader.
- Let feed goroutines call Executor policy.
- Treat dirty hints as authoritative exchange truth.
- Commit Order or Fill truth from WebSocket delivery.

## Required Proof

- BBO exits meet the configured responsive cadence.
- User events mark only their Account dirty.
- Recon waits for the Bot timer.
- Failed recon preserves dirty state.
- Event ordering and loss evidence remain observable.
- Standalone Runner works while Server is stopped.

## Open Decisions

- Exact subscription identity and callback contract.
- Reconnection and resubscription mechanics.
