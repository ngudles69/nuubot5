# BotCycle Package

Status: Implemented as one coordinated Executor unit.
Covers: `internal/botcycle/botcycle.go`
Purpose: Own one admitted activation of every configured Executor.

## Admission

BotCycle initializes every Executor before any Executor submits Orders.

If one initialization fails, BotCycle stops already initialized siblings in
reverse order.

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

## Completion

Flatness alone never requests exit.

An explicit Signal, Risk, Executor, or parent exit starts Stop.

TradeExecutor Stop cancels active Orders, closes exposure, reconciles, and
proves zero Orders and zero position.

Only then does BotCycle complete.

## Result

BotCycleResult contains its cycle number and ordered ExecutorResults.

ExecutorResult preserves ID, role, kind, side, resource, capital, order size,
status, exit reason, and optional AccountResult.
