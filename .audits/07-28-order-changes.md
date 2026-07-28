# Order Changes

## Result

```text
Delete: nested Fill ownership, duplicate Order representations, clones, revisions, transition matrix
Add:    one flat Order struct and one Order row
Change: Order updates one trusted acknowledgement or Venue observation
Keep:   validation, closure, and slippage calculations
```

## 1. Delete

No whole production file is deleted.

Delete types:

- `Input`
- `VenueState`
- `Record`
- `ReconState`
- `ActiveState`
- `Summary`

Delete fields:

- `fills`
- `comparisonState`
- stored `active`
- stored `reconciliationPending`

Delete functions:

- `(*Order).ApplyFill`
- `(*Order).RefreshRecon`
- `(*Order).Record`
- `(*Order).ReconState`
- `(*Order).ComparisonState`
- `(*Order).FillIdentity`
- `(*Order).ActiveState`
- `(*Order).Summary`
- `(*Order).EachFill`
- `(*Order).Fill`
- `(*Order).Clone`
- `(*Order).refreshFills`
- `transitionAllowed`
- `copyInput`

**KEY DECISION:** Order never owns Fill objects.

## 2. Add One Flat Order

Replace every Order state variant with one `Order`.

```text
Identity:
    SweepID when present
    BotID
    Venue
    Network
    Account
    LedgerID
    TradeID
    OrderID
    CLOID
    Venue OID when assigned

Submitted values:
    CycleNumber
    Symbol
    BatchNumber
    OrderPosition
    Role
    Side
    Type
    TimeInForce
    SubmittedQuantity
    SubmittedPrice
    TriggerPrice
    ReduceOnly
    SubmittedMS

Venue and calculated values:
    Status
    RejectReason
    UpdatedMS
    FilledQuantity
    FilledNotional
    AverageFillPrice
    RemainingQuantity
    Fees
    FillCount
    PendingFeeCount
    LastFillMS
    RawJSON
```

`RawJSON` stores the latest raw Order payload as JSON.

**KEY DECISION:** One Order struct equals one Order row.

**KEY DECISION:** Keep only the latest raw Order payload. Add no payload-history system.

## 3. Change Functions

Keep and simplify:

- `New`
- `validRole`
- `validSide`
- `validType`
- `validTimeInForce`
- `validStatus`

Merge:

- `RecordSubmit`
- `ApplyVenueState`

Into:

```go
func (o *Order) Update(update Update) (bool, error)
```

Pseudocode:

```text
Update(update):
    preserve existing Venue OID
    accept Venue OID when first assigned
    reject a changed nonzero Venue OID
    update status, rejection, timestamp, and latest RawJSON

    if nothing changed:
        return false

    return true
```

Add:

```text
IsClosed:
    true for canceled, rejected, expired, or error
    true for filled only when submitted quantity is filled and every Fill fee exists

Slippage:
    buy  = average Fill price - submitted price
    sell = submitted price - average Fill price
    percent = price slippage / submitted price * 100
```

Fill aggregation moves to Ledger `AddFill`.

**KEY DECISION:** Submitted price, trigger price, and actual Fill price remain separate.

**KEY DECISION:** Positive slippage is worse for both buy and sell Orders.

## 4. Change Identity Creation

Pseudocode:

```text
Trade or Ledger creates the child Order identity before Exchange submission
max: database insert assigns OrderID
none: Ledger assigns the equivalent runtime OrderID
Exchange later assigns Venue OID
store OrderID, CLOID, and Venue OID separately
```

**KEY DECISION:** OrderID is the local row identity when persisted. Venue OID and CLOID remain separate.

**KEY DECISION:** Every application Order has a nonblank valid CLOID before Exchange I/O.

## 5. Change Ledger Ownership

Ledger stores each Order once:

```text
Orders[OrderID] -> *Order
OrderByCLOID[CLOID] -> OrderID
OrderByOID[VenueOID] -> OrderID
OrderIDsByTrade[TradeID] -> Order IDs
```

Ledger owns:

- adding Orders;
- applying Fill totals;
- active Order selection;
- Trade relationship lookup;
- persistence;
- recovery indexes.

**KEY DECISION:** Relationship indexes contain IDs, not duplicate Order objects.

## 6. Affected Callers

- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/account/ledger/ledger.go`
- `internal/account/ledger/recon.go`
- `internal/account/ledger/store.go`
- `internal/account/trade/trade.go`
- `internal/executor/grid.go`
- `internal/executor/trade.go`
- Order, Trade, Ledger, Account, and Executor tests

Callers use the flat `Order`. No `Record`, `ReconState`, `ActiveState`, or `Summary` conversion remains.

## 7. Persistence and Recovery

```text
insert local Order intent before Exchange I/O when persistence is enabled
upsert only changed Order rows
store latest RawJSON without translating it
load only scoped Orders required by open or unresolved Trades
trust stored rows
rebuild Ledger relationship indexes
do not add repair logic
```

**KEY DECISION:** Database selects scope recovery by Sweep, Bot, network, Account, and active Trade relationships.

## 8. Proof

```text
no Order Input, VenueState, Record, ReconState, ActiveState, Summary, Clone, or Fill map remains
one Order struct maps to one database row
Order intent exists before Exchange I/O
CLOID, Venue OID, and OrderID survive persistence
latest raw Order JSON round-trips exactly
Fill totals and fee-pending closure remain exact
slippage uses submitted price versus weighted Fill price
active Order callers use Ledger indexes
recovery loads scoped Orders without repair
full tests and vet pass with -tags noasm
git diff --check passes
```

No transition framework, child collection, clone path, identity struct, or payload-history system is added.
