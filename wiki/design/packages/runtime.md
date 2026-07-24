# Runtime Package

Status: Implemented.
Covers: `internal/runtime/runtime.go`
Purpose: Own and sequence one Bot's control components.

## Canonical Source

- `D:/rust/nuubot4/src/runtime.rs`

## Scope & Responsibilities

Runtime owns one Signaler, factory-created Risks, and at most one BotCycle.

- Release due Signals before delivering each BBO.
- Route each BBO through the active BotCycle's Simulator-only ingestion path.
- Assess Risks before the active BotCycle.
- Own graceful stop decisions.

## Program Flow

```text
Runtime

init
  initialize signaler
  create risks
  initialize runtime

start
  start signaler
  start runtime

IngestBBO
  run signaler
  initialize botcycle
  start botcycle
  ingest botcycle bbo
  deliver botcycle bbo

run
  assess risk stops
  check stop request
  run botcycle
  close botcycle
  check max cycles

stop
  request stop
  stop botcycle
  stop risks
  stop signaler
  stop runtime
```

## Notes

- Runtime never selects Signaler calculators, Risks, or Executors.
- Runtime does not know Signaler requirements, OHLCV, calculation, or validation.
- Runtime does not use replay dates to stop; BtRunner owns Reader exhaustion.
- `IngestBBO` remains a domain helper because it accepts one BBO event.
- Runtime initialization receives the Setup context from BtRunner.

## IngestBBO

[`IngestBBO`](../concepts/ingestbbo.md) is separate from normal Executor
`OnBBO` policy.

```text
Runtime.IngestBBO
  call active BotCycle.IngestBBO
```

Runtime does not check whether Simulator is selected. The selected Venue owns
the live no-op or simulated matching behavior.
