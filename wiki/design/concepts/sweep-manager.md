# SweepManager

Status: Approved — SweepManager unimplemented; Sweep template admission implemented.
Covers: `internal/btsweep/**` for template admission. No SweepManager source.
Purpose: Own Server-side Sweep configuration and process-control requests.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/sweeps/sweepmgr.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/sweeprunner.py`

## Scope

SweepManager is an HTTP-agnostic Server use-case service.

The target validates and stores Sweep requests, then asks process supervision
to launch or stop one standalone BtSweep.

Current `internal/btsweep` loads, validates, and expands Sweep templates only.

## Owner and Children

Server owns SweepManager.

SweepManager owns no BtSweep, Controller, or BtBot internals.

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
- Import BtBot.
- Interpret strategy results.
- Mutate Controller descendants.
- Manage live Runners.
- Publish a BtBot result itself.

## Invariants

- Sweep identity and Sweep Bot identity remain distinct.
- Every launched BtBot has one stored Sweep Bot.
- `internal/btsweep` owns expansion; future BtSweep owns bounded workers, cancellation, and aggregation.
- BtBot owns one child Bot replay.
- Sweep status describes the latest execution, not a terminal Sweep identity.
- Stopped and error Sweeps may run again.
- A rerun replays selected Bots from the beginning.
- No partial recovery or resume exists.

## Execution Flow

```text
API
  -> SweepManager
  -> process supervision
  -> standalone BtSweep
  -> bounded BtBot workers
```

Workers receive stored identities only.

They never receive Config objects, Controllers, replay iterators, Accounts, or
result writers across the process boundary.

## Implemented Template Admission

One Sweep template references one complete scalar Bot template.

`[sweep]` owns replay identity and the Bot template reference.

Ordered `sweep.periods` entries own labelled or explicit replay windows.

`[sweep.parameters]` may contain zero dimensions. Nested parameter tables own explicit nonempty value lists.

Paths must exist and be recognized by the referenced Config's exact BotSpec.

Ignored extra Config fields cannot become Sweep dimensions.

Executor arrays-of-tables use stable Config `id` selectors. For example,
`executors.grid.levels` selects the `grid` Executor.

`internal/btsweep` sorts parameter paths, preserves value and period order,
and generates one deterministic Cartesian product. Zero dimensions generate one
unchanged Bot Config per period.

Every combination starts from a fresh Bot map. Complete generated TOML passes
exact `botspec.Validate` and Executor replay-symbol validation before its SHA-256
is returned.

Unknown paths, wrong types, empty lists, missing templates, bad periods,
duplicate periods, Bot field arrays, and generated record IDs fail.

Range syntax and scalar parameter shorthand are unsupported.

The package writes no database and launches no process.

See [BtSweep package](../packages/btsweep.md).

## Template and Record Identity

`v1` templates are mutable operator baselines. They do not preserve schema
history.

Generated Sweep and Bot records are immutable.

Revising a template creates a new Sweep and new Bot IDs.

An unchanged rerun reuses existing IDs and replaces the current result target.

The target Bot ID is globally unique. Sweep ID is optional grouping provenance.

Generated record IDs never belong in Sweep or Bot templates.

## Rerun

First creation allocates immutable Sweep and Bot records.

Every unchanged rerun clears selected current results before work starts.

Rerun reuses the same IDs and replaces one current result per Bot.

A revised template creates a new Sweep and new Bot IDs.

No execution history, checkpoint, partial resume, or worker recovery exists.

A failed child remains visible.

Other submitted children continue.

The latest Sweep execution is error when any selected child fails.

## Calendar Periods

The reusable Calendar toolkit resolves labels such as:

```text
2021
2021-H1
2021-Q1
2021-M01
```

Sweep templates may mix labelled periods and explicit start/end ranges.

Calendar resolution does not inspect market data.

## Deferred Period Analysis

Future analytics may attach searchable regime tags and group results across
tagged periods.

Analytics must show distributions and exact durations, not one misleading
average.

The catalog, tag taxonomy, and analytics remain deferred.

## Future CLI

The documented target import command is:

```text
nuubot-cli create sweep -f <abc.toml>
```

The CLI and record creation remain unimplemented.

## Required Proof

- Invalid Sweep configuration never starts.
- Worker limits remain bounded.
- Failed BtBot results remain visible.
- Stop reaches every active Sweep worker.
- Aggregate completion matches stored current child results.
- A rerun after crash starts selected children from the beginning.
