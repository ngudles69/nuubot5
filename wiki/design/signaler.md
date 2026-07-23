# Signaler

## Purpose

Calculate Signals from complete closed Bars and release them without lookahead.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/signaler.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/signaler.md`
- Nuubot5: `internal/signaler/signaler.go`

## Scope

Signaler owns calculator selection, loaded Bars, calculated Signals, release position, and lifecycle state.

## Owner and Children

Runtime owns one Signaler.

Signaler owns one private calculator selected from configuration.

## Responsibilities

- Select Macross or RSI through the stable factory.
- Declare exact timeframe and warmup requirements.
- Calculate all Signals once before replay.
- Validate Signal timestamp ordering.
- Release due Signals in order.
- Report calculated, released, and pending counts.

## Does Not

- Read files.
- Load Bars.
- Open BotCycles.
- Execute orders.
- Recalculate indicators during replay.

## Lifecycle

`New` selects one calculator.

`Prepare` calculates and validates Signals once.

`Start` permits release.

Repeated `Next` releases due Signals.

`Stop` is idempotent.

## Inputs and Outputs

Inputs are Signaler configuration, loaded Bars, and current BBO timestamp.

Outputs are bar requirements and zero or more ordered Signal values.

## State and Invariants

Every Signal timestamp MUST precede its availability timestamp.

Availability timestamps MUST increase strictly.

A Signal releases only when current BBO time is later than availability.

Each Signal releases at most once.

## Concurrency

Signaler is synchronous and owned by Runtime.

## Persistence

None.

## Errors

Unknown kinds, missing Bars, calculator failures, lifecycle violations, and invalid ordering fail.

## Program Flow

```text
New
  select calculator

BarsNeeded
  return calculator requirements

Prepare
  calculate all Signals
  validate timestamp order
  retain Bars and Signals

Start
  enable release

Next
  reject invalid lifecycle
  return not available when next Signal is not closed
  release one due Signal

Stop
  disable release
  report statistics
```

## Required Proof

- Required Bars include sufficient warmup.
- Signals use closed Bars only.
- Equal-time BBO does not release a Signal.
- Later BBO releases it once.
- Accepted replay reports 55 calculated and released Signals.

## Open Decisions

None.
