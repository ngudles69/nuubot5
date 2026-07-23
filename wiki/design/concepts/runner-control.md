# RunnerControl

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Start, supervise, command, and stop standalone Runner processes.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/server/process.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/service.py`

## Scope

RunnerControl owns operating-system process actions and the narrow authenticated control boundary to one Runner.

## Owner and Children

Server owns RunnerControl.

RunnerControl uses ProcessStore for durable coordination.

RunnerService owns the in-process Runner.

## Responsibilities

- Reserve a Runner action before external effects.
- Spawn one exact Bot generation.
- Verify process identity and health.
- Send start, pause, resume, and stop commands.
- Bound command timeouts and restart attempts.
- Terminate only an exact verified process identity.
- Report command completion or failure through ProcessStore.

## Does Not

- Own Runtime.
- Interpret Bot configuration.
- Implement Runner lifecycle ordering.
- Reach through RunnerService into Runner children.
- Store domain state.
- Select control transport in this page.

## Invariants

- Commands target one Bot generation.
- Unverified process identity is never terminated.
- Restart policy is bounded.
- Runner performs in-process transitions.

## Required Proof

- Duplicate start produces one process.
- Unauthorized or stale control fails.
- PID reuse cannot target another process.
- Unresponsive control reaches bounded failure.
- Restart exhaustion prevents further automatic starts.
