# Trading State Tranche

Status: Implemented for BtRunner with Simulator.
Covers: `internal/runtime`, `internal/botcycle`, `internal/executor`, `internal/account`, `internal/ledger`, `internal/trade`, `internal/order`, `internal/fill`, and `internal/simulator`
Purpose: Define the complete Simulator-first trading path.

## Boundary

Implemented:

- TradeExecutor long and short brackets.
- Account validation, CLOID creation, submission, dirty state, and reconciliation.
- Ledger-owned Trade, Order, and Fill evidence.
- Simulator matching, bracket activation, OCO, reduce-only, fees, and PnL.
- Memory-only Sweep execution with terminal SQLite publication.
- Maximum persistence with durable Ledger and Simulator child-state reload.
- Runtime reconciliation before Risk and Executor policy.

Pending:

- Frozen JSON comparison against `async_hyperliquid`.
- Controlled Hyperliquid testnet mutation parity.
- Live and testnet Account mutation.
- WebSocket user-event dirty hints.
- Shared Sweep terminal summaries.

## Ownership

```text
BtRunner
  Runtime
    BotCycle
      TradeExecutor
        Account
          Ledger
            Trade
              Order
                Fill
          Simulator
```

Each owner calls only direct children.

Runtime receives Account snapshots and terminal results by value.

Runtime never owns Account pointers.

## Event Order

```text
BBO
  Runtime.IngestBBO
  BotCycle.IngestBBO
  TradeExecutor.IngestBBO
  Account.IngestBBO
  Simulator.IngestBBO
  BotCycle.OnBBO
  TradeExecutor.OnBBO

Runtime timer
  BotCycle.Reconcile
  Risk.AssessStop
  BotCycle.OnRecon
  BotCycle.Run
```

Simulator matching always precedes normal Executor BBO policy.

Reconciliation always precedes Risk and Executor decisions.

## Entry

Runtime consumes one standard `enter_long` or `enter_short` Signal.

BotCycle creates TradeExecutor with the latest admitted BBO.

TradeExecutor initializes Account without submitting Orders.

The first successful recon submits entry, take-profit, and stop-loss Orders.

Account creates one Trade, three Orders, and three CLOIDs before Simulator submission.

Reconciliation admits final Order and Fill truth.

## Completion

TP or SL Fill cancels its sibling and leaves the position flat.

TradeExecutor becomes stopping only when its Trade is terminal and no Order is active.

At EOF or parent stop, TradeExecutor:

1. reconciles current truth;
2. cancels active Orders;
3. submits one reduce-only IOC close when exposure remains;
4. reconciles final truth;
5. requires zero exposure and zero active Orders;
6. captures Account result;
7. stops Account.

## Persistence

`none` opens no result database during execution.

Successful shutdown publishes one per-Bot SQLite database through a partial file.

`max` persists every accepted Ledger mutation and Simulator state change.

Recovered Simulator state excludes transient BBO time.

A fresh BBO is required before Account can publish a marked snapshot.

Full Bot resume remains pending Runner-owned replay, Runtime, Signaler, and TradeExecutor cursors.

TradeExecutor fails initialization when persisted Trades exist.

## Proof

Sweep `9`, Bot `13` replayed:

- 7,948,800 ticks.
- 794,880 Runtime passes.
- 50 Trades.
- 151 Orders.
- 100 Fills.
- Zero active terminal Orders.
- 50 terminal Ledgers and Simulator states.

The result database passed SQLite integrity and foreign-key checks.

Final stability proof passed 13/13 runs: 1x, 2x, then 10x.
