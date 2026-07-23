# Order

**Status:** Approved — unimplemented in Nuubot5.

## Purpose

Represent one submitted order leg and its current lifecycle, request facts, and owned Fills.

## Scope

One submitted leg equals one Order.

A batch or bracket is not an Order.

## Responsibilities

- Own Fills for one submitted leg.
- Preserve immutable request and identity fields.
- Store one current lifecycle status.
- Store one active flag derived during status updates.
- Aggregate quantity, average fill price, fees, and last fill time.
- Accept validated lifecycle updates through Ledger recon.

## Does Not

- Submit itself.
- Own sibling Orders.
- Calculate Trade PnL.
- Generate its own CLOID.
- Change immutable request fields after creation.
- Reopen from terminal state through normal recon.

## Identity

`order_id` is the internal datastore identity.

`cloid` is the client Order identity sent to Venue.

Each CLOID MUST identify one Order.

Account creates CLOID. Strategy code MUST NOT supply it.

`batch_no` identifies one Trade submission batch.

`order_pos` identifies one request position inside that batch.

## Logical Relationships

| Relation | Cardinality | Identity | Lifetime | Writer |
|---|---:|---|---|---|
| Trade owns Order | 1 to 1..many | `order_id` and CLOID | Trade | Ledger through Order methods |
| Order owns Fill | 1 to 0..many | Venue Fill identity | Order | Ledger through Order methods |
| Order belongs to batch | Many to 1 logical grouping | `trade_no`, `batch_no` | Submission | Account assigns |

This table defines logical design. It does not define a physical schema.

## Mutable State

Lifecycle fields may include venue Order identity, status, active state, rejection reason, update time, and raw evidence.

Fill-derived totals MUST come from owned Fills, not Order-history totals.

Exact Go fields remain implementation-scoped.

## Invariants

- Every Order MUST belong to one Trade before submission.
- Request identity and terms MUST remain fixed after creation.
- Venue truth wins for validated lifecycle state.
- Different partial Fills MUST remain distinct.
- Duplicate CLOID within the Order store MUST fail.
- Terminal Order state MUST NOT become active through normal recon.

## Reference Evidence

Canonical:

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\cloid.md
```

Supplemental:

```text
D:\rust\nuubot3\wiki\account\order.md
D:\rust\nuubot3\wiki\account\account.md
D:\rust\nuutrader6\src\nuubot\hcbots\exchange.py
D:\rust\nuutrader6\src\nuubot\hcbots\ledger\ledger.py
```

## Conflict

Nuutrader6 uses different Order and Position naming. Nuubot5 keeps the Nuubot3 Trade-to-Order domain relationship.

## Recommendation

Implement only statuses proven by the selected Venue responses and required Executor behavior.
