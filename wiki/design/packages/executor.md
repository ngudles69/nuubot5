# Executor Package

Status: Implemented for current BBO paths.
Covers: `internal/executor/*.go`
Purpose: Create initialized execution policies with explicit event capabilities.

## Ownership

`executor.Create` owns concrete selection and initialization.

BotCycle owns returned Executors.

Concrete files own policy, configuration, status, Accounts, and timers.

Concrete identity never depends on implemented callbacks.

A Grid Executor remains a Grid Executor when it implements `OnBBO`.

## Required Lifecycle

```go
type Executor interface {
    OnInit(Context) error
    OnStop(reason string) error
    Status() Status
    ExitReason() string
}
```

Every Executor implements this minimum contract.

`Create` calls `OnInit` before returning.

BotCycle never initializes the returned Executor again.

Status is the only terminal-state source.

Stopped and error statuses never return to another state.

## Context

Executor initialization receives:

- Logger.
- Cycle and Executor numbers.
- Passive Signaler access.
- Triggering Signal package.
- Concrete Executor configuration.

Executors may query additional Signal history.

The triggering package remains available without another query.

## Optional Capabilities

```go
type BBOHandler interface {
    OnBBO(BBO)
}

type BBOIngestHandler interface {
    IngestBBO(BBO) error
}
```

`BBOHandler` names the capability.

`OnBBO` names its callback.

BotCycle dispatches only supported capabilities.

Unsupported events need no no-op methods.

Recon and user-event handlers remain approved but unimplemented.

Their event types and active paths do not exist yet.

## Admission

`ErrRejected` identifies expected Executor admission rejection.

An Executor may reject missing mandatory Signal fields or other entry requirements.

Executors without gates initialize immediately.

Rejection is not an operational failure.

BotCycle converts it into BotCycle admission rejection.

## Program Flow

```text
Create
  select executor
  initialize executor

Observer.OnInit
  validate config
  admit signal
  initialize observer

Observer.OnStop
  preserve stop reason
  preserve end time
  stop observer
  calculate duration
  report proof

Observer.IngestBBO
  count ingested bbo

Observer.OnBBO
  count received bbo
  record last bbo
  record entry
  assess stop loss
```

## Signal Use

Runtime uses only standard entry triggers.

Executor code and configuration own custom Signal interpretation.

Each Executor owns its history length, timestamp guard, and missing-data policy.

No common `OnSignal` callback exists.

## Timers

Concrete Executor timer methods will register directly with Clock.

No shared timer-handler interface is approved.

No current Executor requires an owned timer.

## BBO Paths

[`IngestBBO`](../concepts/ingestbbo.md) and `OnBBO` remain separate.

`IngestBBO` advances Simulator matching through owned Accounts.

`OnBBO` runs concrete Executor market-data policy.

Neither path performs the other's work.

## Templates

[ObserverExecutor](../concepts/observer-executor.md) is the current starting template.

TradeExecutor may become the richer template after trading ownership exists.
