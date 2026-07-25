# BalancedRisk

Status: Implemented placeholder.
Covers: `internal/risk/balanced.go`
Purpose: Prove the configured Risk call path without requesting an exit.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/risk/balanced.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/risk.md`

## Scope

BalancedRisk counts assessments and always returns `Allow`.

## Owner and Children

Controller owns BalancedRisk through the Risk interface.

BalancedRisk owns no child.

## Responsibilities

- Count assessment calls.
- Return `Allow`.
- Report assessments and zero requested exits once.

## Does Not

- Evaluate balances.
- Read Account snapshots.
- Calculate equity or drawdown.
- Request a real risk exit.
- Claim implemented risk protection.

## Lifecycle

Create, run repeatedly, then stop once.

## Inputs and Outputs

Input is one immutable `RiskInput`.

Output is always `Allow`.

## State and Invariants

Exit decisions MUST remain zero while this object is a placeholder.

Assessment count MUST match Controller assessments reaching it.

## Concurrency

BalancedRisk is synchronous.

## Persistence

None.

## Errors

Current construction, assessment, and stop paths return no error.

## Program Flow

```text
createBalanced
  create risk

Assess
  record assessment
  return Allow

Stop
  stop risk
```

## Required Proof

- Every assessment returns Allow.
- Assessment count increments once per call.
- Repeated stop reports once.
- Logs prove assessment count and zero requested exits.

## Open Decisions

The actual balanced-risk rule is undefined and MUST NOT be inferred from the name.

BalancedRisk is current proof scaffolding only.

It MUST NOT be presented as active protection.

The approved target lets each exact BotSpec construct one real persistent Risk
module containing its supported gates and exit rules.

BalancedRisk remains documented as a stub until an approved implementation
hardcuts it from active defaults.
