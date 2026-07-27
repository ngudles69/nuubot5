# Ledger Package

Status: Implemented for memory, maximum persistence, reload, and summary results.
Covers: `internal/ledger/*.go`
Purpose: Hold one Account's coherent local Trades, Orders, Fills, and reconciliation cursor.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/account/ledger.py`
- `D:/rust/nuubot3/wiki/account/ledger.md`

## Ownership

Account owns one Ledger.

Ledger owns Trades.

Each Trade owns Orders.

Each Order owns Fills.

Venue and Simulator never mutate this tree.

## Lifecycle

Ledger is one concrete mutable component.

It does not need a factory.

`Init` binds one Ledger identity and persistence mode.

`none` starts in memory without opening a database.

`max` opens its durable identity and loads evidence.

`Stop` releases owned resources.

Trade, Order, and Fill remain domain objects without lifecycle phases.

## Program Flow

```text
Init
  bind Ledger inputs
  validate persistence mode
  initialize Ledger
  open Ledger identity when configured
  load Ledger evidence when configured
  rebuild indexes and cached Summary

CreateTrade
  prepare Trade and initial Orders
  persist new Trade and Orders when configured
  publish Trade, Orders, and exact Summary delta

AddOrders
  validate new Orders
  attach validated Orders directly
  persist touched Trade and Orders when configured
  publish Orders and exact Trade Summary delta

RecordSubmit
  validate submitted Orders
  apply ordered outcomes directly
  persist touched submission rows when configured
  refresh touched indexes and exact Trade Summary deltas

Recon
  prepare attempt
  stage selected Fill updates
  stage selected Order updates
  recalculate touched Trade structure
  remark active Trade exposure from stored state
  apply exact old-to-new Trade Summary deltas
  validate candidate index deltas
  persist dirty rows and cursor when configured
  publish recon result without failure

Result
  aggregate terminal flat domain counts
  copy reconciliation cursor and cached Ledger Summary
  return summary-only Ledger result

Stop
  stop Ledger
```

Each indented action becomes one exact source comment during implementation.

## Reconciliation Contract

Account supplies normalized Order, Fill, and account-state values.

Ledger matches only existing local CLOIDs.

Unowned Venue activity becomes a reconciliation diagnostic.

Ledger never adopts an unknown Order into a Trade.

Every matched identity is confirmed before mutation.

Valid forward transitions may share one Venue timestamp.

Conflicting terminal transitions still fail.

Older evidence is ignored.

Duplicate identical Fill evidence is idempotent.

Changed execution fields for one Venue TID fail.

Returned Venue facts may update matching local evidence.

Absence from a bounded Venue response changes nothing.

Ledger never deletes an Order or Fill because it disappeared from Venue history.

Future live comparison uses stable CLOID, OID, and TID indexes.

Routine reconciliation works only on active Orders, new Fills, and touched Trades.

Canonical Recon detects accepted Order changes through the Order-owned transient
mutation revision.

It reads allocation-free Fill ownership instead of constructing detached Order
records.

Ledger owns one cached derived `Summary`.

Trades remain authoritative.

Init or persisted reload rebuilds the cache once from all Trades.

CreateTrade, AddOrders, RecordSubmit, and canonical Recon apply exact old-to-new Trade summary deltas.

`Summary` and `ReconSummary` return the cache without traversing Trades.

An inconclusive exact lookup marks an active Order unresolved. It does not infer a terminal state.

Unresolved-history cleanup may repair only exact CLOID, OID, or TID evidence.

## Atomicity

Ledger validates normalized identities before direct mutation.

It updates owned Trades, Orders, and Fills without cloning the object graph.

Under `max`, one transaction persists only dirty Trade, Order, Fill, Ledger identity, and Fill cursor rows.

Cached Summary is derived and never persisted. The schema remains unchanged.

A failed transaction may leave directly mutated records and cached totals equally untrusted. Sweep exits immediately; live decisions remain blocked.

Account publishes no successful Account Snapshot after persistence failure.

External Venue calls never occur inside Ledger transactions.

Under `max`, one transaction records one complete normalized submission batch.

## Submission Outcomes

Account supplies one normalized outcome for every requested local Order.

Each outcome carries the existing Order identity and preserves request order.

Account expands one payload-wide Venue error before calling Ledger.

Ledger records explicit item errors as terminal `rejected` evidence.

Known local Simulator submission failures become terminal `error` evidence.

Account then marks itself recon-dirty.

Successful acknowledgements move their Orders to submitted.

Reconciliation confirms open, filled, canceled, or expired state.

Only reconciliation creates canonical Fill objects and final filled Order state.

Malformed, incomplete, or unknown response evidence leaves Orders recoverable as `created`.

Mixed success and rejection outcomes apply atomically.

## Cursor

`fills_through_ms` is inclusive.

Account queries from this boundary with bounded time ranges.

The next query repeats the boundary timestamp.

Venue TID uniqueness removes duplicate boundary rows.

The cursor never moves backward.

The cursor never advances past an unproven capped response.

Hyperliquid Fill responses cap at 2,000 rows. Account paginates them before Ledger receives a complete candidate.

Repeated inclusive timestamp boundaries deduplicate by Venue TID.

## Dirty State

Ledger owns no reconciliation-dirty flag.

Ledger operations report whether accepted evidence changed.

Account owns, claims, clears, and restores its recon-dirty state.

Recovery always forces reconciliation before new decisions.

## Capacity

Ledger maps, sets, and reconciliation slices grow dynamically.

Ledger reserves no fixed Trade, Order, Fill, or evidence-buffer capacity.

## Persistence Modes

`none` performs no Ledger, Trade, Order, or Fill database writes.

`max` persists every accepted mutation.

Account passes the configured mode to Ledger.

Account passes store operations only for `max`.

Sweep runs select `none`.

A failed Sweep run has no recovery checkpoint. Its coordinator reruns it.

Durable child-state reload selects `max`.

Full Bot resume remains pending Runner-owned orchestration cursors.

## Terminal Result

`Result` returns one summary-only `ledger.Result`.

It contains Ledger identity, cursor, cached finance totals, and flat Trade, Order, Fill, cancellation, and Stop counts.

Terminal Order traversal for counts, cancellations, and Stop Orders remains unchanged.

It contains no Trade, Order, Fill, slice, map, or copied ownership graph.

## Does Not

- Query Venue.
- Decode raw Hyperliquid JSON.
- Create CLOIDs.
- Select credentials.
- Calculate Simulator matching.
- Guess Trade ownership by symbol.
- Revalidate trusted normalized values.

## Required Proof

- Trade and initial Orders persist atomically.
- Ordered submission outcomes persist atomically.
- One known Simulator failure records one terminal error per requested Order.
- Explicit item rejection becomes terminal without creating a Fill.
- Successful HTTP acknowledgement alone makes no Order terminal.
- Immediate-fill acknowledgement creates no Fill before reconciliation.
- Contradictory recon batches mutate nothing.
- Duplicate recon is idempotent.
- Fill cursor cannot move backward.
- Bounded-history absence deletes nothing.
- One Fill updates its Order and Trade once.
- Persistence failure publishes no success.
- Ledger owns no reconciliation-dirty flag.
- Reload reconstructs the same domain tree.
- A complete-traversal test oracle equals cached Summary after every mutation, Recon, failed validation, and reload.
- `Summary` and `ReconSummary` reads allocate nothing.
