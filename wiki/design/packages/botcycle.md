# BotCycle Package

Status: Implemented.
Covers: `internal/botcycle/botcycle.go`
Purpose: Own configured Executors for one accepted Signal.

## Canonical Source

- `D:/rust/nuubot4/src/botcycle.rs`

## Scope & Responsibilities

BotCycle creates Executors through ExecutorFactory and gives each the same
Signal, BBO stream, and control passes.

## Program Flow

```text
BotCycle(log, cycle, signal, executor_configs)

init
  executors = ExecutorFactory(executor_configs).init(log, cycle, signal)

start
  for executor in executors
    executor.start()

run(bbo)
  for active executor in executors
    executor.ingest(bbo)

run(now)
  for active executor in executors
    executor.run(now)

  return all executors completed

stop
  stop executors in reverse order
```

## Notes

- BotCycle knows the Executor interface, never concrete Executor types.
