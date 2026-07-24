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
create
  select implementation
  validate config
  create observer

start
  start observer

run
  record run

stop
  preserve stop reason
  preserve end time
  stop observer
  calculate duration
  report proof

---

domain
  OnBBO records entry and evaluates stop loss
  Terminal reports completion
  ExitReason reports the terminal reason
  observer long  -> exit at 1% below start price
  observer short -> exit at 1% above start price
```

## Notes

- Current configured Executor is Observer.
- `OnBBO`, `Terminal`, and `ExitReason` remain domain helpers.
- Observer proves control flow without Account, Ledger, Trade, Order, Fill, Simulator, or Venue.
