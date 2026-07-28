# Ledger Changes

## Result

```text
Delete: nested graph helpers, CountOrders, clone/candidate staging
Add:    flat record storage, direct indexes, dirty-row Persist, scoped recovery
Change: Ledger lifecycle, persistence, schema, queries, PnL, Result, and callers
Keep:   Ledger as accounting books: memory plus optional persisted snapshot
```

## 1. Keep Recon Out of This Change

**KEY DECISION:** Ledger stores accounting books. Ledger does not perform reconciliation.

**KEY DECISION:** Recon orchestration, telemetry, cursors, and Recon-specific storage belong to Account Recon.

This file does not redesign Recon flow or delete `internal/account/ledger/recon.go`.

Recon may call Ledger storage and accounting functions. Ledger does not decide what Exchange evidence to request or accept.

## 2. Delete Dead and Bloated Ledger Code

Delete:

- `CountOrders`
- `candidate`
- `persistCandidate`
- `currentCandidate`
- `publish`
- `cloneTrades`
- `indexCLOIDs`
- `addValidatedTradeIndexes`
- full-graph `replaceTradeIndexes`
- full-graph `addTradeIndexes`
- full-graph `refreshTradeIndexes`
- full-graph `removeTradeIndexes`
- separate `save`
- separate `saveMutation`
- separate `saveRecon`
- `sortedTradeIDs` unless deterministic dirty writes prove it necessary

**KEY DECISION:** Do not retain compatibility wrappers, staged graphs, rollback graphs, or repair machinery.

## 3. Replace Nested Ownership With Flat Records

Ledger stores every admitted record once:

```go
type Ledger struct {
    config Config

    trades map[uint64]*trade.Trade
    orders map[uint64]*order.Order
    fills  map[uint64]*fill.Fill // Venue TID

    orderByCLOID map[string]OrderRef
    orderByOID   map[uint64]OrderRef
    fillByTID    map[uint64]FillRef

    orderIDsByTrade map[uint64][]uint64
    fillTIDsByOrder map[uint64][]uint64
    activeTradeIDs  map[uint64]struct{}
    activeOrderIDs  map[uint64]struct{}

    dirtyLedger bool
    dirtyTrades map[uint64]struct{}
    dirtyOrders map[uint64]struct{}
    dirtyFills  map[uint64]struct{}

    summary     Summary
    nextTradeID uint64
    nextTradeNo uint32
    nextOrderID uint64
    store   *ledgerStore
    started bool
    stopped bool
}
```

No Trade owns Order objects.

No Order owns Fill objects.

Indexes contain IDs, not duplicate records.

**KEY DECISION:** Trade, Order, and Fill remain flat. Do not add the nested graph back.

**KEY DECISION:** Ledger is the only mutable owner of its stored records and indexes.

## 4. Change Identity

Store the applicable identity directly on each record:

```text
Trade:
  SweepID, BotID, Venue, Network, Account, LedgerID, TradeID

Order:
  Trade identity, OrderID, CLOID, VenueOID when assigned

Fill:
  Trade and Order identity, VenueTID, VenueOID, CLOID when available
```

Ledger creates each child and obtains its local ID before Exchange submission.

In `max`, the inserted database row assigns that local ID.

In `none`, Ledger allocates the equivalent runtime ID because no row is inserted.

Do not add an identity struct.

**KEY DECISION:** Identity fields stay standard and explicit. They are not wrapped for organization alone.

## 5. Change Ledger Initialization

`(*Ledger).Init`:

```text
validate configuration and lifecycle
initialize empty flat maps and dirty sets
open Bot database
create or hardcut-check schema

if persist mode is none:
    keep memory only
    return

load Ledger row and saved values
load every Trade
load Orders for non-closed Trades using scoped SQL
load Fills for non-closed Trades using scoped SQL
rebuild all memory indexes
perform minimal relationship and decode validation
return
```

Recovery SQL shape:

```sql
SELECT ...
FROM account_order
WHERE sweep_id = ?
  AND bot_id = ?
  AND network = ?
  AND account_name = ?
  AND trade_id IN (
      SELECT trade_id
      FROM account_trade
      WHERE sweep_id = ?
        AND bot_id = ?
        AND network = ?
        AND account_name = ?
        AND status != 'closed'
  );
```

Use the equivalent scoped query for Fills.

`rebuildIndexes()` must be safe to rerun after replacing memory records.

**KEY DECISION:** Let SQLite filter and join before Go processes records.

**KEY DECISION:** Trust correctly scoped database rows. Validate structure, decode, and required relationships only.

**KEY DECISION:** No startup replay, repair engine, or repeated ownership checking.

## 6. Add One Persistence Method

Add:

```go
func (l *Ledger) Persist() error
```

Pseudocode:

```text
Persist():
    if configured mode is none:
        return nil

    if nothing is dirty:
        return nil

    begin transaction
    upsert dirty Ledger row if needed
    upsert dirty Trade rows only
    upsert dirty Order rows only
    upsert dirty Fill rows only
    commit

    clear only committed dirty markers
```

On failure:

```text
return error
keep dirty markers
do not repair or clone memory
```

The caller chooses the persistence boundary.

The normal live boundary is after Account Recon finishes applying accepted changes.

**KEY DECISION:** Callers call `Persist()`. They do not branch on persistence mode.

**KEY DECISION:** `none` keeps complete Ledger state in memory and performs no Ledger database writes.

**KEY DECISION:** `max` writes only dirty records in one atomic database transaction.

**KEY DECISION:** Memory is the current snapshot. Database is its persisted snapshot. Exchange remains truth.

## 7. Change Ledger Mutation API

Keep and flatten:

- `PlanTrade`
- `PlanOrders`
- `CreateTrade`
- `AddOrders`
- `Result`
- `Stop`

Rename:

- `RecordSubmit` to `UpdateOrders`
- `TradeState` to `Trade`
- `OpenTrades` to `ActiveTrades`

Add:

- `AddFill`
- `UpdateFillFee`
- `UpdateAccountPayload`
- `Persist`

Keep storage queries only when a real caller needs them:

- `ActiveOrders`
- `Trade`
- `Order`
- `Fill`
- `TradeOrders`
- `PendingCounts`
- `PendingFillAnchors`
- `HasPendingRecon`
- `Summary`
- `Result`

Pseudocode:

```text
CreateTrade(trade, orders):
    validate parent and child IDs
    store Trade once
    store each Order once
    update direct indexes
    mark new records dirty

AddFill(fill):
    resolve existing Order
    reject unknown parent
    store Fill once by Venue TID
    update Order totals
    update Trade finance
    update Ledger Summary
    mark changed records dirty

UpdateOrders(changes):
    update existing Order records
    update affected indexes
    update affected Trade finance
    update Ledger Summary
    mark changed records dirty
```

**KEY DECISION:** Account and Recon decide what evidence to apply. Ledger only stores accepted record changes and maintains its books.

## 8. Change PnL Ownership

Ledger supports standard accounting calculations:

- realized PnL when closing a Trade;
- unrealized PnL when requested;
- gross PnL;
- fees;
- net PnL;
- cached Ledger totals.

Pseudocode:

```text
Close or update Trade:
    read its stored Orders and Fills by ID indexes
    calculate standard finance
    write calculated values onto Trade
    update Ledger Summary delta

UnrealizedPnL(mark):
    use stored open quantity, side, entry price, and supplied mark
    return calculated value
```

**KEY DECISION:** Ledger knows standard PnL because it owns the books.

**KEY DECISION:** Ledger does not decide when Recon should close or update a Trade.

## 9. Delete `CountOrders`

Delete:

- `(*Ledger).CountOrders`
- `(*Account).CountOrders`
- role-and-status counting for Grid round trips

Change Grid result:

```text
RoundTrips = number of closed Trades
```

Expose this through Ledger `Result` or `Summary`; do not add another general Order-count API.

Affected caller:

- `internal/executor/grid.go`

**KEY DECISION:** A round trip is one closed Trade, not one filled take-profit Order.

## 10. Hardcut the Database Schema

Keep one row per Ledger, Trade, Order, and Fill:

```text
account_ledger
account_snapshot
account_trade
account_order
account_fill
```

Store:

```text
account_ledger:
  identity, saved counters, calculated summary

account_snapshot:
  Account identity, normalized snapshot values, latest raw Account payload

account_trade:
  identity, status, exposure, final finance, timestamps

account_order:
  identity, submitted values, Venue OID, status, calculated execution totals,
  latest raw Order payload

account_fill:
  identity, execution values, fee, liquidity, raw Fill payload
```

Do not store:

- raw Trade payload;
- raw Ledger payload;
- separate request, acknowledgement, and Venue payload histories;
- raw evidence envelopes;
- Recon telemetry;
- Recon cursors;
- repair records.

JSON rule:

```text
store the appropriate API payload as received
persist it as JSON
do not translate it into another evidence format
replace the prior payload with the latest accepted payload
```

**KEY DECISION:** Order, Fill, and Account keep their appropriate latest raw API payload only.

**KEY DECISION:** No separate payload-history system.

**KEY DECISION:** Schema mismatch fails startup. No migration, compatibility reader, dual write, or fallback.

## 11. Change Store Functions

Keep:

- `openLedgerStore`
- `(*ledgerStore).close`
- nullable and decimal SQL helpers that remain used

Replace with:

- `persist`
- `upsertLedger`
- `upsertTrade`
- `upsertOrder`
- `upsertFill`
- `loadLedger`
- `loadTrades`
- `loadActiveTradeOrders`
- `loadActiveTradeFills`

Pseudocode:

```text
load:
    load Ledger and saved values
    load all Trades
    let scoped SQL select active Trade children
    return flat records

persist:
    receive exact dirty ID sets
    write those rows in one transaction
```

**KEY DECISION:** Do not clear and rewrite complete tables for one mutation.

## 12. Upstream and Downstream Changes

### `internal/account/account.go`

- Use flat Ledger records.
- Remove `CountOrders`.
- Call Ledger mutation methods with already accepted data.
- Call `Persist()` at the selected boundary.
- Store the latest raw Account payload through Ledger.

### `internal/account/recon.go`

- Own Recon orchestration and temporary evidence.
- Own Recon telemetry, cursors, and any Recon persistence.
- Apply accepted changes through generic Ledger mutation methods.
- Call Ledger `Persist()` after the complete accepted update.

### `internal/executor/grid.go`

- Set `RoundTrips` from closed Trade count.
- Stop counting filled take-profit Orders.

### `internal/resultpublisher`

- Read the new Ledger Result shape.
- Do not duplicate domain persistence when Ledger `max` already owns it.

### Database setup and rebuild

- Create the hardcut schema from the canonical setup path.
- Rebuild an empty database without migrations.
- Create Observer, Trade, and Grid sweeps from templates.

## 13. Files Changed During Implementation

Production:

- `internal/account/ledger/ledger.go`
- `internal/account/ledger/store.go`
- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/executor/grid.go`
- affected result publication code
- canonical database setup code

Tests:

- Ledger flat-storage tests
- Ledger persistence and recovery tests
- Account integration tests
- Grid round-trip tests
- database rebuild proof

Documentation:

- owning Ledger design page
- rebuild procedure
- `HANDOFF.md`

## 14. Proof

Required:

```text
no CountOrders
no candidate or Clone graph
one in-memory Trade, Order, and Fill record each
none opens the database and creates schema but writes no Ledger records
max writes dirty rows only
persistence failure retains dirty markers
recovery uses scoped SQL for active Trade children
rebuildIndexes is rerunnable
latest Account Snapshot, Order, and Fill raw JSON round-trips exactly
RoundTrips equals closed Trade count
standard realized and unrealized PnL remains exact
schema mismatch fails fast
database rebuild succeeds twice from empty state
Observer, Trade, and Grid sweeps pass
full tests and vet pass with -tags noasm
git diff --check passes
```

No new interface, repository layer, event bus, migration, compatibility bridge, payload-history system, or repair engine is added.
