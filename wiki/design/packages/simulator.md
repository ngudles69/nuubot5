# Simulator Package

Status: Reserved.
Covers: `internal/simulator/doc.go`
Purpose: Simulate responses from the selected trading SDK.

## Canonical Source

- `D:/rust/nuubot3/wiki/account/simulator.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/simulator.py`

## Scope & Responsibilities

Simulator provides the SDK-facing behavior Account expects without contacting a live Venue.

- Accept the same requests as the selected SDK boundary.
- Return equivalent response shapes.
- Keep simulated Venue truth separate from Ledger truth.
- Match eligible existing Orders when `IngestBBO` advances simulated Venue time.

## Program Flow

```text
Simulator(log, ctx, config)

init
  state = simulated venue state

start
  return ready

run(request)
  response = simulate SDK response
  return response

IngestBBO
  accept one validated BBO
  match eligible existing Orders
  record simulated Order and Fill outcomes
  report whether Account state became dirty

stop
  release simulated state
```

## Notes

- [`IngestBBO`](../concepts/ingestbbo.md) is the only BBO route that advances Simulator matching.
- `OnBBO` never advances Simulator or creates simulated Fills.
- Exact matching, latency, fees, order books, and partial-fill mechanics remain intentionally undefined.
