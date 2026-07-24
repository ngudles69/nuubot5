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

  resolve interval
  validate range
  resolve files
  validate files
  create reader

Load(source, interval, start, end) -> rows

  open reader
  load rows
  close reader

Next
  read batch
  filter range
  admit row

Close
  close reader
```

## Notes

- OHLCV owns no replay policy, strategy warmup, or trading lifecycle.
- `Open`, `Load`, `Next`, and `Close` are precise streaming domain operations.
- Bar closure is proven by the next row's start.
- End timestamps are neither decoded nor stored.
- Full Parquet row groups contain 122,880 rows; the final group may be shorter.
- Arrow batches use 122,880 rows so each full row group fits one batch.
- This avoids splitting normal groups and reduced allocation and garbage collection.
- Current intervals are `1s`, `1h`, and `4h`.

Accepted 500x proof:

- Passed 500/500 with zero errors.
- Replay averaged 1,134 ms.
- Allocation averaged 975.697 MB, down 26.1 percent.
- Tests and structural checks passed.
- `wiki/PERFORMANCE.md` owns detailed benchmark history.
- Heap end snapshot rose 11.1 percent; this is not peak memory.
