# OHLCV Package

Status: Implemented.
Covers: `internal/ohlcv/ohlcv.go`
Purpose: Load one validated OHLCV range for one symbol interval.

## Canonical Source

- `D:/rust/nuubot4/src/bars.rs`

## Scope & Responsibilities

OHLCV owns Parquet resolution, decoding, and validation.

- `Load(source, interval, start, end)` returns the complete requested range.
- `Open(source, interval, start, end)` streams the same validated rows.
- `Load` consumes `Open`; both use one decoder and validation path.
- Every interval uses one row shape and one validation path.
- Missing, malformed, invalid, or discontinuous data fails.

## Program Flow

```text
Open(source, interval, start, end) -> Reader

init
  files = monthly files covering start..end

run
  stream start, open, high, low, close, and volume
  validate each OHLCV row
  return one row

stop
  verify the complete range
  close reader and file

Load(source, interval, start, end) -> rows

init
  reader = Open(source, interval, start, end)

run
  append every streamed row

stop
  close reader
  return rows
```

## Notes

- OHLCV owns no replay policy, strategy warmup, or trading lifecycle.
- Bar closure is proven by the next row's start.
- End timestamps are neither decoded nor stored.
- Current intervals are `1s`, `1h`, and `4h`.
