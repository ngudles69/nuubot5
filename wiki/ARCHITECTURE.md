# Nuubot5 Architecture

## Purpose

This page owns system layers, ownership, flows, concurrency, persistence, and deployment boundaries.

[DESIGN.md](DESIGN.md) owns the high-level object catalog.

[`design/**`](design/) owns detailed contracts. [`logic/**`](logic/) remains legacy detail.

## Rules

- Every mutable object has one direct owner.
- A parent controls only direct children.
- Values and narrow intent calls cross ownership boundaries.
- Dependencies point from composition toward domain and adapters.
- Nuubot4 process remains canonical unless the user approves a change.

## Implemented BtRunner

```text
command
`-- BtRunner
    |-- ReplayReader
    |-- TickClock
    `-- Runtime
        |-- Signaler
        |   `-- MacrossSignaler or RsiSignaler
        |-- Risks
        `-- active BotCycle
            `-- Executors
                `-- ObserverExecutor
```

BtRunner owns historical orchestration and exact replay proof.

ReplayReader validates Parquet values before returning BBO values.

TickClock converts replay timestamps into Runtime pass decisions.

Runtime owns signals, risk checks, BotCycle decisions, and graceful shutdown.

BotCycle coordinates Executors. Executor implementations own execution policy.

BalancedRisk is a stub. ObserverExecutor observes BBO values and records simulated exits.

## Canonical BtRunner Flow

```text
main
  parse identities
  load configuration
  create logger
  create BtRunner
  start
  run
  stop
  return one result

BtRunner setup
  load BotSpec
  resolve replay end
  create reader, clock, and Runtime
  load required bars
  prepare Signaler
  calculate expected proof

BtRunner run
  read one validated BBO
  send BBO to Runtime
  advance TickClock
  call Runtime Pass when due
  stop at the configured end
  verify exact replay
```

Detailed behavior remains in [BtRunner](design/packages/btrunner.md) and [Replay](design/concepts/replay.md).

## Approved Live Ownership

This architecture is approved but unimplemented.

```text
Server
|-- DataEngine
|-- ProcessStore
|-- BotManager
|   `-- Runner
|       |-- WallClock
|       |-- subscriptions and local feed state
|       `-- Runtime
|           |-- RuntimeStore
|           |-- Signaler
|           |-- Risks
|           `-- active BotCycle
|               `-- Executors
|                   `-- Accounts
|                       |-- Venue
|                       `-- Ledger
|                           `-- Trades
|                               `-- Orders
|                                   `-- Fills
|-- SweepManager
`-- HTTP application
    |-- API routes
    `-- web routes and assets
```

Server owns shared process resources and service lifecycles.

BotManager owns active Runner lifecycles. SweepManager owns Sweep coordination.

DataEngine owns shared external data acquisition, validation, connection reuse, and multiplexing.

Runner owns subscriptions, local feed state, WallClock, and one Runtime.

DataEngine MUST NOT call Runtime. Runner routes subscribed events through Runtime.

Runtime owns RuntimeStore. RuntimeStore persists Runtime and BotCycle state only.

Runtime owns decisions but does not own Accounts.

BotCycle owns Executors. Each Executor owns its Accounts.

Each Account owns one Venue and one Ledger.

Ledger owns Trades. Each Trade owns Orders. Each Order owns Fills.

Venue owns external or simulated execution truth. Ledger owns local evidence.

Simulator implements Venue behavior and stores venue-shaped truth only.

## Approved Live Flow

### BBO

```text
Venue feed
  -> DataEngine
  -> Runner subscription and local state
  -> Runtime
  -> BotCycle
  -> Executors
```

BBO events support responsive stop-loss and trailing-stop decisions.

Runtime evaluates decisions during its configured synchronous pass.

### User Events and Reconciliation

```text
user event
  -> DataEngine
  -> Runner
  -> Runtime
  -> BotCycle
  -> Executor
  -> Account marks its Ledger dirty

reconciliation cadence
  -> Runtime
  -> BotCycle
  -> Executor
  -> Account reconciles Venue into Ledger
  -> AccountSnapshot returns upward
```

Dirty state clears only after successful reconciliation.

Runtime receives snapshots. It MUST NOT reach into Account, Ledger, Trade, Order, or Fill state.

## Concurrency

Current BtRunner execution is synchronous.

Future DataEngine readers may use owned goroutines for external connections.

Every goroutine MUST have one owner, stop condition, context, and error path.

Runner serializes external events and clock events into Runtime calls.

Runtime policy remains synchronous.

This is not an HFT design. Bounded polling and clear ownership take priority.

## Data Boundaries

Parquet files, database rows, and venue messages are untrusted inputs.

Boundary packages validate shape, identity, timestamps, prices, quantities, and sequence before returning trusted Go values.

Runtime MUST NOT decode Parquet, query Sweep storage, or parse venue messages.

Signaler receives validated bars. Indicator code MUST NOT read files.

Venue normalizes external outcomes. Account reconciles them into Ledger evidence.

## Persistence Boundaries

Current BtRunner reads Bot configuration from SQLite and market data from Parquet.

Approved live persistence separates:

- ProcessStore for process and manager state.
- RuntimeStore for Runtime, BotCycle, and executor records.
- Account persistence for Ledger, Trade, Order, and Fill evidence.
- Simulator persistence for venue-shaped simulated state.

These are logical boundaries, not one database graph.

SQLite remains the backtesting datastore.

PostgreSQL is approved for future live, simulator, and paper operation.

Physical tables, keys, migrations, and transaction boundaries require later approval.

## Deployment

Windows BtRunner execution is proven.

Ubuntu 24 is the intended VPS target.

The standard Go toolchain and pure-Go boundary preserve portable builds.

Linux runtime and deployment remain unproven.
