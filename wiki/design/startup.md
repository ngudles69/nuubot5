# Runner Startup

Status: Design approved; implementation pending.
Covers: Runner startup and recovery for `simnet`, `testnet`, and `mainnet`.
Purpose: Start every Runner Bot through one deterministic path from persisted state to reconciled live operation.

## Scope

This design belongs only to the Runner execution path.

BtBot uses replay input, TickClock, and its separate bounded replay lifecycle.

BtBot never loads runtime state and never recovers an interrupted backtest.

A crashed backtest is discarded and rerun from the beginning.

Runner uses the same startup and recovery flow for:

- `simnet` through Simulator;
- `testnet` through Hyperliquid testnet; and
- `mainnet` through Hyperliquid mainnet.

The selected network changes the Venue. It does not change startup sequencing or recovery behavior.

## Decision

Runner startup handles both new Bots and recoverable Bots.

Terminal `error` and `stopped` Bots never restart.

Rerunning terminal behavior requires cloning the Bot into a new Bot ID.

There is one Runner Init path and one Runner Start path.

No caller selects a new-Bot path or recovery path.

Loaded Ledger state naturally determines the required reconciliation work.

## Why

Separate new and recovery paths duplicate lifecycle behavior and eventually drift.

Separate network startup paths would duplicate the same persistence, reconciliation, and lifecycle decisions.

One path proves every Runner Bot loads persisted truth, establishes reconciliation boundaries, and resumes safely.

A new Runner Bot is only a Bot whose normal persistence queries return zero rows.

A recovered Runner Bot is only a Bot whose same queries return persisted rows and lifecycle intent.

## Init

Every Runner Bot executes:

```text
Init
  load Bot state and mutable status
  select configured Venue and network
  load Ledger
  load Trades
  load Orders
  load Fills
  construct initialized runtime objects
```

A new Runner Bot loads zero Ledger, Trade, Order, and Fill rows.

A recovered Runner Bot loads its persisted Ledger tree through the same queries.

Initialization never marks an uninitialized object Running or Stopping.

An Executor may begin in-memory as `Configured` while retaining separate persisted lifecycle intent.

Init performs no startup reconciliation.

## Start

Every Runner Bot executes:

```text
Start
  start selected Venue connection
  run Account reconciliation
  continue persisted lifecycle intent
```

Start never decides whether the Bot is new or recovered.

Start never selects different lifecycle logic for `simnet`, `testnet`, or `mainnet`.

Account reconciliation reads the loaded Ledger and selects the required work.

## Reconciliation

Venue means Simulator on `simnet` and Hyperliquid on `testnet` or `mainnet`.

```text
Ledger has no active Trades
  skip Venue Order pull
  skip Venue Fill pull
  advance reconciliation cursors to current time
  return

Ledger has active Trades
  pull Venue Orders
  pull Venue Fills
  reconcile Venue truth into Ledger
  advance reconciliation cursors
  return
```

The no-active-Trade path avoids irrelevant historical Venue reads.

Advancing cursors establishes the current boundary for future Order and Fill reconciliation.

Active Trades require Venue truth because Orders, Fills, exposure, or cleanup may remain incomplete.

The same reconciliation code serves clean startup, normal restart, crash recovery, and stopping cleanup.

## Lifecycle Continuation

Runner resumes persisted intent only after initialization and startup reconciliation.

```text
configured
  continue normal Start

starting
  continue idempotent Start

running
  complete idempotent Start
  continue Running

stopping
  complete idempotent Start
  continue idempotent Stop

stopped
  do not restore as active
```

A recovered stopping Bot completes remaining cancellation, position closure, reconciliation, and terminal persistence.

## Ownership

Runner owns startup sequencing and failure reporting.

Runner never clears persisted Bot, Ledger, Trade, Order, Fill, or Venue data.

Failed data remains available for review, troubleshooting, strategy analysis, and code analysis.

A code defect does not authorize automatic deletion or repair.

Explicit backend repair is separate operator work. Otherwise the Bot remains `error` with its evidence intact.

Each component loads and initializes its directly owned state.

Account owns reconciliation selection and cursor advancement.

Ledger owns loaded Trade, Order, and Fill evidence.

The selected Venue supplies reconciliation truth.

Venue truth repairs active persisted state. It does not create a separate startup path.

## Invariants

- This contract applies only to Runner.
- Runner contains no `clearData` operation.
- Terminal `error` and `stopped` Bots cannot restart.
- A rerun requires a cloned Bot with a new Bot ID.
- Failed and terminal data remains durable evidence.
- BtBot never restores interrupted runtime state.
- BtBot crash handling is a complete rerun from the beginning.
- `simnet`, `testnet`, and `mainnet` execute identical startup sequencing.
- New and recovered Runner Bots execute identical Init code.
- New and recovered Runner Bots execute identical Start code.
- Init always loads the complete persisted Ledger tree.
- Every Start invokes Account reconciliation.
- No active Trades means no Venue Order or Fill pull.
- The no-active-Trade path advances reconciliation cursors to current time.
- Active Trades select normal Venue reconciliation.
- Persisted lifecycle intent never bypasses initialization.
- Start and Stop continuation are idempotent.
- Stopping recovery finishes cleanup instead of resuming normal operation.

## Required Proof

- All three networks execute the same Runner startup sequence.
- Empty persistence returns zero rows through normal Init queries.
- New-Bot Start performs no Venue Order or Fill pull.
- New-Bot Start advances reconciliation cursors to current time.
- Active persisted Trades trigger normal Order and Fill reconciliation.
- Recovered `starting` and `running` states duplicate no startup effects.
- Recovered `stopping` completes partial cleanup.
- Interrupted Start and Stop remain safe when repeated.
