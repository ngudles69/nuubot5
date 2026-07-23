# Fill

**Status:** Approved — unimplemented in Nuubot5.

## Purpose

Record one actual Venue or Simulator execution for one Order.

## Scope

One real partial execution equals one Fill.

Fill is a current evidence record, not an event stream or accounting engine.

## Responsibilities

- Preserve venue Fill identity.
- Preserve execution side, quantity, price, and source timestamp.
- Preserve ownership identity through Order, Trade, Ledger, Account, and BotCycle.
- Accept later fee, liquidity, timestamp, or raw-evidence enrichment.
- Reject conflicting execution facts for the same venue Fill.

## Does Not

- Own Order.
- Calculate Order totals or Trade PnL.
- Invent internal or venue identity.
- Merge distinct partial Fills.
- Change side, quantity, or price after creation.

## Identity

`fill_id` is the internal datastore identity.

Venue Fill identity is the external deduplication key.

Hyperliquid commonly exposes that identity as `tid`.

Exact scoped uniqueness MUST include network, account, symbol, source time, and venue Fill identity.

## Logical Relationships

| Relation | Cardinality | Identity | Lifetime | Writer |
|---|---:|---|---|---|
| Order owns Fill | 1 to 0..many | Venue Fill identity; optional internal `fill_id` | Order | Ledger through Fill methods |
| Fill references Trade | Many to 1 | `trade_id` | Fill | Immutable ownership |
| Fill references Ledger | Many to 1 | Ledger identity | Fill | Immutable ownership |

This table defines logical design. It does not define a physical schema.

## Invariants

- Same venue Fill identity with changed execution facts MUST fail.
- Same execution facts with older or equal evidence MUST remain idempotent.
- Different venue Fill identities MUST remain different Fills.
- Source timestamp MUST retain Venue or Simulator meaning.
- Late enrichment MUST NOT alter execution identity.

## Reference Evidence

Canonical:

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\ownership.md
```

Supplemental:

```text
D:\rust\nuubot3\wiki\account\fill.md
D:\rust\nuubot3\wiki\account\ledger.md
D:\rust\nuutrader6\src\nuubot\hcbots\recon.py
D:\rust\nuutrader6\src\nuubot\hcbots\ledger\ledger.py
```

## Conflict

Nuutrader6 uses `fill_key` around its current Position journal. Nuubot5 retains the Fill identity contract without copying that schema.

## Recommendation

Select the exact Hyperliquid Fill key only after the approved Venue response model is fixed.
