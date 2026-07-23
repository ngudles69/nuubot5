# Bar

## Purpose

Carry validated closed OHLCV history for one Signaler timeframe.

## Status

Implemented.

## Canonical Sources

- Nuubot4 market value: `D:/rust/nuubot4/src/market.rs`
- Nuubot4 loader: `D:/rust/nuubot4/src/bars.rs`
- Nuubot5: `internal/bars/bars.go`

## Scope

`bars.Data` holds aligned timestamp and OHLCV columns plus one warmup boundary.

## Owner and Children

BtRunner loads Bars requested by Runtime.

Signaler retains loaded Bars during replay.

Bar data owns no lifecycle child.

## Responsibilities

- Identify timeframe.
- Preserve requested warmup count.
- Preserve start, end, open, high, low, close, and volume columns.
- Load complete monthly Parquet ranges.
- Validate schema, values, duration, sequence, and exact range.

## Does Not

- Select timeframes independently.
- Calculate indicators.
- Release Signals.
- Stream replay ticks.
- Repair missing or invalid rows.

## Lifecycle

Load once before Runtime starts, retain through Signaler lifetime, then release.

## Inputs and Outputs

Inputs are tick path, replay bounds, and Signaler requirements.

Output is validated `[]bars.Data`.

## State and Invariants

All columns MUST have one row per closed Bar.

Bar end MUST equal start plus timeframe duration.

Open, high, low, and close MUST be finite and positive.

Volume MUST be finite and non-negative.

High MUST cover open and close. Low MUST not exceed open or close.

Start timestamps MUST form an exact timeframe sequence.

Loaded range MUST include requested warmup and end exactly.

## Concurrency

Loading and indicator consumption are synchronous.

## Persistence

Bars read monthly Parquet files and write nothing.

## Errors

Missing files or columns, wrong types, nulls, invalid values, gaps, range mismatch, and decoder failures fail preparation.

## Program Flow

```text
Load
  for each Signaler requirement
    calculate warmup start
    select monthly files
    read required columns
    validate and append in-range rows
    verify exact count and boundaries
  return Bars
```

## Required Proof

- Warmup and replay boundaries match exactly.
- Invalid OHLCV fails.
- Gaps, duplicates, nulls, and missing columns fail.
- Macross receives both requested timeframes.
- RSI receives its requested timeframe.

## Open Decisions

None.
