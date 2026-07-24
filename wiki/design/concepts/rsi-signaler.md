# RSI Signaler

Status: Implemented.
Covers: `internal/signaler/rsi.go`
Purpose: Create volume-confirmed RSI Signals from closed Bars.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/signaler/rsi.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/signaler.md`

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

Inputs are validated OHLCV and configured RSI and volume periods.

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

Unknown intervals or missing required OHLCV fail.

## Program Flow

```text
createRSI
  parse interval

Requirements
  create requirements

Calculate
  find rows
  calculate indicators
  calculate signals
```

## Required Proof

- Threshold boundaries include 30 and 70.
- Low volume blocks Signals.
- Repeated identical sides are suppressed.
- Missing timeframe fails.

## Open Decisions

None.
