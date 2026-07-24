# Filesystem

Status: Approved — partially implemented.
Covers: `.gitignore`, `workspace/config/config.toml`, and `internal/toolkit/logging/logging.go`
Purpose: Keep every mutable Nuubot file under one portable workspace root.

## Root Contract

`workspace/` is the only mutable filesystem root.

Source, binaries, tests, and wiki pages remain outside it. Runtime code must not
write configuration, databases, logs, data, results, or temporary state outside
`workspace/`.

`workspace/` is the future Docker mount. The application image remains
immutable. The exact container mount path is unresolved.

## Layout

```text
workspace/
|-- config/
|   |-- config.toml
|   `-- credentials.toml
|-- db/
|   |-- <main datastore files>
|   `-- sweeps/
|       `-- sweep_<sweep_id>/
|           `-- bot_<bot_id>.db
|-- logs/
`-- data/
```

## Directory Ownership

| Path | Contents | Git |
|---|---|---|
| `workspace/config/config.toml` | Shared non-secret application configuration. | Tracked |
| `workspace/config/credentials.toml` | Local credentials and secrets. | Ignored |
| `workspace/db/` | Main live and Sweep-catalog datastore files. | Ignored |
| `workspace/db/sweeps/` | Per-Bot Sweep result SQLite databases. | Ignored |
| `workspace/logs/` | Runtime, Server, Bot, and test-run logs. | Ignored |
| `workspace/data/` | Market and other runtime data files. | Ignored |

## Configuration and Secrets

`config.toml` must begin with a prominent `NO SECRETS ALLOWED IN THIS FILE`
warning.

`credentials.toml` must remain ignored and untracked. Secret values must not
enter source, shared configuration, wiki pages, logs, tests, or prompts.

## Databases

Main datastore files live directly under `workspace/db/`.

The main datastore may hold live data, Sweep definitions, Bot status, relative
result paths, and small terminal summaries.

Live operation may use one shared database because its expected write pressure
is bounded.

Each Sweep Bot writes detailed results to:

```text
workspace/db/sweeps/sweep_<sweep_id>/bot_<bot_id>.db
```

Each result database has one writer. It contains high-volume Trade, Order, Fill,
and detailed replay evidence for that Bot.

Workers must not write detailed results into one shared Sweep database. A
coordinator may serialize small catalog and terminal-summary updates.

Completed result databases become read-only evidence. Sweep aggregation reads
them after their owning Bots finish.

PocketBase remains an unresolved consideration. Its adoption may change the
main datastore engine, schema, filename, and access path.

No design may assume PocketBase until the user approves it.

## Logs

Logs remain under `workspace/logs/`.

Current identity naming includes `server.log`, `bot_<sweep_id>_<bot_id>.log`,
and timestamped `rtest` result logs.

## Data

Market data and other runtime datasets belong under `workspace/data/`.

Exact source, symbol, timeframe, and retention subdirectories remain unresolved.

## Current Drift

Setup loads `workspace/config/config.toml` and
`workspace/config/credentials.toml`.

Current BtRunner reads `workspace/datastore/nuubot5_sweeps.db`.

Current shared market data may resolve outside this repository.

Logging already writes under `workspace/logs/`.

These facts describe current implementation. They do not override the approved
target layout.

## Does Not

- Select PocketBase or another main datastore engine.
- Define database schemas or migrations.
- Move the current datastore into the target database layout.
- Define Docker image or container paths.
- Authorize secrets in tracked files.

## Required Proof

- Shared configuration is trackable and contains no secrets.
- Credentials remain ignored and untracked.
- Runtime writes stay below `workspace/`.
- Each Sweep Bot writes only its own result database.
- Main datastore workers do not receive high-volume result writes.
