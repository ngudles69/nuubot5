# BtRunner Package

Status: Implemented with exact BotSpec admission and first-class results.
Covers: `internal/btrunner/btrunner.go`
Purpose: Run one bounded historical Bot replay and prove exact completion.

## Ownership

BtRunner owns:

- Setup admission;
- one ReplayReader;
- one TickClock;
- one Controller;
- exact replay proof; and
- ordered telemetry samples;
- terminal RunReport construction; and
- terminal ResultPublisher invocation.

BtRunner owns no Signaler, Risk, BotCycle, Executor, Account, or Ledger
directly.

The command may profile the complete BtRunner execution boundary when explicitly requested.

Profiling belongs to `cmd/nuubot-btrunner`, not this domain package.

The command writes CPU, runtime trace, heap, allocations, block, and mutex artifacts.
BtRunner receives no profiling configuration and contains no profiling lifecycle.

## Initialization

```text
Setup loads AppConfig, stored BotConfig, ReplayInput, and Meta
initialize TickClock and ReplayReader
build exact compiled BotSpec
initialize Controller from BotDefinition
calculate expected replay proof
```

Caller context reaches Setup and Meta admission.

Setup creates no background context.

## Loop

Every admitted BBO receives its replay symbol.

BtRunner sends the BBO to Controller, advances TickClock, and runs Controller
through the registered timer.

Reader exhaustion is normal completion.

## Proof

BtRunner verifies exact tick count, control-pass count, first timestamp, and
last timestamp.

Failure prevents successful publication.

## Result

`btrunner.Result` contains:

- immutable `controller.Result`; and
- replay counts, range, historical-data-loop elapsed time, completion, and publication proof; and
- one terminal `runreport.Run`.

`btrunner_historical_data_loop_elapsed_ms` measures `BtRunner.Loop()` from entry through replay verification.

It excludes `Init`, `Start`, `Stop`, result publication, and shutdown.

Test scripts measure `btrunner_elapsed_ms` from fresh BtRunner process launch through exit.

BtRunner samples Controller telemetry after every successful control pass.

One terminal sample follows Controller shutdown.

BtRunner samples Go memory before RunReport construction and result publication.

The memory field names explicitly contain `before_publication`.

ResultPublisher writes the same terminal hierarchy to the per-Bot SQLite
database.

## Standalone Boundary

BtRunner runs with Server stopped.

It reads the exact saved BotConfig from Datastore.

TOML import files never select trading behavior after Bot creation.
