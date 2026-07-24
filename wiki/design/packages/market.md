# Market Package

Status: Implemented.
Covers: `internal/market/market.go`
Purpose: Carry one admitted best-price market event across the replay and Runtime boundary.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/market.rs`

## Scope

Current BBO contains one normalized timestamp and one price.

## Owner and Children

Replay Reader creates BBO values.

BtRunner, Runtime, BotCycle, and Executors consume copies.

BBO owns no child.

## Responsibilities

- Preserve one positive millisecond timestamp.
- Preserve one finite positive price.
- Establish trusted market input after boundary validation.

## Does Not

- Track market history.
- Represent separate bid and ask values.
- Calculate midpoint or spread.
- Trigger policy by itself.
- Own lifecycle.

## Lifecycle

Admit once, pass by value, then discard.

## Inputs and Outputs

Inputs are normalized timestamp and decoded close price.

Output is one trusted `market.BBO`.

## State and Invariants

Timestamp MUST be positive.

Price MUST be finite and positive.

Replay Reader separately enforces exact one-second sequence.

## Concurrency

BBO is an immutable value after creation.

## Persistence

None.

## Errors

Invalid timestamp or price fails admission.

## Program Flow

```text
CreateBBO
  validate bbo
  create bbo
```

## Required Proof

- Zero timestamps fail.
- Non-positive, NaN, and infinite prices fail.
- Valid values retain exact timestamp and price.

## Open Decisions

Live BBO may require separate bid and ask fields. That contract is not approved.
