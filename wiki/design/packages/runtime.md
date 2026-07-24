# Runtime Package

Status: Implemented with reconciliation-first trading control.
Covers: `internal/runtime/runtime.go`
Purpose: Run one configured Bot and own its direct control children.

## Ownership

Runtime is the running Bot.

It owns one passive Signaler, configured Risks, and at most one BotCycle.

Runtime reads the Bot configuration and symbol.

It owns standard entry admission and graceful Bot stop decisions.

## Signal Polling

BtRunner's Runtime timer calls `Runtime.Run`.

Runtime requests the latest Signal package at the scheduled time.

Runtime stores the latest consumed timestamp.

A package is consumed before BotCycle admission begins.

The same package is never retried.

Runtime reads only `enter_long` and `enter_short`.

It does not interpret custom Signal fields.

## Entry Admission

```text
read latest package
  duplicate timestamp
    ignore
  no entry trigger
    wait
  active BotCycle
    consume and skip entry
  no BotCycle
    initialize BotCycle
```

One entry trigger must be active.

BotCycle rejection is handled and logged as nonfatal.

Runtime keeps `r.cycle` empty and waits for the next package.

Unexpected BotCycle initialization errors remain fatal.

## Program Flow

```text
Init
  create signaler
  create risks
  initialize runtime

Start
  start runtime

Run
  check stop request
  reconcile botcycle
  assess risk stops
  check stop request
  deliver recon event
  check botcycle
  check max cycles
  read signal
  consume signal
  open botcycle

Stop
  request stop
  capture and stop botcycle
  stop risks
  stop signaler
  stop runtime

IngestBBO
  ingest botcycle bbo
  deliver botcycle bbo

openCycle
  initialize botcycle

closeCycle
  stop botcycle and collect result
  retain immutable result values

Result
  return immutable runtime result
```

## BBO Flow

```text
Runtime.IngestBBO
  BotCycle.IngestBBO
  BotCycle.OnBBO
```

Runtime does not inspect Executor capabilities.

BotCycle owns capability detection.

See [IngestBBO](../concepts/ingestbbo.md).

## Cycle Closure

Runtime checks BotCycle completion during timed `Run`.

It clears `r.cycle` before stopping the completed cycle.

Reader exhaustion ends input.

Runtime stop closes any remaining cycle gracefully.

Reaching `max_cycles` stops Runtime after the final cycle closes.

## Trading Control Order

Runtime stores the latest admitted BBO.

It passes that value into new Executor initialization.

```text
Run
  check stop request
  reconcile botcycle
  assess risk
  check stop request
  deliver recon event
  check botcycle
  check max cycles
  read signal
  consume signal
  open botcycle
```

Reconciliation precedes Risk and Executor decisions.

Failed reconciliation prevents both.

During control passes, Runtime receives AccountSnapshot values only.

It never owns or reaches into Account state.

During closure, Runtime receives immutable BotCycle result values.

It retains no Account, Ledger, Simulator, Executor, or BotCycle pointer.

Each `closeCycle` appends the returned `botcycle.Result`.

After Stop, `Runtime.Result` returns one immutable `runtime.Result` containing every closed cycle.

See [Trading State Tranche](../concepts/trading-state.md).
