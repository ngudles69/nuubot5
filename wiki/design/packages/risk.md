# Risk Package

Status: Implemented.
Covers: `internal/risk/*.go`
Purpose: Create and assess configured stop policies behind a stable factory.

## Canonical Source

- `D:/rust/nuubot4/src/risk.rs`

## Scope & Responsibilities

RiskFactory selects concrete Risks. Runtime knows only the common Risk
contract.

- Each Risk assesses coherent Runtime state.
- New policies add one concrete file and one factory case.

## Program Flow

```text
create
  select implementation
  create risk

assess stop
  record assessment

stop
  stop risk
```

## Notes

- Current BalancedRisk records assessments and never requests exit.
- Risk has no separate Init or Start work, so those phases are omitted.
- Risk uses `AssessStop`, not `Run`, because future policies may assess other
  actions such as position reduction.
