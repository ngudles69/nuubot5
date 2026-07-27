# Order Package

Status: Implemented for request identity, Venue state, and Fill aggregation.
Covers: `internal/order/*.go`
Purpose: Represent one immutable submitted request and its changing Venue lifecycle.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/account/order.py`
- `D:/rust/nuubot3/wiki/account/order.md`

## Ownership

Trade owns Order.

Order owns Fills.

Order never queries Account, Venue, Simulator, or storage.

## Domain Shape

Order is a mutable domain object.

It is not a lifecycle component.

Order owns one transient mutation revision.

Canonical Recon reads that scalar before and after mutation.

Identical duplicate evidence leaves the revision unchanged.

It has no `Init`, `Start`, `Run`, or `Stop`.

Request identity and terms remain immutable.

Venue state and Fill aggregates may advance.

## States

```text
created -> submitted -> open -> partially_filled -> filled
created -> rejected
submitted/open/partially_filled -> canceled
submitted/open/partially_filled -> expired
created/submitted/open/partially_filled -> error
```

`canceled`, `rejected`, `expired`, and `error` are locally terminal.

Venue `filled` remains reconciliation-pending and active until quantity and every Fill fee are complete.

## Immutable Fields

- Ledger, Trade, cycle, and local Order identity.
- Batch number and position.
- CLOID.
- Symbol, role, side, type, and time-in-force.
- Requested quantity, requested price, and trigger price.
- Reduce-only flag and submission timestamp.

## Mutable Fields

- Venue Order identity.
- Status, active flag, reconciliation-pending flag, and rejection reason.
- Venue update timestamp.
- Filled quantity, remaining quantity, average Fill price, and fees.
- Raw Venue evidence.

## Operations

```text
New
  keep admitted request
  initialize created state

ApplyVenueState
  reject invalid transition
  preserve Venue identity
  advance lifecycle
  preserve raw evidence

ApplyFill
  validate Fill ownership
  reject changed execution
  add or enrich Fill
  refresh Fill totals

RefreshRecon
  keep fee-incomplete or Fill-incomplete acknowledgements pending
  complete only quantity-exact fee-complete Orders

ComparisonState
  return the allocation-free mutation revision

FillIdentity
  return allocation-free immutable ownership for one Fill update

Record
  return one flat Order database value

EachFill
  visit directly owned Fills without copying them
```

These are domain operations, not lifecycle phases.

## Fill Aggregation

Filled quantity is the sum of unique owned Fills.

Average Fill price is quantity-weighted.

Fees are summed once.

Missing fees keep the Order reconciliation-pending. Zero and negative fees remain complete evidence.

Remaining quantity never becomes negative.

One Venue TID may enrich metadata but cannot change side, quantity, or price.

## Persistence

See [Trading Schema](../concepts/trading-schema.md).

One Order row contains no Fill collection. Fills remain separate rows linked by `order_id`.

## Required Proof

- Invalid state transitions fail.
- Duplicate identical Fills are idempotent.
- Identical duplicate Order and Fill evidence leaves comparison state unchanged.
- New Fill, fee enrichment, and terminal transitions advance comparison state.
- Comparison and Fill ownership reads allocate nothing.
- Changed execution for one Venue TID fails.
- Partial and complete Fill totals are correct.
- Fee-incomplete and Fill-incomplete acknowledgements remain reconciliation-pending.
- Locally terminal Orders remain terminal.
- Request fields never change during reconciliation.
