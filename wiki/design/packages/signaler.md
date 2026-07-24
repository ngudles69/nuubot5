# Signaler Package

Status: Implemented.
Covers: `internal/signaler/*.go`
Purpose: Initialize and run one Signaler with one configured calculator.

## Canonical Source

- `D:/rust/nuubot4/src/signaler.rs`

## Scope & Responsibilities

Runtime owns one concrete Signaler.

- Signaler owns calculator selection, OHLCV requirements, loading, and Signal calculation.
- Macross and RSI are calculator implementations, not separate Signalers.
- New strategies add one calculator file and one selection case.

## Program Flow

```text
init
  select calculator
  resolve requirements
  load ohlcv
  calculate signals
  validate signals
  initialize signaler

start
  start signaler

run
  release signal

stop
  stop signaler

---

domain
  macross = 1h EMA 9/21 cross filtered by closed 4h EMA 200 regime
```

## Notes

- Current configured calculator is Macross.
- Signaler Init loads required OHLCV, calculates Signals, and validates them.
- Backtest calculates wholesale but releases only after the next row starts.
- Live Signaler is approved but unimplemented.
- Live initialization loads warmup plus buffer and generates only its required tail.
- Live ingestion appends closed rows, prunes equally, recalculates the frame, and generates only its required tail.
- Each concrete Signaler owns its generation tail length.
