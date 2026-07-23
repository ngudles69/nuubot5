# Risk

## Purpose

Request graceful Runtime stop when coherent state breaches a configured risk rule.

## Status

Implemented.

No real risk rule is implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/risk.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/risk.md`
- Nuubot5: `internal/risk/risk.go`

## Scope

Risk defines one synchronous assessment and shutdown boundary selected through configuration.

## Owner and Children

Runtime owns configured Risks.

The factory creates concrete Risks.

## Responsibilities

- Select the configured Risk implementation.
- Return whether Runtime must stop.
- Preserve Runtime as the only graceful-stop owner.
- Stop and report implementation-owned statistics.

## Does Not

- Stop Runtime directly.
- Own Account references.
- Read mutable Ledger state.
- Query exchanges.
- Execute orders.
- Invent missing snapshots.

## Lifecycle

Construct through the factory, assess during each Runtime pass, then stop.

## Inputs and Outputs

Current input is one timed assessment call without domain state.

Current output is a Boolean stop request.

Canonical future input is coherent owned Account snapshots.

## State and Invariants

Runtime MUST assess Risk before active BotCycle pass work.

Risk MUST return intent only. Runtime owns the resulting stop.

Future snapshot input MUST be immutable and coherent.

## Concurrency

Risk assessment is synchronous in Runtime.

## Persistence

None.

## Errors

Unknown factory kinds fail construction.

The current interface cannot return assessment or stop errors.

## Program Flow

```text
New
  select configured Risk

Runtime Pass
  assess each Risk
  request risk stop when any returns true

Runtime Stop
  stop Risks in reverse order
```

## Required Proof

- Unknown kinds fail.
- Every Runtime pass assesses each configured Risk.
- A true result uses Runtime graceful stop.
- Risk stop runs once.

## Open Decisions

The future snapshot type and fallible assessment contract require Account design.
