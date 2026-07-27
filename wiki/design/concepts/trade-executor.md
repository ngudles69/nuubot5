# TradeExecutor

Status: Implemented for Simulator.
Covers: `internal/executor/trade.go`
Purpose: Execute one fixed-side bracket policy through Account.

## Config

TradeExecutor receives one typed immutable Spec:

- fixed side and role;
- distinct Venue, network, physical Account, and symbol;
- declared capital;
- fixed order size;
- take-profit and stop-loss percentages;
- fee and slippage assumptions;
- persistence mode; and
- admitted Meta and minimum notional.

Signal contains no direction.

## Lifecycle

```text
OnInit
  bind TradeExecutor inputs and log init
  reject terminal TradeExecutor state
  validate TradeExecutor config
  retain TradeExecutor identity
  initialize Account
  log init completed

OnStart
  log start
  reject terminal TradeExecutor state
  read latest BBO
  continue loaded TradeExecutor state
  log start completed

OnStop
  log stop
  validate stop state
  mark TradeExecutor stopping
  read current time and latest BBO
  reconcile current Account truth
  cancel active Orders
  close open exposure
  reconcile final Venue truth
  capture terminal Account result
  stop Account
  cache terminal Account result
  log stop completed

OnRecon
  submit bracket when no Trade exists
  check owned Trade completion
```

Every lifecycle entry and completion log identifies cycle, Executor number, Executor ID, kind, and side.

`OnInit` performs no Order mutation.

The first accepted `OnRecon` submits one entry, take-profit, and stop-loss
bracket.

Simulator receives BBO through its MarketData subscription.

TradeExecutor reads the latest BBO only during Start, first bracket policy, and Stop.

Reconciliation updates Venue and Ledger truth.

Bracket completion makes the Executor terminal.

## Stop

Signal, Risk, parent, and normal completion use one Stop path.

Stop cancels remaining Orders, closes exposure, forces reconciliation, and
rejects non-flat terminal state.

It then captures immutable AccountResult.

## Equity

Controller supplies the resource's current starting equity.

Declared capital and order size remain unchanged.

Terminal Account equity returns through ExecutorResult for the next cycle and
Bot-level risk calculations.
