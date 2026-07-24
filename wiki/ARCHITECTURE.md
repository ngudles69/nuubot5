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

## Approved Target Bot Architecture

Status: Approved design. Not implemented.

One configured Bot selects one exact compiled BotSpec and stores its exact
BotConfig TOML in the database.

Start creates one immutable BotGeneration and admitted BotDefinition.

```text
Runner or BtRunner
`-- Controller
    |-- Signaler
    |-- Risk
    `-- zero or one active BotCycle
        `-- coordinated Executors
```

Runner, BtRunner, and SweepRunner are standalone programs.

Server may launch and supervise them, but execution never requires Server.

Live cross-process Account ownership, claims, persistence, and WebSocket
sharing remain TBD.

Controller owns Signaler and Risk for its complete generation.

They remain active across BotCycles and flat intervals.

Signaler reports strategy state like a traffic light.

Risk reports gates and exits like a second signal source.

Neither component performs lifecycle or trading mutations.

Controller alone arbitrates:

- Strategy entry and exit signals.
- Risk gates and exits.
- Whether one BotCycle is active.
- Account-symbol availability.
- Capital, Meta, and market-data admission.
- BotCycle and Controller Stop.

One BotCycle is one exchange-style Bot campaign.

Every Executor in that cycle starts as one unit and receives the same strategy Signal.

Executors may monitor or place Orders independently according to their fixed BotSpec roles.

Every Executor uses one distinct Account-symbol resource.

The same Account-symbol cannot appear twice inside one Bot.

An explicit exit condition starts BotCycle Stop.

Flatness never triggers exit.

BotCycle completes only after Stop and authoritative proof of zero active Orders
and zero positions for every used Account-symbol.

See [BotSpec](design/concepts/bot-spec.md),
[Runtime and approved Controller hardcut](design/packages/runtime.md),
[Signaler](design/packages/signaler.md), [Risk](design/packages/risk.md), and
[BotCycle](design/packages/botcycle.md).

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

TickClock invokes BtRunner's registered Runtime callback from replay timestamps.

Runtime owns signals, risk checks, BotCycle decisions, and graceful shutdown.

BotCycle coordinates Executors. Executor implementations own execution policy.

BalancedRisk is a stub. ObserverExecutor observes BBO values and records simulated exits.

## Proposed BtRunner Trading Slice

This architecture is designed but unimplemented.

```text
BtRunner
`-- Runtime
    `-- active BotCycle
        `-- TradeExecutor
            `-- Account
                |-- Simulator
                `-- Ledger
                    `-- Trades
                        `-- Orders
                            `-- Fills
```

Simulator produces Hyperliquid-shaped Venue truth.

Account validates and translates that truth.

Ledger reconciles it into domain evidence.

Immediate HTTP responses and WebSocket events are non-authoritative hints.

Forced reconciliation restores Venue-authoritative Order, Fill, position, and balance truth.

See [Trading State Tranche](design/concepts/trading-state.md).

## Permanent Parity Probe

```text
parity-probe command
`-- Parity Probe
    `-- Info probe
        |-- Hyperliquid testnet
        `-- Simulator simnet
```

The command runs one selected operation through production code.

Testnet clearinghouse-state is implemented.

Simnet activates after Simulator implements real clearinghouse behavior.

See [Hyperliquid Parity Probe](design/hyperliquid/parity.md).

## Canonical BtRunner Flow

```text
main
  open Server logger
  parse identities
  open Bot logger
  create BtRunner
  initialize
  start
  loop
  stop
  log one result

BtRunner init
  initialize Setup
  load BotSpec
  resolve replay end
  create and initialize TickClock
  register Runtime timer
  initialize Reader
  create and initialize Runtime
  calculate expected proof

BtRunner loop
  read one validated BBO
  send BBO to Runtime
  advance TickClock
  registered timer callback runs Runtime
  stop at Reader exhaustion or Runtime request
  verify exact replay
```

Detailed behavior remains in [BtRunner](design/packages/btrunner.md) and [Replay](design/concepts/replay.md).

## Approved Process Boundaries

```text
direct
  Runner -> Controller
  BtRunner -> Controller
  SweepRunner -> bounded BtRunner workers

optional Server
  API -> BotManager -> process supervision -> Runner
  API -> SweepManager -> process supervision -> SweepRunner
```

Server is the master PocketBase-style application process.

It owns WebServer, thin API, BotManager, SweepManager, Datastore, bootstrap
checks, and process supervision.

API forwards requests.

Managers own domain validation and choose their data sources.

BotManager never constructs Controller.

SweepManager never expands permutations or imports BtRunner.

SweepRunner owns expansion, cancellation, bounded workers, and aggregation.

BtRunner owns one child Bot replay.

Server failure does not automatically stop healthy standalone execution.

Shared exchange WebSockets, monitoring, safety switches, process reconnection,
and live Account claims remain TBD.

## Target Live Flow

Status: Transport ownership remains TBD.

```text
validated Venue event
  -> standalone Runner local feed state
  -> Controller
  -> BotCycle
  -> Executors
```

Runner must obtain required live inputs without requiring Server.

A future Server-owned shared feed may serve supervised Runners only if direct
Runner execution remains complete.

```text
reconciliation cadence
  -> Controller
  -> BotCycle
  -> Executor
  -> Account reconciles Venue into Ledger
  -> immutable AccountSnapshot returns upward
```

Dirty state clears only after successful reconciliation.

Controller receives values. It never reaches into Account, Ledger, Trade,
Order, or Fill state.

## Concurrency

Current BtRunner execution is synchronous.

Future live transport readers may use owned goroutines for external
connections.

Every goroutine MUST have one owner, stop condition, context, and error path.

Runner serializes external events and clock events into Controller calls.

Controller policy remains synchronous.

This is not an HFT design. Bounded polling and clear ownership take priority.

## Data Boundaries

Parquet files, database rows, and venue messages are untrusted inputs.

Boundary packages validate shape, identity, timestamps, prices, quantities, and sequence before returning trusted Go values.

Controller MUST NOT decode Parquet, query Sweep storage, or parse venue messages.

Each concrete Signaler loads validated OHLCV through the `ohlcv` package.

Venue normalizes external outcomes. Account reconciles them into Ledger evidence.

## Persistence Boundaries

Current BtRunner reads Bot configuration from SQLite and market data from Parquet.

Earlier live persistence planning separates:

- ProcessStore for process and manager state.
- RuntimeStore for Controller, BotCycle, and Executor records.
- Account persistence for Ledger, Trade, Order, and Fill evidence.
- Simulator persistence for venue-shaped simulated state.

These are logical boundaries, not one database graph.

Account passes one configured `persist_mode` to Ledger and Simulator.

`none` keeps both in memory until successful result publication.

`max` persists every accepted Ledger mutation and Simulator state change.

Neither child detects Runner, Sweep, paper, or live mode.

`none` opens no database during Account execution.

ResultPublisher atomically creates its final database only after success.

Before teardown, Account evidence moves upward as immutable owned values.

Controller retains these values, not descendant pointers.

ResultPublisher writes `none` evidence only after successful Controller shutdown.

The existing SQLite Sweep database remains the read-only backtesting datastore.

One Server-owned PocketBase application owns the Server writable SQLite
database.

PocketBase queues writes through one write connection. SQLite WAL permits
concurrent reads while a write transaction runs.

Runner, BtRunner, and SweepRunner must remain independently executable while
Server is stopped.

Their exact saved-Config reads and status writes remain TBD.

Nuubot owns domain transactions, conditional transitions, generations,
idempotency, and unique trading identities.

PocketBase owns physical database concurrency. Nuubot MUST NOT add a generic
write queue or database mutex.

Physical tables, keys, migrations, and transaction boundaries require later approval.

## Deployment

Windows BtRunner execution is proven.

Ubuntu 24 is the intended VPS target.

The standard Go toolchain and pure-Go boundary preserve portable builds.

Linux runtime and deployment remain unproven.
