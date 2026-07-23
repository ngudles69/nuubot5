# Executor

## Covers

- `internal/executor/executor.go`
- `internal/executor/observer.go`
- `internal/botcycle/botcycle.go`
- `internal/runtime/runtime.go`

## Intent

An Executor MUST run one execution policy inside one BotCycle.

## Ownership

```text
Runtime
`-- BotCycle
    `-- []Executor
```

BotCycle MUST create configured Executors through the approved
configuration-selected factory. The factory MUST return BotCycle's
consumer-owned Executor interface.

BotCycle MUST pass the same accepted Signal to each Executor. Runtime MUST NOT
know concrete Executor types.

The consumer-owned Executor interface separates BotCycle from the selected
concrete policy.

The interface MUST contain only BotCycle's current lifecycle and event
requirements.

It MUST NOT expand for testing or hypothetical policies.

The current source contract uses this exact control path:

```text
Runtime.MainLoop(now)
  BotCycle.MainLoop(now)
    Executor.MainLoop(now)
```

`MainLoop` is a pre-contract source name. Its separately approved migration
target is `Pass`; this page MUST retain `MainLoop` until the source changes.

## ObserverExecutor

ObserverExecutor MUST prove the execution control path without Account, Ledger,
Trade, Order, Fill, Simulator, or Venue.

```text
first BBO after Signal availability
  record entry timestamp and price
  calculate stop-loss price

long
  stop when price <= entry * (1 - stopLossPct)

short
  stop when price >= entry * (1 + stopLossPct)

Stop(reason)
  preserve an existing stop-loss reason
  otherwise use the parent reason
  record final timestamp and price
  become terminal
```

## Evidence

ObserverExecutor MUST report:

- Signal and availability timestamps;
- side and Signal price;
- configured stop-loss percentage;
- entry, stop, exit, and final prices;
- start, end, and duration;
- ticks and passes;
- terminal reason.

BotCycle MUST report its own work separately.
