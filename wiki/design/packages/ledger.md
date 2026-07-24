# Ledger Package

Status: Implemented for memory, maximum persistence, reload, and final publication.
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
  index active evidence

CreateTrade
  stage Trade and initial Orders
  persist staged tree when configured
  publish Trade and Orders

AddOrders
  stage existing Trade
  persist staged Orders when configured
  publish Orders

RecordSubmit
  stage submitted Orders
  apply ordered outcomes
  persist submission evidence when configured
  publish submit result

Recon
  index active local Orders
  match incoming Venue evidence
  validate complete recon batch
  persist affected trees and cursor when configured
  publish recon result

Result
  copy Trades Orders and Fills
  copy reconciliation cursor and snapshot
  return immutable Ledger result

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

## Atomicity

Ledger stages the complete reconciliation result before publishing it.

Under `max`, one transaction persists affected Trades, Orders, Fills, Ledger snapshot, and Fill cursor.

A failed transaction publishes no success cursor or snapshot.

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

## Dirty State

Ledger owns no reconciliation-dirty flag.

Ledger operations report whether accepted evidence changed.

Account owns, claims, clears, and restores its recon-dirty state.

Recovery always forces reconciliation before new decisions.

## Persistence Modes

`none` retains memory state until one successful final export.

`max` persists every accepted mutation.

Account passes the configured mode to Ledger.

Account passes store operations only for `max`.

Sweep runs select `none`.

A failed Sweep run has no recovery checkpoint. Its coordinator reruns it.

Durable child-state reload selects `max`.

Full Bot resume remains pending Runner-owned orchestration cursors.

## Terminal Result

`Result` returns one immutable `ledger.Result`.

It contains Ledger identity, cursor, snapshot, Trades, Orders, and Fills.

Every slice and map is newly owned. No value aliases mutable Ledger state.

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
