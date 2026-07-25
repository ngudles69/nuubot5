# Executor Package

Status: Implemented for ObserverExecutor and Simulator TradeExecutor.
Covers: `internal/executor/*.go`
Purpose: Own fixed execution roles inside one BotCycle.

## Spec

One immutable Executor Spec contains:

- ID, role, kind, and fixed side;
- `(venue, network, physical_account_id, symbol)` resource;
- declared capital and order size;
- strategy-specific percentages;
- Meta and minimum notional; and
- result persistence path.

BotSpec admission rejects duplicate resources inside one Bot.

Direction never comes from Signal.

## Lifecycle

```go
type Executor interface {
    OnInit(Context) error
    OnStop(string) error
    Status() Status
    ExitReason() string
    Result() (Result, error)
}
```

Optional narrow capabilities handle BBO, reconciliation, and Account
snapshots.

BotCycle owns capability dispatch.

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

## Resource Rules

- Same Account with different symbols is allowed.
- Different Accounts with the same symbol are allowed.
- The same Account-symbol resource is forbidden, regardless of side.
- Current BotSpecs use Simulator `simnet`.
- Live Venue admission and cross-process claims remain deferred.
