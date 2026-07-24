# Trade Package

Status: Implemented for Simulator trading evidence.
Covers: `internal/trade/*.go`
Purpose: Represent one trading intent and derive its state and PnL from owned Orders and Fills.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/account/trade.py`
- `D:/rust/nuubot3/wiki/account/trade.md`

## Ownership

Ledger owns Trade.

Trade owns Orders.

Trade never queries Account, Venue, Simulator, or storage.

## Domain Shape

Trade is a mutable domain object.

It is not a lifecycle component.

It has no `Init`, `Start`, `Run`, or `Stop`.

Ledger constructs it with admitted identity and persisted state.

## States

```text
pending -> open -> closing -> closed
pending -> canceled
pending/open/closing -> error
```

`closed`, `canceled`, and `error` are terminal.

Terminal Trade values never reopen.

## Responsibilities

- Attach only Orders with matching Ledger, Account, cycle, symbol, and Trade identity.
- Derive position size and average entry price.
- Aggregate realized PnL and fees.
- Mark open exposure from current BBO.
- Derive status from Fill and active Order evidence.
- Lock final values when terminal.

## PnL

Trade calculates domain PnL from owned Fill evidence.

Simulator and Hyperliquid PnL fields remain raw comparison evidence.

```text
gross_pnl = realized_pnl + unrealized_pnl
net_pnl   = gross_pnl - fees
```

Long open exposure marks at bid.

Short open exposure marks at ask.

Fill processing must reject a Trade reversal.

## Operations

```text
New
  keep admitted identity
  keep admitted state

AddOrder
  validate Order ownership
  attach Order
  refresh Trade

Refresh
  order Fills by event time
  calculate exposure
  calculate PnL
  derive status
  lock terminal values

Snapshot
  return immutable Trade values
```

These are domain operations, not lifecycle phases.

## Persistence

See [Trading Schema](../concepts/trading-schema.md).

SQLite stores decimal values as canonical text.

## Required Proof

- Entry Fills create correct long and short exposure.
- Multiple same-side Fills calculate weighted entry price.
- Closing Fills calculate gross realized PnL.
- Fees reduce net PnL once.
- Over-closing Fill batches fail before mutation.
- Terminal Trades cannot reopen.
