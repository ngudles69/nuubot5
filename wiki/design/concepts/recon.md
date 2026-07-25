# Reconciliation

Status: Implemented for Simulator-backed TradeExecutor.
Covers: `internal/controller`, `internal/botcycle`, `internal/executor`,
`internal/account`, and `internal/ledger`
Purpose: Rebuild coherent Account truth before Risk and execution policy.

## Flow

```text
Controller Run
  BotCycle reconciles each capable Executor Account
    Account queries Simulator or Venue truth
    Ledger admits Order and Fill evidence atomically
    Account returns immutable Snapshot
  Controller builds RiskInput
  Risk decides
  BotCycle delivers accepted OnRecon
```

Failed reconciliation prevents Risk-dependent admission and Executor policy.

Controller receives values only.

It retains no Account, Ledger, Trade, Order, Fill, or Simulator pointer.

## Dirty State

Submission, Simulator mutation, and future user events mark Account dirty.

Normal reconciliation may skip a clean Account.

Forced reconciliation always queries truth.

WebSocket evidence remains an optional dirty hint.

## Completion

TradeExecutor Stop cancels Orders, closes exposure, forces reconciliation, and
proves zero active Orders and zero position.

Flatness alone never requests BotCycle exit.
