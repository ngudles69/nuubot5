# Signaler

## Covers

- `internal/ohlcv/ohlcv.go`
- `internal/signaler/signaler.go`
- `internal/signaler/macross.go`
- `internal/signaler/rsi.go`
- `internal/runtime/runtime.go`

## Intent

Signaler MUST remain passive.

It calculates immutable flat packages and serves timestamp-bounded history.

## Ownership

Runtime MUST own one factory-selected Signaler.

`signaler.Create` MUST select Macross or RSI and complete initialization.

Macross and RSI are complete Signalers.

They are not calculators behind another Signaler.

Each concrete Signaler owns:

- Configuration and requirements.
- Replay OHLCV loading.
- Calculations and custom fields.
- Ordered package history.
- Cleanup.

## Runtime Contract

```text
Runtime.Init
  create initialized Signaler

Runtime.Run
  request latest available package
  consume unseen timestamp
  inspect standard entry triggers

Runtime.Stop
  stop Signaler
```

No `Start`, `Run`, or `Release` method exists on Signaler.

## Data Contract

Every package contains standard fields.

Concrete fields remain flat beside those fields.

Package time MUST use the next admitted bar start.

Queries MUST return no future package.

Returned history MUST remain chronological.

## Implementations

Macross MUST calculate:

- Signal and regime EMAs.
- Closed-regime alignment.
- Regime-filtered crossover triggers.
- One package per signal bar.

RSI MUST calculate:

- Smoothed relative strength.
- Volume moving-average confirmation.
- Threshold transition triggers.
- One package per signal bar.

## Evidence

Signaler MUST report kind, symbol, timeframes, loaded rows, and package count.

Runtime MUST separately report packages read and skipped entry triggers.
