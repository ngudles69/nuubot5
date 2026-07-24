# SweepManager

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Own Server-side Sweep configuration and process-control requests.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/sweeps/sweepmgr.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/sweeprunner.py`

## Scope

SweepManager is an HTTP-agnostic Server use-case service.

It validates and stores Sweep requests, then asks process supervision to launch
or stop one standalone SweepRunner.

## Owner and Children

Server owns SweepManager.

SweepManager owns no SweepRunner, Controller, or BtRunner internals.

## Responsibilities

- Create, clone, read, list, update, and delete Sweeps.
- Validate Sweep configuration before persistence.
- Start and stop one Sweep through its lifecycle owner.
- Report Sweep status, summary, and completed Bot results.
- Preserve exact Sweep and child Bot identities.

## Does Not

- Execute replay directly.
- Expand parameter permutations.
- Create or manage a worker pool.
- Import BtRunner.
- Interpret strategy results.
- Mutate Controller descendants.
- Manage live Runners.
- Publish a BtRunner result itself.

## Invariants

- Sweep identity and Sweep Bot identity remain distinct.
- Every launched BtRunner has one stored Sweep Bot.
- SweepRunner owns expansion, bounded workers, cancellation, and aggregation.
- BtRunner owns one child Bot replay.
- Sweep status describes the latest execution, not a terminal Sweep identity.
- Stopped and error Sweeps may run again.
- A rerun replays selected Bots from the beginning.
- No partial recovery or resume exists.

## Execution Flow

```text
API
  -> SweepManager
  -> process supervision
  -> standalone SweepRunner
  -> bounded BtRunner workers
```

Workers receive stored identities only.

They never receive Config objects, Controllers, replay iterators, Accounts, or
result writers across the process boundary.

## Expansion

One Sweep targets one exact BotSpecID and one complete base BotConfig.

Explicit ordered value lists define parameter dimensions.

Paths must exist and be recognized by the exact BotSpec.

Ignored extra Config fields cannot become Sweep dimensions.

SweepRunner sorts parameter paths, preserves value order and date-range order,
and generates one deterministic Cartesian product.

Every generated Bot passes normal admission.

Range syntax remains deferred.

## Rerun

First execution creates stable child Bot identities.

Every new execution clears selected current results before work starts.

Rerun replaces one current result per child.

Clone copies only the Sweep definition when both result sets must remain.

No execution history, checkpoint, partial resume, or worker recovery exists.

A failed child remains visible.

Other submitted children continue.

The latest Sweep execution is error when any selected child fails.

## Deferred Period Analysis

A future reusable period catalog may define labels such as:

```text
2021
2021-H1
2021-Q1
2021-M01
```

Each period keeps exact start and end timestamps, description, and searchable
tags such as `event:black_swan`, `regime:bearish`, or `volatility:high`.

SweepManager may select periods but will not own that catalog.

Backtest results will preserve the selected period snapshot.

Future analytics may group Calmar, CAGR, return, drawdown, Executor activation,
and Executor contribution across tagged periods.

Analytics must show distributions and exact durations, not one misleading
average.

The catalog, tag taxonomy, and analytics remain deferred.

## Required Proof

- Invalid Sweep configuration never starts.
- Worker limits remain bounded.
- Failed BtRunner results remain visible.
- Stop reaches every active Sweep worker.
- Aggregate completion matches stored current child results.
- A rerun after crash starts selected children from the beginning.
