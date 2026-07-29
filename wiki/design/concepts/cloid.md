# CLOID

Status: Canonical application Order identity.
Covers: `internal/cloid/*.go`
Purpose: Encode one canonical Order key for Venue use.

## Scope

CLOID is the Venue-facing identity of one Order.

Account creates each CLOID after Ledger allocates the Order.

Strategy code and Venue implementations never invent CLOIDs.

## Layout

```text
0x + 16 lowercase hexadecimal LedgerID digits
   + 16 lowercase hexadecimal OrderID digits
```

The result is exactly one 128-bit Hyperliquid CLOID.

## Identity

The canonical Order key is `(LedgerID, OrderID)`.

CLOID encodes that key directly.

It contains no Trade, Order-set, role, side, level, timestamp, or strategy
metadata.

An Order set has no identity.

## Responsibilities

- Reject zero Ledger or Order identities.
- Produce exactly 32 lowercase hexadecimal digits after `0x`.
- Preserve both canonical key values without truncation.

## Does Not

- Allocate identities.
- Decode unused metadata.
- Group Orders.
- Identify Trades or Fills.
- Submit or cancel Orders.
- Provide compatibility layouts.

## Invariants

- One CLOID identifies one Order.
- One Order keeps one CLOID.
- CLOID identity never depends on Order role or lifecycle state.
- Account creates CLOID only after Ledger allocates `OrderID`.
