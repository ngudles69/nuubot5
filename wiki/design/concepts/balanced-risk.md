# BalancedRisk

Status: Stub.
Covers: `internal/risk/balanced.go`
Purpose: Prove the configured Risk call path without requesting an exit.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/risk/balanced.rs`
- Nuubot4 contract: `D:/rust/nuubot4/wiki/logic/risk.md`

## Scope

BalancedRisk counts Runtime runs and always declines a stop request.

## Owner and Children

Runtime owns BalancedRisk through the Risk interface.

BalancedRisk owns no child.

## Responsibilities

- Count assessment calls.
- Return `false`.
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

Input is one Runtime Run call.

Output is always `false`.

## State and Invariants

Exit requests MUST remain zero while this object is a stub.

Assessment count MUST match Runtime Run calls reaching it.

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

Run
  record assessment

Stop
  stop risk
```

## Required Proof

- Every assessment returns false.
- Assessment count increments once per call.
- Repeated stop reports once.
- Logs prove assessment count and zero requested exits.

## Open Decisions

The actual balanced-risk rule is undefined and MUST NOT be inferred from the name.
