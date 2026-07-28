# Ledger Stack Review

## TODO

1. Remove the clone chain (`Fill.Clone <- Order.Clone <- Trade.Clone <- cloneTrades <- nobody calls it`). This is dead code.
2. Simplify Fill to one `Fill` struct. Keep `New`, `Update`, and `HasFee`. Validate inside `New`. Delete `Input`, `Record`, `State`, and `sameExecution`. `Update` receives only fee evidence.
3. Simplify Order to one local Exchange record. Keep `New`, `Update`, `IsClosed`, and validation helpers. Remove state variants, versioning, transition matrix, clones, and nested Fill traversal.
4. Make Ledger the canonical owner of complete Trade, Order, and Fill records. Store each once in flat maps. Link them through ID indexes. Move `AddFill`, Fill lookup, and Fill iteration to Ledger. Order stores no Fill objects.

## Proposed Account Stack Hardcut

Status: Proposed. Requires adversarial review before implementation.

Scope: `internal/account/{ledger,trade,order,fill}` and direct Account, Executor,
ResultPublisher, reporting, persistence, test, and design callers.

This section is the implementation target. It replaces the current nested
Trade-Order-Fill object graph without compatibility code.

### Main Ideas

1. Ledger owns one flat local copy of every admitted Trade, Order, and Fill.
2. Each Trade, Order, and Fill is represented by one Go struct and one SQLite row.
3. There are no separate Input, Record, State, ReconState, metrics, or clone representations.
4. Struct comments group fields. Field groups are not nested structs.
5. Every record carries complete applicable Nuubot identity for direct routing,
   SQL filtering, logging, and debugging.
6. Exchange JSON is temporary untrusted evidence.
7. Account validates and normalizes Exchange JSON once.
8. Ledger matches normalized evidence to trusted local records.
9. CLOID is the durable bridge between local Order intent and Exchange Order truth.
10. Venue TID identifies one Fill. Raw Exchange Fill evidence may omit CLOID.
    Resolve by CLOID when present, otherwise by Account and Venue OID.
11. Account creates Trade and Order intent before Exchange I/O.
12. Live commits Trade and Order intent before `place_order`.
13. Backtest keeps the same intent in memory and performs no runtime database write.
14. Submitted Order values remain separate from actual Fill values.
15. Slippage derives from exact submitted price and actual weighted Fill price.
16. Raw submitted and received JSON is retained for debugging.
17. Typed fields contain only identity, routing, decisions, finance, indexing,
    SQL filtering, and frequently used report values.
18. Raw JSON prevents duplicating every Exchange field into typed columns.
19. All Trade records remain in memory because they are cheap and own final finance.
20. Orders and Fills remain in memory while their Trade is active or unresolved.
21. Closed Trade records store complete final finance and become immutable.
22. After durable Trade closure, its Orders and Fills remain in SQLite but leave memory.
23. Closed Trade finance is trusted. The runtime does not retain children as a
    defensive fallback for bad calculations.
24. Calculation defects require code and proof fixes.
25. Relationship indexes contain IDs or compact references, never duplicate objects.
26. Routine Recon updates only relevant evidence and dirty parents.
27. Downloaded Order, Fill, and Account JSON is attempt-local and discarded after Recon.
28. Live SQLite contains complete durable history.
29. Backtest `persist_mode=none` contains complete in-memory history and reruns after failure.
30. Live persistence and memory publication use the same Trade, Order, and Fill structs.
31. Every Recon updates mark-derived finance for every active Trade, even when
    no Order or Fill changed.
32. Fill cursors advance only across Exchange time ranges proven complete.
33. Exact raw Exchange envelopes are durable debugging evidence, including
    decode failures that cannot attach to a domain record.

### Requirements

#### General

- Use standard Go, `shopspring/decimal`, SQLite, and existing packages.
- Add no dependency, generic repository layer, interface, factory, event bus,
  object pool, compatibility adapter, or dual behavior.
- Use `-tags noasm`.
- Preserve exact accepted finance and behavior.
- Validate untrusted JSON at Account or Hyperliquid ingress.
- Trust concrete internal records and resolved pointers after ingress.
- Do not pass identity into an update only to verify the caller selected the right record.
- Unknown Exchange Orders are diagnostic evidence. Recon never adopts them.
- Recon creates a Fill only when CLOID or Venue OID resolves an existing local Order.
- Recon never creates a Trade or Order.
- Every admitted local Fill has a CLOID copied from its resolved local Order.
- Venue OID is the normal bridge when raw Fill evidence omits CLOID.

#### Identity

Every record carries the complete applicable identity set.

```text
Trade
  SweepID (optional)
  BotID
  AccountID
  LedgerID
  TradeID

Order
  SweepID (optional)
  BotID
  AccountID
  LedgerID
  TradeID
  OrderID

Fill
  SweepID (optional)
  BotID
  AccountID
  LedgerID
  TradeID
  OrderID
```

Cycle, Executor, Venue, network, symbol, CLOID, Venue OID, and Venue TID remain
where required for routing, uniqueness, recovery, or proof.

Identity repetition is intentional. It removes repeated tree traversal and makes
every record independently searchable and diagnosable.

Identity sources and target types are:

```text
SweepID    *uint64  optional; from setup.Nuubot.Bot.SweepID when present
BotID       uint64  from setup.Nuubot.Bot.BotID
AccountID   string  Venue Account identifier from Account Config.Name
LedgerID    uint64  from Account Config.LedgerID
TradeID     uint64  Ledger allocated
OrderID     uint64  Ledger allocated
VenueTID    uint64  Exchange allocated and used as Fill identity
```

After implementation and profiling, an identity field may be removed only when
callers, SQL, logs, recovery, and proof show it has no use.

#### Submitted Values and Slippage

Account stores exact local intent before Exchange I/O.

Order stores the exact values encoded into the Exchange request:

```text
submitted quantity
submitted price when present
submitted trigger price when present
submitted Order type
submitted time-in-force
submitted reduce-only value
submitted timestamp
submitted request JSON
```

`SubmittedPrice` means the exact `px` encoded into the request. For a market
submission, this is its encoded aggressive IOC price. `TriggerPrice` remains
separate. Slippage is request-price versus actual weighted Fill price.

Fill stores actual Exchange execution:

```text
actual quantity
actual price
actual timestamp
actual fee when present
actual liquidity when present
raw Fill JSON
```

Nuubot submits every Order with CLOID. Raw Fill evidence does not need to return
it. Resolve by raw CLOID when present. Otherwise resolve the unique local Order
by Account and Venue OID, then copy the local Order CLOID onto the admitted Fill.

Order stores exact calculated execution totals:

```text
filled quantity
filled notional
average Fill price
remaining quantity
fees
Fill count
pending-fee count
last Fill timestamp
```

Slippage is derived, not independently mutable.

```text
Buy price slippage  = average Fill price - submitted price
Sell price slippage = submitted price - average Fill price
Slippage percent    = price slippage / submitted price * 100
```

Positive slippage is worse for both sides.

Slippage is unavailable when submitted price or actual Fill quantity is absent.

#### Raw JSON

Typed records remain lean because raw JSON preserves complete evidence.

```text
Order.SubmitJSON
  exact per-Order JSON value sent by Account

Order.AckJSON
  exact per-Order acknowledgement JSON value returned for submission

Order.VenueJSON
  latest accepted authoritative Venue Order observation

Fill.VenueJSON
  latest accepted authoritative Venue Fill observation

Ledger.AccountJSON
  latest accepted clearinghouse Account state

RawExchangeEvidence
  exact complete request or response envelope
  operation kind, timestamp, and success or decode error
  retain the latest 100 envelopes per Ledger and operation
  not copied into every Order or Fill
```

Fee enrichment replaces `Fill.VenueJSON` with the later fee-complete official
Fill row. The durable fee field proves the accepted current value.

Ingress captures the exact response bytes before decoding. Domain `VenueJSON`
uses `json.RawMessage` from that payload; it is not re-encoded from a typed value.

Invalid or unmatched JSON is not admitted into a Trade, Order, or Fill. Live
persists its exact envelope and error as `RawExchangeEvidence`. Backtest keeps
it in terminal result evidence.

Live transaction ownership is exact:

```text
submission request envelope
  commit with local Trade and Order intent before Exchange I/O

submission response envelope
  on valid decode, commit with acknowledgement outcome
  on decode failure, commit exact bytes and error alone

Recon request envelope
  commit before each Exchange call

Recon response envelope
  on valid decode, commit with admitted rows in the Recon transaction
  on decode failure, commit exact bytes and error alone

retention
  delete envelopes older than the latest 100 for that Ledger and operation
  in the same insert transaction
```

Network receipt and SQLite commit cannot be one atomic operation. A crash
between them may lose that response envelope; durable intent and CLOID recovery
preserve trading correctness.

### Ownership and Memory

```text
Account
`-- Ledger
    |-- all Trade records
    |-- Orders for active or unresolved Trades
    |-- Fills for active or unresolved Trades
    |-- compact identity indexes
    |-- active and pending ID sets
    `-- cached Ledger Summary
```

Ledger is the only mutable owner.

Trade, Order, and Fill are flat records with calculation methods. They do not
own child objects, maps, database handles, Venue clients, schedulers, or lifecycle
components.

#### Live Memory

Live retains:

```text
all Trade records
all Orders belonging to active or unresolved Trades
all Fills belonging to active or unresolved Trades
all compact CLOID, Venue OID, and Venue TID references
active Trade IDs
active Order IDs
pending Order IDs
pending-fee Fill IDs
relationship indexes for active or unresolved Trades
cached Ledger Summary
next local identity values
reconciliation cursors
```

Closed Trade records remain because Ledger can cheaply sum exact realized,
unrealized, gross, fee, and net values.

Closed Trade children remain only in SQLite.

#### Backtest Memory

Backtest retains every Trade, Order, and Fill until terminal publication.

`persist_mode=none` performs no Ledger, Trade, Order, or Fill runtime database work.

#### Durable Closure

A Trade group leaves Live memory only after:

```text
Trade.IsClosed() is true
every Order belonging to the Trade is closed
every Fill fee is complete
no pending reconciliation work references the Trade
the complete SQLite transaction commits
Ledger Summary includes the final Trade values
```

After commit:

```text
keep final Trade record in memory
remove its Order records from memory
remove its Fill records from memory
remove active relationship indexes
retain compact CLOID, Venue OID, and Venue TID references
```

### Target Records

The exact field types must follow existing IDs and decimal conventions. These
examples define grouping and ownership, not final Go spelling.

#### Trade

```go
type Trade struct {
    // Identity (denormalized for faster retrieval)
    SweepID     *uint64 // optional
    BotID       uint64
    AccountID   string
    LedgerID    uint64
    TradeID     uint64

    // General immutable fields
    TradeNumber uint32
    CycleNumber int
    Symbol      string

    // Updated and final Exchange-derived position
    Status            Status
    Side              string
    OpenQuantity      decimal.Decimal
    AverageEntryPrice decimal.Decimal

    // Calculated metadata
    RealizedPnL   decimal.Decimal
    UnrealizedPnL decimal.Decimal
    GrossPnL      decimal.Decimal
    Fees          decimal.Decimal
    NetPnL        decimal.Decimal
    OpenedMS      uint64
    ClosedMS      uint64
    UpdatedMS     uint64
}
```

Trade contains no Order collection.

Closed Trade fields are complete final finance and immutable.

#### Order

```go
type Order struct {
    // Identity (denormalized for faster retrieval)
    SweepID     *uint64 // optional
    BotID       uint64
    AccountID   string
    LedgerID    uint64
    TradeID     uint64
    OrderID     uint64

    // General immutable fields
    CycleNumber int
    Symbol      string
    CLOID       string
    Role        string
    Side        string

    // Submitted values
    Type              string
    TimeInForce       string
    SubmittedQuantity decimal.Decimal
    SubmittedPrice    *decimal.Decimal
    TriggerPrice      *decimal.Decimal
    ReduceOnly        bool
    SubmittedMS       uint64
    SubmitJSON        string

    // Updated and final Exchange values
    VenueOrderID uint64
    Status       Status
    RejectReason string
    UpdatedMS    uint64
    AckJSON      string
    VenueJSON    string

    // Calculated metadata
    FilledQuantity  decimal.Decimal
    FilledNotional  decimal.Decimal
    AverageFillPrice decimal.Decimal
    RemainingQuantity decimal.Decimal
    Fees             decimal.Decimal
    FillCount         int
    PendingFeeCount   int
    LastFillMS        uint64
}
```

Order contains no Fill collection, mutation revision, active flag, or separate
reconciliation-pending flag.

`IsClosed` derives closure from stored status, quantity, and fee completeness.

#### Fill

```go
type Fill struct {
    // Identity (denormalized for faster retrieval)
    SweepID     *uint64 // optional
    BotID       uint64
    AccountID   string
    LedgerID    uint64
    TradeID     uint64
    OrderID     uint64

    // General immutable fields
    CycleNumber int
    Symbol      string
    CLOID       string
    VenueOrderID uint64
    VenueTID     uint64
    Side         string

    // Updated and final Exchange values
    Quantity    decimal.Decimal
    Price       decimal.Decimal
    TimestampMS uint64
    Fee         *decimal.Decimal
    Liquidity   string
    VenueJSON   string
}
```

Fill has no Input, Record, State, metadata subobject, or clone.

`Fee == nil` means missing. Zero fee and negative rebate remain present values.

`SweepID` is optional routing metadata. Missing `SweepID` does not invalidate a
record and must not be required for matching, updating, closing, or recovery.

### Target Ledger Structure

```go
type OrderRef struct {
    TradeID uint64
    OrderID uint64
}

type FillRef struct {
    TradeID uint64
    OrderID uint64
    VenueTID uint64
}

type Ledger struct {
    // General identity and policy
    Config Config

    // Canonical records
    Trades map[uint64]*trade.Trade
    Orders map[uint64]*order.Order
    Fills  map[uint64]*fill.Fill // keyed by Venue TID

    // Complete compact identity indexes
    OrderByCLOID map[string]OrderRef
    OrderByOID   map[uint64]OrderRef
    FillByTID    map[uint64]FillRef

    // Active relationship indexes
    OrderIDsByTrade map[uint64][]uint64
    FillTIDsByOrder map[uint64][]uint64

    // Selected working sets
    ActiveTradeIDs  map[uint64]struct{}
    ActiveOrderIDs  map[uint64]struct{}
    PendingOrderIDs map[uint64]struct{}
    PendingFillTIDs map[uint64]struct{}

    // Calculated metadata
    Summary Summary

    // Local identity and reconciliation progress
    NextTradeID    uint64
    NextTradeNo    uint32
    NextOrderID    uint64
    FillsThroughMS uint64
    LastReconMS    uint64
    AccountJSON    string

    Store *ledgerStore
}
```

Actual fields remain private unless another package must read them directly.
The flat structure is the requirement; capitalization is illustrative.

### Local Intent and Submission

```text
Account validates the requested batch
Account allocates Trade, Order, and CLOID identity
Account builds exact submitted Order values
Account encodes exact official Order JSON
Ledger creates complete local Trade and Order records

if persist_mode=max
    begin SQLite transaction
    insert Trade and Orders
    update Ledger next IDs
    commit

publish Trade and Orders to Ledger memory
call Exchange place_order
decode and validate acknowledgement
match each result to existing Order by request position and CLOID
update existing Order acknowledgement fields

if persist_mode=max
    write changed Orders transactionally

mark Account recon-dirty
```

The Exchange response never creates the local Trade or Order.

Unknown or missing acknowledgement preserves recoverable local intent.

### Temporary Recon Evidence

Downloaded evidence exists only for one Recon attempt.

```go
type reconEvidence struct {
    Orders []OrderEvidence
    Fills  []FillEvidence

    OrderByCLOID map[string]int
    OrderByOID   map[uint64]int
    FillByTID    map[uint64]int

    OrderResponseJSON string
    FillResponseJSON  string
    AccountJSON       string
}
```

Slices own each decoded evidence value once.

Maps contain slice indexes, not duplicate evidence structs.

The evidence object is released after success or failure. Admitted row JSON is
copied into its domain record. Exact envelopes and failures go to the raw
evidence owner. No downloaded slice, index map, or decoder object becomes
Ledger storage.

### Recon Flow

```text
download current open Orders
download exact status for active local Orders missing from open Orders
download Fill history using complete bounded time windows
download bounded missing-fee repair windows
download Account balance and position

decode and validate JSON
build temporary evidence indexes

prepare one private Ledger attempt containing dirty copies of the same
Trade, Order, and Fill structs

for each downloaded Fill
    lookup local Fill by Venue TID
    if missing
        resolve owning Order by Fill CLOID
        otherwise resolve the unique local Order by Account and Venue OID
        if the local Order has no Venue OID after a lost acknowledgement
            resolve Fill Venue OID to downloaded Order evidence
            resolve that downloaded Order CLOID to the local Order
            bind the Venue OID to that local Order
        copy the resolved local Order CLOID and Order ID onto the Fill
        ignore or diagnose evidence not owned by this Ledger
        create one Fill
        add Fill to dirty records
        add Fill identity indexes
        update owning Order calculated totals
    else if local Fill fee is missing and downloaded fee is present
        update only fee and raw Fill JSON
        update owning Order fee total and pending-fee count
    mark owning Order and Trade dirty

for each active local Order
    lookup downloaded Order by CLOID, then Venue OID
    update when evidence exists
    use exact orderStatus when absent from open Orders
    calculate IsClosed
    mark owning Trade dirty when changed
    compare temporary Venue executed quantity with admitted local Fill quantity
    halt on unexplained missing execution

for each dirty Trade
    read its active relationship ID indexes
    calculate exact exposure, realized PnL, unrealized PnL, gross PnL,
    fees, net PnL, status, and timestamps
    calculate IsClosed
    apply exact old-to-new Ledger Summary delta

for each active Trade
    update mark-derived unrealized, gross, and net PnL
    apply exact old-to-new Ledger Summary delta

update Account Snapshot candidate

if persist_mode=max
    write dirty Fills, Orders, Trades, Ledger cursors, and Account JSON
    commit one SQLite transaction

publish dirty records, indexes, Summary, cursors, and Account Snapshot
archive children of newly durable closed Trades
release all temporary evidence
```

#### Complete Fill Windows

`userFillsByTime` has a bounded response. One full response does not prove the
requested time range is complete.

```text
query [cursor, attempt boundary] with a fixed end time
if returned rows are below the Exchange limit
    range is complete
else
    split the time range and query both halves
    continue splitting any full result

overlap timestamp boundaries
deduplicate by Venue TID
if one millisecond still returns the Exchange limit
    halt; the range is not retrievable

exact-query every active or unresolved local Order
compare Venue cumulative executed quantity with admitted local Fill quantity
if Venue quantity is greater after all available Fill windows
    halt with a retention gap

advance FillsThroughMS only after every active or unresolved local Order
has no unexplained execution
```

An empty old range is not itself a gap; no Fill may have occurred. Local Order
cumulative execution is the conservative proof. Live never adopts unrelated
Account fills. A detected or unsplittable gap does not advance the cursor or
publish possibly incomplete finance.

### Matching Cost

Example:

```text
Local:
  500 Fills
  30 pending-fee Fills
  600 Orders
  150 active Orders

Downloaded for the shared Exchange Account:
  2,000 Orders
  2,000 Fills
```

Matching:

```text
Build downloaded Fill TID index once.
Loop downloaded Fills once to discover locally owned new Fills.
Loop 30 pending Fill IDs and lookup downloaded evidence by TID.
Loop 150 active Order IDs and lookup downloaded evidence by CLOID or OID.
Loop only dirty Orders.
Loop only dirty Trades.
Loop active Trades once for mark-derived finance.
```

Complexity:

```text
O(downloaded Orders + downloaded Fills + active Orders + pending Fills
  + dirty Orders + dirty Trades + active Trades)
```

There is no active-child times downloaded-history nested search.

### Recovery

Live recovery performs:

```text
load Ledger identity, next IDs, cursors, and Account JSON
load every Trade record
load every Order belonging to active or unresolved Trades
load every Fill belonging to those Orders
load compact CLOID, Venue OID, and Venue TID indexes
rebuild active and pending ID sets
rebuild active relationship indexes
sum every cheap Trade record into exact Ledger Summary
validate complete loaded relationships and finance
force normal Exchange reconciliation
allow decisions only after successful Recon
```

Closed Trade children remain in SQLite and are not loaded.

Closed Trade records contain complete final finance and are trusted.

### Deferred Periodic Integrity Check

This is outside the account-stack hardcut. Add it only after the flat design
passes recovery and Recon proof.

Runner would own the timer. Ledger would own the check mechanics. No interval
or scheduler is added by this implementation.

```text
stream durable Trade rows without retaining closed children
rebuild exact Trade totals
compare with cached Ledger Summary
compare active database rows with active memory records
verify foreign keys and unique identity constraints
verify cursors do not move backward
discard streamed verification rows
```

The periodic check is not routine Recon and does not query bounded Exchange
history to prove the complete archive.

### Persistence Modes

#### `none`

```text
all Trades, Orders, and Fills remain in memory
no runtime Ledger database opens
no runtime Ledger row writes
failed backtest reruns from the beginning
terminal publication remains the backtest evidence owner
```

#### `max`

```text
Trade and Order intent commits before Exchange I/O
every accepted mutation commits transactionally
complete history remains durable
memory publishes only after durable commit
closed Trade children leave memory after commit
restart loads all Trades and active Trade children
```

Live must require `max`.

### Proposed SQLite Shape

The implementation must use the exact Go record fields that are operationally
required. Raw JSON carries remaining Exchange detail.

Required keys and indexes:

```text
account_trade
  primary key: ledger_id, trade_id
  unique: sweep_id, bot_id, account_id, trade_id
  index: ledger_id, status

account_order
  primary key: ledger_id, order_id
  unique: cloid
  unique when nonzero: ledger_id, venue_order_id
  index: ledger_id, trade_id
  index: ledger_id, status

account_fill
  primary key: ledger_id, venue_tid
  index: ledger_id, trade_id
  index: ledger_id, order_id
  index: ledger_id, fee

account_exchange_raw
  primary key: ledger_id, evidence_id
  index: ledger_id, operation, timestamp_ms
  columns: direction, exact_json, success, error
  retention: latest 100 rows per ledger_id and operation
```

Trading decimals remain canonical text.

Foreign keys preserve:

```text
Order -> Trade
Fill -> Order
```

Required raw JSON columns:

```text
account_ledger.account_state_json
account_order.submit_json
account_order.ack_json
account_order.venue_json
account_fill.venue_json
account_exchange_raw.exact_json
```

Do not add columns for Exchange values used only during debugging when the raw
JSON already preserves them.

### Function Design

#### Fill Target API

```text
New
  validate one complete new Fill
  return the admitted Fill

Update
  accept later fee evidence and replacement raw Venue JSON
  leave execution identity, quantity, price, side, and timestamp unchanged

HasFee
  return Fee != nil

Validate
  validate complete new Fill identity and execution values
  use small private validation subfunctions when clearer
```

Current Fill function disposition:

```text
New             KEEP and simplify to one Fill struct
Enrich          RENAME to Update
State           DELETE
HasFee          KEEP
Clone           DELETE
sameExecution   DELETE
validateInput   RENAME or MERGE into Validate
```

#### Order Target API

```text
New
  validate one complete locally submitted Order
  create its initial local status and calculated totals

Update
  apply trusted acknowledgement or normalized Venue status values
  update Venue OID, status, rejection, timestamps, and raw JSON
  return changed

IsClosed
  return true for canceled, rejected, expired, or error
  return true for filled only when submitted quantity is filled and every fee is complete

Validate
  validate submitted identity and request terms
  validate known incoming status
  use private role, side, type, time-in-force, and status helpers

Slippage
  derive signed price and percent slippage from submitted price and weighted Fill price
  return unavailable when required values are absent
```

Current Order function disposition:

```text
New                  KEEP and simplify to one Order struct
RecordSubmit         MERGE into Update
ApplyVenueState      MERGE into Update
ApplyFill            MOVE to Ledger AddFill
RefreshRecon         DELETE; IsClosed and Ledger updates own the result
Record               DELETE
ReconState           DELETE
ComparisonState      DELETE; mutations return changed
FillIdentity         DELETE
ActiveState          DELETE
Summary              DELETE
EachFill             DELETE
Fill                 DELETE
Clone                DELETE
refreshFills         MOVE into Ledger AddFill calculations
validateInput        RENAME or MERGE into Validate
validRole            KEEP private
validSide            KEEP private
validType            KEEP private
validTimeInForce     KEEP private
validStatus          KEEP private
isTerminal           REPLACE with IsClosed
transitionAllowed    DELETE
copyInput            DELETE
```

#### Trade Target API

```text
New
  validate one complete Trade identity
  create one pending Trade record

Update
  calculate exact exposure and finance from Ledger-resolved Orders and Fills
  update only active or unresolved Trade records
  return changed

UpdateMark
  update unrealized, gross, and net PnL from stored exposure and current mark
  never read Orders or Fills

IsClosed
  return true for closed, canceled, or error

Validate
  validate identity and exact calculated invariants
  validate persisted closed finance during recovery
```

Current Trade function disposition:

```text
New                 KEEP and simplify to one Trade struct
AddOrder            MOVE to Ledger AddOrders
Refresh             MERGE into Update
RefreshRecon        MERGE into Update
RefreshMark         RENAME to UpdateMark
MarkedRecord        DELETE
Record              DELETE
ReconState          DELETE
Summary             DELETE; Trade fields are the summary source
Clone               DELETE
Order               DELETE
EachOrder           DELETE
record              DELETE
executions          MOVE into Ledger relationship traversal or Trade Update input preparation
calculate           KEEP private and simplify
calculateFinance    KEEP private
isCloseRole         KEEP private or move beside Order role definitions
isTerminal          REPLACE with IsClosed
sameMetrics         DELETE; Update returns changed
sameSign            KEEP private
```

#### Ledger Target API

```text
Init
  initialize empty backtest memory or restore Live durable state

PlanTrade
  return next Trade and initial Order identities without mutation

PlanOrders
  return later Order identities without mutation

CreateTrade
  store one Trade and initial Orders
  commit Live intent before Exchange I/O

AddOrders
  store later local Order intent under one existing active Trade
  commit Live intent before Exchange I/O

UpdateOrders
  apply trusted ordered acknowledgement values to existing Orders

AddFill
  resolve one existing local Order
  create a new Fill or complete an existing missing fee
  update exact Order calculated totals
  mark owning Order and Trade dirty

Recon
  own one private attempt
  update relevant Fills, Orders, Trades, Summary, indexes, cursors, and persistence
  publish only after success

Trade
  return one Trade record by ID

Order
  return one in-memory Order record by ID

Fill
  return one Fill record by Venue TID from memory

ArchivedOrder
  explicitly read one archived Order from durable storage

ArchivedFill
  explicitly read one archived Fill from durable storage

ActiveOrders
  return current active Order records

PendingCounts
  return current pending Order and Fill counts

PendingFillAnchors
  return missing-fee repair identities and timestamps

HasPendingRecon
  report whether pending work exists

Summary
  return cached exact Ledger totals

Result
  return terminal Ledger identity, cursors, counts, and Summary

Stop
  close owned store resources and stop admission
```

Current Ledger function disposition:

```text
CreateTrade               KEEP and flatten ownership
AddOrders                 KEEP and flatten ownership
RecordSubmit              RENAME to UpdateOrders
Result                    KEEP
Stop                      KEEP
PlanTrade                 KEEP
PlanOrders                KEEP
ActiveOrders              KEEP; return Order records
TradeState                RENAME to Trade
NextBatchNo               KEEP
OpenTrades                RENAME to ActiveTrades if still externally required
CountOrders               KEEP through cached counts or durable query
TradeCount                KEEP
TradeOrders               KEEP only for explicit read/report paths
Orders                    REPLACE with direct Order lookup or explicit batch read
FillsThroughMS            KEEP
Fill                      KEEP
PendingCounts             KEEP
PendingFillAnchors        KEEP
ReconFillChanges          RETURN through Recon result; delete separate read method
HasPendingRecon           KEEP
Summary                   KEEP
validateAddedOrders       KEEP private and simplify for trusted flat records
validateNewOrderIndexes   KEEP private
validateSubmitOutcomes    KEEP private
nextTradeInput            KEEP private
candidate                 DELETE
persistCandidate          DELETE
currentCandidate          DELETE
publish                   DELETE
rebuildIndexes            KEEP for recovery
replaceTradeIndexes       REPLACE with direct flat index updates
addTradeIndexes           REPLACE with indexTrade, indexOrder, and indexFill
refreshTradeIndexes       REPLACE with dirty direct index updates
addValidatedTradeIndexes  DELETE
replaceTradeSummary       KEEP concept; rename updateSummary
removeTradeIndexes        REPLACE with archiveTrade
cloneTrades               DELETE
indexCLOIDs               DELETE
sortedSet                 KEEP private if selected-ID ordering still requires it
tradeIsActive             REPLACE with Trade.IsClosed
```

#### Ledger Recon Function Disposition

```text
Init                    KEEP
Recon                   KEEP as the only public Ledger Recon mutation
PrepareRecon            MAKE private
UpdateReconFills        MAKE private updateFills
UpdateReconOrders       MAKE private updateOrders
evidenceOrder           KEEP private resolver
UpdateReconTrades       MAKE private updateTrades
ReconSummary            DELETE; Recon returns Summary
CommitRecon             MAKE private commit
stageOrder              REPLACE with dirty flat Order copy
refreshReconSummaries   DELETE
sameTradeReconState     DELETE
markedTradeSummary      DELETE
```

#### Store Function Disposition

```text
openLedgerStore        KEEP
close                  KEEP
save                   MERGE into one transaction writer
saveMutation           MERGE into one transaction writer
saveRecon              MERGE into one transaction writer
storeLedgerIdentity    KEEP as one upsert helper
storeReconTrade        RENAME upsertTrade
storeReconOrder        RENAME upsertOrder
storeReconFill         RENAME upsertFill
load                   SPLIT into explicit recovery reads
loadOrders             REPLACE with loadActiveTradeOrders
loadFills              REPLACE with loadActiveTradeFills
nullableUint           KEEP if schema still needs it
nullableText           KEEP if schema still needs it
nullableDecimal        KEEP if schema still needs it
nullableDecimalPointer KEEP
parseOptionalDecimal   KEEP
sortedTradeIDs         DELETE unless deterministic writes still require it
```

New recovery reads:

```text
loadLedger
loadTrades
loadActiveTradeOrders
loadActiveTradeFills
loadIdentityIndexes
```

### Failure and Atomicity

Backtest:

```text
direct memory mutation is allowed
any failure terminates the run
rerun reconstructs from the beginning
```

Live:

```text
prepare dirty copies of only changed flat records
validate the complete attempt
write dirty rows and Ledger cursor in one SQLite transaction
commit
publish the same record values into memory
publish Account Snapshot last
```

No full graph clone or rollback graph exists.

A Live persistence failure publishes neither memory changes nor Account Snapshot.

### Hardcut and Cutover

- Replace the nested object graph in one coherent implementation.
- Do not preserve old Input, Record, State, ReconState, Clone, callback traversal,
  candidate, or public staged-Recon APIs.
- Do not maintain both nested and flat ownership.
- Update Account, Executor, ResultPublisher, report, tests, schemas, and design
  in the same implementation.
- Existing Live data compatibility or migration requires separate explicit user
  approval.
- Backtest stored templates and accepted trading behavior remain unchanged.

### Proof Requirements

#### Static

```text
gofmt
project diagnostics
git diff --check
no old nested ownership references
no Clone functions
no Input, Record, State, or ReconState domain duplicates
no public staged-Recon methods
no unsafe, CGO, assembly, or new dependency
```

#### Focused

- Fill New, Update, HasFee, zero fee, and negative rebate.
- Order submitted-versus-filled values and exact slippage.
- Order IsClosed across filled, fee-pending, canceled, rejected, expired, and error.
- Trade exact long, short, partial close, complete close, fees, and immutable final finance.
- Mark-only Recon updates every active Trade and exact Ledger Summary.
- Ledger flat record uniqueness and full identity.
- CLOID-first and Account-scoped Venue-OID Fill routing.
- Lost acknowledgement routes Fill OID through downloaded Order CLOID.
- OID-resolved Fill stores the canonical local Order CLOID and Order ID.
- Venue TID Fill deduplication.
- Full Fill response splits time windows and never advances an unproven cursor.
- Retention gaps halt without publishing incomplete finance.
- One-millisecond full Fill windows halt without further splitting.
- Venue cumulative execution proves no locally owned Fill is missing.
- Unknown Exchange Order rejection without adoption.
- Dirty Fill to Order to Trade propagation.
- Exact cached Summary deltas.
- Live intent commit before Exchange I/O.
- Live dirty-row transaction before memory publication.
- Recovery loads all Trades and only active Trade children.
- Closed Trade children remain in SQLite and absent from memory.
- Exact raw submitted, acknowledgement, Order, Fill, Account, batch, and failed JSON round-trip.
- `none` opens no runtime database.
- `max` survives restart with identical active state and Summary.

#### Integrated

```text
CGO_ENABLED=0 go test -count=1 -tags noasm ./...
CGO_ENABLED=0 go vet -tags noasm ./...
bash -n stest.sh
git diff --check
./stest.sh -bot 9
./stest.sh -bot 13
./stest.sh -bot 15
```

Require exact accepted Trade, Order, Fill, finance, equity, drawdown, Recon, and
result proof.

#### Performance

Measure:

```text
downloaded Order rows
downloaded Fill rows
locally matched rows
new Fills
fee completions
active Orders
dirty Orders
dirty Trades
rows written
Recon duration
Recon allocation
Live recovery rows retained versus streamed
```

Prove routine Recon performs no nested active-record times downloaded-history
search and no complete historical child load.

### Decisions Still Requiring Implementation-Time Confirmation

None.

Live uses a hardcut schema version. An incompatible existing database fails
startup with an explicit version error. No automatic migration, compatibility
reader, dual write, or destructive reset is part of this implementation.

These decisions must not reopen the confirmed ownership, memory, CLOID, raw JSON,
closed-Trade immutability, or persistence boundaries.

## Adversarial Review

Independent read-only review completed before implementation.

Round 1: FAIL.

- Added active-Trade mark updates so mark-only Recon refreshes PnL and Summary.
- Added complete bounded Fill windows and safe cursor advancement.
- Added lost-ack OID-to-downloaded-Order-to-CLOID recovery.
- Added exact raw envelope and failed-payload ownership.
- Removed synthetic FillID; Venue TID is the Fill identity.
- Deferred periodic integrity scheduling.
- Made archive reads explicit.

Round 2: FAIL.

- Added conservative missing-Fill proof from Venue cumulative Order execution.
- Added fail-stop behavior for a full one-millisecond Fill window.
- Bounded raw envelopes to the latest 100 per Ledger and operation.
- Defined request, response, success, and failure transaction ownership.
- Made normal Order lookup memory-only.
- Chose fail-fast hardcut schema versioning with no implicit migration.

User clarification:

- Nuubot always submits Orders with CLOID.
- Raw Exchange Fill evidence may omit CLOID.
- Every admitted local Fill stores CLOID.
- Resolve Fill by CLOID when present, otherwise by Account and Venue OID.
- After OID resolution, copy the local Order CLOID and Order ID onto the Fill.

Reference implementation checked:

- `nuutrader6` requires CLOID before Order submission.
- `nuutrader6` accepts raw Fill evidence with empty CLOID.
- `nuutrader6` resolves Fill by CLOID first, then unique Account-scoped OID.

Round 3: PASS.

No production implementation was performed.
