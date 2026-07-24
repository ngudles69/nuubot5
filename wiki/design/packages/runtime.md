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

## Approved Controller Hardcut

Status: Approved target design. Not implemented.

`Runtime` will become `Controller` through one direct hard rename.

The target uses:

```text
internal/controller
Controller
controller.Result
```

No Runtime wrapper, type alias, compatibility package, old Config field, or
dual log name will remain.

The existing synchronous `Run` control pass remains.

BtRunner retains repeated `Loop`.

Reconciliation-first ordering remains.

See [BotSpec](../concepts/bot-spec.md).

## Approved Controller Meaning

Controller runs one admitted BotGeneration built from one complete BotSpec.

Controller owns for its entire lifetime:

- Signaler.
- Risk.
- BotCycle admission.
- Zero or one active BotCycle.
- Stop decisions.
- Immutable Result evidence.

Live cross-process Account ownership remains TBD.

Signaler and Risk start once with Controller.

They remain active across every BotCycle and every flat interval.

They stop only when Controller stops.

BotCycle completion never resets Signaler indicators or Risk state.

## Approved Decision Model

Signaler behaves like a traffic light.

It reports strategy state but owns no lifecycle action.

Risk behaves like a second signal source and traffic gate.

It reports gates and exit decisions but owns no lifecycle action.

Controller alone decides whether to:

- Ignore a strategy signal.
- Block cycle admission.
- Start the BotCycle.
- Stop the BotCycle.
- Stop the complete Controller.

Entry admission considers:

- Whether one BotCycle is already active.
- Risk gates.
- Account-symbol ownership.
- Clean Venue state.
- Capital and margin.
- Meta and market-data readiness.
- Controller stop state.

Current BotSpecs always allow exactly one active BotCycle.

There is no configurable `max_concurrent_cycles`.

Users start another Bot when they want another concurrent strategy.

A future multi-position TradeBot requires a separate BotSpec design.

While the BotCycle runs, `StartCycle` does nothing.

After completion, Controller rechecks the current strategy action on the next
control event.

It may start another BotCycle without a fresh crossover.

It never restarts in the same event or timestamp.

There is no queued Signal.

## Approved Target Flow

```text
Controller Start
  start Signaler
  start Risk

Controller Run
  reconcile admitted Account truth
  build immutable RiskInput
  read one Risk decision
  read current strategy action
  arbitrate decisions
  admit, stop, or run the BotCycle

Controller Stop
  stop active BotCycle
  prove required Account flatness
  stop Risk
  stop Signaler
  stop Controller
```

Risk or Signaler failure is fatal to Controller.

A fatal Executor failure stops and flattens the BotCycle, then ends the
Controller generation after cleanup.

No new BotCycle starts after a fatal component failure.

Component decisions cross ownership boundaries as values.

Signaler and Risk never call Controller, BotCycle, Executor, Account, or Venue.

## Approved Result

`ControllerResult` is one immutable value.

It contains:

- Bot and generation identity.
- Ordered BotCycleResults.
- Signaler decisions.
- Risk decisions.
- Controller exit source and reason.
- Bot capital, net Bot PnL, Bot equity, and drawdown summaries when available.

It contains no live child pointer.

BacktestResult wraps ControllerResult with replay, Config, Meta, data, code, and
timing provenance.

## Deferred Telemetry

Possible Bot-run samples include:

- Timestamp.
- Bot run identity.
- Bot capital.
- Realized and unrealized PnL.
- Fees and funding.
- Net Bot PnL.
- Bot equity.
- Drawdown.

The final schema, cadence, persistence, physical Account telemetry, and GUI
queries remain deferred.

Risk uses current immutable in-memory facts.

Risk never reads chart rows or telemetry storage.
