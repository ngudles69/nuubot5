# Replay Process

Status: Implemented.
Covers: `internal/btrunner/btrunner.go`, `internal/replay/*.go`, `internal/runtime/runtime.go`
Purpose: Drive one exact historical market sequence through the canonical Runtime path.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/btrunner.rs`
- Nuubot4: `D:/rust/nuubot4/src/replay.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/btrunner.md`

## Participants

- BtRunner owns orchestration and proof.
- OHLCV owns Parquet decoding and row admission.
- Replay Reader owns BBO conversion and iteration.
- TickClock owns pass timing.
- Runtime owns Bot decisions and stop.

## Preconditions

- Configuration and Bot specification are valid.
- Market paths remain inside shared data.
- Required OHLCV files exist.
- Runtime OHLCV and Signals are prepared.
- BtRunner and Runtime are started.
- Replay starts at `BotSpec.ReplayStart`.

`BotSpec.StartAt` may exist, but current BtRunner intentionally ignores it.

## Ordered Flow

```text
start Reader and OHLCV at ReplayStart
open one streaming 1s OHLCV reader
read one BBO
send BBO to Runtime
record tick evidence
advance TickClock
when due
  run one Runtime Pass
when Runtime requests stop
  end replay
when Reader completes
  stop Runtime with end_date
verify count, passes, and range
```

## Decisions

Reader decides whether decoded data is admissible.

TickClock decides when a Runtime pass is due.

Runtime decides Signal acceptance, cycle lifecycle, and graceful stop.

BtRunner decides whether replay evidence is exact.

## State Changes

Reader advances file, batch, and row positions.

BtRunner records served ticks, triggered passes, and replay boundaries.

Runtime updates Signals, cycles, and stop state.

`StartAt` causes no replay state change.

## Failure Handling

Any decode, validation, child, lifecycle, or verification error terminates `Run`.

The command still calls BtRunner `Stop` after a successful `Start`.

Malformed data MUST NOT be skipped, repaired, or accepted.

## Recovery

Replay has no in-process retry.

Restart creates a fresh process and replays from the beginning.

## Completion

Completion requires exact expected ticks, passes, first timestamp, last timestamp, successful child teardown, and semantic terminal statistics.

Exit zero alone is insufficient.

## Does Not

- Persist checkpoints.
- Resume partial replay.
- Model exchange execution.
- Read live feeds.
- Bypass Runtime.

## Required Proof

- Sweep 6 Bot 9 serves 7,948,800 ticks.
- TickClock triggers 794,880 passes.
- Runtime produces 55 Signals and 18 closed cycles.
- Reader and BtRunner ranges match expected boundaries.
- Canonical `noasm` stability gate passes.

## Open Decisions

No alternate replay loader is approved.
