# Macross Signaler

Status: Implemented.
Covers: `internal/signaler/macross.go`
Purpose: Create regime-filtered EMA crossover Signals from closed Bars.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/signaler/macross.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/signaler.md`

## Scope

Macross owns its timeframes, EMA periods, closed-bar alignment, and Signal interpretation.

## Owner and Children

Signaler owns one Macross calculator when selected.

Macross owns no lifecycle child.

## Responsibilities

- Request signal OHLCV with slow-EMA warmup.
- Request regime OHLCV with regime-EMA warmup.
- Calculate fast, slow, and regime EMAs.
- Backward-align only closed regime values.
- Detect confirmed crossover direction.
- Apply matching regime filter.
- Return ordered Signals.

## Does Not

- Load Bars.
- Release Signals.
- Use an unclosed regime Bar.
- Own lifecycle or replay state.
- Place orders.

## Lifecycle

Construct once, report requirements, calculate once, then remain immutable.

## Inputs and Outputs

Inputs are validated signal and regime OHLCV plus configured EMA periods.

Output is ordered `[]Signal`.

## State and Invariants

Signal and regime timeframes MUST differ.

Long requires upward crossover and close above the latest closed regime EMA.

Short requires downward crossover and close below the latest closed regime EMA.

Signals begin after configured warmup.

## Concurrency

Calculation is synchronous.

## Persistence

None.

## Errors

Unknown intervals, equal intervals, or missing required OHLCV fail.

## Program Flow

```text
Requirements
  request signal timeframe warmup
  request regime timeframe warmup

Calculate
  find required Bars
  calculate three EMA series
  align latest closed regime EMA
  scan signal OHLCV after warmup
  detect crossover
  apply regime filter
  append Signal
```

## Required Proof

- Regime values never look ahead.
- Long and short conditions match canon.
- Missing timeframe fails.
- Accepted dataset produces the canonical Signal sequence.

## Open Decisions

None.
