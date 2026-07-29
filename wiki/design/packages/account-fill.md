# Fill Package

Status: Implemented.
Covers: `internal/account/fill/*.go`
Purpose: Preserve one actual Venue or Simulator execution as domain evidence.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/account/fill.py`
- `D:/rust/nuubot3/wiki/account/fill.md`

## Ownership

Ledger owns each flat Fill.

Order owns no Fill objects.

Ledger creates Fill from normalized Venue evidence and links it by keys.

Fill never queries or mutates another component.

## Domain Shape

Fill is an execution value.

It is not a lifecycle component.

It has no `Init`, `Start`, `Run`, or `Stop`.

## Immutable Execution

- Local and Venue identity.
- Ledger, Trade, Order, Account, cycle, and symbol identity.
- Venue TID.
- CLOID when supplied by Exchange.
- Venue Order identity when supplied by Exchange.
- Side, quantity, price, and execution time.

These fields never change after admission.

`FillID` is the Nuubot local identity.

Venue TID is the immutable Exchange identity.

Memory-only Ledger allocates `FillID` and indexes Venue TID to that key.

Ledger never copies parent CLOID or OID into Fill evidence.

Incoming Fill CLOID and OID remain unchanged and are saved when present.

Either supplied identity may match the parent Order.

When both are supplied, both must match the same Order.

Missing identity remains missing.

Unknown or conflicting identity rejects the Fill without mutation.

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

Update
  reject conflicting fee evidence
  accept later fee and raw evidence
```

These are domain operations, not lifecycle phases.

## Persistence

One FillID identifies one Fill inside one Ledger.

Venue TID remains unique Exchange identity.

SQLite stores decimals as canonical text.

See [Trading Schema](../concepts/trading-schema.md).

## Required Proof

- Missing identity, nonpositive quantity, nonpositive price, and invalid side fail.
- Duplicate identical evidence is idempotent.
- Late fee and liquidity enrich the same Fill.
- Changed side, quantity, or price fails.
- Fill values survive persistence round-trip exactly.
