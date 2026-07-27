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

## Program Flow

```text
BotCycle Init
  retain BotCycle inputs and log init
  create Executors
  log init completed

BotCycle Start
  log start
  start Executors after every sibling initializes
  mark BotCycle running
  log start completed

BotCycle Run
  validate run state
  record run
  deliver current Signal to running Executors
  check coordinated completion

BotCycle Stop
  log stop
  ignore repeated stop request
  mark BotCycle not running
  stop Executors
  collect immutable Executor results
  mark BotCycle completed and stopped
  resolve exit reason
  calculate terminal result
  report results and stats

BotCycle AcctRecon
  read current Clock time
  reconcile running Executor Accounts
  reject partial Account snapshots

BotCycle OnBBO
  record BotCycle time
```

## Coordination

Every Executor starts as one unit.

Each Executor may remain monitoring until its own trigger places Orders.

One normal Executor terminal decision closes the entire BotCycle.

A fatal Executor error fails the BotCycle and propagates.

BotCycle Stop reaches every sibling with one parent reason.

## Reconciliation

`AcctRecon` reconciles every capable running Executor Account before completing one barrier.

The result reports any failure and the maximum consecutive Account Recon failure count.

Any failure suppresses the complete Snapshot barrier. Controller receives no partial snapshots.

A successful barrier returns every immutable Account Snapshot to Controller.

Only after Risk allows the successful barrier does BotCycle deliver `OnRecon`.

`OnRecon` lets each supported running Executor act on accepted Account truth.

TradeExecutor may submit or complete its bracket. GridExecutor may submit initial
Orders or re-enter completed Levels.

`Run(signal)` receives the complete unchanged Signal package from Controller.

It passes that package to supported running Executors, then checks Executor
statuses and reports coordinated BotCycle completion or failure.

BotCycle does not split standard and custom Signal fields.

BotCycle retains Nuubot and reads `nuubot.Clock.NowMS()` when current time is
required. `Run` and `OnRecon` receive no timestamp argument.

## Market Data

BotCycle receives BBO timing evidence only from Controller's MarketData subscription.

BotCycle does not forward BBO values to Executors, Accounts, or Simulator.

Executors and Simulator subscribe directly to exact MarketData keys when required.

Current BtBot admits one replay symbol. Multi-source deterministic replay remains deferred.

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

## Testing

BotCycle has no isolated unit-test file.

Its lifecycle, Signal delivery, reconciliation barrier, Executor coordination,
and shutdown require the real integrated runtime path.

Observer, Trade, and Grid system runs prove those paths through `stest.sh`.

## Telemetry

`Telemetry()` returns the cycle number, status, and one current snapshot per Executor.

It performs no reconciliation, trading action, logging, or persistence.
