# Live Runner

## Status

This is the approved live-flow contract. It is not implemented.

## Intent

Runner MUST run one live, paper, simulator, or testnet Bot until a stop
condition or user stop request closes it.

BtRunner MUST remain the separate bounded historical replay owner. Live Runner
MUST NOT copy Parquet replay logic.

## Ownership

```text
Server
`-- BotManager
    `-- Runner
        |-- WallClock
        |-- BBO feed
        |-- user-event feed
        `-- Runtime
```

Server, API, web server, BotManager, and SweepManager MUST live outside Runtime.
Runner MUST own its feeds and Runtime.

## Program Flow

```text
Runner.New(...)
  create Runtime
  create WallClock
  create required feeds

Runner.Init()
  authenticate venue clients
  register BBO and user-event inputs

Runner.Start()
  start Runtime
  start feeds
  start Clock

Runner.Loop()
  wait for feed events or Clock cadence

  on BBO
    update latest owned BBO state

  on user event
    mark Account and Ledger dirty

  on fast Clock pass
    Runtime checks BBO stop-loss and trailing-stop rules

  on reconciliation Clock pass
    if Account or Ledger is dirty
      reconcile venue truth
      clear dirty state only after success

  stop when Runtime, end date, context, or user requests stop

Runner.Stop()
  stop new admission
  stop feeds
  stop Clock
  stop Runtime
  close active BotCycle
```

## Concurrency

- WebSocket readers MUST run in Runner-owned goroutines.
- Each reader MUST accept `context.Context`, publish typed events, and terminate
  when Runner stops.
- Feed goroutines MUST NOT place orders, stop BotCycles, reconcile accounts, or
  mutate Runtime policy.
- Runtime MUST make decisions synchronously on its Clock pass.
- Account and Ledger dirty flags MUST be concurrency-safe.
- BBO stop checks MUST use the configured fast cadence. The default MUST be
  5 seconds.
- Dirty reconciliation MUST use the configured reconciliation cadence. The
  default MUST be 10 seconds.

The cadence values remain configuration policy. This is not an HFT system.

## Stop Contract

Bot start and end timestamps MUST be valid configuration.

A missing end timestamp means no time-based stop.

At the end timestamp, Runtime MUST use the same graceful stop path as a user
request.

An active BotCycle MUST close and record its final state.

## Dependencies

All dependencies MUST be pure Go. CGO MUST NOT be introduced without explicit
user approval recorded in `AGENTS.md`.
