# Runtime

## Purpose

Own synchronous Bot decisions, active-cycle lifecycle, Risk assessment, and graceful stop.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/runtime.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/btrunner.md`
- Nuubot5: `internal/runtime/runtime.go`

## Scope

Runtime consumes admitted BBO values and timed passes for one configured Bot.

## Owner and Children

BtRunner owns Runtime.

Runtime directly owns:

- one Signaler;
- configured Risks; and
- at most one active BotCycle.

## Responsibilities

- Declare Signaler bar requirements.
- Prepare Signaler from loaded bars.
- Release every due Signal during BBO ingestion.
- Open one BotCycle only when none is active.
- Count Signals skipped while a cycle is active.
- Deliver BBO values to the active BotCycle.
- Assess Risks before cycle pass work.
- Close completed cycles.
- Own all graceful stop reasons.
- Close an active BotCycle during stop.
- Stop remaining children in reverse ownership order.

## Does Not

- Read Parquet or databases.
- Drive its Clock.
- Calculate concrete indicators.
- Implement execution policy.
- Persist results.
- Claim Account integration.

## Lifecycle

`New` creates Signaler and Risks.

`PrepareBars` calculates Signals before start.

`Start` starts Signaler.

Canonical target `Pass` performs one timed control pass.

`Stop` is idempotent and closes all active work.

Current Go names `Pass` as `MainLoop`. This is documented pre-contract drift.

## Inputs and Outputs

Inputs are Runtime configuration, optional end date, Bars, BBO values, and pass timestamps.

Outputs are stop decisions, child actions, statistics, and errors.

## State and Invariants

At most one BotCycle may be active.

The first stop reason wins.

An end-date BBO requests graceful stop.

Every started cycle MUST close exactly once.

Risk assessment occurs before BotCycle pass work.

## Concurrency

Runtime decisions are synchronous.

Future feed goroutines MUST NOT execute Runtime policy.

## Persistence

Runtime owns no datastore.

## Errors

Invalid lifecycle calls fail.

Child construction, start, pass, and stop failures propagate.

Shutdown MUST still attempt remaining direct children.

## Program Flow

```text
New
  create Signaler
  create configured Risks
  store optional end timestamp

PrepareBars
  calculate and validate Signals

Start
  start Signaler
  mark started

Ingest
  release every due Signal
  open cycle when none exists
  otherwise count skipped Signal
  send BBO to active cycle
  request end_date stop when reached

Pass
  assess every Risk
  stop when requested
  run active BotCycle pass
  close completed cycle
  stop at maximum cycles

Stop
  latch first reason
  close active BotCycle
  stop Risks in reverse order
  stop Signaler
  report Runtime statistics
```

## Required Proof

- No concurrent active cycles exist.
- Due Signals are released without lookahead.
- Risk and end-date exits use Runtime `Stop`.
- Active BotCycle closes during stop.
- Started and closed cycle counts match.
- Current accepted statistics remain exact.

## Open Decisions

Future Account ownership and reconciliation inputs require their detailed design before implementation.
