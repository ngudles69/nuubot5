# Fill Package

Status: Reserved.
Covers: `internal/fill/doc.go`
Purpose: Record one actual Venue or Simulator execution for one Order.

## Canonical Source

- `D:/rust/nuubot3/wiki/account/fill.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/simulator.py`

## Scope & Responsibilities

Fill preserves one execution as immutable trading evidence.

- Keep Venue identity, side, quantity, price, fee, and time.
- Accept later metadata without changing execution facts.

## Program Flow

```text
Fill(order, execution)

init
  keep order identity and execution

start
  attach to Order

run(metadata)
  enrich missing metadata
  return fill

stop
  keep immutable execution

---

domain
  one partial execution = one Fill
```

## Notes

- Order owns Fill. Ledger creates or updates it from validated Venue evidence.
- Fill does not calculate Order totals or Trade PnL.
