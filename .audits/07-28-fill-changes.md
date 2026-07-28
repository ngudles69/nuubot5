# Fill Changes

## Result

```text
Delete: Input, Record, State, Clone, duplicate execution comparison
Add:    one flat Fill struct and one Fill row
Change: Update accepts later fee evidence and latest raw Fill payload
Keep:   New and HasFee
```

## 1. Delete

No whole production file is deleted.

Delete types:

- `Input`
- `Record`

Delete functions:

- `(*Fill).Enrich`
- `(*Fill).State`
- `(*Fill).Clone`
- `(*Fill).sameExecution`

Rename or merge:

- `validateInput` into `Validate`

Delete separate immutable and mutable field copies inside `Fill`.

**KEY DECISION:** Fill has one representation.

## 2. Add One Flat Fill

Replace all Fill state with one `Fill`.

```text
Identity:
    database ID
    SweepID when present
    BotID
    Venue
    Network
    Account
    LedgerID
    TradeID
    OrderID
    Venue TID
    Venue OID
    CLOID when available from Exchange, otherwise copied from local Order

Execution:
    CycleNumber
    Symbol
    Side
    Quantity
    Price
    TimestampMS
    Fee
    Liquidity
    RawJSON
```

`Fee == nil` means fee evidence is missing. Zero fee and negative rebate are present fees.

`RawJSON` stores the raw Fill payload as JSON.

**KEY DECISION:** One Fill struct equals one Fill row.

**KEY DECISION:** Keep the raw Fill payload. Add no payload-history system.

## 3. Change Functions

Keep and simplify:

- `New`
- `HasFee`

Rename:

- `Enrich` to `Update`

Pseudocode:

```text
New(fill):
    validate complete parent and Venue identity
    validate side, quantity, price, and timestamp
    preserve exact execution and RawJSON
    return one Fill
```

```text
Update(fee, rawJSON):
    if existing fee conflicts:
        fail

    set later fee when provided
    replace RawJSON when provided
    return whether anything changed
```

`Update` receives fee evidence and its raw payload only. It does not receive a second complete Fill to compare.

**KEY DECISION:** Fill execution identity, quantity, price, side, and timestamp never change after admission.

## 4. Change Identity Creation

Pseudocode:

```text
Ledger resolves the existing parent Order
Venue TID identifies the Fill
database insert assigns physical row ID
copy TradeID, OrderID, CLOID, Venue OID, and Account scope from the parent
store all exact IDs
```

No synthetic FillID is added.

**KEY DECISION:** Venue TID is the Fill identity.

**KEY DECISION:** Parent Order supplies trusted child relationship identity.

## 5. Change Ledger Ownership

Move Fill admission and lookup from Order to Ledger.

```text
AddFill(evidence):
    find existing Fill by Venue TID

    if existing:
        update later fee evidence only
    else:
        resolve existing Order by CLOID
        otherwise resolve by scoped Venue OID
        create Fill with parent IDs and CLOID
        store Fill once
        index Venue TID

    update owning Order totals
    mark owning Trade dirty
```

Ledger stores:

```text
Fills[VenueTID] -> *Fill
FillByTID[VenueTID] -> TradeID + OrderID
FillTIDsByOrder[OrderID] -> Venue TIDs
```

**KEY DECISION:** Unknown Fill evidence never creates a Trade or Order.

**KEY DECISION:** Raw Fill may omit CLOID. Resolve by scoped Venue OID and copy the local Order CLOID.

## 6. Affected Callers

- `internal/account/recon.go`
- `internal/account/ledger/ledger.go`
- `internal/account/ledger/recon.go`
- `internal/account/ledger/store.go`
- `internal/account/order/order.go`
- `internal/account/trade/trade.go`
- Fill, Order, Trade, Ledger, and Account tests

Callers use the flat `Fill`. No `Input`, `Record`, or `State` conversion remains.

## 7. Persistence and Recovery

```text
insert one new Fill row
upsert only a Fill whose fee or raw payload changed
retain every Fill row in the database
load Fills only for scoped open or unresolved Trades
trust stored rows
rebuild Venue TID and Order relationship indexes
do not replay Fills to repair records
do not add excessive validation
```

**KEY DECISION:** Database is trusted persisted snapshot storage. Wrong writes are fixed in code.

## 8. Proof

```text
no Fill Input, Record, State, Clone, or sameExecution remains
one Fill struct maps to one database row
Venue TID deduplicates Fill admission
CLOID-first and scoped Venue-OID routing resolve the correct Order
zero fee and negative rebate count as present
missing fee can be completed once
execution values cannot change
raw Fill JSON round-trips exactly
Order totals and Trade finance update from the admitted Fill
recovery loads scoped Fills without repair
full tests and vet pass with -tags noasm
git diff --check passes
```

No synthetic Fill ID, child collection, clone path, repair engine, identity struct, or payload-history system is added.
