# WallClock

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Drive live Bot callbacks from UTC wall time.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/wiki/logic/runner.md`
- Nuubot3: `D:/rust/nuubot3/nuubot/core/clock.py`

## Scope

WallClock owns live time advancement and registered callback schedules for one Runner.

## Owner and Children

Runner owns WallClock.

WallClock owns no Runtime, Account, feed, or trading object.

## Responsibilities

- Return current UTC milliseconds.
- Register named timer callbacks before Start.
- Advance timers on one bounded internal tick.
- Dispatch each due callback once per advance.
- Prevent overlapping dispatch for one clock.
- Stop dispatch before Runner tears down inputs.

## Does Not

- Decide callback business order.
- Execute Runtime policy outside registered callbacks.
- Read market data.
- Reconcile Accounts.
- Persist state.
- Retry failed callbacks silently.

## Lifecycle

`NewWallClock` constructs a stopped clock.

`Start` opens dispatch.

`Loop` advances wall time until cancellation or failure.

`Stop` closes dispatch and releases owned timer work.

## Invariants

- Runner registers callback cadence and intent.
- A callback receives its admitted `now_ms` value.
- Callback failure reaches Runner.
- Stop is idempotent after successful Start.

## Required Proof

- Fast BBO checks and slower recon requests follow configured cadences.
- Callback failure stops supervision.
- Cancellation ends Loop without leaked work.
- No callback runs after Stop completes.
