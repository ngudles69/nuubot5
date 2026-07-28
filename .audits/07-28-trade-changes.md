# Trade Changes

## Result

```text
Delete: nested Order ownership, duplicate Trade representations, clones, traversal APIs
Add:    one flat Trade struct and one Trade row
Change: Trade calculations consume Ledger-supplied records
Keep:   standard realized and unrealized PnL calculations
```

## 1. Delete

No whole production file is deleted.

Delete types:

- `Input`
- `Record`
- `ReconState`
- `Summary`
- `metrics`

Delete fields:

- `orders`
- `finalized`

Delete functions:

- `(*Trade).AddOrder`
- `(*Trade).Refresh`
- `(*Trade).RefreshRecon`
- `(*Trade).MarkedRecord`
- `(*Trade).Record`
- `(*Trade).ReconState`
- `(*Trade).Summary`
- `(*Trade).Clone`
- `(*Trade).Order`
- `(*Trade).EachOrder`
- `(*Trade).record`
- `(*Trade).executions`
- `sameMetrics`

**KEY DECISION:** Trade never owns Orders or Fills.

## 2. Add One Flat Trade

Replace all Trade state variants with one `Trade`.

```text
Identity:
    SweepID when present
    BotID
    Venue
    Network
    Account
    LedgerID
    TradeID

Trade data:
    TradeNumber
    CycleNumber
    Symbol
    Status
    Side
    OpenQuantity
    AverageEntryPrice
    RealizedPnL
    UnrealizedPnL
    GrossPnL
    Fees
    NetPnL
    OpenedMS
    ClosedMS
    UpdatedMS
```

Trade stores no Exchange raw payload.

**KEY DECISION:** One Trade struct equals one Trade row.

**KEY DECISION:** Trade is an application record. It has no Exchange payload.

## 3. Change Functions

Keep and simplify:

- `New`
- `calculate`
- `calculateFinance`
- `isCloseRole`
- `sameSign`

Rename:

- `RefreshMark` to `UpdateMark`
- `isTerminal` to `IsClosed`

Add:

```go
func (t *Trade) Update(orders []*order.Order, fills []*fill.Fill) (bool, error)
```

Pseudocode:

```text
Update(orders, fills):
    sort Fills by time, then Venue TID
    calculate exposure, status, realized PnL, fees, and timestamps
    calculate unrealized, gross, and net PnL

    if Trade is already closed and calculated values differ:
        fail

    if nothing changed:
        return false

    write calculated fields into this Trade
    return true
```

```text
UpdateMark(mark):
    if Trade is closed or flat:
        return unchanged

    calculate unrealized, gross, and net PnL
    update changed fields
```

**KEY DECISION:** Ledger supplies related Orders and Fills. Trade only calculates its own finance.

**KEY DECISION:** Closing a Trade and requesting unrealized PnL remain standard Trade methods.

## 4. Change Identity Creation

Pseudocode:

```text
Ledger creates Trade intent
max: database insert assigns TradeID
none: Ledger assigns the equivalent runtime TradeID
store one TradeID
```

No clone, identity wrapper, or alternate state object is added.

**KEY DECISION:** TradeID is the local row identity when persisted, not a second duplicate ID.

## 5. Change Round Trips

Delete:

- `Account.CountOrders`
- `Ledger.CountOrders`
- Grid use of Take-Profit Filled Order count

Replace with:

```text
RoundTrips = number of closed Trades
```

Affected caller:

- `internal/executor/grid.go`

**KEY DECISION:** RoundTrips counts closed Trades, not Orders.

## 6. Affected Callers

- `internal/account/ledger/ledger.go`
- `internal/account/ledger/recon.go`
- `internal/account/ledger/store.go`
- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/executor/grid.go`
- `internal/executor/trade.go`
- `internal/resultpublisher`
- Trade, Ledger, Account, Executor, and report tests

Callers receive the flat `Trade` value or pointer. No `Record`, `ReconState`, or `Summary` conversion remains.

## 7. Persistence and Recovery

```text
insert or upsert one Trade row
load all scoped Trade rows
trust stored rows
rebuild Ledger indexes and summary
do not reconstruct Trade by replaying children
do not add repair logic
```

Closed Trades remain in memory. Their Orders and Fills may remain only in the database.

**KEY DECISION:** Closed Trade finance is stored and trusted.

## 8. Proof

```text
no Trade Input, Record, ReconState, Summary, Clone, or child map remains
one Trade struct maps to one database row
long and short PnL remain exact
partial and complete close remain exact
fees, gross PnL, net PnL, and mark-only unrealized PnL remain exact
closed Trade values cannot change
RoundTrips equals closed Trade count
recovery loads scoped Trades without repair
full tests and vet pass with -tags noasm
git diff --check passes
```

No repository layer, identity struct, child collection, clone path, or payload-history system is added.
