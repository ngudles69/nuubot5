# Ledger Package

Status: Reserved.
Covers: `internal/ledger/doc.go`
Purpose: Hold one Account's coherent local Trades, Orders, and Fills.

## Canonical Source

- `D:/rust/nuubot3/wiki/account/ledger.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/ledger/ledger.py`

## Scope & Responsibilities

Ledger owns the local trading domain tree for one Account.

- Create and retain Trades, Orders, and Fills.
- Reconcile validated Venue responses.
- Return one coherent Account snapshot.

## Program Flow

```text
Ledger(log, ctx, account)

init
  state = load(account)

start
  return ready

run(order_requests)
  trade  = Trade(order_requests)
  orders = trade.orders
  save(trade, orders)
  return trade

run(venue_response)
  apply fills
  apply order state
  refresh affected trades
  save changes
  return snapshot

stop
  save pending changes

---

domain
  Ledger
    Trade
      Order
        Fill
```

## Notes

- Ledger never queries Venue. Account supplies normalized responses.
