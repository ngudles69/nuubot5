# BotCycle Package

Status: Implemented.
Covers: `internal/botcycle/botcycle.go`
Purpose: Own configured Executors for one accepted Signal.

## Canonical Source

- `D:/rust/nuubot4/src/botcycle.rs`

## Scope & Responsibilities

BotCycle creates Executors through ExecutorFactory and gives each the same
Signal, BBO stream, and timed Runs.

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

OnBBO
  record cycle time
  ingest executor bbo
```

## Notes

- BotCycle knows the Executor interface, never concrete Executor types.
- `OnBBO` remains a domain helper because it accepts one market event.
