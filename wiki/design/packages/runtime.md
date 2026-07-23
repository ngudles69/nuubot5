# Runtime Package

Status: Implemented.
Covers: `internal/runtime/runtime.go`
Purpose: Own and sequence one Bot's control components.

## Canonical Source

- `D:/rust/nuubot4/src/runtime.rs`

## Scope & Responsibilities

Runtime owns one factory-created Signaler, factory-created Risks, and at most
one BotCycle.

- Release due Signals before delivering each BBO.
- Run Risks before the active BotCycle.
- Own graceful stop decisions.

## Program Flow

```text
Runtime(log, ctx)

init
  signaler = SignalerFactory(ctx.config.signaler).init(log, ctx)
  risks    = RiskFactory(ctx.config.risks).init(log, ctx)

start
  signaler.start()

run(bbo)
  for signal in signaler.release(bbo.time)
    if no botcycle
      botcycle = BotCycle(log, signal, ctx.config.executors)
      botcycle.start()

  botcycle.ingest(bbo)

run(now)
  if any risk requests exit
    stop active botcycle
    return stop

  if botcycle.run(now) completes
    stop botcycle

  return continue

stop
  botcycle.stop()
  risks.stop()
  signaler.stop()
```

## Notes

- Runtime uses factories. It never selects concrete Signalers, Risks, or Executors.
- Runtime preparation belongs inside Runtime initialization.
