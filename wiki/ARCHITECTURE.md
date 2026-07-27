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

Status: Implemented for standalone historical replay.

One configured Bot selects one exact compiled BotSpec and stores its exact
BotConfig TOML in the database.

Setup transforms one exact stored BotConfig into one immutable typed BotSpec and returns one shared Nuubot harness.

```text
Runner or BtBot
`-- Controller
    |-- Signaler
    |-- Risk
    `-- zero or one active BotCycle
        `-- coordinated Executors
```

Runner, BtBot, and BtSweep are standalone programs.

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
- Capital, Meta, and market-data validation.
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
[Controller](design/packages/controller.md),
[Signaler](design/packages/signaler.md), [Risk](design/packages/risk.md), and
[BotCycle](design/packages/botcycle.md).

## Implemented BtBot

```text
command
`-- BtBot
    |-- ReplayReader
    |-- TickClock
    |-- MarketData
    `-- Controller
        |-- Signaler
        |   `-- Macross or RSI
        |-- Risks
        `-- active BotCycle
            `-- Executors
                |-- ObserverExecutor
                |-- TradeExecutor
                |   `-- Account
                `-- GridExecutor
                    `-- Account
                        |-- Simulator
                        `-- Ledger
```

BtBot owns historical orchestration and exact replay proof.

ReplayReader validates Parquet values before returning BBO values.

TickClock invokes BtBot's registered Controller callback from replay
timestamps.

Setup transforms exact BotConfig TOML into one typed BotSpec.

Controller receives one shared Nuubot harness containing Setup infrastructure and BotSpec.

Controller constructs Signaler and Risks, then owns Signal, Risk, BotCycle,
capital, drawdown, and graceful shutdown decisions.

BotCycle coordinates Executors. Executor implementations own execution policy.

GridExecutor owns arithmetic Levels, repeated Trades, boundary exits, and coordinated flattening.

BalancedRisk remains a non-protective stub.

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

## Canonical BtBot Flow

```text
main
  open Server logger
  parse identities
  open Bot logger
  create BtBot
  initialize
  start
  loop
  stop
  log one result

BtBot init
  prepare shared Nuubot harness
  reset Bot status for fresh replay
  clear replay data
  retain replay and result inputs
  resolve replay range
  initialize ReplayReader
  create and initialize TickClock
  attach TickClock to Nuubot
  create and attach MarketData to Nuubot
  initialize Controller from Nuubot
  register Controller timer
  initialize replay stats
  log init completed

BtBot loop
  read one validated BBO
  publish BBO to MarketData
  update replay stats
  advance TickClock
  registered timer callback runs Controller
  stop at Reader exhaustion or Controller request
  verify replay completion
```

Detailed behavior remains in [BtBot](design/packages/btbot.md) and [Replay](design/concepts/replay.md).

## Implemented Sweep Template Validation

```text
Sweep template
  -> internal/btsweep
  -> referenced Bot template
  -> exact botspec.Validate
  -> ordered generated Bot Config values
```

`internal/btsweep` validates replay inputs, ordered date ranges, optional
explicit parameter dimensions, exact parameter paths, and generated Configs.

It sorts parameter paths, preserves list and date-range order, emits complete
TOML, and hashes the exact emitted bytes.

It creates no Sweep or Bot record, writes no database, and launches no process.

`cmd/nuubot-bt-sweep` remains an `Under Construction.` placeholder.

## Approved Process Boundaries

```text
direct
  Runner -> Controller
  BtBot -> Controller
  BtSweep -> bounded BtBot workers

optional Server
  API -> BotManager -> process supervision -> Runner
  API -> SweepManager -> process supervision -> BtSweep
```

Server is the master PocketBase-style application process.

It owns WebServer, thin API, BotManager, SweepManager, Datastore, bootstrap
checks, and process supervision.

API forwards requests.

Managers own domain validation and choose their data sources.

BotManager never constructs Controller.

SweepManager never expands permutations or imports BtBot.

Implemented `internal/btsweep` owns template validation and deterministic
expansion.

Future BtSweep composes that package and owns record loading, cancellation,
bounded workers, and aggregation.

BtBot owns one child Bot replay.

Server failure does not automatically stop healthy standalone execution.

Runner owns one process-local shared WebSocket for its Bot. Monitoring, safety
switches, process reconnection, and live Account claims remain TBD.

## Target Live Flow

Status: Runner transport ownership approved; transport implementation remains pending.

```text
validated Venue event
  -> standalone Runner local feed state
  -> Controller
  -> BotCycle
  -> Executors
```

Runner must obtain required live inputs without requiring Server.

Runner owns its shared Info and WebSocket endpoints.

Any component may request supported data through those shared Nuubot objects.

A future Server optimization cannot make standalone Runner depend on Server.

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

Current BtBot execution is synchronous.

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

Current BtBot reads Bot configuration from SQLite and market data from Parquet.

Earlier live persistence planning separates:

- ProcessStore for process and manager state.
- Candidate ControllerStore for Controller, BotCycle, and Executor records.
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

Runner, BtBot, and BtSweep must remain independently executable while
Server is stopped.

Sweep template validation and expansion are implemented without persistence.

Immutable Sweep and Bot creation, saved-Config writes, status writes, and
unchanged-rerun ID reuse remain TBD.

Nuubot owns domain transactions, conditional transitions, generations,
idempotency, and unique trading identities.

PocketBase owns physical database concurrency. Nuubot MUST NOT add a generic
write queue or database mutex.

Physical tables, keys, migrations, and transaction boundaries require later approval.

## Deployment

Windows BtBot execution is proven.

Ubuntu 24 is the intended VPS target.

The standard Go toolchain and pure-Go boundary preserve portable builds.

Linux runtime and deployment remain unproven.
