# SweepManager

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Own Sweep configuration, lifecycle, and aggregate progress.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/sweeps/sweepmgr.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/sweeprunner.py`

## Scope

SweepManager validates Sweep requests and controls one SweepRunner or equivalent bounded backtest supervisor.

## Owner and Children

Server owns SweepManager.

SweepManager owns no Runtime or BtRunner internals.

## Responsibilities

- Create, clone, read, list, update, and delete Sweeps.
- Validate Sweep configuration before persistence.
- Start and stop one Sweep through its lifecycle owner.
- Report Sweep status, summary, and completed Bot results.
- Preserve bounded worker limits and exact Bot identities.

## Does Not

- Execute replay directly.
- Interpret strategy results.
- Mutate Runtime descendants.
- Manage live Runners.
- Publish a BtRunner result itself.

## Invariants

- Sweep identity and Sweep Bot identity remain distinct.
- Every launched BtRunner has one stored Sweep Bot.
- A Sweep finishes only after every terminal Bot result is accounted for.

## Required Proof

- Invalid Sweep configuration never starts.
- Worker limits remain bounded.
- Failed BtRunner results remain visible.
- Stop reaches every active Sweep worker.
- Aggregate completion matches stored terminal results.
