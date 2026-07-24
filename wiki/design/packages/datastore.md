# Datastore Package

Status: Implemented.
Covers: `internal/datastore/*.go`
Purpose: Load one validated Bot replay specification and define datastore ownership expectations.

## Canonical Sources

- Nuubot4 boundary: `D:/rust/nuubot4/src/datastore.rs`
- Nuubot4 store: `D:/rust/nuubot4/src/datastore/sweep.rs`
- Nuubot4 model: `D:/rust/nuubot4/src/datastore/models.rs`

## Scope

Datastore reads one Bot JSON configuration by exact Sweep and Bot identity.

The implemented path remains read-only. Future writable datastore behavior is
approved only at the ownership level described below.

## Owner and Children

Setup calls Datastore.

Datastore opens one short-lived read-only SQLite connection.

## Responsibilities

- Open the configured SQLite database read-only and immutable.
- Query one exact Bot row.
- Decode stored JSON.
- Parse replay dates and optional Bot dates.
- Preserve optional `StartAt` and `EndAt` in `BotSpec`.
- Validate symbol, tick path, and ordered dates.
- Return one normalized `BotSpec`.

## Does Not

- Modify Sweep data.
- Retain a database connection.
- Resolve shared-data containment.
- Load market rows.
- Persist replay results.
- Select or implement the future main datastore.

## Lifecycle

`LoadBot` opens, queries, validates, closes, and returns.

## Inputs and Outputs

Inputs are database path, Sweep ID, and Bot ID.

Output is one `datastore.BotSpec`.

## State and Invariants

Symbol and tick path MUST be non-empty.

Replay start MUST precede replay end.

When both exist, Bot start MUST precede Bot end.

Optional Bot dates accept RFC3339 or `YYYY-MM-DD`.

Datastore parses and returns `StartAt`. Current BtRunner intentionally ignores it.

## Concurrency

Each call owns its local connection.

## Persistence

The SQLite database is read-only and immutable for backtesting.

## Approved Target

The canonical mutable layout is defined by
[`Filesystem`](../concepts/filesystem.md).

```text
workspace/db/
|-- <main datastore files>
`-- sweeps/
    `-- sweep_<sweep_id>/
        `-- bot_<bot_id>.db
```

Main datastore expectations:

- Live tables may share one main datastore.
- Sweep definitions and Bot configuration remain centrally discoverable.
- Sweep and Bot status updates stay small.
- Result database paths are stored relative to `workspace/`.
- Small terminal summaries may return to the main datastore.
- High-volume Trade, Order, Fill, and replay rows do not enter it.

Per-Bot Sweep result expectations:

- Each `(sweep_id, bot_id)` owns one SQLite result database.
- Each worker writes only its owned result database.
- Workers never share a result database writer.
- Completed result databases become read-only evidence.
- Sweep aggregation reads completed databases after Bot termination.

One coordinator may serialize shared Sweep-catalog updates. The design must not
rely on SQLite WAL to make high-volume shared writes safe.

PocketBase remains an unresolved consideration. Adopting it may change the main
datastore engine, schema, filename, migrations, and access path.

Datastore must not assume PocketBase until the user approves that decision.

## Errors

Open, query, JSON, date, and validation failures return errors.

## Program Flow

```text
LoadBot
  open database
  query bot
  decode bot
  parse dates
  validate bot
  return bot
```

## Required Proof

- Known Sweep and Bot return expected values.
- Missing identity fails.
- Invalid JSON, dates, fields, and ordering fail.
- The database remains unchanged.

## Open Decisions

- PocketBase adoption.
- Main datastore engine and filename.
- Main schemas, migrations, and transaction boundaries.
- Sweep catalog and terminal-summary schema.
- Live datastore access and write serialization.
- Per-Bot result schema and aggregation contract.
