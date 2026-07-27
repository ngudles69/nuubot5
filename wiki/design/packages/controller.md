# Controller Package

Status: Implemented for standalone historical replay.
Covers: `internal/controller/controller.go`
Purpose: Construct and run one configured Bot from one shared Nuubot harness.

## Ownership

Controller receives one shared `*setup.Nuubot` and retains it as `c.nuubot`.

Nuubot supplies Logger, App Config, Bot identity and provenance, replay input,
typed BotSpec, Clock, MarketData, Meta, and ResultPath.

Controller retains `nuubot.Log` as `c.log` for convenient logging.

Controller does not duplicate Nuubot, BotSpec, Bot identity, App Config, Meta, or ResultPath.

Controller constructs and owns one persistent Signaler, configured Risks, and
zero or one active BotCycle.

BotCycle constructs configured Executors when a cycle starts.

Controller never reparses App Config, BotConfig, or Setup infrastructure.

## Program Flow

```text
Controller Init
  retain Nuubot and log init
  set Signaler replay range
  create Signaler
  create Risks
  retain Controller components
  initialize Controller state
  subscribe to MarketData timing
  initialize resource capital
  log init completed

Controller Start
  log start
  validate start state
  mark Controller started
  log start completed

Controller Run
  validate run state
  read Clock, record run, and check stop request
  reconcile active BotCycle
  assess Risks
  apply Risk decisions
  deliver accepted reconciliation
  read current Signal
  record new Signal
  run active BotCycle with current Signal
  apply Signal action

Controller Stop
  log stop
  ignore repeated stop request
  request Controller stop
  stop MarketData subscriptions
  close active BotCycle
  stop Risks
  stop Signaler
  mark Controller stopped
  log stopped results and stats
  return stop error
```

## MarketData Callback

```text
onBBO
  read latest BBO
  record Controller market time
  record active BotCycle market time
  record accepted tick
```

## Control Pass

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

Controller never queues or splits a Signal package.

Controller reads current time from `nuubot.Clock`.

Controller reads the complete current package, passes it unchanged into active
`BotCycle.Run(signal)`, then applies its standard lifecycle Action.

Controller passes no timestamp into `AcctRecon`, `OnRecon`, or `BotCycle.Run`.

Controller subscribes to MarketData only to record accepted ticks, BBO gaps, and active BotCycle market time.

Controller owns no latest-BBO map and performs no Executor or Simulator BBO forwarding.

Controller has no isolated unit tests.

RTest proves Controller through the real `Setup -> Nuubot -> Controller -> BtBot`
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

Controller stops its MarketData timing subscriptions, then stops the active BotCycle.

BotCycle stops and flattens every Executor Account.

Controller then stops Risks and Signaler.

Fatal child errors propagate to BtBot.

## Deferred

- Live cross-process Account claims.
- Multi-source replay merging.
- Physical Account and global risk.
- Runner telemetry persistence and API queries.
