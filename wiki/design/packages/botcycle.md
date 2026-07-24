# BotCycle Package

Status: Implemented with optional Account reconciliation and result collection.
Covers: `internal/botcycle/botcycle.go`
Purpose: Own Executors for one admitted entry Signal.

## Ownership

Runtime creates one BotCycle after a standard entry trigger.

BotCycle owns its initialized Executors.

BotCycle knows Executor interfaces, never concrete types.

## Admission

BotCycle asks the Executor factory to initialize every configured Executor.

Any Executor may return admission rejection.

BotCycle then stops previously initialized Executors and rejects the cycle.

Runtime consumes that Signal, keeps no cycle, and waits for the next Signal.

Unexpected Executor errors remain fatal.

Executors without gates initialize normally.

## Program Flow

```text
Init
  create executors
  initialize botcycle

Run
  check completion

Stop
  stop executors
  collect cached account results
  resolve exit reason
  calculate duration
  report proof

Reconcile
  reconcile capable Executor Accounts

OnRecon
  deliver accepted recon event

Result
  return immutable BotCycle result

IngestBBO
  ingest executor bbo

OnBBO
  record cycle time
  deliver executor bbo
```

## Event Dispatch

BotCycle uses capability assertions.

```go
if handler, ok := activeExecutor.(executor.BBOHandler); ok {
    handler.OnBBO(bbo)
}
```

Only running Executors receive events.

An Executor without a capability is silently skipped.

## BBO Order

```text
Runtime.IngestBBO
  BotCycle.IngestBBO
    each running BBOIngestHandler
  BotCycle.OnBBO
    each running BBOHandler
```

Simulator ingestion completes before `OnBBO`.

See [IngestBBO](../concepts/ingestbbo.md).

## Completion

BotCycle completes when every Executor is stopping, stopped, or in error.

Runtime closes a completed cycle during its next timed `Run`.

Runtime clears `r.cycle` before stopping the old cycle.

## Trading Extension

BotCycle remains the capability dispatcher.

It never receives or exposes Account references.

```text
Reconcile
  reconcile capable Executors
  collect Account snapshots
  return one completed barrier

OnRecon
  deliver accepted recon event
```

Runtime calls `Reconcile` before Risk.

Runtime calls `OnRecon` only after reconciliation and Risk succeed.

One Executor reconciliation failure stops the control pass.

BotCycle returns no partial success barrier.

BotCycle stops every Executor before collecting supported `AccountResultProvider` values.

Collected results are immutable values.

`Stop` returns one `botcycle.Result` after every Executor has stopped.

That value contains ordered `account.Result` values captured during Executor shutdown.

BotCycle returns no Account, Ledger, Simulator, or Executor pointer.

See [Trading State Tranche](../concepts/trading-state.md).

## Approved Target Meaning

Status: Approved target design. Not implemented.

A BotCycle is one exchange-style Bot campaign.

One accepted strategy Signal starts the complete BotCycle.

BotCycle owns one coordinated Executor unit.

Every configured Executor:

- Initializes before any Venue mutation.
- Starts with the BotCycle.
- Receives the same immutable strategy Signal.
- Runs its own configured market and Order logic.
- Stops with the BotCycle.

No Executor starts independently.

No Executor responds to a separate strategy Signal.

An Executor may remain running while it only monitors.

Starting an Executor does not require placing an Order.

Examples include:

- Grid plus hedge.
- Grid plus grid.
- DCA plus hedge.
- BTC trade plus ETH hedge.

## Approved Exit Model

Flat Account state does not trigger BotCycle completion.

An explicit exit condition triggers BotCycle Stop.

Exit sources may include:

- Signaler regime change.
- Risk cycle exit.
- Executor price bounds.
- Executor strategy objective.
- User Stop.
- Runner or replay boundary.
- Fatal component failure.

The detecting component returns an exit decision.

It never stops siblings directly.

Controller accepts the decision and starts BotCycle Stop.

## Approved Stop and Completion

```text
accept exit decision
enter stopping
reject new strategy and Order actions
request every Executor stop
cancel remaining Orders through their owners
close remaining positions through their owners
reconcile every used Account-symbol
prove zero active Orders
prove zero position
finish every Executor stop
complete BotCycle
```

Flatness is a terminal proof, not an exit trigger.

A BotCycle may monitor while flat for its entire active duration.

A BotCycle may exit before placing any Order.

BotCycle cannot complete successfully until authoritative reconciliation proves:

- Zero active Orders for every used Account-symbol.
- Zero position for every used Account-symbol.

Cleanup failure produces unsuccessful terminal evidence.

Signaler and Risk are Controller children.

BotCycle Stop never stops or resets them.

They remain active across BotCycles.

Normal Executor exit conditions stop only the BotCycle.

A fatal Executor error stops every Executor and ends the Controller generation
after cleanup.

No automatic restart follows a fatal Executor error.

## Approved Result Evidence

BotCycle Result records:

- Bot and generation identity.
- Cycle identity.
- Exit source.
- Exit reason.
- Exit timestamp.
- Final reconciliation timestamp.
- Final flatness proof.
- Cleanup success or failure.

Result hierarchy is:

```text
ControllerResult
  BotCycleResults
    ExecutorResults
```

Every result is an immutable value.

It preserves Bot, generation, cycle, and Executor identity.

Account and Ledger evidence remains attributable to its exact Executor
identity.

Live cross-process Account ownership remains TBD.
