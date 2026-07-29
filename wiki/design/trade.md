# Trade

Status: Canonical application business logic.
Purpose: Define one Trade and its complete Order and Fill relationships.

## Identity

Every flat Trade, Order, and Fill record stores its complete parent identity.

```text
SweepID
└── BotID
    └── CycleNumber
        └── Account
            └── LedgerID
                ├── TradeID
                ├── OrderID
                └── FillID
```

The current Account identity is the configured Account name.

Inside one live Ledger, maps use local IDs because the Ledger supplies the
parent scope.

One Trade has one `TradeID`.

Every Order created for that Trade carries the same `TradeID`.

Every Order has its own `OrderID`.

Every Fill has its own Venue TID and references its parent `OrderID` and
`TradeID`.

Every Fill also has one Nuubot `FillID`.

Memory-only Ledger allocates `TradeID`, `OrderID`, and `FillID` from counters.

Venue TID remains the immutable Exchange Fill identity and deduplication key.

```text
TradeID
├── OrderID: entry 1
│   └── FillID / Venue TID
├── OrderID: entry 2
│   └── FillID / Venue TID
├── OrderID: entry 3
│   └── FillID / Venue TID
├── OrderID: exit 1
│   └── FillID / Venue TID
└── OrderID: exit 2
    └── FillID / Venue TID
```

Relationships use these canonical keys.

No synthetic relationship key replaces them.

## Order Sets

The application submits one set of Orders for one Trade.

It never submits Orders for multiple Trades in one request.

An Order set may contain entry, take-profit, and stop-loss Orders.

These roles are examples, not a fixed Order count or required shape.

One Trade may accumulate any number of Orders during its lifecycle.

The same role may appear on several Orders.

Every Order keeps its independently requested and Venue-rounded quantity.

Account validates minimum notional independently for every Order.

Account rejects an Order below minimum notional.

It never copies, equalizes, or silently increases quantities across an Order
set.

The Venue receives that set concurrently.

The set exists only for submission.

It is not a stored domain entity.

It has no batch number, batch key, or batch lifecycle.

`TradeID` groups every Order in the set.

Each `OrderID` identifies one Order.

CLOID is always the Venue-facing identity of one Order.

## Trade Evolution

Orders remain attached to the Trade after their state changes.

Later Orders append to the same Trade.

They do not replace earlier Orders.

## Exchange Snapshot Authority

Order and Fill records preserve synchronized Exchange facts.

Local submission and cancellation requests do not determine Exchange status.

Only synchronized Exchange Order evidence changes `Order.Status`.

Fill attachment never changes `Order.Status`.

Ledger may calculate Fill totals, exposure, and finance from accepted snapshots.

Derived calculations never rewrite Exchange identity, status, or payload
evidence.

A later Exchange snapshot may revise a previously closed Trade snapshot.

`Trade.IsClosed()` describes current derived state. It is not an irreversible
lifecycle lock.

Cancellation request metadata remains separate from synchronized Exchange
status.

Typical evolution:

```text
entry + take-profit + stop-loss

entry filled + take-profit + stop-loss

entry filled + take-profit filled + stop-loss canceled
```

Alternative stop-loss completion:

```text
entry filled + take-profit canceled + stop-loss filled
```

Explicit close completion:

```text
entry filled + take-profit canceled + stop-loss canceled + close

entry filled + take-profit canceled + stop-loss canceled + close filled
```

Partial close completion:

```text
entry filled + take-profit canceled + stop-loss canceled + half close

entry filled + take-profit canceled + stop-loss canceled + half close filled

entry filled + take-profit canceled + stop-loss canceled
+ half close filled + remaining close

entry filled + take-profit canceled + stop-loss canceled
+ half close filled + remaining close filled
```

DCA entry with staged exits:

```text
entry 1 + entry 2 + entry 3 + exit 1 + exit 2

entry 1 filled + entry 2 filled + entry 3 filled + exit 1 + exit 2

entry 1 filled + entry 2 filled + entry 3 filled
+ exit 1 filled + exit 2

entry 1 filled + entry 2 filled + entry 3 filled
+ exit 1 filled + exit 2 filled
```

Every Order shown above carries the same `TradeID`.

The Trade derives its current state and finance from all its Orders and Fills.

One Trade may append several close Orders.

Each close Order has its own `OrderID` and may have several Fills.

Filled entry quantities increase exposure.

Filled exit quantities reduce exposure.

Order role counts do not determine exposure.

Entry and exit Order counts need not match.

Individual entry and exit quantities need not match.

Aggregate signed Fill quantity determines current Trade size.

The Trade model imposes no fixed number of Orders or Fills.

It imposes no fixed count for entry, exit, take-profit, stop-loss, or close
Orders.

Venue request limits are transport constraints, not Trade identity or shape.

## Closure

A Trade closes when every Order is resolved and aggregate Trade size is zero.

Zero size does not close a Trade while any Order reports `IsClosed() == false`.

Exact Venue Order status is not the Trade closure contract.

Order exposes the derived `IsClosed()` contract.

Order owns the mapping from its detailed statuses and evidence into that
result.

Any attached Fill without fee evidence keeps `Order.IsClosed()` false.

A filled Order closes only after its full quantity and every Fill fee arrive.

Canceled, rejected, expired, or error Orders close only when every attached
Fill has fee evidence.

Trade never inspects Fill fee completeness directly.

Trade consumes only `Order.IsClosed()`.

This includes completed entry and exit Fills whose signed quantities net to
zero.

This also includes a Trade whose Orders all cancel before any Fill.

An all-canceled, zero-size Trade has `Status == Canceled` and
`IsClosed() == true`.

`Canceled` means no execution completed.

`Closed` means executed exposure returned to zero.

Both outcomes are closed meta states.

Canceled Trades do not count as completed round trips.

The number of entry, exit, take-profit, stop-loss, or close Orders is
irrelevant.
