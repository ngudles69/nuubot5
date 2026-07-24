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

## Approved Target Contract

Status: Approved target design. Not implemented.

Risk is a second persistent signal source.

It behaves like a traffic cop.

It monitors whether a BotCycle exists or not.

Controller owns Risk for the complete BotGeneration.

Risk starts once with Controller and stops only when Controller stops.

Risk state persists across BotCycles.

The exact BotSpec selects one exact Risk module.

Config supplies only that module's supported thresholds.

Controller contains no individual Risk policy logic.

## Responsibilities

The exact Risk module owns:

- Cycle-entry gates.
- Maximum drawdown logic.
- Losing-cycle limits.
- Exposure limits.
- Cycle-stop conditions.
- Controller-stop conditions.
- Persistent Risk state.
- Decision reasons and supporting values.

Examples:

```text
per-cycle loss threshold -> stop BotCycle
per-cycle duration -> stop BotCycle
maximum Bot drawdown -> stop Controller
maximum losing cycles -> stop Controller
```

## RiskInput

Risk receives one immutable `RiskInput`.

It contains only facts required by the exact implemented Risk module.

It contains no database handle, Venue client, network client, mutable pointer,
generic registry, or untyped extension map.

Likely first facts include:

- Bot capital.
- Net Bot PnL.
- Bot equity.
- Controller-run drawdown.

Exact content remains tied to real Risk rules and may change before
implementation.

Physical Account and global portfolio Risk are not part of Bot Risk.

## Decision Contract

Risk returns exactly one decision:

```text
Allow
BlockCycleStart
StopCycle
StopController
```

`Allow` permits normal Controller processing.

`BlockCycleStart` leaves an existing BotCycle running and blocks a new one.

`StopCycle` stops and flattens the active BotCycle.

`StopController` stops the BotCycle and ends the Controller generation.

Risk never:

- Creates or stops a lifecycle owner directly.
- Places, cancels, or closes Orders.
- Mutates Account, Ledger, Executor, or Venue.
- Selects trading direction.

Controller executes every accepted Risk decision.

Risk gates fail closed.

Account retains mechanical Order, symbol, margin, and Venue safety checks.

## Evaluation Order

```text
reconcile authoritative Account truth
build one coherent snapshot
evaluate Risk
block actions or accept exit decisions
run allowed strategy and Executor actions
```

Risk remains active while a BotCycle stops and Accounts flatten.

Risk may record its decision reason and supporting values.

No RiskManager or generic configurable policy engine is approved.

## Server Safety Boundary

Possible Server drawdown, Black Swan, monitoring, and emergency switches remain
deferred.

They are operational Server safeguards, not Bot Risk.

An isolated backtest cannot evaluate complete physical Account or cross-Bot
Risk.

A future Server safeguard may ask BotManager to block starts or stop supervised
Runners.

It must not call Executors or Venues directly.

A Server-controlled switch cannot protect against Server death.
