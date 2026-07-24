# ObserverExecutor

Status: Implemented.
Covers: `internal/executor/observer.go`
Purpose: Prove execution control and stop-loss behavior without orders or account state.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/executor/observer.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/executor.md`

## Scope

ObserverExecutor observes BBO values, records an entry, calculates stop price, and becomes terminal after stop loss.

## Owner and Children

BotCycle owns ObserverExecutor through the Executor interface.

ObserverExecutor owns no child.

## Responsibilities

- Validate configured stop-loss percentage.
- Record the first BBO after Signal availability as entry.
- Calculate side-specific stop-loss price.
- Track last timestamp and price.
- Trigger at the inclusive stop boundary.
- Preserve final evidence during parent stop.
- Report terminal statistics once.

## Does Not

- Place or cancel orders.
- Create Account, Ledger, Trade, Order, Fill, or Simulator state.
- Model slippage, fees, liquidity, or fills.
- Decide Runtime stop policy.

## Lifecycle

Create, `Start`, repeated `OnBBO` and timed `Run`, then idempotent `Stop`.

Invalid `Start` returns an error.

`OnBBO` silently ignores inactive or terminal states.

`Run` returns terminal state successfully and has no error path.

## Inputs and Outputs

Inputs are Signal, stop-loss percentage, BBO values, Run timestamps, and parent stop reason.

Outputs are terminal state, exit reason, and execution evidence in terminal statistics.

## State and Invariants

Stop-loss percentage MUST be greater than zero and less than one.

Long stops when price is at or below entry multiplied by one minus stop percentage.

Short stops when price is at or above entry multiplied by one plus stop percentage.

Parent stop preserves last BBO as final time and price.

## Concurrency

ObserverExecutor is synchronous.

## Persistence

None.

## Errors

Invalid construction and invalid `Start` calls fail.

`OnBBO` silently ignores inactive or terminal states.

Current `Run` and `Stop` return no operational error.

`Stop` is idempotent.

## Program Flow

```text
createObserver
  validate config
  create observer

Start
  start observer

OnBBO
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
```

## Required Proof

- Long and short inclusive stop boundaries trigger.
- Entry comes from first received BBO.
- Parent shutdown after Reader exhaustion records final evidence.
- Stop is idempotent.
- Accepted replay reports 17 stop-loss exits and one parent-stop closure.

## Open Decisions

None.
