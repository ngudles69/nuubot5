# Reconciliation

Status: Implemented for Simulator-backed Executors. Approved live behavior remains unimplemented.
Covers: `internal/controller`, `internal/botcycle`, `internal/executor`,
`internal/account`, and `internal/ledger`
Purpose: Rebuild coherent Account truth before Risk and execution policy.

## Current Implementation

```text
Controller Run
  BotCycle reconciles each capable Executor Account
    Account queries Simulator truth
    Ledger admits Order and Fill evidence atomically
    Account returns immutable Snapshot
  Controller builds RiskInput
  Risk decides
  BotCycle delivers accepted OnRecon
```

Failed reconciliation prevents Risk-dependent admission and Executor policy.

Controller receives values only.

It retains no Account, Ledger, Trade, Order, Fill, or Simulator pointer.

Submission, Simulator mutation, and open-position marks make Account dirty.

Normal reconciliation may skip a clean Account. Forced reconciliation always queries truth.

TradeExecutor Stop cancels Orders, closes exposure, forces reconciliation, and
proves zero active Orders and zero position.

Flatness alone never requests BotCycle exit.

## Approved Live Heartbeat

Runner owns one scheduler timer with a ten-second heartbeat.

Each heartbeat reads time once. The same value determines which configured work is due.

Due work may include:

- normal reconciliation;
- unresolved-Order cleanup;
- balance and equity calculation;
- telemetry append; and
- stopping action.

All intervals are configurable. Due times advance from scheduled boundaries, not work completion, to avoid timer drift.

Sweep uses its existing deterministic cadence. It does not gain live failure tolerance.

## Normal Live Reconciliation

Normal reconciliation reads, in order:

1. `openOrders`.
2. `userFillsByTime` from the last committed inclusive cursor.
3. Exact `orderStatus` for each locally active Order absent from `openOrders`.
4. Account state.

Comparison uses stable CLOID, OID, and TID indexes.

Work is limited to active Orders, new Fills, and touched Trades. It does not traverse the complete Ledger graph.

The Fill cursor advances only after the complete reconciliation succeeds.

Fill retrieval handles 2,000-row pagination. Repeated inclusive boundaries deduplicate by Venue TID.

Hyperliquid history has these boundaries:

- `userFills` and `userFillsByTime` have no symbol filter and cap responses at 2,000 rows;
- `userFillsByTime` retains only the latest 10,000 Fills;
- `historicalOrders` returns the latest 2,000 Orders; and
- `openOrders` has no documented 2,000-row cap.

Routine Order reconciliation never queries `historicalOrders`.

## Unresolved Active Orders

An active local Order remains unresolved when it is absent from `openOrders` and exact `orderStatus` provides no conclusive evidence.

Future live behavior marks the Order unresolved. Its owning Grid level becomes stuck or quarantined.

The level receives no replacement, reuse, or assumed outcome.

Other levels may continue only within an approved safety boundary. That boundary remains unresolved.

Sweep fails immediately on the first unresolved Order or reconciliation error.

## Unresolved-History Cleanup

Future live cleanup runs only when the unresolved set is nonempty and its configurable interval is due.

A possible interval is 30 minutes or one hour. No default is approved.

Cleanup reads the latest 2,000 historical Orders and Fills.

It matches CLOID, OID, and TID, repairs only exact evidence, and releases levels only when their evidence is complete.

Hot-path state tracks primitive counters, unresolved high-water count, oldest age, and attempts.

The escalation age and attempt threshold remain unresolved.

## Failure Policy

A transient whole-reconciliation failure in live execution retains the last published generation.

The first and second consecutive failures publish failure telemetry without domain state or cursor changes.

A successful complete reconciliation resets the consecutive-failure count.

The third consecutive failure begins stoppage.

Sweep fails on the first error.

## Publication Atomicity

Reconciliation stages exact deltas, then validates the Ledger and Account candidate.

Maximum persistence writes only dirty rows in one SQL transaction.

After commit, memory publication must be non-failing.

Failure before publication changes no domain state, cursor, or successful Account snapshot.

The design does not deep-clone the complete object graph for rollback.
