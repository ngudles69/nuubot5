# Executor

## Covers

- `internal/executor/executor.go`
- `internal/executor/observer.go`
- `internal/botcycle/botcycle.go`
- `internal/runtime/runtime.go`

## Intent

An Executor MUST own one execution policy inside one BotCycle.

## Ownership

```text
Runtime
`-- BotCycle
    `-- []Executor
```

BotCycle MUST create configured Executors through `executor.Create`.

The factory MUST call `OnInit` before returning.

Runtime MUST NOT know concrete Executor types.

## Required Contract

Every Executor MUST implement:

- `OnInit`.
- `OnStop`.
- `Status`.
- `ExitReason`.

Status MUST be the only terminal-state source.

## Optional Capabilities

Executors implement only events they consume.

Current capabilities are:

- `BBOHandler.OnBBO`.
- `BBOIngestHandler.IngestBBO`.

BotCycle MUST use type assertions and skip unsupported capabilities.

Concrete identity MUST NOT derive from a callback name.

## Admission

An Executor MAY reject BotCycle admission through `executor.ErrRejected`.

BotCycle MUST stop earlier initialized Executors.

Runtime MUST consume the triggering Signal and wait for the next Signal.

Unexpected initialization errors MUST remain fatal.

## Signal Access

Executor initialization receives the triggering package and passive Signaler.

Concrete Executors own custom fields, history length, guards, and missing-data policy.

Runtime owns only standard entry admission.

## ObserverExecutor

Observer MUST implement `BBOHandler.OnBBO`.

```text
first BBO
  record entry timestamp and price
  calculate stop-loss price

long
  stop when price <= entry * (1 - stopLossPct)

short
  stop when price >= entry * (1 + stopLossPct)
```

Observer MUST report separate ingestion and normal BBO counts once during stop.
