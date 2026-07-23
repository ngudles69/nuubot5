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

stop
  release simulated state
```

## Notes

- Matching, fills, latency, fees, order books, and detailed mechanics remain intentionally undefined.
