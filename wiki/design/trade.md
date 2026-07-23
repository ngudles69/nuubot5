# Trade

## Purpose

Represent one strategy-managed trading lifecycle and its Orders, metrics, and PnL.

## Status

Approved — unimplemented.

## Canonical Sources

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\cloid.md
```

Supplemental behavior:

```text
D:\rust\nuubot3\wiki\account\trade.md
D:\rust\nuubot3\wiki\account\ledger.md
D:\rust\nuutrader6\src\nuubot\hcbots\ledger\ledger.py
```

## Scope

Trade is one mutable current-state domain object for one Account and symbol.

It is not Venue exposure or an execution plan.

## Owner and Children

Ledger owns Trades. Each Trade owns one or more Orders.

## Responsibilities

- Store internal `trade_id` and BotCycle-local `trade_no`.
- Track current and terminal lifecycle state.
- Aggregate Order-level execution totals.
- Calculate gross and net PnL from local evidence.
- Lock final metrics after terminal completion.

## Does Not

- Query Venue.
- Own Account or Ledger.
- Inspect raw Venue payloads.
- Re-scan Fills when Order rollups exist.
- Use Venue PnL as canonical domain PnL.
- Reverse exposure through zero.

## Lifecycle

```text
pending -> open -> closing -> closed
pending -> canceled
any valid state -> error
```

Terminal states are `closed`, `canceled`, and `error`.

## Inputs and Outputs

Inputs are owned Orders, Order rollups, timestamps, and current BBO for open PnL.

Outputs are Trade status, metrics, PnL, and terminal evidence.

## State and Invariants

- `trade_id` is the internal datastore identity.
- `trade_no` is the BotCycle-local CLOID number.
- `trade_no` MUST NOT equal or replace `trade_id`.
- A Trade with any Fill MUST NOT become `canceled`.
- Finalized Trades remain locked.
- Long exposure marks at bid; short exposure marks at ask.

## Concurrency

Only the owning Ledger control path mutates Trade.

Trade contains no locks or shared mutable references.

## Persistence

Ledger persists Trade with its initial Orders in one transaction.

Later Trade changes persist through Ledger.

## Errors

Identity mismatch, exposure reversal, invalid lifecycle transition, or impossible accounting returns an error.

## Program Flow

```text
RuntimeStore reserves trade_no
Account builds Order intents
Ledger creates Trade
Ledger attaches Orders
Ledger persists Trade and Orders atomically
recon updates Orders
Trade refreshes metrics and status
terminal Trade locks final values
```

## Required Proof

- `trade_no` remains distinct from `trade_id`.
- Initial Trade and Orders persist atomically.
- Status follows owned Order evidence.
- PnL uses local Fill evidence and fees.
- Finalized Trade cannot reopen.

## Open Decisions

Nuutrader6 uses Position-centered evidence. Nuubot5 retains Trade unless a separate VenuePosition requirement is approved.
