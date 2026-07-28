# Backtest Package

Status: Implemented with typed BotSpec construction and first-class results.
Covers: `internal/backtest/backtest.go`
Purpose: Run one bounded historical Bot replay and prove exact completion.

## Ownership

Backtest `Run` owns:

- one shared Nuubot harness;
- one ReplayReader;
- one TickClock;
- one MarketData service;
- one Controller;
- exact replay proof; and
- ordered telemetry samples;
- terminal RunReport construction; and
- terminal ResultPublisher invocation.

Backtest `Run` owns no Signaler, Risk, BotCycle, Executor, Account, or Ledger
directly.

`backtest.Execute` may profile the complete Backtest execution boundary when explicitly requested.

Shared profiling mechanics belong to `internal/runharness`, not Backtest lifecycle code or the command parser.

The harness writes CPU, runtime trace, heap, allocations, block, and mutex artifacts.
Backtest `Run` receives no profiling configuration and contains no profiling lifecycle.

## Initialization

```text
prepare shared Nuubot harness
select Backtest runtime policy
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

Backtest `Run` publishes the BBO to every configured exact MarketData key, advances TickClock, and runs Controller through the registered timer.

The timer uses selected `nuubot.Runtime.ControllerIntervalMS`. Account reads selected Recon and Recon sweep cadence.

MarketData completes synchronous Simulator and strategy callbacks before Backtest `Run` advances to the next replay tick.

Reader exhaustion is normal completion.

## Stop

Stop logs every request before checking stopped state.

The first request marks Backtest `Run` stopped, stops Reader, Controller, and MarketData, builds results, and logs results and stats.

Successful Stop logs final `btbot stopped.` immediately before returning.

A repeated request logs `btbot stopping - ignoring stop request` and returns
without repeating teardown.

This guard keeps Stop idempotent.

The entry and ignored-request logs expose duplicate, late, rogue, or unexpected
Stop calls.

They may also reveal lifecycle sequencing or timing defects without claiming
concurrent execution.

## Crash Model

Backtest `Run` does not recover interrupted runtime state.

Backtest `Run` does not load persisted Ledger, Trade, Order, Fill, Controller, BotCycle, or Executor state.

Every invocation resets loaded Bot status to `configured`, clears stale attempt data, and replays the complete requested historical range.

`clearData` always deletes the attempt database and SQLite sidecars before Controller initialization.

The previous completed result remains isolated until successful atomic replacement.

If the Backtest process crashes, that Run fails and is discarded.

The next attempt reruns the backtest from the beginning.

Live Startup recovery does not apply to Backtest.

See [Live Startup](../startup.md) for the intentionally different live execution model.

## Proof

Backtest `Run` verifies exact tick count, control-pass count, first timestamp, and
last timestamp.

Failure prevents successful publication.

## Result

`backtest.Result` contains:

- immutable `controller.Result`; and
- replay counts, range, historical-data-loop elapsed time, completion, and publication proof; and
- one terminal `report.Run`.

`btbot_historical_data_loop_elapsed_ms` measures `backtest.Run.Loop()` from entry through replay verification.

It excludes `Init`, `Start`, `Stop`, result publication, and shutdown.

Test scripts measure `btbot_elapsed_ms` from fresh BtBot process launch through exit.

Backtest `Run` checks telemetry cadence after every successful control pass.

It collects only when selected Backtest telemetry cadence is due.

Backtest retains samples in memory and writes them once during terminal publication.

One terminal sample follows Controller shutdown.

Backtest `Run` samples Go memory before RunReport construction and result publication.

The memory field names explicitly contain `before_publication`.

ResultPublisher writes the same terminal hierarchy to the per-Bot SQLite
database.

## Standalone Boundary

Backtest runs with Server stopped.

It reads the exact saved BotConfig from Datastore.

TOML import files never select trading behavior after Bot creation.
