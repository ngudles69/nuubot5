# Executor Package

Status: Implemented for ObserverExecutor, Simulator TradeExecutor, and Simulator GridExecutor.
Covers: `internal/executor/*.go`
Purpose: Own fixed execution roles inside one BotCycle.

## Spec

One immutable Executor Spec contains:

- ID, role, kind, and fixed side;
- `(venue, network, physical_account_id, symbol)` resource;
- declared capital and order size;
- strategy-specific levels, thresholds, and percentages;
- strategy-specific persistence policy.

Meta, application-wide minimum notional, Logger, and ResultPath come from the
shared Nuubot harness. They are not Executor specification fields.

BotSpec admission rejects duplicate resources inside one Bot.

Direction never comes from Signal.

## Lifecycle

```go
type Executor interface {
    OnInit(Context) error
    OnStop(string) error
    Status() Status
    ExitReason() string
    Telemetry() Telemetry
    Result() (Result, error)
}
```

Optional narrow capabilities handle complete Signal packages, BBO,
reconciliation, and Account snapshots.

`SignalHandler.OnSignal` receives the unchanged `signaler.Package`. Each
Executor reads only the fields required by its BotSpec.

BotCycle owns capability dispatch.

`Telemetry()` returns current status and optional Account state.

It performs no reconciliation, mutation, logging, or persistence.

## Testing

Executor runtime lifecycle has no isolated unit-test files.

ObserverExecutor, TradeExecutor, and GridExecutor runtime behavior require the
real BotCycle, Account, Ledger, Simulator, Clock, Signal, and shutdown path.

Only pure deterministic Grid calculations retain unit tests in
`internal/executor/grid_test.go`.

## ObserverExecutor

Observer proves fixed-side lifecycle and stop-loss behavior without Orders.

It remains a replay proof Executor.

## TradeExecutor

TradeExecutor owns one Simulator-backed Account for its BotCycle.

It submits one bracket after the first accepted reconciliation.

Normal completion, Signal exit, Risk exit, and parent Stop all use the same
flat cleanup path.

Controller carries terminal Account equity into the same resource's next
cycle.

## GridExecutor

GridExecutor owns one Simulator-backed Account and one bottom-up Level table.

Thirty total Levels contain two non-enterable boundaries and 28 active Levels.

Current `macross_grid_bot` uses arithmetic spacing and equal capital slices.

It deploys 95 percent of cycle-start equity across active Levels.

Initial long entries use the lower of current and Grid price.

Initial short entries use the higher value.

Each active Level submits one GTC entry and one reduce-only TP.

Completed Levels re-enter at their stored Grid price.

Every calculation is stored before submission.

Each expected PnL deducts entry and exit commissions.

Zero, negative, or threshold-failing expected PnL stops Controller startup.

Both boundaries stop the BotCycle.

Adverse boundaries report `stop_loss`; favorable boundaries report `take_profit`.

Shutdown cancels TP, SL, then entry Orders in one ordered batch.

It then closes each open Trade and proves zero Orders and zero position.

TradeExecutor and GridExecutor use the same `stop` Order role for shutdown exposure.

One submission plus two proven-safe retries precedes fatal graceful exit.

Accepted or uncertain mutation outcomes never retry.

Every start logs one header and one record per validated Level.

Geometric spacing, scaled sizing, and minimum price-gap filtering are absent.

Those behaviors require another exact BotSpec.

## Resource Rules

- Same Account with different symbols is allowed.
- Different Accounts with the same symbol are allowed.
- The same Account-symbol resource is forbidden, regardless of side.
- Current BotSpecs use Simulator `simnet`.
- Live Venue admission and cross-process claims remain deferred.
