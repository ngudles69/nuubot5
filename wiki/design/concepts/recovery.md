# Recovery Process

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Rebuild durable Bot state, reconcile external truth, and reopen admission safely after interruption.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/runtime/store.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/server/process.py`

## Participants

- Server identifies recoverable processes.
- ProcessStore owns process and restart evidence.
- BotManager selects the stored Bot.
- Runner recreates direct children.
- RuntimeStore returns durable Runtime and BotCycle state.
- Runtime recreates its active domain subtree.
- Accounts reconcile external truth before admission.

## Ordered Flow

```text
Server startup
  identify recoverable Bot
  reserve recovery generation
  start Runner in recovery mode

Runner Init
  load stored Bot and Runtime state
  recreate Runtime and active BotCycle
  recreate Account bindings

Runner Start
  bootstrap market data
  reconcile every active Account
  evaluate recovered state
  start feeds and Clock
  open Runtime admission last
  mark running
```

## Decisions

ProcessStore decides whether a generation may recover.

RuntimeStore decides what durable domain state exists.

Runtime decides whether reconciled state can resume or must stop.

## Failure Handling

- Missing or contradictory durable state fails recovery.
- Account recon failure prevents admission.
- Exhausted restart policy prevents automatic retry.
- Partial recovery uses normal idempotent shutdown.

## Does Not

- Infer missing Trades, Orders, or Fills.
- Reopen admission before full recon.
- Reuse stale process identity.
- Convert failed recovery into a fresh Bot silently.

## Required Proof

- Active BotCycle identity survives restart.
- External truth wins over stale local observations.
- Every active Account reconciles before decisions resume.
- Repeated failure reaches restart exhaustion.
- Recovery and clean start remain distinguishable.
