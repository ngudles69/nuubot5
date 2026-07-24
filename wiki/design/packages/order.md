# Order Package

Status: Reserved. Proposed next-tranche design.
Covers: `internal/order/doc.go`
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

`filled`, `canceled`, `rejected`, `expired`, and `error` are terminal.

Terminal Orders never return active.

## Immutable Fields

- Ledger, Trade, cycle, and local Order identity.
- Batch number and position.
- CLOID.
- Symbol, role, side, type, and time-in-force.
- Requested quantity, requested price, and trigger price.
- Reduce-only flag and submission timestamp.

## Mutable Fields

- Venue Order identity.
- Status, active flag, and rejection reason.
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

Snapshot
  return immutable Order values
```

These are domain operations, not lifecycle phases.

## Fill Aggregation

Filled quantity is the sum of unique owned Fills.

Average Fill price is quantity-weighted.

Fees are summed once.

Remaining quantity never becomes negative.

One Venue TID may enrich metadata but cannot change side, quantity, or price.

## Persistence

See [Trading Schema](../concepts/trading-schema.md).

## Required Proof

- Invalid state transitions fail.
- Duplicate identical Fills are idempotent.
- Changed execution for one Venue TID fails.
- Partial and complete Fill totals are correct.
- Terminal Orders remain terminal.
- Request fields never change during reconciliation.
