# BtBot Package

Status: Implemented with typed BotSpec construction and first-class results.
Covers: `internal/btbot/btbot.go`
Purpose: Run one bounded historical Bot replay and prove exact completion.

## Ownership

BtBot owns:

- one shared Nuubot harness;
- one ReplayReader;
- one TickClock;
- one MarketData service;
- one Controller;
- exact replay proof; and
- ordered telemetry samples;
- terminal RunReport construction; and
- terminal ResultPublisher invocation.

BtBot owns no Signaler, Risk, BotCycle, Executor, Account, or Ledger
directly.

The command may profile the complete BtBot execution boundary when explicitly requested.

Profiling belongs to `cmd/nuubot-bt-bot`, not this domain package.

The command writes CPU, runtime trace, heap, allocations, block, and mutex artifacts.
BtBot receives no profiling configuration and contains no profiling lifecycle.

## Initialization

```text
prepare shared Nuubot harness
reset Bot status for fresh replay
clear replay data
retain replay and result inputs
resolve replay range
initialize ReplayReader
create TickClock
initialize TickClock
attach TickClock to Nuubot
create and attach MarketData to Nuubot
initialize Controller from Nuubot
register Controller timer
initialize replay stats
log init completed
```

Caller context reaches Setup and Meta loading.

Setup creates no background context.

## Loop

Every validated BBO receives its replay symbol.

BtBot publishes the BBO to every configured exact MarketData key, advances TickClock, and runs Controller through the registered timer.

MarketData completes synchronous Simulator and strategy callbacks before BtBot advances to the next replay tick.

Reader exhaustion is normal completion.

## Stop

Stop logs every request before checking stopped state.

The first request marks BtBot stopped, stops Reader, Controller, and MarketData, builds results, and logs results and stats.

Successful Stop logs final `btbot stopped.` immediately before returning.

A repeated request logs `btbot stopping - ignoring stop request` and returns
without repeating teardown.

This guard keeps Stop idempotent.

The entry and ignored-request logs expose duplicate, late, rogue, or unexpected
Stop calls.

They may also reveal lifecycle sequencing or timing defects without claiming
concurrent execution.

## Crash Model

BtBot does not recover interrupted runtime state.

BtBot does not load persisted Ledger, Trade, Order, Fill, Controller, BotCycle, or Executor state.

Every invocation resets loaded Bot status to `configured`, clears stale attempt data, and replays the complete requested historical range.

`clearData` always deletes the attempt database and SQLite sidecars before Controller initialization.

The previous completed result remains isolated until successful atomic replacement.

If the BtBot process crashes, that run fails and is discarded.

The next attempt reruns the backtest from the beginning.

Runner Startup recovery does not apply to BtBot.

See [Runner Startup](../startup.md) for the intentionally different live execution model.

## Proof

BtBot verifies exact tick count, control-pass count, first timestamp, and
last timestamp.

Failure prevents successful publication.

## Result

`btbot.Result` contains:

- immutable `controller.Result`; and
- replay counts, range, historical-data-loop elapsed time, completion, and publication proof; and
- one terminal `report.Run`.

`btbot_historical_data_loop_elapsed_ms` measures `BtBot.Loop()` from entry through replay verification.

It excludes `Init`, `Start`, `Stop`, result publication, and shutdown.

Test scripts measure `btbot_elapsed_ms` from fresh BtBot process launch through exit.

BtBot samples Controller telemetry after every successful control pass.

One terminal sample follows Controller shutdown.

BtBot samples Go memory before RunReport construction and result publication.

The memory field names explicitly contain `before_publication`.

ResultPublisher writes the same terminal hierarchy to the per-Bot SQLite
database.

## Standalone Boundary

BtBot runs with Server stopped.

It reads the exact saved BotConfig from Datastore.

TOML import files never select trading behavior after Bot creation.
