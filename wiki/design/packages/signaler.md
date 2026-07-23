# Signaler Package

Status: Implemented.
Covers: `internal/signaler/*.go`
Purpose: Create and run one configured Signal strategy behind a stable factory.

## Canonical Source

- `D:/rust/nuubot4/src/signaler.rs`

## Scope & Responsibilities

SignalerFactory selects one concrete Signaler. Runtime knows only the common
Signaler contract.

- Concrete Signalers own configuration, OHLCV requirements, and Signal calculation.
- New strategies add one concrete file and one factory case.

## Program Flow

```text
SignalerFactory(kind, log, ctx) -> Signaler

init
  concrete = select kind
  rows       = load complete Backtest range
  indicators = concrete.calculate(rows)
  candidates = concrete.generate(rows, all)

start
  no-op

run(now)
  when now crosses candidate next-row start
    set candidate availability to now
    return candidate

stop
  no-op

---

domain
  macross = 1h EMA 9/21 cross filtered by closed 4h EMA 200 regime
```

## Notes

- Current configured Signaler is Macross.
- Backtest calculates wholesale but releases only after the next row starts.
- Live Signaler is approved but unimplemented.
- Live initialization loads warmup plus buffer and generates only its required tail.
- Live ingestion appends closed rows, prunes equally, recalculates the frame, and generates only its required tail.
- Each concrete Signaler owns its generation tail length.
