# Filesystem

Status: Approved — partially implemented.
Covers: `.gitignore`, `.audits/`, `workspace/config/config.toml`, `internal/toolkit/logging/logging.go`, and `stest.sh`
Purpose: Keep every mutable Nuubot file under one portable workspace root.

## Root Contract

`workspace/` is the only mutable filesystem root.

Source, binaries, tests, and wiki pages remain outside it. Controller code must not
write configuration, databases, logs, data, results, or temporary state outside
`workspace/`.

One operator-tool exception exists.

`parity-probe` may create permanent, tracked API evidence under
`wiki/design/hyperliquid/json`.

It validates path segments, refuses overwrite, and writes no credentials.

This exception does not apply to Server, Runner, BtBot, Controller, or Simulator.

`.audits/` is reserved for current tracked engineering review evidence.

The directory is currently empty except for `.gitkeep` after an authorized
stale-report purge. Removed reports remain recoverable from Git history.

Authorized development work may write current reports there. Runtime programs
must not.

`workspace/` is the future Docker mount. The application image remains
immutable. The exact container mount path is unresolved.

## Layout

```text
workspace/
|-- config/
|   |-- config.toml
|   `-- credentials.toml
|-- db/
|   |-- nuubot.db
|   `-- bots/
|       `-- bot_<bot_id>.db
|-- logs/
|-- perf/
|   `-- profiles/
`-- data/
```

## Directory Ownership

| Path | Contents | Git |
|---|---|---|
| `workspace/config/config.toml` | Shared non-secret application configuration. | Tracked |
| `workspace/config/credentials.toml` | Local credentials and secrets. | Ignored |
| `workspace/db/nuubot.db` | Central Config, Meta, commands, acknowledgements, process generations, status, and health. | Ignored |
| `workspace/db/bots/` | One isolated execution SQLite database per globally unique Bot ID. | Ignored |
| `workspace/logs/` | Controller, Server, Bot, and test-run logs. | Ignored |
| `workspace/perf/profiles/` | Opt-in command performance diagnostics. Never user or account profiles. | Ignored |
| `workspace/data/` | Market and other runtime data files. | Ignored |

## Configuration and Secrets

`config.toml` must begin with a prominent `NO SECRETS ALLOWED IN THIS FILE`
warning.

`credentials.toml` must remain ignored and untracked. Secret values must not
enter source, shared configuration, wiki pages, logs, tests, or prompts.

## Databases

The configured shared datastore is `workspace/db/nuubot.db`.

It holds Sweep definitions, Bot configuration, mainnet Meta, commands, acknowledgements, process generations, lifecycle status, and health.

Every globally unique Bot ID owns one execution database:

```text
workspace/db/bots/bot_<bot_id>.db
```

Each result database has one writer. It contains high-volume Trade, Order, Fill,
and detailed replay evidence for that Bot.

Workers must not write detailed results into one shared Sweep database. A
coordinator may serialize small catalog and terminal-summary updates.

Completed Backtest databases become read-only evidence. Live reopens its database across process recovery without clearing historical evidence.

Sweep workers keep detailed Ledger and Simulator state in memory while running.

Only successful runs export their per-Bot result database.

ResultPublisher builds `bot_<bot_id>.db.partial` beside the final path.

One successful committed export closes and atomically renames it to `.db`.

Only `.db` is completed evidence.

Failed runs retain no recoverable partial state and are rerun.

PocketBase-owned SQLite is approved for optional Server persistence.

Runner, BtBot, and BtSweep remain independently executable while Server
is stopped.

Standalone saved-Config reads, status writes, physical schemas, migrations,
and database access paths remain unresolved.

## Logs

Logs remain under `workspace/logs/`.

Current identity naming includes `server.log`, `bot_<sweep_id>_<bot_id>.log`,
and timestamped `stest` result logs.

## Performance Diagnostics

`workspace/perf/profiles/` owns disposable `stest.sh -pp` CPU, trace, memory, block, and mutex diagnostics.

These files describe command execution. They are not user, account, identity, or configuration profiles.

## Data

Market data and other runtime datasets belong under `workspace/data/`.

Exact source, symbol, timeframe, and retention subdirectories remain unresolved.

## Current External Data

Shared market data may resolve outside this repository through
`paths.shared_data`.

## Does Not

- Define PocketBase schemas or standalone access paths.
- Define database schemas or migrations.
- Move the current datastore into the target database layout.
- Define Docker image or container paths.
- Authorize secrets in tracked files.

## Required Proof

- Shared configuration is trackable and contains no secrets.
- Credentials remain ignored and untracked.
- Controller writes stay below `workspace/`.
- Each Sweep Bot writes only its own result database.
- Main datastore workers do not receive high-volume result writes.
