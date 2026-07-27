# GridExecutor

Status: Implemented for Simulator.
Covers: `internal/executor/grid.go`
Purpose: Own one arithmetic Grid inside one coordinated BotCycle.

## Identity

`macross_grid_bot` contains exactly one GridExecutor.

One Level number identifies calculations, runtime state, Trades, Orders, logs, and result rows.

Levels always number bottom-up.

Long processes active Levels bottom-up.

Short preserves identity but processes active Levels top-down.

TradeExecutor and future HedgeExecutor use Level zero.

## Geometry

Cycle-start BBO defines starting price.

Lower and upper bounds equal starting price minus and plus configured range.

Arithmetic spacing divides the rounded bounds into `levels - 1` intervals.

Thirty total Levels produce 28 active Levels.

Level zero and the highest Level are non-enterable boundaries.

The starting price need not equal one stored Level.

## Capital

Configured capital defines the first cycle's resource equity.

Later cycles receive the prior cycle's terminal equity.

GridExecutor deploys 95 percent of cycle-start equity.

Equal capital slices divide deployed capital by active Level count.

Quantity uses entry notional plus entry commission.

Final rounded entry and exit notionals must each meet Venue minimum.

Account normalization must not increase the calculated quantity.

## Level

Immutable calculation fields are:

- Level number and boundary flag;
- Grid, initial-entry, re-entry, and exit prices;
- quantity and entry notionals;
- entry and exit commissions;
- initial and re-entry expected PnL; and
- intended action.

Mutable runtime fields are:

- current Trade ID, number, and status;
- Level status;
- initial-submission completion;
- submission attempts; and
- last submitted and completed timestamps.

Normal processing uses stored values.

## Orders

Initial long entry uses `min(current_price, grid_price)`.

Initial short entry uses `max(current_price, grid_price)`.

Re-entry uses stored Grid price.

Long exits at the next higher Level.

Short exits at the next lower Level.

Each active Level submits one GTC entry and one reduce-only trigger TP.

Current Grid Trades have no individual SL.

The adverse Grid boundary owns coordinated stop-loss behavior.

Simulator matches crossed Orders through its direct MarketData subscription.

Live matching and Fill timing remain exchange-owned.

## Economics Gate

Expected PnL deducts both entry and exit commissions.

Initial and re-entry expected PnL must strictly exceed `min_expected_pnl_usdc`.

Failure is fatal before BotCycle admission completes.

## Lifecycle

```text
OnInit
  bind GridExecutor inputs and log init
  validate GridExecutor state
  validate GridExecutor config
  validate fixed side
  retain GridExecutor identity and equity
  initialize Account
  log init completed

OnStart
  log start
  validate start state
  read latest BBO
  calculate Grid levels
  subscribe to MarketData
  submit initial Grid at cycle-start BBO
  mark GridExecutor running
  log start completed

OnStop
  log stop
  validate stop state
  mark GridExecutor stopping
  stop MarketData subscription
  read current time and latest BBO
  reconcile current Account truth
  cancel active Orders
  close open Trades
  reconcile final Venue truth
  capture terminal Account result
  stop Account
  cache terminal Account result
  log stop completed

onBBO
  read latest BBO
  assess Grid boundaries

OnRecon
  read current Nuubot time
  re-enter completed levels
```

Source action comments follow this order.

## Stop

Long lower bound reports `stop_loss`.

Long upper bound reports `take_profit`.

Short reverses those meanings.

Signal, Risk, boundary, and parent stops use one cleanup path.

Cancellation sends TP first, SL second, entry third, then other active Orders.

One ordered cancel batch prevents parent cancellation from invalidating requested children.

GridExecutor submits one reduce-only IOC closure per open Trade.

Every shutdown closure uses the canonical `stop` Order role.

Completion requires every Trade closed, zero active Orders, and zero Account position.

Future HedgeExecutor completion remains part of the same BotCycle barrier.

## Failure

Submission receives one initial attempt and up to two proven-safe retries.

Accepted or uncertain mutation outcomes never retry.

Exhaustion attempts graceful cancel and flatten.

The failure is fatal to BotCycle and Controller.

`max` persistence may retain evidence.

Persisted Grid Trades fail loudly because Runner recovery is not implemented.

## Logging and Results

Every Start logs one table header and one record per validated Level.

Every lifecycle entry and completion log identifies cycle, Executor number, Executor ID, kind, and side.

Each record includes Grid, entry, exit, side, size, notional, and intended action.

Terminal result stores all Level calculations and runtime state.

## Deferred

Minimum price-gap-distance is not configured or implemented.

Geometric spacing is not configured or implemented.

Linear or logarithmic size scaling is not configured or implemented.

Each future calculation contract requires another exact BotSpec.

## Proof

- Ten Levels create eight initial Trades and 16 Orders.
- Four midpoint long entries fill while four remain resting.
- Adverse bound cancels Orders and closes eight open Trades.
- Final Account is flat and inactive.
- CLOIDs preserve active Level identity.
- A three-month replay publishes complete Grid evidence.
