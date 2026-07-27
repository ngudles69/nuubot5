# DataEngine

Status: Superseded for standalone Runner; possible Server optimization remains unapproved.
Covers: No implemented source.
Purpose: Record the earlier shared live-event candidate.

## Canonical Sources

- Nuubot3 Runner boundary: `D:/rust/nuubot3/nuubot/runner/runner.py`
- Nuutrader6 evidence: `D:/rust/nuutrader6/src/nuubot/services/data_engine.py`

## Scope

Standalone Runner owns its process-local shared WebSocket.

If DataEngine is later approved for Server optimization, it may own shared WebSocket connections,
subscription reference counts, message admission, reconnection, and typed event
distribution for Server-supervised Runners.

It cannot become a required dependency of standalone Runner.

## Owner and Children

Possible Server ownership remains unapproved.

Standalone Runner must be able to own the live transport required for its Bot.

Public, private, venue, network, and Account sharing boundaries remain TBD.

## Responsibilities

- Share one upstream subscription across identical Runner requirements.
- Reference-count subscriptions by admitted identity.
- Validate every external message before publication.
- Normalize admitted messages into typed BBO, bar, or user events.
- Multiplex typed events only to matching subscribers.
- Reconnect and restore active subscriptions.
- Report connection, subscription, rejection, and reconnect evidence.

## Does Not

- Own Runner or Controller.
- Store Bot-local mutable feed state.
- Place or cancel orders.
- Mark domain objects dirty directly.
- Reconcile Accounts.
- Execute stop-loss, signal, risk, or execution policy.
- Select transport, SDK, or persistence technology here.

## Lifecycle

`NewDataEngine` constructs one stopped service.

`Start` opens owned stream supervision.

`Subscribe` admits one Runner requirement and returns its owned subscription.

`Unsubscribe` releases one Runner requirement.

`Stop` closes admission, subscriptions, streams, and distribution work.

## Data Flow

```text
upstream message
  validate shape and identity
  normalize typed event
  find matching subscriptions
  deliver to Runner-owned local feed state
```

## Invariants

- External data stays untrusted until validation succeeds.
- One subscriber cannot receive another Bot's user event.
- Slow subscribers cannot mutate shared stream state.
- Reconnect restores only active subscriptions.
- Subscription cleanup is idempotent.

## Required Proof Before Approval

- Shared BBO subscription creates one upstream requirement.
- Invalid messages are rejected and counted.
- Reconnect restores active subscriptions once.
- Unsubscribe removes delivery without affecting remaining subscribers.
- User events remain isolated by admitted account identity.
- Standalone Runner remains fully functional while Server is stopped.
