# RSI Signaler

## Purpose

Create volume-confirmed RSI Signals from closed Bars.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/signaler/rsi.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/signaler.md`
- Nuubot5: `internal/signaler/rsi.go`

## Scope

RSI owns its timeframe, RSI period, volume period, indicators, and Signal interpretation.

## Owner and Children

Signaler owns one RSI calculator when selected.

RSI owns no lifecycle child.

## Responsibilities

- Request sufficient RSI and volume warmup.
- Calculate smoothed relative strength.
- Calculate volume moving average.
- Require current volume above its average.
- Emit long at RSI 30 or below.
- Emit short at RSI 70 or above.
- Suppress repeated identical sides.

## Does Not

- Load Bars.
- Release Signals.
- Own replay state.
- Place orders.

## Lifecycle

Construct once, report requirements, calculate once, then remain immutable.

## Inputs and Outputs

Inputs are validated Bars and configured RSI and volume periods.

Output is ordered `[]Signal`.

## State and Invariants

Indicator decisions use only the current closed Bar and earlier Bars.

Signals begin after warmup.

Repeated identical sides remain suppressed until the state changes.

## Concurrency

Calculation is synchronous.

## Persistence

None.

## Errors

Unknown timeframes or missing required Bars fail.

## Program Flow

```text
BarsNeeded
  request maximum indicator warmup

Calculate
  find required Bars
  calculate RSI
  calculate volume average
  scan Bars after warmup
  apply volume confirmation
  select threshold side
  suppress repeated side
  append Signal
```

## Required Proof

- Threshold boundaries include 30 and 70.
- Low volume blocks Signals.
- Repeated identical sides are suppressed.
- Missing timeframe fails.

## Open Decisions

None.
