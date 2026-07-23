# Executor

## Purpose

Define one execution policy boundary inside a BotCycle.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/executor.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/executor.md`
- Nuubot5: `internal/executor/executor.go`

## Scope

Executor defines lifecycle, BBO delivery, timed passes, terminal state, and exit reason.

## Owner and Children

`internal/executor` defines and owns the current Executor interface.

BotCycle owns concrete Executor instances returned by the factory.

This producer-owned interface placement is pre-contract drift from the consumer-owned target.

## Responsibilities

- Provide the current minimal execution behavior contract.
- Select a concrete implementation through one factory.
- Accept the same Signal as sibling Executors.
- Report terminal state and exit reason.

## Does Not

- Own BotCycle.
- Release Signals.
- Assess Runtime Risk.
- Expose concrete implementation types.
- Add methods without a proven consumer need.

## Lifecycle

`Start` begins one Executor.

`OnBBO` accepts one admitted market value.

Canonical target `Pass` advances one timed step.

`Stop` terminates and records the parent reason.

Current Go names `Pass` as `MainLoop`. This is documented pre-contract drift.

## Inputs and Outputs

Factory inputs are logger, cycle identity, Executor identity, Signal, and configuration.

Runtime inputs are BBO values, pass timestamps, and stop reasons through BotCycle.

Outputs are terminal state, exit reason, and errors.

## State and Invariants

The current interface remains in `internal/executor`.

The coding contract requires consumer-owned interfaces. Moving this interface requires a separately confirmed source change.

Factory kinds MUST be validated before execution.

Terminal Executors MUST receive no further BBO or pass work.

## Concurrency

Executors are synchronous under BotCycle ownership.

## Persistence

The current Executor contract owns no persistence.

## Errors

Unknown factory kinds and implementation lifecycle failures propagate through BotCycle.

## Program Flow

```text
New
  select configured implementation
  return Executor

BotCycle Start
  Executor Start

BotCycle OnBBO
  Executor OnBBO when non-terminal

BotCycle Pass
  Executor Pass when non-terminal

BotCycle Stop
  Executor Stop
```

## Required Proof

- Factory selects ObserverExecutor.
- Unknown kinds fail.
- BotCycle calls only non-terminal Executors.
- Stop records a stable exit reason.

## Open Decisions

Future real Executors may create Accounts. Their ownership contract requires detailed design before implementation.
