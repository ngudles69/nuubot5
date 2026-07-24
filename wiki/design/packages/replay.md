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
init
  open ohlcv
  initialize reader

Next
  read next ohlcv
  create bbo
  record proof

stop
  close ohlcv
  report proof
```

## Notes

- Replay streams; Signalers retain complete ranges through `ohlcv.Load`.
- Reader is synchronous and has one caller.
- `Next` remains a domain helper because it advances one streaming reader.
