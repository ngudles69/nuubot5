# BotCycle Package

Status: Implemented as one coordinated Executor unit.
Covers: `internal/botcycle/botcycle.go`
Purpose: Own one admitted activation of every configured Executor.

## Admission

BotCycle initializes every Executor before any Executor submits Orders.

If one initialization fails, BotCycle stops already initialized siblings in
reverse order.

After every Executor initializes, BotCycle starts each supported Executor.

Grid submits its initial Orders only inside this start barrier.

Admission rejection is nonfatal to Controller.

Unexpected initialization failure is fatal.

## Coordination

Every Executor starts as one unit.

Each Executor may remain monitoring until its own trigger places Orders.

One normal Executor terminal decision closes the entire BotCycle.

A fatal Executor error fails the BotCycle and propagates.

BotCycle Stop reaches every sibling with one parent reason.

## Reconciliation

BotCycle reconciles capable Executor Accounts as one barrier.

It returns immutable Account snapshots to Controller.

Only after Risk allows the pass does BotCycle deliver `OnRecon`.

## Market Data

BotCycle distributes symbol-qualified BBO values.

Executors ignore values for other symbols.

Current BtRunner admits one replay symbol.

Multi-source deterministic replay remains deferred.

Stopping Executors continue Venue BBO ingestion until coordinated cleanup begins.

## Completion

Flatness alone never requests exit.

An explicit Signal, Risk, Executor, or parent exit starts Stop.

TradeExecutor Stop cancels active Orders, closes exposure, reconciles, and
proves zero Orders and zero position.

GridExecutor Stop cancels ordered active batches, closes every open Trade, and proves the same flat state.

Only then does BotCycle complete.

## Result

BotCycleResult contains its cycle number and ordered ExecutorResults.

ExecutorResult preserves ID, role, kind, side, resource, capital, order size,
status, exit reason, and optional AccountResult.

Grid ExecutorResult also preserves cancellation, closure, retry, round-trip, and Level evidence.

## Telemetry

`Telemetry()` returns the cycle number, status, and one current snapshot per Executor.

It performs no reconciliation, trading action, logging, or persistence.
