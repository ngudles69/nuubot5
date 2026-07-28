# Datastore Package

Status: Implemented.
Covers: `internal/datastore/*.go`
Purpose: Load one exact stored BotConfig, mutable Bot status, and validated ReplayInput.

## Canonical Sources

- Nuubot4 boundary: `D:/rust/nuubot4/src/datastore.rs`
- Nuubot4 store: `D:/rust/nuubot4/src/datastore/sweep.rs`
- Nuubot4 model: `D:/rust/nuubot4/src/datastore/models.rs`

## Scope

Datastore reads one Bot row by exact Sweep and Bot identity.

The implemented path remains read-only. Future writable datastore behavior is
approved only at the ownership level described below.

## Owner and Children

Setup calls Datastore.

Datastore opens one short-lived read-only SQLite connection.

## Responsibilities

- Open the configured SQLite database read-only and immutable.
- Query one exact Bot row.
- Read exact BotSpecID, BotConfig TOML, Config SHA-256, mutable status, and replay JSON.
- Verify Config TOML against the stored SHA-256.
- Validate stored Bot lifecycle status.
- Decode replay JSON.
- Parse replay dates and optional Bot dates.
- Preserve optional `StartAt` and `EndAt` in `ReplayInput`.
- Validate symbol, tick path, and ordered dates.
- Return one normalized `datastore.Bot`.

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

Output is one `datastore.Bot` containing immutable configuration, mutable status, and ReplayInput.

## State and Invariants

Symbol and tick path MUST be non-empty.

Replay start MUST precede replay end.

When both exist, Bot start MUST precede Bot end.

Optional Bot dates accept RFC3339 or `YYYY-MM-DD`.

Datastore parses and returns `StartAt`. Current BtBot intentionally ignores it.

## Concurrency

Each call owns its local connection.

## Persistence

`LoadBot` opens the shared SQLite database read-only and immutable.

## Approved Target

The canonical mutable layout is defined by
[`Filesystem`](../concepts/filesystem.md).

```text
workspace/db/
|-- nuubot.db
`-- bots/
    `-- bot_<bot_id>.db
```

Main datastore expectations:

- `paths.database` names the shared database.
- The shared database contains Sweep, Bot, Meta, command, acknowledgement, and process-supervision tables.
- Execution evidence never enters the central control database.
- Sweep definitions and Bot configuration remain centrally discoverable.
- Sweep and Bot status updates stay small.
- Result database paths are stored relative to `workspace/`.
- Small terminal summaries may return to the main datastore.
- High-volume Trade, Order, Fill, and replay rows do not enter it.

Per-Bot execution database expectations:

- Each globally unique `bot_id` owns one SQLite execution database.
- Each worker writes only its owned result database.
- Workers never share a result database writer.
- Workers retain detailed Sweep state in memory during execution.
- One successful run exports its result database after completion.
- Final export uses one transaction and an atomic temporary-file rename.
- Failed runs retain no recovery checkpoint and are rerun.
- Completed Backtest databases become read-only evidence.
- Live reopens its durable database during recovery without clearing evidence.
- Sweep aggregation reads completed Backtest databases after Bot termination.

TradeExecutor result tables are defined by
[Trading Schema](../concepts/trading-schema.md).

One coordinator may serialize shared Sweep-catalog updates. The design must not
rely on SQLite WAL to make high-volume shared writes safe.

PocketBase is approved for future Server persistence.

Server owns that PocketBase application and its writable SQLite database.

Runner, BtBot, and BtSweep must remain independently executable while
Server is stopped.

How standalone processes read saved Config and publish current status while
preserving PocketBase write ownership remains TBD.

The per-Bot Sweep result DDL remains independent from the future PocketBase migration.

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

- Remaining main schemas, migrations, and transaction boundaries.
- Sweep catalog and terminal-summary schema.
- Live datastore access and write serialization.
- Sweep aggregation contract.

## BotSpec Hardcut

Status: Implemented.

Replay fields use `datastore.ReplayInput`.

BotSpec means the complete compiled Controller design described by
[BotSpec](../concepts/bot-spec.md).

## BotConfig Persistence

One configured Bot stores:

```text
bot_spec_id
config_toml
config_hash
```

TOML is the authoritative stored BotConfig representation.

The existing replay JSON remains ReplayInput only.

It owns no Bot behavior.

Start copies the exact saved TOML and hash into one immutable BotGeneration.

The active generation never rereads the Bot row or original import file.

Later Bot-row changes cannot mutate an active generation.

Server-side BotManager owns Config validation for Server requests.

Standalone execution cannot require Server availability.

Controller never receives or writes a database handle.

## Approved Active Resource Claims

One active resource is identified by:

```text
venue
network
physical_account_id
symbol
```

Its active owner records Bot ID and BotGeneration ID.

The resource tuple is unique while one Bot is starting, running, or stopping.

Configured, stopped, and error Bots do not hold active claims.

Direction never changes exclusivity.

One BotGeneration cannot claim the same tuple more than once.

Every Executor uses a distinct Account-symbol resource.

Fresh Start additionally requires authoritative Venue proof of:

- Zero active Orders for the symbol.
- Zero position for the symbol.
- Sufficient available funds and margin for the complete Bot.

Existing Orders and positions are cleared manually by the user.

Fresh Start never adopts, cancels, or flattens them.

Live cross-process Account ownership, claim storage, and stale-process cleanup
remain TBD.

## Approved Sweep Execution

Sweep definitions are reusable and have no terminal lifecycle.

`configured`, `starting`, `running`, `stopping`, `stopped`, and `error`
describe only the current or latest execution.

Stopped and error Sweeps may run again.

Sweep execution has no recovery, checkpoint, or partial replay resume.

Every rerun starts selected child Bots from the beginning.

Selected current results are cleared when a new execution starts.

Child Bot identities remain stable.

Rerun replaces one current result per child.

Clone the Sweep first when both result sets must remain.
