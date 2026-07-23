# RuntimeStore

## Purpose

Persist durable Bot, Runtime, BotCycle, signal, telemetry, and recovery evidence.

## Status

Approved — unimplemented.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/runtime/store.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`

## Scope

RuntimeStore owns durable state transitions required by one live or simulated Runtime path.

## Owner and Children

Runner owns one RuntimeStore handle.

Runtime and Runner call narrow store operations. They do not own persistence mechanics.

## Responsibilities

- Load one stored Bot and admitted configuration.
- Persist lifecycle transitions with expected prior state.
- Record Signals and BotCycle identity.
- Persist terminal BotCycle and Bot outcomes.
- Write and read operator telemetry.
- Return recovery state needed before admission.
- Reject stale or contradictory transitions.

## Does Not

- Decide Runtime policy.
- Own process identity or operating-system liveness.
- Reconcile Accounts.
- Store secrets.
- Expose datastore rows as mutable domain state.
- Define database schema in this page.

## Invariants

- Durable transitions are conditional and monotonic.
- Recovery reads one coherent stored state.
- Runtime errors cannot silently become successful terminal state.
- Store failures propagate to their lifecycle owner.

## Required Proof

- Invalid prior states reject transitions.
- Terminal writes are idempotent or reject duplicates clearly.
- Recovery state recreates the same Bot identity and active cycle.
- Telemetry cannot overwrite lifecycle truth.

