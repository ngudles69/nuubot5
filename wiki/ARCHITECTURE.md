# Nuubot5 Architecture

## Scope

This page owns system layers, ownership, flows, concurrency, persistence, and deployment boundaries.

[DESIGN.md](DESIGN.md) owns object contracts. `wiki/logic/**` owns detailed algorithms.

## Current Layers

```text
Command
  cmd/nuubot-btrunner

Orchestration
  btrunner -> setup -> replay.Reader + TickClock + Runtime

Runtime domain
  Runtime -> Signaler + []Risk + active BotCycle
  BotCycle -> []Executor

Data and adapters
  config + datastore + bars + replay

Shared values and mechanics
  market.BBO + common.Logger + common helpers
```

Dependencies MUST point downward.

Domain packages MUST NOT import commands, servers, or concrete future infrastructure.

## Current Ownership

```text
main
`-- BtRunner
    |-- replay.Reader
    |-- TickClock
    `-- Runtime
        |-- Signaler
        |   `-- calculator
        |-- []Risk
        `-- active BotCycle
            `-- []Executor
```

Each mutable object has one direct owner.

A parent creates, starts, calls, and stops only its direct children.

Values cross boundaries without exposing child mutable state.

## Current BtRunner Flow

```text
main
  parse sweep and bot identities
  load configuration
  create process logger
  create BtRunner
  start
  run
  stop
  return one terminal error

BtRunner.New
  load BotSpec
  resolve the effective end date
  create TickClock
  create replay.Reader
  create Runtime
  load required bars
  prepare Signaler
  calculate expected proof

BtRunner.Run
  read one validated BBO
  pass BBO to Runtime
  advance TickClock
  run one Runtime pass when due
  stop Runtime at the end date
  verify exact replay
```

[BtRunner Logic](logic/btrunner.md) owns replay and completion details.

## Runtime Flow

Runtime receives validated BBO values and releases available signals.

Runtime owns the decision to open or close one active BotCycle.

Runtime runs Risk before BotCycle work during each timed pass.

Runtime owns every graceful stop reason and closes the active BotCycle before its remaining children.

Signaler owns indicator calculation and ordered signal release.

BotCycle coordinates configured Executors. Executor implementations own execution policy.

[Signaler Logic](logic/signaler.md), [Executor Logic](logic/executor.md), and [Risk Logic](logic/risk.md) own detailed behavior.

## Approved Live Architecture

The live architecture is approved design. It is not implemented.

```text
server composition
  |-- process and datastore resources
  |-- BotManager
  |   `-- Runner per active Bot
  |       |-- WallClock
  |       |-- BBO feed
  |       |-- user-event feed
  |       `-- Runtime
  |-- SweepManager
  `-- HTTP application
      |-- API routes
      `-- web routes and assets
```

The server composition function owns shared resources, managers, clock services, and HTTP application assembly.

API and web packages are thin route modules. They are not required to become classes or stateful service objects.

BotManager owns active Runner lifecycles.

SweepManager owns Sweep orchestration. It MUST NOT manage Runtime internals.

Runner owns one live Bot input path, its feeds, its clock, and one Runtime.

Runtime remains synchronous at decision boundaries.

[Live Runner Logic](logic/runner.md) owns the approved event and cadence contract.

Current Nuubot3 live wiring is incomplete. Nuubot5 MUST NOT claim copied or proven live behavior.

## Approved Account Boundary

Account integration is approved design but incomplete in the reference and absent from Nuubot5.

```text
Runtime authoritative active Account list
`-- Account
    |-- Ledger
    |   `-- Trades
    |       `-- Orders
    |           `-- Fills
    `-- Exchange adapter
        |-- live or testnet adapter
        `-- Simulator
```

Executors create required Accounts and register the same objects with Runtime.

Runtime owns the authoritative active Account list for reconciliation, BBO delivery, close, and stop.

Account owns one Ledger and one selected Exchange adapter.

Ledger owns local Trade, Order, and Fill evidence.

Exchange or Simulator owns external execution truth.

Simulator stores exchange-shaped state. It MUST NOT store domain Order or Fill objects.

Whether the project names this protocol `Venue` or `Exchange` remains unresolved.

No Nuubot5 implementation may silently settle that naming decision.

## Concurrency

Current BtRunner execution is synchronous.

Live WebSocket readers will run in Runner-owned goroutines.

Each goroutine MUST have one owner, stop condition, context, and error path.

BBO and user-event readers publish typed events or update their explicitly owned feed state.

Feed goroutines MUST NOT place orders, reconcile Accounts, stop BotCycles, or execute Runtime policy.

BBO updates drive responsive stop checks on the configured fast Runtime cadence.

User events mark Account and Ledger dirty.

The configured reconciliation pass reads authoritative Exchange or Simulator truth and clears dirty state only after success.

This is not an HFT architecture. Clear ownership and bounded polling take priority over unnecessary concurrency.

## Data Boundaries

Parquet files and database rows are untrusted inputs.

Boundary packages validate shape, range, identity, timestamps, and sequence before returning trusted Go values.

BtRunner receives normalized BBO values. Runtime MUST NOT decode Parquet or query Sweep storage.

Signaler receives validated bars. Indicator code MUST NOT read files.

External exchange responses will enter through the Account-facing adapter boundary.

## Persistence

Current backtesting reads Bot configuration from SQLite.

Current market data comes from read-only Parquet files.

SQLite remains the backtesting datastore.

PostgreSQL is the approved future datastore for live, simulator, and paper operation.

Result persistence remains an unimplemented boundary.

No result store object, schema, or publication contract is approved.

## Logging

Current source uses `common.Logger` and formatted messages.

That implementation predates the coding contract and remains documented drift.

The approved target is standard `log/slog`, configured once at each executable or server boundary.

Logging migration requires a separately confirmed source change.

## Deployment

Current execution proof is Windows-only.

Ubuntu 24 is the intended VPS target.

Pure-Go dependencies and the standard Go toolchain preserve the intended portable build boundary.

Linux runtime and deployment remain unproven.
