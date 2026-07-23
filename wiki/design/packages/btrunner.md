# BtRunner Package

Status: Implemented.
Covers: `internal/btrunner/btrunner.go`
Purpose: Run one bounded historical Bot replay and prove exact completion.

## Canonical Source

- `D:/rust/nuubot4/src/btrunner.rs`

## Scope & Responsibilities

BtRunner owns Setup context, TickClock, TickReader, Runtime, and replay proof.

- Serve every admitted tick to Runtime.
- Trigger due Runtime passes.
- Stop direct children in reverse order.

## Program Flow

```text
BtRunner(log, sweep_id, bot_id)

init
  ctx      = nuubot_setup(log, sweep_id, bot_id)
  clock    = TickClock(log, ctx.config.btrunner)
  ticks    = TickReader(log, ctx.bot, ctx.config.btrunner)
  runtime  = Runtime(log, ctx)
  proof    = ReplayProof(ctx.bot, ctx.config.btrunner)

start
  runtime.start()

run
  for tick in ticks
    runtime.ingest(tick)
    proof.record(tick)

    if clock.advance(tick.time)
      proof.record_pass()
      if runtime.run(tick.time)
        break

  proof.verify()

stop
  runtime.stop()
  ticks.stop()
  clock.stop()

---

domain
  ticks  = expected ticks
  passes = expected passes
  first  = expected first timestamp
  last   = expected last timestamp
```

## Notes

- BtRunner knows Runtime, never Runtime-owned Signaler, Risk, BotCycle, or Executor.
- Canonical builds and tests use `-tags noasm`.
