# ObserverExecutor

Status: Implemented.
Covers: `internal/executor/observer.go`
Purpose: Provide the complete starting template without placing trades.

## Benchmark Role

Observer is the full-replay throughput baseline.

Its backtest consumes the complete historical dataset, including all 7.9 million BBO ticks in the current Bot 9 replay.

Observer has no artificial BBO-count or loop-count exit.

The benchmark excludes Account, Order, Fill, Trade, Venue execution, and conditional trading decisions.

Its lightweight Signal, BotCycle, subscription, logging, and configured stop-loss lifecycle remain active.

Replay completion stops the parent Controller and its active Observer.

## Ownership

BotCycle owns ObserverExecutor through the Executor interface.

Observer owns no Account or trading state.

Executor factory constructs it and calls `OnInit`.

## Lifecycle

Observer uses the canonical Executor status.

```text
configured
  starting
    running
      stopping
        stopped
```

Invalid initialization enters error.

Stopped and error states never transition.

No separate terminal flag exists.

`OnStop` is idempotent.

## Admission

Observer requires one exact market identity: Venue, network, and symbol.

Legacy exact Observer Config without Venue and network receives the defined `simulator/simnet` default.

Observer requires exactly one standard entry trigger.

Missing or conflicting triggers reject BotCycle admission.

Valid long or short entry starts Observer immediately.

## Capabilities

Observer implements `StartHandler` and owns one MarketData subscription while running.

It implements no BBO ingestion or unused event method.

## Program Flow

```text
OnInit
  bind ObserverExecutor base inputs and log init
  validate ObserverExecutor state
  bind ObserverExecutor inputs and mark starting
  validate ObserverExecutor config
  log init completed

OnStart
  log start
  validate ObserverExecutor state
  read latest BBO
  subscribe to MarketData
  mark ObserverExecutor running
  log start completed

OnStop
  log stop
  stop MarketData subscription
  ignore terminal stop request
  mark ObserverExecutor stopping
  preserve stop reason
  preserve end time
  mark ObserverExecutor stopped
  calculate duration
  log stop completed

onBBO
  read latest BBO
  record received BBO
  record current BBO
  assess stop loss
```

## Stop Loss

The latest BBO read during Start becomes the observed entry.

Later subscribed BBO values assess stop loss.

Long stops at or below entry multiplied by one minus stop percentage.

Short stops at or above entry multiplied by one plus stop percentage.

Stop loss moves Observer to stopping.

Controller closes the owning BotCycle during its next timed pass.

## Logging

Observer never logs each BBO.

Every lifecycle entry and completion log identifies cycle, Executor number, Executor ID, kind, and side.

Its final summary reports:

- Triggering Signal facts.
- Entry and final prices.
- Stop-loss price.
- Duration and reason.
- `on_bbo_count`.

## Does Not

- Place or cancel Orders.
- Create Account, Ledger, Trade, Fill, Simulator, or Venue state.
- Match Orders or create simulated Fills.
- Model fees, liquidity, or slippage.
- Directly stop Controller.
