# Macross Signaler

Status: Implemented.
Covers: `internal/signaler/macross.go`
Purpose: Produce regime-filtered EMA Signal packages from closed Bars.

## Ownership

Signaler factory creates and initializes Macross.

Macross is a complete Signaler.

It owns configuration, requirements, OHLCV loading, calculation, packages, and cleanup.

## Responsibilities

- Load signal OHLCV with slow-EMA warmup.
- Load regime OHLCV with regime-EMA warmup.
- Calculate fast, slow, and regime EMAs.
- Backward-align only closed regime values.
- Detect confirmed crossover direction.
- Apply matching regime filter.
- Produce one package for every admitted signal bar.
- Serve timestamp-bounded package history.

## Package Fields

Standard fields include entry triggers and regime.

Custom fields include:

- `bar_start_ms`.
- `signal_price`.
- `fast_ma`.
- `slow_ma`.
- `regime_ma`.

## Program Flow

```text
Init
  configure macross
  load macross data
  calculate macross signals
  validate packages
  initialize signaler

Calculate
  find rows
  calculate emas
  align regime
  calculate signals

configure
  parse intervals
  validate config
```

## Invariants

Signal and regime timeframes MUST differ.

Long requires upward crossover and bullish closed regime.

Short requires downward crossover and bearish closed regime.

Package time uses the next admitted signal-bar start.

## Required Proof

- Regime values never look ahead.
- Every admitted signal bar produces one package.
- Crossover entry triggers match the accepted baseline.
- Flat JSON contains standard and custom fields.
