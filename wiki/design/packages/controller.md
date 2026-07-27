# Controller Package

Status: Implemented for standalone historical replay.
Covers: `internal/controller/controller.go`
Purpose: Construct and run one configured Bot from complete Setup and typed BotSpec.

## Ownership

Controller receives complete `setup.Infrastructure` and one immutable
`botspec.Spec`.

Complete Setup supplies App Config, Bot identity and provenance, replay inputs,
Meta, and ResultPath.

BotSpec supplies typed Controller, Signaler, Risk, and Executor specifications.

Controller constructs and owns one persistent Signaler, configured Risks, and
zero or one active BotCycle.

BotCycle constructs configured Executors when a cycle starts.

Controller never reparses App Config or loads Setup infrastructure.

## Control Pass

```text
reconcile active BotCycle
build immutable RiskInput
assess every Risk
deliver accepted reconciliation
close a terminal BotCycle
read current Signaler action
start or stop one BotCycle
```

Reconciliation precedes Risk and Executor policy.

A failed barrier with maximum consecutive count one or two ends that control pass.

Those passes run no Risk assessment, Executor `OnRecon`, or BotCycle `Run`.

A failed barrier with any count at least three returns an error and fails the Sweep.

Persistence and execution failures outside Account Recon remain immediately fatal.

Risk decisions are `Allow`, `BlockCycleStart`, `StopCycle`, and
`StopController`.

Every Risk assesses the same immutable RiskInput before Controller acts.

`StopController` dominates `StopCycle`, which dominates start blocking.

`StopCycle` prevents same-pass cycle start even when no cycle is active.

Signaler actions are `NoAction`, `StartCycle`, and `StopCycle`.

Controller never queues a Signal.

Controller has no isolated unit tests.

RTest proves Controller through the real `Setup -> BotSpec -> Controller -> BtBot`
path using `./stest.sh -bot 9`.

After completion, the current `StartCycle` action may start another cycle on
the next control event.

Controller never restarts a cycle in the event that closed it.

## Capacity

Exactly one BotCycle may be active.

There is no configurable concurrency limit.

`max_cycles` limits total configured cycles for one Controller generation.

For example, `max_cycles = 999` permits 999 sequential cycles.

It does not permit concurrent cycles.

## Capital and Risk

Controller tracks declared capital and current simulated equity per distinct
Executor resource.

Completed Account equity becomes the next cycle's starting equity.

RiskInput contains current Account snapshots, Bot capital, net PnL, Bot
equity, peak equity, and drawdown.

BalancedRisk currently always returns `Allow`. It is not protection.

## Result

`controller.Result` contains:

- exact Bot identity;
- ordered Signal and changed Risk decisions;
- ordered BotCycle and Executor results;
- Controller exit reason; and
- Bot capital, net PnL, equity, peak equity, and drawdown.

It retains no live child pointer.

## Telemetry

`Telemetry()` composes current Controller counters, capital, equity, PnL, drawdown, and active BotCycle state.

It calls the active BotCycle telemetry path once.

It performs no mutation, reconciliation, trading decision, logging, or persistence.

## Shutdown

Controller stops the active BotCycle first.

BotCycle stops and flattens every Executor Account.

Controller then stops Risks and Signaler.

Fatal child errors propagate to BtBot.

## Deferred

- Live cross-process Account claims.
- Multi-source replay merging.
- Physical Account and global risk.
- Runner telemetry persistence and API queries.
