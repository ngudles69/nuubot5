# RuntimeStore

Status: Candidate only. Ownership and scope TBD.
Covers: No implemented source.
Purpose: Retain the earlier RuntimeStore candidate without approving recovery,
telemetry persistence, or standalone datastore ownership.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/runtime/store.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`

## Scope

Standalone Runner persistence remains unresolved.

Recovery and telemetry persistence remain deferred.

## Owner and Children

Ownership is not approved.

If retained, Runner and Controller may call narrow store operations.

## Responsibilities

Possible responsibilities requiring later approval:

- Persist lifecycle transitions with expected prior state.
- Record Signal and BotCycle evidence.
- Persist terminal BotCycle and Bot outcomes.
- Reject stale or contradictory transitions.

## Does Not

- Decide Controller policy.
- Own process identity or operating-system liveness.
- Reconcile Accounts.
- Store secrets.
- Expose datastore rows as mutable domain state.
- Define database schema in this page.
- Approve recovery or telemetry persistence.

## Invariants

- Durable transitions are conditional and monotonic.
- Controller errors cannot silently become successful terminal state.
- Store failures propagate to their lifecycle owner.

## Required Proof

- Invalid prior states reject transitions.
- Terminal writes are idempotent or reject duplicates clearly.
- Standalone Runner remains functional while Server is stopped.
- Any later telemetry cannot overwrite lifecycle truth.

## Open Decisions

- Whether RuntimeStore exists.
- Standalone Runner datastore ownership and writes.
- Recovery state and policy.
- Telemetry schema, cadence, and persistence.
