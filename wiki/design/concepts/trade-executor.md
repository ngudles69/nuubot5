# TradeExecutor

Status: Implemented for BtRunner with Simulator.
Covers: `internal/executor/trade.go`
Purpose: Define the first Account-owning Executor and complete simulated bracket-trade path.

## Scope

TradeExecutor is the first trading Executor template.

The first implementation runs only under BtRunner with Simulator.

Live and testnet mutation remain blocked.

## Required Inputs

| Input | Owner | Purpose |
|---|---|---|
| Trigger Signal | Runtime | Choose long or short entry |
| Latest BBO | Runtime | Supply current entry and shutdown reference |
| Executor config | Config | Supply Account and bracket policy |
| Credentials catalog | Setup | Resolve later live Accounts |
| Clock time | Runtime | Preserve deterministic event time |

The current Signal needs no ATR or new custom field.

## Initial Config

```text
kind
account_name
network
order_notional_usdc
take_profit_pct
stop_loss_pct
simulator_equity_usdc
simulator_fee_pct
simulator_slippage_pct
persist_mode
```

BtRunner accepts only `network = "simnet"`.

`persist_mode` is `none` or `max`.

TradeExecutor passes it to Account.

Percentages and USDC values use canonical decimal text.

Account applies Meta rounding and the configured USDC 11 minimum.

## Lifecycle

`executor.Create` selects TradeExecutor and calls `OnInit`.

`OnInit` validates configuration, Signal, current BBO, and Simulator admission.

`OnInit` initializes one Account.

`OnInit` performs no Venue mutation.

`OnInit` fails when persisted Trades exist because full Runner recovery is pending.

The Account starts reconciliation-dirty.

The first successful recon event submits the bracket.

`OnStop` cancels active exits, closes exposure, reconciles, captures terminal evidence, and stops Account.

Stopped and error remain terminal.

## Optional Capabilities

```go
type BBOIngestHandler interface {
    IngestBBO(market.BBO) error
}

type BBOHandler interface {
    OnBBO(market.BBO)
}

type AccountReconciler interface {
    Reconcile(nowMS uint64, forced bool) (account.Snapshot, bool, error)
}

type ReconHandler interface {
    OnRecon(nowMS uint64) error
}
```

`AccountReconciler` refreshes Account truth only.

`ReconHandler` runs policy only after Runtime accepts the completed recon barrier.

TradeExecutor implements all four capabilities.

## Program Flow

```text
OnInit
  validate trade config
  admit trigger Signal
  admit current BBO
  initialize Account
  reject persisted Trades
  initialize TradeExecutor

IngestBBO
  ingest Account BBO

OnBBO
  record current BBO

Reconcile
  reconcile Account when dirty or forced
  return Account snapshot

OnRecon
  submit bracket when no Trade exists
  check owned Trade completion

OnStop
  reconcile current Account truth
  cancel active Orders
  close open exposure
  reconcile final Venue truth
  capture terminal Account result
  stop Account
  cache terminal Account result
  stop TradeExecutor

AccountResult
  return cached terminal Account result
```

Each indented action becomes one exact source comment during implementation.

## Entry Plan

Long entry uses current ask.

Short entry uses current bid.

Configured notional divided by entry reference produces requested quantity.

Account performs Meta size rounding.

Take-profit and stop-loss prices derive from configured percentages.

Account performs Meta price rounding.

The bracket contains one entry, one take-profit, and one stop-loss Order.

All three Orders belong to one Trade.

The batch preserves entry, take-profit, then stop-loss order.

## Simulator Submission

The entry is an explicit market-like IOC request.

It may fill immediately from the supplied current reference.

TP and SL remain inactive until entry execution exists.

Resting submission returns `resting`, `waitingForFill`, and `waitingForFill`.

Immediate entry Fill returns `filled`, `waitingForTrigger`, and `waitingForTrigger`.

Resting and trigger Orders cannot consume an already-processed BBO.

The next admitted BBO drives later matching.

## Completion

TradeExecutor remains running while any owned Trade is nonterminal.

It also remains running while any owned Order is active.

It stops after every owned Trade is terminal and every owned Order is inactive.

## Shutdown

TradeExecutor first cancels active TP, SL, and other exits.

It submits one reduce-only market-like close for remaining exposure.

Simulator may execute that explicit close from the latest admitted BBO reference.

Account immediately reconciles the close outcome.

TradeExecutor captures `Account.Result` after that final reconciliation.

It stops Account, then exposes only the cached immutable value.

Shutdown fails if exposure remains or no required BBO exists.

No terminal state may hide an open simulated position.

## Does Not

- Read raw Hyperliquid JSON.
- Mutate Ledger children directly.
- Calculate Venue fills.
- Interpret another Executor's Account.
- Select live credentials during Simulator operation.
- Retry unknown mutation outcomes.
- Depend on ATR during the first slice.

## Required Proof

- Long and short plans produce correct sides and percentages.
- Invalid or conflicting entry Signals reject BotCycle admission.
- BtRunner rejects live and testnet Accounts.
- One bracket creates one Trade and three Orders.
- Entry execution arms TP and SL.
- TP or SL execution cancels its sibling.
- Account preserves one result for each ordered bracket request.
- Completion requires terminal Trade and inactive Orders.
- Shutdown finishes flat and reconciled.
- Terminal result includes shutdown Orders, Fills, and flat exposure.
- No credential appears in logs or evidence.
