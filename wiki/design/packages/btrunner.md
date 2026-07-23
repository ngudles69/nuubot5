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

init(log, sweep_id, bot_id)

init
  ctx      = nuubot_setup(log, sweep_id, bot_id)
  clock    = TickClock(log)
  clock.register_timer(ctx.config.btrunner.timer_interval, runtime_run)
  ticks    = TickReader(log, ctx.bot, ctx.config.btrunner)
  runtime  = Runtime(log, ctx)
  proof    = ReplayProof(ctx.bot, ctx.config.btrunner)

start
  runtime.start()

loop
  for tick in ticks
    runtime.ingest(tick)
    proof.record(tick)

    clock.advance(tick.time)
    if runtime requested stop
      break

runtime_run(time)
  proof.record_pass()
  remember runtime.run(time) stop request
  propagate runtime error

  proof.verify()

stop
  clock.stop()
  ticks.stop()
  runtime.stop()

---

domain
  ticks  = expected ticks
  passes = expected passes
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
