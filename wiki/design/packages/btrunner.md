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
- terminal ResultPublisher invocation.

BtRunner owns no Signaler, Risk, BotCycle, Executor, Account, or Ledger
directly.

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
- replay counts, range, elapsed time, completion, and publication proof.

ResultPublisher writes the same terminal hierarchy to the per-Bot SQLite
database.

## Standalone Boundary

BtRunner runs with Server stopped.

It reads the exact saved BotConfig from Datastore.

TOML import files never select trading behavior after Bot creation.
