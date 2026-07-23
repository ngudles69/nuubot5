# Trade Package

Status: Reserved.
Covers: `internal/trade/doc.go`
Purpose: Represent one trading intent from entry through final exit.

## Canonical Source

- `D:/rust/nuubot3/wiki/account/trade.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/ledger/ledger.py`

## Scope & Responsibilities

Trade owns the Orders belonging to one trading intent.

- Aggregate Order execution and fees.
- Calculate current and final PnL.
- Derive lifecycle from owned Orders.

## Program Flow

```text
Trade(identity, order_requests)

init
  orders = Orders(order_requests)

start
  status = pending

run(order_updates)
  refresh orders
  refresh exposure, fees, pnl, status
  return snapshot

stop
  finalize metrics

---

domain
  pending -> open -> closing -> closed
  pending -> canceled
```

## Notes

- Ledger owns Trade. Trade owns Orders.
- Finalized Trade values do not reopen through normal reconciliation.
