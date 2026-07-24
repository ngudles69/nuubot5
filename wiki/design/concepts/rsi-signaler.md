# RSI Signaler

Status: Implemented.
Covers: `internal/signaler/rsi.go`
Purpose: Produce volume-confirmed RSI Signal packages from closed Bars.

## Ownership

Signaler factory creates and initializes RSI.

RSI is a complete Signaler.

It owns configuration, requirements, OHLCV loading, calculation, packages, and cleanup.

## Responsibilities

- Load OHLCV with RSI and volume warmup.
- Calculate smoothed relative strength.
- Calculate volume moving average.
- Require current volume above its average.
- Trigger long at RSI 30 or below.
- Trigger short at RSI 70 or above.
- Suppress repeated identical sides.
- Produce one package for every admitted signal bar.
- Serve timestamp-bounded package history.

## Package Fields

Standard fields include entry triggers and neutral regime.

Custom fields include:

- `bar_start_ms`.
- `signal_price`.
- `rsi`.
- `volume_ratio`.
- `oversold`.
- `overbought`.

## Program Flow

```text
Init
  configure rsi
  load rsi data
  calculate rsi signals
  validate packages
  initialize signaler

Calculate
  find rows
  calculate indicators
  calculate signals

configure
  parse interval
  validate config
```

## Invariants

Indicator decisions use only the current closed Bar and earlier Bars.

Repeated identical sides remain suppressed until state changes.

Package time uses the next admitted signal-bar start.

## Required Proof

- Threshold boundaries include 30 and 70.
- Low volume blocks entry triggers.
- Every admitted signal bar produces one package.
- Flat JSON contains standard and custom fields.
