# Replay Package

Status: Implemented.
Covers: `internal/replay/*.go`
Purpose: Stream one validated BBO sequence through OHLCV.

## Canonical Source

- Nuubot4 boundary: `D:/rust/nuubot4/src/replay.rs`
- Nuubot4 Parquet reader: `D:/rust/nuubot4/src/replay/parquet.rs`

## Scope & Responsibilities

Replay owns BBO conversion and replay statistics.

- Open one streaming `1s` OHLCV reader.
- Return one ordered BBO per `Next`.
- Never decode Parquet directly.

## Program Flow

```text
Reader(source, start, end)

init
  rows = ohlcv.Open(source, "1s", start, end)

Next
  row = rows.Next()
  convert row close into BBO
  return BBO

Stop
  close rows
  report statistics
```

## Notes

- Replay streams; Signalers retain complete ranges through `ohlcv.Load`.
- Reader is synchronous and has one caller.
