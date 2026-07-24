# Executor Package

Status: Implemented.
Covers: `internal/executor/*.go`
Purpose: Create and run configured execution policies behind a stable factory.

## Canonical Source

- `D:/rust/nuubot4/src/executor.rs`

## Scope & Responsibilities

`executor.go` owns the stable Executor contract and configured factory.
Concrete Executor files own implementation state and policy.

- Every Executor receives the same accepted Signal and BBO stream.
- New policies add one concrete file and one factory case.
- `Create` selects, constructs, and initializes the configured Executor.
- BotCycle starts, runs, and stops the returned configured Executor.
- Every concrete Executor follows the common lifecycle and method contract.

## Factory

```text
Create
  select implementation
  construct implementation
  initialize implementation
  return configured Executor
```

Concrete constructors do not form a second public factory.

## Concrete Executor Structure

```text
concrete Executor
  configuration and identity
  lifecycle status
  policy state
  owned Accounts when required

Start
  start owned children
  enter running state

IngestBBO
  forward BBO through owned Accounts when present
  record observation when the Executor owns no Account

OnBBO
  consume BBO for Executor policy

Run
  reconcile-dependent Executor decisions

Stop
  stop owned children
  enter terminal state
```

[`IngestBBO`](../concepts/ingestbbo.md) and `OnBBO` are separate paths.

```text
Executor.IngestBBO
  call each owned Account.IngestBBO

Executor.OnBBO
  run Executor BBO policy
```

`IngestBBO` never runs Executor policy. `OnBBO` never advances Simulator,
matches Orders, creates simulated Fills, or reconciles Accounts.

## Templates

[`ObserverExecutor`](../concepts/observer-executor.md) is the current starting
template. It proves the complete Executor shape without Accounts or trades.
Its `IngestBBO` records delivery count without simulating matching.

TradeExecutor may become the richer starting template after Account, Trade,
Order, Fill, and Simulator behavior exists.

A complex Executor may receive its own concept page when its implemented design
cannot be explained clearly here. Do not create speculative implementation
pages.

## Current Observer Behavior

```text
Start
  start observer

IngestBBO
  count ingested bbo

OnBBO
  count received bbo
  record last bbo
  record entry
  check stop loss

Run
  record run

Stop
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
