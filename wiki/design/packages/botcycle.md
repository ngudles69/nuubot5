# BotCycle Package

Status: Implemented.
Covers: `internal/botcycle/botcycle.go`
Purpose: Own configured Executors for one accepted Signal.

## Canonical Source

- `D:/rust/nuubot4/src/botcycle.rs`

## Scope & Responsibilities

BotCycle creates Executors through ExecutorFactory and gives each the same
Signal, BBO stream, and timed Runs.

BotCycle routes Simulator-only BBO ingestion through every active Executor.

## Program Flow

```text
init
  create executors
  initialize botcycle

start
  start executors
  start botcycle

run
  run executors
  check completion

stop
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

## IngestBBO

[`IngestBBO`](../concepts/ingestbbo.md) follows direct ownership.

```text
BotCycle.IngestBBO
  call each active Executor.IngestBBO
```

This route completes before BotCycle delivers the same BBO through `OnBBO`.
BotCycle does not select or access Venue implementations.

## Notes

- BotCycle knows the Executor interface, never concrete Executor types.
- `OnBBO` remains a domain helper because it accepts one market event.
