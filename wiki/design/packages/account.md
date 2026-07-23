# Account Package

Status: Reserved.
Covers: `internal/account/doc.go`
Purpose: Give one Executor a trading boundary backed by one Venue and one local Ledger.

## Canonical Source

- `D:/rust/nuubot3/wiki/account/account.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/account.py`

## Scope & Responsibilities

Account owns one Venue connection and one Ledger for one configured account.

- Submit and cancel requests through Venue.
- Reconcile Venue responses into Ledger.
- Hide live, testnet, and simulated Venue differences from Executor.

## Program Flow

```text
Account(log, ctx, config)

init
  venue  = Venue(config)
  ledger = Ledger(log, ctx, config)

start
  venue.start()
  ledger.start()

run(request)
  response = venue.submit(request)
  ledger.run(response)
  return response

run(recon)
  response = venue.fetch()
  snapshot = ledger.run(response)
  return snapshot

stop
  ledger.stop()
  venue.stop()

---

domain
  Executor owns Account
  Account owns Venue and Ledger
  Venue is external truth
  Ledger is reconciled local truth
```

## Notes

- Account does not calculate Trade PnL or mutate Ledger-owned children directly.
