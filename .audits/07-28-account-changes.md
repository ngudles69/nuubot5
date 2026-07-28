# Account Changes

Date: 2026-07-28

## Result

```text
Delete: 8 dead or wrong Account functions
Add:    explicit persistence call and closed-Trade result count
Change: Account lifecycle organization, Order preparation, raw Account evidence, and callers
Keep:   Account-owned CLOID, Venue requests, event-driven Recon entrypoint, and flat Ledger boundary
```

No production or test file changed during this design pass.

## 1. Delete Dead or Wrong Functions

Delete:

- `Result.Clone`
- `CountOrders`
- `TradeOrders`
- `Order`
- `Fill`
- `PositionQuantity`
- `HasPendingRecon`
- `copySpec`

Caller changes:

```text
GridExecutor CountOrders(TP, filled)
  -> terminal Account result
  -> Ledger closed-Trade count
  -> RoundTrips
```

Account tests using `Order` or `Fill` inspect same-package Ledger state.

**KEY DECISION:** Test convenience never creates a production Account method.

**KEY DECISION:** `RoundTrips` equals the number of closed Trades.

## 2. Keep Account Event-Driven

Account has no `Start` method.

Account starts no goroutine, timer, worker, or polling loop.

Keep these event operations:

```text
Init
PlaceOrders
CancelOrders
Reconcile
Persist
Stop
```

Pseudocode:

```text
Init(config):
    validate identity and policy
    initialize Ledger
    initialize selected Venue
    mark Account initialized

event arrives:
    caller invokes one Account method
    Account performs one bounded operation
    return
```

Move `Init` from `recon.go` to `account.go`.

Keep the Recon process in `recon.go`.

**KEY DECISION:** Account is event-driven. It owns no running process.

## 3. Keep Order and Venue Ownership

Account remains the only Order-submission owner.

Pseudocode:

```text
PlaceOrders(specs):
    prepare and validate one batch
    ask Ledger for Trade and Order identities
    create every CLOID
    create local flat Order records
    create exact Hyperliquid Venue requests
    store local intent
    submit ordered Venue batch
    decode exact Venue response
    store submission evidence
    return flat Order results
```

Keep:

- `prepareOrderBatch`, renamed from `normalizeSpecs`
- `createCLOID`
- `venueOrderRequest`
- `venueGrouping`
- `markPrice`

Executor supplies Order intent and `OrderLevel`.

Account creates CLOID, batch identity, local Order intent, and Venue request.

`cloid` performs mechanical encoding only.

**KEY DECISION:** Parent Account creates the CLOID before submitting its child Venue Order.

**KEY DECISION:** Simulator and live Hyperliquid return shapes follow the same Account decode path.

## 4. Keep Ledger as Books

Ledger owns:

- flat Trade, Order, and Fill records;
- memory storage and indexes;
- SQLite load and save;
- generated database row IDs;
- scoped queries;
- standard PnL calculations requested by callers; and
- terminal summaries.

Ledger does not own:

- Recon scheduling;
- Venue calls;
- matching;
- status decisions;
- repair policy;
- Account lifecycle; or
- background processing.

Account Recon supplies validated evidence to Ledger.

Ledger records that evidence and calculates requested accounting values.

**KEY DECISION:** Ledger is the accounting books, not the accounting process.

**KEY DECISION:** Recon telemetry and telemetry storage remain Recon-owned.

## 5. Replace Order Counting

Delete:

```go
func (a *Account) CountOrders(role string, status order.Status) uint64
```

Delete the matching Ledger `CountOrders` traversal.

Add one terminal result value:

```text
Ledger.Result.ClosedTrades
```

Pseudocode:

```text
Ledger.Result:
    count Trades whose status is closed
    publish ClosedTrades

GridExecutor Stop:
    result = Account.Result()
    RoundTrips = result.Ledger.ClosedTrades
```

Affected callers:

- `internal/executor/grid.go`
- `internal/executor/executor.go`
- `internal/resultpublisher/resultpublisher.go`

**KEY DECISION:** Do not infer completed trading from Take-Profit Order counts.

## 6. Store Latest Raw Account Evidence

Account Recon receives the relevant raw Account API payload.

Store only the latest payload as JSON.

Pseudocode:

```text
download Account state:
    retain exact response JSON
    decode normalized Account values
    pass both to Ledger

Ledger:
    replace current account_snapshot.raw_json
```

Rules:

- Order stores its latest raw Order JSON.
- Fill stores its raw Fill JSON.
- Account snapshot stores its latest relevant raw Account JSON.
- Trade stores no Exchange payload.
- Ledger itself has no Exchange payload.
- JSON is stored without translation into another evidence format.
- No raw-payload history table exists.

**KEY DECISION:** Normalized fields support operation. Raw JSON preserves Exchange evidence.

## 7. Add Explicit Persistence

Add persistence to the private Venue contract:

```go
Persist(mode persist.Mode) error
```

Add one Account method:

```text
Account.Persist:
    read canonical configured mode
    call Venue Persist(mode)
    return the error
```

Account does not decide whether Simulator state is dirty.

Simulator `Persist(mode)` owns that check.

The Bot loop chooses the call boundary.

Current target is once after loop work and once before final close.

`Stop` closes Venue and Ledger resources. It starts no hidden persistence policy.

**KEY DECISION:** Persistence is an explicit event call, not a background Account process.

## 8. Keep the Required Account Menu

Keep:

- `Init`
- `PlaceOrders`
- `CancelOrders`
- `Reconcile`
- `Result`
- `Telemetry`
- `ReconciliationTelemetry`
- `Persist`
- `Stop`
- `ActiveOrders`
- `Trade`
- `OpenTrades`

Keep `ActiveOrders`, `Trade`, and `OpenTrades` only because current Executor paths require them.

Add nothing for possible future callers.

**KEY DECISION:** If a new Account method is needed later, add it when a real caller exists.

## 9. Upstream and Downstream Impact

### Executors

- Trade and Grid keep submitting typed Order intent.
- Grid reads `ClosedTrades` for `RoundTrips`.
- Remove Account result clone calls when returned values are already detached.
- Keep shutdown checks through Snapshot, `ActiveOrders`, and `OpenTrades`.

### BotCycle and Controller

- Keep Account Recon invocation event-driven.
- Add the smallest loop-end Account persistence call.
- Add no Account worker, queue, mutex, or timer.

### Simulator and Hyperliquid

- Both implement the same private Venue request and response boundary.
- Both return exact Hyperliquid-shaped payloads.
- Simulator implements `Persist(mode)`.
- Account never reads Simulator private structures.

### Ledger

- Remove `CountOrders`.
- Add terminal closed-Trade count.
- Keep latest Account raw JSON.
- Keep storage and accounting calculations separate from Recon decisions.

### ResultPublisher

- Preserve the existing `RoundTrips` output field.
- Its meaning changes to closed Trade count.

## 10. Implementation Files

Production:

- `internal/account/account.go`
- `internal/account/recon.go`
- `internal/account/ledger/ledger.go`
- `internal/account/ledger/store.go`
- `internal/executor/grid.go`
- affected Bot-loop persistence caller
- affected private Venue implementations

Tests:

- Account caller tests for removed methods
- Grid round-trip result proof
- Account raw-payload replacement proof
- Account persistence delegation proof

Documentation:

- Account package design
- Account Ledger package design
- Executor package design
- owning persistence and trading-schema pages
- `HANDOFF.md`

## 11. Proof

Required:

```text
Account has no Start, goroutine, timer, or background loop
Account creates every submitted CLOID
Venue requests remain exact Hyperliquid shapes
dead Account methods have no production callers
CountOrders is absent
RoundTrips equals closed Trade count
latest Account raw JSON replaces the prior value
no Account payload-history table exists
Ledger performs storage and requested accounting only
Recon behavior remains Recon-owned
Persist delegates once with the canonical mode
focused Account, Ledger, Executor, and ResultPublisher tests pass with -tags noasm
full tests pass with -tags noasm
full vet passes with -tags noasm
Observer, Trade, and Grid sweeps pass
two clean rebuild.sh rounds recreate and run all three template sweeps
```

No interface, repository layer, event bus, Account worker, migration, repair engine, compatibility bridge, or payload-history system is added.
