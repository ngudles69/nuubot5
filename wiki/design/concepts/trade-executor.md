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
  bind TradeExecutor inputs
  validate TradeExecutor state
  validate TradeExecutor config
  validate fixed side
  retain current BBO and identity
  initialize Account
  mark TradeExecutor running
  log init completed

OnStop
  validate stop state
  mark TradeExecutor stopping
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

`OnInit` performs no Order mutation.

The first accepted `OnRecon` submits one entry, take-profit, and stop-loss
bracket.

Subsequent BBO and reconciliation passes update Venue and Ledger truth.

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
