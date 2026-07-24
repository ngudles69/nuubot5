# BotCycle Package

Status: Implemented with optional Account reconciliation and result collection.
Covers: `internal/botcycle/botcycle.go`
Purpose: Own Executors for one admitted entry Signal.

## Ownership

Runtime creates one BotCycle after a standard entry trigger.

BotCycle owns its initialized Executors.

BotCycle knows Executor interfaces, never concrete types.

## Admission

BotCycle asks the Executor factory to initialize every configured Executor.

Any Executor may return admission rejection.

BotCycle then stops previously initialized Executors and rejects the cycle.

Runtime consumes that Signal, keeps no cycle, and waits for the next Signal.

Unexpected Executor errors remain fatal.

Executors without gates initialize normally.

## Program Flow

```text
Init
  create executors
  initialize botcycle

Run
  check completion

Stop
  stop executors
  collect cached account results
  resolve exit reason
  calculate duration
  report proof

Reconcile
  reconcile capable Executor Accounts

OnRecon
  deliver accepted recon event

Result
  return immutable BotCycle result

IngestBBO
  ingest executor bbo

OnBBO
  record cycle time
  deliver executor bbo
```

## Event Dispatch

BotCycle uses capability assertions.

```go
if handler, ok := activeExecutor.(executor.BBOHandler); ok {
    handler.OnBBO(bbo)
}
```

Only running Executors receive events.

An Executor without a capability is silently skipped.

## BBO Order

```text
Runtime.IngestBBO
  BotCycle.IngestBBO
    each running BBOIngestHandler
  BotCycle.OnBBO
    each running BBOHandler
```

Simulator ingestion completes before `OnBBO`.

See [IngestBBO](../concepts/ingestbbo.md).

## Completion

BotCycle completes when every Executor is stopping, stopped, or in error.

Runtime closes a completed cycle during its next timed `Run`.

Runtime clears `r.cycle` before stopping the old cycle.

## Trading Extension

BotCycle remains the capability dispatcher.

It never receives or exposes Account references.

```text
Reconcile
  reconcile capable Executors
  collect Account snapshots
  return one completed barrier

OnRecon
  deliver accepted recon event
```

Runtime calls `Reconcile` before Risk.

Runtime calls `OnRecon` only after reconciliation and Risk succeed.

One Executor reconciliation failure stops the control pass.

BotCycle returns no partial success barrier.

BotCycle stops every Executor before collecting supported `AccountResultProvider` values.

Collected results are immutable values.

`Stop` returns one `botcycle.Result` after every Executor has stopped.

That value contains ordered `account.Result` values captured during Executor shutdown.

BotCycle returns no Account, Ledger, Simulator, or Executor pointer.

See [Trading State Tranche](../concepts/trading-state.md).
