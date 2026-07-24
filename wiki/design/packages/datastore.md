# Datastore Package

Status: Implemented.
Covers: `internal/datastore/*.go`
Purpose: Load one validated Bot replay specification from the read-only Sweep database.

## Canonical Sources

- Nuubot4 boundary: `D:/rust/nuubot4/src/datastore.rs`
- Nuubot4 store: `D:/rust/nuubot4/src/datastore/sweep.rs`
- Nuubot4 model: `D:/rust/nuubot4/src/datastore/models.rs`

## Scope

Datastore reads one Bot JSON configuration by exact Sweep and Bot identity.

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
- Implement PostgreSQL live storage.

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

Result persistence and PostgreSQL live stores require separate designs.
