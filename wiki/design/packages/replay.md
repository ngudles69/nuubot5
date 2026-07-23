# Replay Package

Status: Implemented.
Covers: `internal/replay/*.go`
Purpose: Stream one validated BBO sequence from monthly Parquet files.

## Canonical Sources

- Nuubot4 boundary: `D:/rust/nuubot4/src/replay.rs`
- Nuubot4 Parquet reader: `D:/rust/nuubot4/src/replay/parquet.rs`

## Scope

Reader owns Parquet file sequencing, Arrow batches, row admission, resource release, and reader statistics.

## Owner and Children

BtRunner owns one Reader.

Reader owns one open Parquet file and record reader at a time.

## Responsibilities

- Resolve monthly input filenames.
- Require every selected file.
- Project `close_time_us` and `close`.
- Stream bounded rows without loading all ticks.
- Validate timestamp shape, sequence, and price.
- Convert close timestamps into exact next-second milliseconds.
- Release Arrow and file resources.

## Does Not

- Advance TickClock.
- Call Runtime.
- Load OHLCV bars.
- Repair, skip, or reorder malformed in-range rows.
- Write source data.

## Lifecycle

`NewReader` validates file presence and prepares range state.

Repeated `Next` returns one BBO, completion, or error.

`Stop` is idempotent and closes current resources.

## Inputs and Outputs

Inputs are tick directory, inclusive start, and exclusive end.

Each `Next` outputs one admitted `market.BBO`, completion flag, and error.

## State and Invariants

In-range close timestamps MUST end within microseconds `999000..999999`.

Normalized BBO timestamps MUST advance exactly one second.

Prices MUST be finite and positive.

Rows outside the requested range are ignored.

## Concurrency

Reader is synchronous and MUST have one caller.

## Persistence

Reader opens monthly Parquet files read-only.

## Errors

Missing files, schemas, types, nulls, unequal columns, decoder failures, invalid values, gaps, duplicates, and overflow fail the replay.

Validation MUST remain even when the decoder changes.

## Program Flow

```text
NewReader
  build monthly paths
  require regular files
  store range

Next
  read next batch when exhausted
  open next file when needed
  read timestamp and price
  skip rows outside range
  validate and normalize row
  verify exact sequence
  return BBO

Stop
  release record reader
  close file
  report statistics
```

## Required Proof

- Exact first, last, and count values match the replay range.
- Invalid timestamp fractions fail.
- Gaps and duplicate seconds fail.
- Invalid prices fail.
- `Stop` releases resources after success and failure.

## Open Decisions

No loader replacement is approved. DuckDB remains benchmark-only until separately approved.
