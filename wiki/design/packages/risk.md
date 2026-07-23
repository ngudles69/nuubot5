# Risk Package

Status: Implemented.
Covers: `internal/risk/*.go`
Purpose: Create and run configured stop policies behind a stable factory.

## Canonical Source

- `D:/rust/nuubot4/src/risk.rs`

## Scope & Responsibilities

RiskFactory selects concrete Risks. Runtime knows only the common Risk
contract.

- Each Risk assesses coherent Runtime state.
- New policies add one concrete file and one factory case.

## Program Flow

```text
RiskFactory(kind, log, risk, config) -> Risk

init
  concrete = select kind
  concrete.init(log, risk, config)

start
  ready to assess

run
  return concrete.assess()

stop
  concrete.stop()
```

## Notes

- Current BalancedRisk records assessments and never requests exit.
