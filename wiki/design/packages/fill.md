# Fill Package

Status: Implemented.
Covers: `internal/fill/*.go`
Purpose: Preserve one actual Venue or Simulator execution as domain evidence.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/account/fill.py`
- `D:/rust/nuubot3/wiki/account/fill.md`

## Ownership

Order owns Fill.

Ledger creates Fill from normalized Venue evidence.

Fill never queries or mutates another component.

## Domain Shape

Fill is an execution value.

It is not a lifecycle component.

It has no `Init`, `Start`, `Run`, or `Stop`.

## Immutable Execution

- Local and Venue identity.
- Ledger, Trade, Order, Account, cycle, and symbol identity.
- CLOID and Venue Order identity.
- Side, quantity, price, and execution time.

These fields never change after admission.

## Late Enrichment

Fee, liquidity, and raw evidence may arrive later.

Later evidence may fill missing metadata.

It cannot change execution identity or economics.

## Operations

```text
New
  validate complete execution identity
  keep immutable execution
  keep available metadata

Enrich
  reject changed execution
  accept later metadata
```

These are domain operations, not lifecycle phases.

## Persistence

One Venue TID identifies one Fill inside one Ledger.

SQLite stores decimals as canonical text.

See [Trading Schema](../concepts/trading-schema.md).

## Required Proof

- Missing identity, nonpositive quantity, nonpositive price, and invalid side fail.
- Duplicate identical evidence is idempotent.
- Late fee and liquidity enrich the same Fill.
- Changed side, quantity, or price fails.
- Fill values survive persistence round-trip exactly.
