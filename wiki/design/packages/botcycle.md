# BotCycle Package

Status: Implemented.
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
  resolve exit reason
  calculate duration
  report proof

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
