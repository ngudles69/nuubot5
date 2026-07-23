# Executor Package

Status: Implemented.
Covers: `internal/executor/*.go`
Purpose: Create and run configured execution policies behind a stable factory.

## Canonical Source

- `D:/rust/nuubot4/src/executor.rs`

## Scope & Responsibilities

ExecutorFactory selects concrete Executors. BotCycle knows only the common
Executor contract.

- Every Executor receives the same accepted Signal and BBO stream.
- New policies add one concrete file and one factory case.

## Program Flow

```text
ExecutorFactory(kind, log, cycle, executor, signal) -> Executor

init
  concrete = select kind
  concrete.init(log, cycle, executor, signal)

start
  concrete.start()

run(bbo)
  concrete.ingest(bbo)

run(now)
  return concrete.run(now)

stop
  concrete.stop()

---

domain
  observer long  -> exit at 1% below start price
  observer short -> exit at 1% above start price
```

## Notes

- Current configured Executor is Observer.
- Observer proves control flow without Account, Ledger, Trade, Order, Fill, Simulator, or Venue.
