# Order Package

Status: Reserved.
Covers: `internal/order/doc.go`
Purpose: Represent one submitted order leg and its owned Fills.

## Canonical Source

- `D:/rust/nuubot3/wiki/account/order.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/ledger/ledger.py`

## Scope & Responsibilities

Order keeps one immutable request and its current Venue lifecycle.

- Own all Fills for one submitted leg.
- Aggregate filled quantity, price, and fees.
- Preserve local and Venue identity.

## Program Flow

```text
Order(identity, request)

init
  keep identity and request

start
  status = created

run(venue_order)
  update lifecycle

run(fill)
  add fill
  refresh totals
  return snapshot

stop
  keep terminal state

---

domain
  Trade owns Order
  Order owns Fills
```

## Notes

- Account submits Orders. Ledger applies validated Venue responses.
- Request identity and terms remain fixed after creation.
