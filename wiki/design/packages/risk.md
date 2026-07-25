# Risk Package

Status: Interface implemented. BalancedRisk remains a non-protective stub.
Covers: `internal/risk/*.go`
Purpose: Return typed gates and exits from immutable Controller facts.

## Contract

```go
type Risk interface {
    Assess(Input) Decision
    Stop()
}
```

Decisions are:

- `Allow`;
- `BlockCycleStart`;
- `StopCycle`; and
- `StopController`.

## RiskInput

Input currently contains:

- timestamp;
- active-cycle state;
- completed-cycle count;
- immutable Account snapshots;
- Bot capital;
- net Bot PnL;
- Bot equity;
- peak equity; and
- current and maximum drawdown.

Controller constructs one input after reconciliation.

Every configured Risk assesses that same value before Controller acts.

Decision precedence is `StopController`, `StopCycle`, `BlockCycleStart`, then
`Allow`.

`StopCycle` also blocks cycle admission on that assessment pass.

Risk owns no lifecycle action and calls no Controller descendant.

## BalancedRisk

BalancedRisk returns `Allow` and counts assessments.

It must not be advertised as protection.

Real policies belong in their concrete Risk implementation.

No RiskManager or global Risk framework exists.
