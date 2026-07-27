# MarketData

Status: Implemented for BtBot; Runner WebSocket publication awaits live transport.
Purpose: Own current market values, BBO ingestion, buffering, and optional subscriptions across BtBot and Runner.

## Permanent Decision

This contract must not be weakened, relocated, or replaced without explicit user approval.

MarketData is shared Nuubot infrastructure.

BtBot and Runner own MarketData and its lifecycle.

Nuubot carries the owned MarketData reference for shared access.

Consumers either subscribe for every requested update or read the latest buffered BBO through `nuubot.MarketData`.

MarketData owns BBO ingestion, latest-value buffers, subscription registration, and update notification.

Simulator and strategy components consume MarketData only when required.

Executor, Account, BotCycle, Controller, Simulator, and Venue do not form an ingestion chain.

## Why

The current ingestion chain mixes market transport, Simulator matching, Account routing, and Executor capability plumbing.

```text
Controller
  BotCycle
    Executor
      Account
        Venue
          Simulator
```

That chain makes every intermediate component carry behavior it does not own.

Live Venue implementations would require meaningless no-op ingestion methods.

Executors would implement ingestion even when their strategy does not require BBO updates.

A shared buffer removes duplicated `lastBBO` ownership and eliminates BBO values from initialization context.

Optional subscriptions preserve strategy choice without forcing every component to implement BBO handling.

## Ownership

```text
BtBot or Runner
  owns MarketData lifecycle
  attaches MarketData to Nuubot

MarketData
  owns latest BBO buffers
  owns subscription registry
  validates and publishes updates

ReplayReader or WebSocket
  publishes BBO updates

Subscribers
  consume only requested market streams
```

Nuubot carries one shared MarketData reference.

`BotCycleContext` does not carry BBO data.

Controller does not retain a duplicate latest-BBO map.

## Market Identity

One BBO buffer is identified by:

```text
Venue + Network + Symbol
```

Physical Account is excluded because BBO is market data, not Account data.

The key distinguishes Simulator, testnet, mainnet, and future Venue streams without execution-mode checks.

## API Intent

The exact implementation may remain small.

Required behavior is equivalent to:

```go
type Key struct {
    Venue   string
    Network string
    Symbol  string
}

IngestBBO(Key, BBO) error
LatestBBO(Key) (BBO, bool)
SubscribeBBO(Key, Callback) (Subscription, error)
Subscription.Stop() error
```

MarketData returns detached immutable BBO values.

Subscribers never receive pointers into mutable buffers.

## Ingestion

BtBot and Runner use the same ingestion operation.

```text
receive BBO
validate identity, timestamp, and price
replace latest matching buffer
notify matching subscribers
return callback failure to producer
```

The buffer updates before subscribers run.

Invalid BBO data changes no buffer and triggers no callback.

No subscribers still means the latest buffer updates.

There is one ingestion owner: MarketData.

## BtBot Source

```text
ReplayReader reads Parquet tick
BtBot publishes BBO to MarketData
MarketData updates buffer and notifies subscribers
BtBot advances TickClock
Controller timer may later create BotCycle
```

The latest BBO at Executor Start comes from Parquet replay.

Every replay BBO requiring subscriber delivery completes before the next replay tick.

BtBot still clears fresh attempt data and never restores interrupted runtime state.

## Runner Source

```text
WebSocket receives valid live BBO
Runner publishes BBO to MarketData
MarketData updates buffer and notifies subscribers
WallClock later triggers Controller
Controller may later create BotCycle
```

Runner Start waits for required first valid BBO data before opening timed Controller execution.

Runner Start does not create BotCycle.

Controller alone decides when a new or recovered BotCycle exists.

The latest BBO at Executor Start comes from the current WebSocket-backed buffer.

Runner restart creates empty MarketData and waits for new live market truth.

MarketData recovery never loads stale BBO as current truth.

## Executor Startup

Executor Init initializes configuration and owned objects without BBO data.

Executor Start requests current BBO data directly from Nuubot MarketData.

```text
Observer Start
  read latest BBO
  establish entry and stop-loss state

Trade Start
  read latest BBO
  establish current strategy price

Grid Start
  read latest BBO
  calculate Grid and starting Orders
```

Missing or stale required BBO data prevents that Executor from starting.

BBO is not passed through `BotCycleContext` or lifecycle parameters.

## Executor Subscription

BBO subscription is optional strategy behavior.

The base Executor interface contains no BBO method.

Executors without tick-level strategy requirements do not subscribe and implement no empty callback.

An Executor needing every BBO subscribes during Start and unsubscribes idempotently during Stop.

Examples:

```text
Observer
  subscribe for stop-loss checks

Grid
  subscribe for boundary exits

Trade
  read latest BBO when required
  subscribe only if a configured strategy later requires tick updates
```

Executors consume BBO data as strategy input. They never own ingestion or global buffering.

## Simulator Subscription

Simulator subscribes during its initialization for each required market key.

The subscription callback reads the latest BBO from `nuubot.MarketData` and performs Simulator matching.

```text
MarketData ingests BBO
  update latest buffer
  invoke Simulator callback

Simulator callback
  read latest buffered BBO
  process active Orders
  create Venue Orders and Fills
  mark changed state for reconciliation
```

Simulator does not receive BBO through Executor, Account, BotCycle, or Controller forwarding.

Simulator must process every subscribed BBO update. Delivery cannot silently coalesce ticks.

Simulator retains its subscription handle and unsubscribes idempotently during Stop.

Mainnet or testnet without Simulator has no Simulator subscription and performs no Simulator callback.

## Two Consumer Modes

MarketData supports two intentional consumer modes.

### Latest-value read

Use when a component needs only current market state.

```go
bbo, found := nuubot.MarketData.LatestBBO(key)
```

Examples include Executor Start, periodic strategy evaluation, and current monitoring.

### Subscription

Use when a component must process every update.

Examples include Simulator matching, trailing stops, and immediate BBO exit conditions.

Subscription is not added until one current requirement needs it.

## Ordering

For one admitted BBO:

```text
validate BBO
publish latest buffer
notify every matching subscriber
complete required subscriber delivery
return from IngestBBO
```

The buffer is readable before callbacks execute.

Simulator matching for one BBO completes before BtBot admits the next replay BBO.

Executor policy never consumes unpublished Simulator Fill truth.

Account reconciliation remains the only path from Venue truth into Ledger evidence.

## Callback Contract

A subscriber registers one exact market key and one callback.

Callbacks process only their subscribed stream.

Callback failure returns through MarketData ingestion to BtBot or Runner supervision.

Subscription Stop is idempotent.

MarketData removes stopped subscriptions before later notification.

The delivery implementation may be synchronous or queued only if every-update ordering and failure propagation remain proven.

Silent update loss, reordering, and coalescing are prohibited for every-update subscriptions.

## Time

BBO timestamp is market-event evidence.

```text
BBO.TimestampMS
  when the market observation occurred

Nuubot Clock.NowMS
  when lifecycle, Order, cancellation, or reconciliation work occurs
```

Consumers may use BBO timestamp for freshness validation and deterministic market evidence.

Consumers must not use BBO timestamp as a general replacement for current Clock time.

## Persistence and Recovery

The latest BBO buffer is process-local current market state.

BtBot rebuilds it from Parquet during every fresh replay.

Runner rebuilds it from new WebSocket data after process start or recovery.

Runner does not restore a persisted BBO as current live truth.

Durable Ledger, Trade, Order, Fill, and Venue evidence remains separate from MarketData.

## Same Code Across Execution Paths

```text
BtBot + Simulator
  Parquet publishes
  Simulator subscribes

Runner + Simulator
  WebSocket publishes
  Simulator subscribes

Runner + testnet
  WebSocket publishes
  only required strategy subscribers attach

Runner + mainnet
  WebSocket publishes
  only required strategy subscribers attach
```

No component checks `if live` or `if backtest` to consume BBO data.

Execution mode changes the producer and persistence preparation, not the MarketData contract.

## Prohibited Architecture

Do not restore the forwarding chain:

```text
Controller.IngestBBO
  BotCycle.IngestBBO
  Executor.IngestBBO
  Account.IngestBBO
  Venue.IngestBBO
```

Do not retain `BBOIngestHandler` on Executor.

Do not require every Executor to implement BBO methods.

Do not add live Venue no-op ingestion methods.

Do not carry BBO through `BotCycleContext`.

Do not retain duplicate Controller or Executor latest-BBO stores when MarketData owns the value.

Do not make Simulator matching depend on Controller timer cadence.

Do not allow asynchronous delivery to collapse replay ticks.

## Implemented Hardcut

The implementation removes:

- Controller-owned `latestBBOs`;
- `BotCycleContext.LatestBBO`;
- `BBOIngestHandler`;
- Observer, Trade, and Grid `IngestBBO` methods;
- BotCycle BBO ingestion forwarding;
- Account BBO ingestion forwarding; and
- Venue BBO ingestion requirements.

BtBot and Runner create one owned MarketData object and attach it to Nuubot.

Controller subscribes only for tick and BotCycle timing evidence.

Simulator subscribes during initialization.

Observer and Grid subscribe during Start for required strategy policy.

Trade reads the latest BBO only when startup, policy, or shutdown needs it.

## Completed Proof

- BtBot and Runner create and attach one MarketData object.
- Parquet source publishes through the shared ingestion operation.
- LatestBBO returns one detached latest value by exact market key.
- Invalid BBO changes no buffer and invokes no callback.
- No subscribers still updates the latest buffer.
- Simulator subscribes and receives every replay BBO.
- Simulator unsubscribes idempotently during Stop.
- Simulator matching completes before the next replay tick.
- Executor Start reads the latest BBO without Context carriage.
- Executors without BBO requirements create no subscription.
- Strategy subscriptions receive every required update.
- Callback failures reach BtBot supervision.
- BBO event time and Clock operation time remain distinct.

- Focused MarketData, Simulator, Account, Executor, BotCycle, Controller, BtBot, and Runner tests pass.
- Observer Bot 9 passes with Start-time latest-BBO entry: 63 cycles and 16 stop-loss exits.
- Trade Bot 13 preserves 193 Trades, 626 Orders, and 386 Fills.
- Grid Bot 15 preserves 1,982 Trades, 4,697 Orders, 2,636 Fills, and 585 round trips.

## Pending Runner Proof

- WebSocket publishes through the same MarketData ingestion operation.
- Callback failures reach Runner supervision.
- Runner simnet, testnet, and mainnet use the same MarketData implementation.

Runner WebSocket publication remains unproven because live transport is not implemented and Runner was not executed.