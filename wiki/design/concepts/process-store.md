# ProcessStore

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Persist standalone process identity, command reservation, health, and bounded restart evidence.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/server/process.py`
- Nuubot3 store evidence: `D:/rust/nuubot3/nuubot/runtime/store.py`

## Scope

ProcessStore owns durable coordination records used by Server and RunnerControl.

## Owner and Children

Server owns ProcessStore.

ProcessStore owns no operating-system process and no Runner.

## Responsibilities

- Reserve one start or control action atomically.
- Record process identity and generation.
- Record health observations and command completion.
- Expose restart eligibility from durable evidence.
- Mark dead, exhausted, failed, or terminal processes.
- Reject stale generations and duplicate actions.

## Does Not

- Spawn, terminate, or signal processes.
- Probe operating-system liveness.
- Implement restart policy.
- Store Runtime domain state.
- Define database schema in this page.

## Invariants

- Process identity includes enough evidence to reject PID reuse.
- One generation owns one active command.
- Command completion matches its reservation.
- Restart counters are durable and bounded.

## Required Proof

- Concurrent starts reserve one winner.
- Stale generations cannot complete current commands.
- PID reuse fails identity checks.
- Restart exhaustion remains durable.
