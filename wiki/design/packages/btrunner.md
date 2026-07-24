# BtRunner Package

Status: Implemented.
Covers: `internal/btrunner/btrunner.go`
Purpose: Run one bounded historical Bot replay and prove exact completion.

## Canonical Source

- `D:/rust/nuubot4/src/btrunner.rs`

## Scope & Responsibilities

BtRunner owns Setup context, TickClock, TickReader, Runtime, and replay proof.

- Serve every admitted tick to Runtime.
- Register and own the Runtime timer callback.
- Stop timer and input admission before Runtime.

## Program Flow

```text
BtRunner

init
  prepare setup
  set replay range
  create clock
  initialize clock
  register runtime timer
  initialize replay reader
  initialize runtime
  create proof

start
  start clock
  start runtime

loop
  read replay
  ingest runtime bbo
  record proof
  advance clock
  check stop request
  verify replay

runtime_run(time)
  run runtime
  remember stop request

stop
  stop clock
  stop replay reader
  stop runtime
  report proof
  return stop errors

---

domain
  ticks  = expected ticks
  runs   = expected Runtime runs
  first  = expected first timestamp
  last   = expected last timestamp
```

## Notes

- BtRunner knows Runtime, never Runtime-owned Signaler, Risk, BotCycle, or Executor.
- The command owns BtRunner lifecycle order; BtRunner owns each phase.
- The command owns one BtRunner value; Init fills it and returns only an error.
- Init binds its logger before any fallible preparation.
- Lifecycle messages already identify BtRunner and their stage.
- BtRunner adds no duplicate component, event, or status fields.
- BtRunner names its explicit logger `log`.
- BtRunner constructs each complete message before calling `log`.
- Replay range selection, validation, and proof conversion stay together.
- Initialization is logged only after every owned child and replay proof succeeds.
- Failures return to the command boundary and are logged once.
- Canonical builds and tests use `-tags noasm`.
