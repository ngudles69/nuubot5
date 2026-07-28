# Simulator Audit and Implementation Blueprint

Date: 2026-07-28
Status: FAIL — Backtest behavior is proven, but the approved Simulator target is not implemented.
Scope: `internal/simulator/*.go`, every upstream and downstream caller, canonical design, persistence, MarketData, Clock, Hyperliquid wire ownership, and proof.

This report contains:

- current-state evidence;
- all 42 production-function dispositions;
- the user-approved target;
- holistic upstream and downstream impact;
- exact implementation details;
- implementation order; and
- required proof.

No production or canonical design file changed during this audit.

## Approved Target Summary

Account continues to own Simulator selection and lifetime.

One durable simulated account record is identified by:

```text
SweepID + BotID + Network + AccountName
```

Symbol and Executor identity are not part of that account identity.

Simulator remains one Account-owned object per current Executor-cycle.

For the first implementation, Simulator intentionally ignores `PersistMode=none` and writes dirty records to SQLite.

This temporary behavior preserves balance, Orders, Fills, positions, and Venue counters across current Account recreation.

A later Account/Simulator lifetime redesign must restore true memory-only Backtest behavior.

Simulator memory owns:

- one open-Order collection;
- one closed-Order collection;
- one Fill-history collection;
- per-symbol position state;
- account balance;
- Venue OID and TID allocation; and
- dirty state per changed record.

Persistence is incremental:

```text
Persist(mode)
  ignore mode temporarily
  return when nothing is dirty
  begin one SQLite transaction
  upsert dirty account and position rows
  insert or update dirty Order rows
  insert dirty Fill rows
  commit
  assign generated database IDs
  clear committed dirty flags
```

Three changed Orders produce three Order upserts, even when 8,000 clean Orders exist.

Simulator always opens its database and loads scoped state during `Init`.

Recovery trusts the database:

- check schema version;
- select by SweepID, BotID, Network, and AccountName;
- decode required values;
- partition Orders by status;
- rebuild memory indexes;
- derive next Venue OID and TID from stored maximums; and
- continue.

Recovery adds no repair engine, replay validation, or repeated ownership checks.

Nuubot supplies its canonical Clock.

Simnet paper/test receives WallClock.

Backtest receives TickClock.

Simulator never reads system time directly.

Simulator returns exact official Hyperliquid JSON through canonical `internal/hyperliquid` wire types.

Simulator never returns its private storage structures.

## Confirmed Temporary Architectural Debt

Current Account and Simulator objects die at every BotCycle boundary.

Controller currently carries only ending equity into the next cycle.

Without database writes, current `none` mode loses Simulator Order history, Fill history, OID/TID continuity, and private state.

The approved temporary hardcut is:

```text
none and max both persist dirty Simulator records
```

Ledger persistence policy is unchanged by this temporary Simulator exception.

Future work must choose a Bot-lifetime Account/Simulator or an explicit in-memory state handoff.

That future change must remove the temporary `none` write behavior.

## Application Intent

Nuubot runs one configured Bot through Backtest or Live orchestration.

Executors own Accounts.

Each Account owns one Ledger and one selected Venue.

Current trading Accounts admit only `simulator/simnet`.

Account sends official Hyperliquid-shaped actions to Simulator.

Simulator returns detached official-shaped JSON.

Account decodes that untrusted JSON and reconciles it into local Ledger truth.

MarketData owns BBO ingestion and synchronous subscriber notification.

Backtest is synchronous and proven.

Live execution remains unavailable.

## Current Simulator Responsibility

Simulator currently owns one simulated Account-symbol Venue:

- official Order admission;
- Venue OID, TID, and private batch allocation;
- canonical private Order and Fill records;
- CLOID and active-Order indexes;
- bracket arming and sibling cancellation;
- BBO-driven matching;
- signed position, entry, realized PnL, and fees;
- official-shaped query JSON;
- one MarketData subscription;
- transient latest-BBO state; and
- optional schema-version-3 SQLite persistence.

Simulator owns no Ledger, local Trade, local Order, local Fill, role, purpose, or terminal Account result.

Account owns Simulator lifetime and its dirty callback.

## Current Ownership and Call Flow

```text
Backtest Run
  -> MarketData.IngestBBO
     -> Simulator.onBBO
        -> match active Venue Orders
        -> append Venue Fills
        -> update Venue position
        -> Account dirty callback

Executor
  -> Account.PlaceOrders
     -> venue.PlaceOrders
        -> Simulator.PlaceOrders

Executor
  -> Account.CancelOrders
     -> venue.CancelOrders
        -> Simulator.CancelOrders

Account.Reconcile
  -> Simulator.OpenOrders
  -> Simulator.OrderStatus
  -> Simulator.Fills
  -> Simulator.AccountState
  -> Ledger atomic Recon

Account.Stop
  -> Simulator.Stop
  -> Ledger.Stop
```

`internal/account.initializeVenue` creates the concrete Simulator and stores it behind Account's private `venue` interface.

No production caller imports Simulator except Account.

## Current Public API

```go
func (s *Simulator) Init(Config) error
func (s *Simulator) PlaceOrders(hyperliquid.PlaceOrderAction, uint64) ([]byte, error)
func (s *Simulator) CancelOrders(hyperliquid.CancelByCLOIDAction, uint64) ([]byte, error)
func (s *Simulator) OpenOrders(string) ([]byte, error)
func (s *Simulator) Fills(string, uint64, uint64) ([]byte, error)
func (s *Simulator) OrderStatus(string, string) ([]byte, error)
func (s *Simulator) AccountState(string) ([]byte, error)
func (s *Simulator) Stop() error
func (s *Simulator) SetFillFeeAvailableForTest(uint64, bool) error
```

All mutation and query methods require successful `Init` and reject a stopped Simulator.

`SetFillFeeAvailableForTest` is the sole exception. It performs no lifecycle check.

## Nuutrader6 Compatibility Reference

Required reference:

```text
D:\rust\nuutrader6\src\nuubot\hcbots\simulator.py
D:\rust\nuutrader6\src\nuubot\hcbots\exchange.py
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
D:\rust\nuutrader6\.project\plans\runtime\hcbot-account-exchange-simulator-parity.md
```

Nuutrader6 establishes these external expectations:

- Account alone owns Simulator selection and lifetime.
- Strategy code does not call Simulator directly.
- Simulator accepts the same Account-facing order, cancel, and truth-read operations as live Exchange.
- Submit and cancel use Hyperliquid-shaped status envelopes.
- Unknown cancellation produces an item error status.
- A market tick is required before submission.
- Limit buys cross best ask; limit sells cross best bid.
- Market Fills use best ask or best bid.
- Trigger Fills use trigger price plus configured adverse slippage.
- Trigger children expose `waitingForFill` and `waitingForTrigger` internally.
- BBO processing fills crossed Orders in retained insertion order.
- Reduce-only quantity is capped by current exposure.
- Fills, Order history, position, fees, and state survive configured persistence.
- Unsupported commands fail loudly.

Current Nuubot-approved behavior intentionally differs:

- Nuubot MarketData currently supplies one exact price, not separate bid and ask.
- Nuubot returns only official `resting`, `filled`, or `error` submit statuses.
- Nuubot keeps trigger waiting and arming private.
- Nuubot numeric Venue TID is the official Fill identity.
- Nuubot Fill JSON may omit CLOID; Account resolves OID and enriches locally.
- Nuubot realized PnL excludes fees; fees remain separate and net PnL subtracts them once.
- Nuubot latest BBO is transient across recovery.
- Nuubot accepted Trade and Grid counts and finance are existing behavior proof.

Those approved Nuubot differences remain compatibility constraints.

Nuutrader6 implementation structure is not a porting target.

## Complete Function Inventory and Disposition

`KEEP` means retain current responsibility and semantics.

`DEAD` means no production requirement justifies the current function.

`MODIFY` means preserve stated external compatibility while changing confirmed behavior or internal mechanics.

### `internal/simulator/simulator.go`

| Function | Behavior | Production callers | Test callers | Class | Why and affected callers |
|---|---|---|---|---|---|
| `(*Simulator).Init` | Validates Config, initializes state, optionally restores v3, subscribes to MarketData, and reads latest BBO. | `account.initializeVenue`. | `newSimulator`; indirectly every Account test that calls `Account.Init`. | `MODIFY` | Always open the Bot database, load the scoped account, inject Clock, and initialize only when no stored account exists. |
| `(*Simulator).PlaceOrders` | Validates one official action, allocates private batch/OIDs, stages Orders, immediately matches eligible Orders, persists, commits, and returns ordered JSON statuses. | `Account.PlaceOrders` through `venue.PlaceOrders`. | Direct: `TestSimulatorOwnsCanonicalOrdersAndDetachedOfficialJSON`, `TestSimulatorPersistsEachCanonicalRecordOnce`, `TestSimulatorTreatsOfficialCLOIDAsOpaque`, `TestSimulatorPersistenceFailureDoesNotAdmitMutation`. Indirect: Account bracket, recovery, uncertainty, and immediate-Fill tests. | `MODIFY` | Preserve official status envelope and current proven finance; align missing-market and filled-fee behavior with the reference where Nuubot proof does not differ. Internal staging is a confirmed scaling issue. |
| `(*Simulator).CancelOrders` | Validates all active CLOIDs, stages cancellation, cascades limit-parent cancellation to trigger children, persists, commits, and returns successes. | `Account.CancelOrders` through `venue.CancelOrders`. | No direct Simulator test. Indirect Trade/Grid shutdown system paths. | `MODIFY` | Nuutrader6 returns per-item error statuses; current code returns a batch error before JSON. Parent-child cascade is current Nuubot-approved behavior. Affects Account cancel decoding and shutdown proof. |
| `(*Simulator).Stop` | Stops subscription, persists, closes store, then marks stopped; completed repeats are nil. | `Account.Stop`. | Direct: `TestSimulatorPersistsEachCanonicalRecordOnce`. Indirect Account/system shutdown. | `MODIFY` | Connection parity remains implementation-tested. Stop must not own persistence policy; caller performs final `Persist(mode)` before close. |
| `(*Simulator).onBBO` | Reads latest BBO, warms first state, stages ordered matching, persists change, commits mark state, and notifies Account. | MarketData callback registered by `Init`; Backtest via `MarketData.IngestBBO`. | `ingestSimulatorBBO`, used by all four core Simulator behavior tests as applicable; Account tests via `ingestAccountBBO`. | `MODIFY` | Preserve every-update synchronous delivery and accepted one-price results. Remove confirmed full-history staging and repeated-scan cost. Live serialization remains an affected caller contract. |
| `(*Simulator).OpenOrders` | Sorts all nonterminal Orders by OID and returns detached official-shaped JSON. | `Account.downloadOrderEvidence`. | Direct: `TestSimulatorOwnsCanonicalOrdersAndDetachedOfficialJSON`. Indirect every executed Account Recon test. | `MODIFY` | External meaning is compatible; repeated allocation/sort is measured. Affects Recon and profile counts only. |
| `(*Simulator).Fills` | Linearly filters all retained Fills by inclusive time and returns detached official-shaped JSON. | `Account.pullFillEvidence`. | Direct: `TestSimulatorOwnsCanonicalOrdersAndDetachedOfficialJSON`. Indirect Account immediate-Fill and delayed-fee tests. | `MODIFY` | Preserve current/Nuutrader6 inclusive returned evidence; internal range lookup and official retention limits need explicit proof. Affects Recon pagination and fee repair. |
| `(*Simulator).OrderStatus` | Resolves CLOID and returns `unknownOid` or exact official-shaped status JSON. | `Account.downloadOrderEvidence` exception path. | Direct: `TestSimulatorOwnsCanonicalOrdersAndDetachedOfficialJSON`. Indirect missing/submitting Order Account tests. | `MODIFY` | One timestamp and terminal `sz` currently diverge from official shape. Affects Account exact-status validation and raw evidence. |
| `(*Simulator).AccountState` | Marks maintained position to latest price and returns clearinghouse-shaped equity, margin, position, and withdrawal JSON. | `Account.downloadAccountState`. | Direct: `TestSimulatorTreatsOfficialCLOIDAsOpaque`. Indirect every successful Account Recon test. | `MODIFY` | Preserve finance, but use per-symbol positions, injected Clock, and canonical Hyperliquid wire encoders. |
| `(*Simulator).SetFillFeeAvailableForTest` | Mutates one Fill fee-availability flag without lifecycle, staging, or persistence. | None. | `account.setFillFeeAvailableForTest`, used by delayed-fee Account tests. | `DEAD` | Test-only production API violates the real exchange surface. Only Account test setup is affected. |
| `(*Simulator).validateAccount` | Rejects invalid lifecycle or nonmatching Account. | `OpenOrders`, `Fills`, `OrderStatus`, `AccountState`. | Indirect through the direct query tests above and all Account Recon tests. | `KEEP` | Exact single-Account boundary is required by both implementations. |
| `(*Simulator).validateRequest` | Validates asset, opaque mandatory CLOID, positive decimals, one type, supported TIF, and trigger kind. | `PlaceOrders`. | Indirect through all direct `PlaceOrders` tests and Account placement tests. | `MODIFY` | `positionTpsl` and trigger-market inputs are admitted elsewhere but lack distinct semantics. Affected caller is Account request shaping. |
| `(*Simulator).newOrder` | Parses one request into a private open Order and exact comparison key. | `PlaceOrders`. | Indirect through all direct `PlaceOrders` tests and Account placement tests. | `MODIFY` | Current timestamp conflates submit, arm, and terminal events. Affects open/status JSON, matching eligibility, persistence, and recovery. |
| `(*Simulator).match` | Repeatedly sorts open Orders and fills/cancels the first eligible OID, at most one leg per batch per BBO. | `onBBO`. | Direct path: `TestSimulatorOwnsCanonicalOrdersAndDetachedOfficialJSON`; indirect Trade/Grid system proof. | `MODIFY` | Preserve current OID order and one-leg-per-batch Nuubot behavior; measured sort/scan cost is material. |
| `(*Simulator).matchAdded` | Immediately executes IOC or marketable newly added Orders in request order, one leg per batch. | `PlaceOrders`. | Direct path: all four `PlaceOrders` Simulator tests; indirect Account immediate-Fill tests. | `MODIFY` | Preserve Nuubot batch behavior; align filled response evidence and mutation staging. |
| `(*Simulator).fill` | Applies slippage and fee, appends one Fill, updates position, terminalizes Order, and changes bracket children. | `match`, `matchAdded`. | Indirect canonical, persistence, Account finance, Trade, and Grid tests. | `MODIFY` | Preserve accepted Nuubot finance, which correctly separates fees from realized PnL unlike Nuutrader6. Internal mutation/persistence and event timestamps affect callers. |
| `(*Simulator).cancel` | Removes one active Order and marks it canceled and unarmed. | `CancelOrders`, `match`, `matchAdded`, `cancelChildren`. | Indirect canonical bracket and system shutdown paths. | `MODIFY` | Behavior remains; active-index and event-time changes affect all four callers. |
| `(*Simulator).armChildren` | Arms active trigger children for one private batch and changes matching timestamp. | `fill` after limit Fill. | Indirect `TestSimulatorOwnsCanonicalOrdersAndDetachedOfficialJSON` and bracket system proof. | `MODIFY` | Responsibility remains, but arming time must stop overwriting submitted time. |
| `(*Simulator).cancelChildren` | Sorts and cancels every trigger child in one batch. | `CancelOrders`, `fill`. | Indirect canonical bracket and shutdown proof. | `MODIFY` | Preserve Nuubot OCO/cascade behavior; sorting/index impact reaches both callers. |
| `(*Simulator).executableQuantity` | Returns remaining quantity or exposure-capped reduce-only quantity. | `match`, `matchAdded`. | Indirect canonical reduce-only bracket and Account/system tests. | `KEEP` | Matches Nuutrader6 and current accepted behavior. |
| `(*Simulator).sortedActiveOrders` | Allocates and sorts active map values by OID. | `match`, `cancelChildren`, `OpenOrders`. | Indirect canonical, Account Recon, Trade/Grid profiles. | `DEAD` | Its work is the measured 4.516s Grid flat hotspot; active ordering is an index responsibility. All three callers are affected. |
| `(*Simulator).position` | Replays retained Fills, validates derivations/TIDs, and rebuilds position. | `restore`. | Indirect `TestSimulatorPersistsEachCanonicalRecordOnce`. | `DEAD` | Recovery trusts persisted account and position rows. Full Fill replay and repair validation are explicitly rejected. |
| `(*Simulator).persist` | No-ops for `none`; otherwise writes a complete JSON snapshot. | `onBBO`, `PlaceOrders`, `CancelOrders`, `Stop`. | Direct max paths: persistence and failure-atomicity tests; indirect max Account recovery tests. | `MODIFY` | Export as `Persist(mode)`. Temporarily ignore mode, transactionally upsert only dirty records, and clear dirty flags after commit. |
| `(*Simulator).storedState` | Copies all Orders and Fills into v3 persistence DTOs. | `Init`, `persist`. | Indirect persistence, recovery, and failure tests. | `DEAD` | Normalized dirty rows replace the complete JSON snapshot and its full-history copy. |
| `(*Simulator).stage` | Returns self for `none`; clones all mutable history and indexes for `max`. | `PlaceOrders`, `CancelOrders`, `onBBO`. | Indirect all mutation tests; max failure test directly proves its atomic intent. | `DEAD` | Atomicity is required; complete-state cloning is not. All three mutation callers and max proof are affected. |
| `(*Simulator).commit` | Replaces live state with a full staged clone in max mode. | `PlaceOrders`, `CancelOrders`, `onBBO`. | Indirect all mutation and max failure tests. | `DEAD` | Live memory mutates directly. SQLite commit controls only dirty-flag clearing and generated database-ID assignment. |
| `(*Simulator).restore` | Validates counters/identity, rebuilds indexes/Fills, and replays position. | `Init`. | Indirect `TestSimulatorPersistsEachCanonicalRecordOnce` and max Account recovery tests. | `MODIFY` | Replace JSON replay with scoped SQL loads, status partitioning, index rebuild, and maximum Venue-ID derivation. |
| `restoreOrder` | Parses one stored Order, validates core fields, and rebuilds comparison key. | `restore`. | Indirect persistence round-trip and max Account recovery tests. | `MODIFY` | Decode one trusted SQL row and rebuild its comparison key. Add no exhaustive semantic validation. |
| `restoreFill` | Parses one stored Fill into private Venue evidence. | `restore`. | Indirect persistence round-trip and max Account recovery tests. | `MODIFY` | Decode one trusted SQL row. Add no repair or invariant replay. |
| `orderPrice` | Returns trigger price for trigger Orders, otherwise submitted price. | `OpenOrders`, `newOrder`, `fill`; `decimalCrosses` test helper. | Direct exact crossing tests through `decimalCrosses`; indirect canonical JSON and Fill tests. | `DEAD` | Keep submitted limit, trigger threshold, and Fill execution price separate. Callers select the exact field they require. |
| `crosses` | Applies IOC or exact limit/TP/SL threshold comparison. | `match`, `matchAdded`. | `TestKeyCrossingMatchesDecimalCrossing`, `TestComparisonKeyCrossingAllocatesNothing`, `FuzzComparisonKeyMatchesDecimal`, `BenchmarkComparisonKeyCrossing`. | `KEEP` | Exact results and zero allocation are directly proven; hotspot time comes from 77.7m calls, not incorrect comparison mechanics. |
| `newComparisonKey` | Creates a positive-decimal comparison key. | `onBBO`, `newOrder`, `restoreOrder`. | Direct comparison, crossing, allocation, fuzz, benchmark, and persistence-key checks. | `KEEP` | Exact ordering and zero-allocation crossing are proven. |
| `compareComparisonKeys` | Compares two positive keys without decimal arithmetic. | `crosses`. | Direct comparison, crossing, allocation, fuzz, benchmark, and persistence-key checks. | `KEEP` | Exact behavior is proven. |
| `closePnL` | Calculates realized PnL on only closed quantity. | `fill`, `position`. | Indirect canonical, persistence, Account finance, Trade, and Grid proof. | `KEEP` | Current Nuubot finance is approved and avoids Nuutrader6's fee-in-realized plus separate-fee ambiguity. |
| `fillDirection` | Maps prior position and side to official-style direction text. | `fill`, `position`. | Indirect canonical Fill and persistence proof. | `KEEP` | Matches required response meaning. |
| `sideCode` | Maps buy/sell to `B`/`A`. | `OpenOrders`, `Fills`, `OrderStatus`. | Indirect canonical JSON and Account Recon tests. | `KEEP` | Matches Nuutrader6 and Hyperliquid shape. |
| `sameSign` | Tests whether two nonzero decimals share sign. | `fill`, `position`. | Indirect finance, persistence, Trade, and Grid proof. | `KEEP` | Minimal shared finance mechanic. |
| `validCLOID` | Validates opaque 128-bit hex shape without domain decoding. | `validateRequest`. | Direct `TestSimulatorTreatsOfficialCLOIDAsOpaque`; indirect all placements. | `KEEP` | Preserves Account-created CLOID opacity and current project requirement. |

### `internal/simulator/store.go`

| Function | Behavior | Production callers | Test callers | Class | Why and affected callers |
|---|---|---|---|---|---|
| `openSimulatorStore` | Opens one one-connection SQLite store and creates the v3 JSON table. | `Simulator.Init` for `max`. | Indirect persistence/failure Simulator tests and max Account recovery tests. | `MODIFY` | Open for every mode and create normalized account, position, Order, and Fill tables. |
| `(*simulatorStore).close` | Closes the SQLite handle and wraps failure. | `Simulator.Init` cleanup; `Simulator.Stop`. | Indirect persistence round-trip and Stop paths. | `KEEP` | Concrete close ownership is correct; caller lifecycle handling is the issue. |
| `(*simulatorStore).save` | Marshals all state and upserts one Account-symbol row using wall-clock update time. | `Simulator.Init`; `Simulator.persist`. | Indirect persistence, failure atomicity, and max Account recovery tests. | `MODIFY` | Replace with one transaction that writes only dirty rows and uses supplied Clock time. |
| `(*simulatorStore).load` | Loads one v3 row, rejects schema mismatch, decodes JSON, and validates Config identity/policy. | `Simulator.Init`. | Direct inspection in `TestSimulatorPersistsEachCanonicalRecordOnce`; indirect max Account recovery tests. | `MODIFY` | Query scoped normalized rows, trust database constraints, decode, and return. |

### Exact Account integration test callers

These tests call the concrete Simulator through Account:

```text
TestAccountRunsOneReconciledBracket
TestAccountSchedulesDirtyReconAndCleanSweep
TestAccountReconFailurePublishesOnlyTelemetry
TestAccountMaxPersistenceRecoversDirtyVenueState
TestAccountMaxPersistenceRecoversSimulatorSubmitFailure
TestAccountDoesNotMarkAcceptedSimulatorOrderRetriable
TestAccountReconRetainsImmediateFillAfterSubmitPersistenceFailure
TestAccountReconTelemetryReportsNoBulkOrderStatusQueries
TestAccountReconTelemetryPreservesFailedOrderStatusQuery
TestAccountMaxPersistenceRepairsMissingSubmittingOrder
TestAccountRepairsCursorAdvancedMissingFee
TestAccountKeepsFeeIncompleteClosurePendingAndFinalFinanceStatic
```

Exact operation coverage:

- All listed tests initialize a concrete Simulator before any test-local Venue interception.
- Concrete placement occurs in every listed test except dirty/clean scheduling.
- Every listed successful Recon calls `OpenOrders`, `Fills`, and `AccountState`.
- Missing-Order and failed-status cases call `OrderStatus`.
- Dirty Venue recovery and submit-failure recovery call schema-v3 load.
- Delayed-fee tests call `SetFillFeeAvailableForTest`.
- Completed tests call `Simulator.Stop` through `Account.Stop`.
- `TestAccountSendsOnlyOfficialVenueOrderAction` uses `venueRecorder`, not Simulator.
- `TestEnrichFillCLOIDsRejectsAmbiguousVenueIdentity` is pure Account evidence logic and does not call Simulator.

## Holistic Module Impact

| Area | Confirmed current owner | Required implementation impact | Risk | Wiki update with code |
|---|---|---|---|---|
| Bot identity | `setup.Nuubot` carries stored Sweep and Bot IDs. | Pass SweepID and BotID into every Simulator Config. | Wrong account checkpoint selection. | Setup, entities, Simulator. |
| Account ownership | Each Executor-cycle creates and stops one Account and Simulator. | Keep ownership. Load prior scoped Simulator rows during each Account Init. | `none` must temporarily write for continuity. | Account, Simulator, startup. |
| Persistence mode | Each ExecutorSpec owns `none|max`; Account copies it. | Keep one value in Account. Remove stored mode from Simulator Config. Call `Persist(mode)`. | Conflicting duplicated policy. | Account, Simulator, schema. |
| Backtest loop | MarketData publishes before TickClock advances. | Admit Clock time before Simulator callback, then fire Controller timers. | Backtest Fill time becomes one tick stale. | Clock, Backtest, MarketData. |
| Paper/test loop | Live WallClock and future WebSocket work may run independently. | Route events through one Bot loop before Simulator mutation. | Concurrent Account and BBO mutation. | Live, MarketData, Simulator. |
| Hyperliquid wire | Simulator and Hyperliquid duplicate official response fields. | Move exact response encoding into `internal/hyperliquid`. | Silent response drift. | Hyperliquid, Simulator parity. |
| SQLite | One JSON row stores complete Account-symbol history. | Replace with account, position, Order, and Fill rows. | Full-history rewrites and wrong account scope. | Trading schema, Simulator. |
| Account queries | Account decodes detached `[]byte` responses. | Keep the boundary unchanged. | Exposing Simulator storage objects. | Account and Hyperliquid pages. |
| Tests | Account tests mutate Simulator production state. | Delete the test-only method and fabricated mutation tests. | Tests determine production behavior. | Testing boundary if needed. |
| Logging | Account owns lifecycle reporting. | Add no Simulator logger. | Duplicate lifecycle messages. | None. |

## Required Target Flow

### Account Init

```text
Account.Init
  validate Account and persistence config
  initialize Ledger
  construct Account-owned Simulator
  Simulator.Init
    validate SweepID, BotID, Network, AccountName, Clock, MarketData, and path
    open Bot SQLite database
    create current schema
    load scoped account row
    when found:
      load positions
      load Orders
      load latest 10,000 Fills into memory
      partition Orders by status
      rebuild indexes
      derive next OID, TID, and batch ID
    when absent:
      insert initial account row
      use configured starting equity
    subscribe to required MarketData key
  publish initialized Account
```

The initial account insert is required even when the supplied mode is `none`.

### BBO Mutation

```text
admit event time into Nuubot Clock
publish BBO into MarketData
Simulator callback reads latest BBO
scan open Orders only
move completed Orders into closed Orders
append new Fills
update balance and per-symbol position
mark only changed records dirty
mark Account dirty
run due Controller timers
call Simulator Persist(mode) at loop end
```

BBO receipt without a state mutation does not mark Simulator state dirty.

### Order Submission

```text
Account validates typed Order intent
Account creates mandatory CLOID
Simulator receives valid typed action
Simulator allocates Venue OID
Simulator creates one open or rejected Order record
Simulator marks that record dirty
Simulator returns canonical Hyperliquid JSON
Account decodes the response
loop-end Persist writes the dirty record
```

A valid Venue-rule rejection consumes CLOID and Venue OID.

It creates one rejected Order and immediately moves it into `closedOrders`.

Malformed or blank CLOID proves application corruption and fails the run.

### Cancellation

```text
Simulator processes each requested CLOID in request order
known open Order:
  cancel it
  move it to closed Orders
  mark it dirty
  return success
valid unknown or already-closed Order:
  return item error
```

Ordinary batch siblings do not cascade.

Cancelling an unfilled or partially filled bracket parent cancels its children internally.

Automatic child cancellations are not extra response entries.

### Loop-End Persistence

```go
func (s *Simulator) Persist(mode string) error
```

First implementation:

```text
validate mode is none or max
ignore the mode choice temporarily
return nil when no record is dirty
write dirty records in one transaction
retain dirty state on failure
clear committed dirty state on success
```

Call placement:

- BBO and Controller work complete first.
- Each Bot loop calls persistence once.
- Account shutdown performs one final explicit persistence call.
- `Stop` closes resources after persistence.

No mutation performs synchronous full-history persistence.

No persistence failure rolls back valid Simulator memory.

The caller receives the error and decides whether the Bot stops.

## Target Memory Model

```go
type Simulator struct {
    config           Config
    store            *simulatorStore
    clock            clock.Clock
    openOrders       []*simOrder
    closedOrders     []*simOrder
    ordersByOID      map[uint64]*simOrder
    fills            []*simFill
    positions        map[string]*position
    dirtyOrders      map[*simOrder]struct{}
    dirtyFills       map[*simFill]struct{}
    dirtyPositions   map[string]*position
    accountDirty     bool
    nextVenueOID     uint64
    nextVenueTID     uint64
    nextBatchID      uint64
}
```

The exact Go container may change during implementation.

Required behavior does not:

- `openOrders` contains waiting, armed, and resting Orders.
- `closedOrders` contains filled, cancelled, and rejected Orders.
- Closing moves the same pointer and preserves database ID.
- Fills use one immutable history collection.
- Dirty registration is idempotent.
- Matching never scans closed Orders or Fills.
- Database retains every Order and Fill.
- Memory retains the latest 10,000 Fills.
- Each Fill response returns at most 2,000 rows.

Closed-Order retention has no approved cap.

Keep all closed Orders until a separate measured requirement sets one.

## Target SQLite Schema

This is implementation DDL, not a migration.

Existing Simulator schema is hardcut.

Unsupported schema versions fail immediately.

```sql
CREATE TABLE simulator_account (
    id              INTEGER PRIMARY KEY,
    sweep_id        INTEGER NOT NULL,
    bot_id          INTEGER NOT NULL,
    network         TEXT NOT NULL,
    account_name    TEXT NOT NULL,
    schema_version  INTEGER NOT NULL,
    equity          TEXT NOT NULL,
    observed_ms     INTEGER NOT NULL,
    updated_ms      INTEGER NOT NULL,
    UNIQUE (sweep_id, bot_id, network, account_name)
);

CREATE TABLE simulator_position (
    id                    INTEGER PRIMARY KEY,
    simulator_account_id  INTEGER NOT NULL REFERENCES simulator_account(id),
    symbol                TEXT NOT NULL,
    size                  TEXT NOT NULL,
    entry_price           TEXT,
    realized_pnl          TEXT NOT NULL,
    fees                  TEXT NOT NULL,
    updated_ms            INTEGER NOT NULL,
    UNIQUE (simulator_account_id, symbol)
);

CREATE TABLE simulator_order (
    id                    INTEGER PRIMARY KEY,
    simulator_account_id  INTEGER NOT NULL REFERENCES simulator_account(id),
    venue_oid             INTEGER NOT NULL,
    cloid                 TEXT NOT NULL,
    asset                 INTEGER NOT NULL,
    symbol                TEXT NOT NULL,
    batch_id              INTEGER NOT NULL,
    kind                  TEXT NOT NULL,
    is_buy                INTEGER NOT NULL,
    limit_price           TEXT NOT NULL,
    quantity              TEXT NOT NULL,
    reduce_only           INTEGER NOT NULL,
    time_in_force         TEXT NOT NULL,
    trigger_price         TEXT,
    status                TEXT NOT NULL,
    armed                 INTEGER NOT NULL,
    remaining_quantity    TEXT NOT NULL,
    filled_quantity       TEXT NOT NULL,
    average_fill_price    TEXT,
    fees                  TEXT NOT NULL,
    submitted_ms          INTEGER NOT NULL,
    eligible_ms           INTEGER NOT NULL,
    status_ms             INTEGER NOT NULL,
    UNIQUE (simulator_account_id, venue_oid)
);

CREATE INDEX simulator_order_cloid
    ON simulator_order (simulator_account_id, cloid);

CREATE TABLE simulator_fill (
    id                    INTEGER PRIMARY KEY,
    simulator_account_id  INTEGER NOT NULL REFERENCES simulator_account(id),
    venue_oid             INTEGER NOT NULL,
    venue_tid             INTEGER NOT NULL,
    symbol                TEXT NOT NULL,
    is_buy                INTEGER NOT NULL,
    quantity              TEXT NOT NULL,
    price                 TEXT NOT NULL,
    execution_ms          INTEGER NOT NULL,
    start_position        TEXT NOT NULL,
    closed_pnl            TEXT NOT NULL,
    direction             TEXT NOT NULL,
    fee                   TEXT NOT NULL,
    liquidity             TEXT NOT NULL,
    UNIQUE (simulator_account_id, venue_tid)
);
```

CLOID is indexed but not unique inside Simulator.

Account must still create a valid nonblank CLOID for every request.

New rows use `INSERT ... RETURNING id`.

Returned IDs remain local until transaction commit succeeds.

After commit, Simulator assigns generated IDs and clears only committed dirty records.

Venue OID and TID are separate Simulator-assigned identities.

Next Venue IDs derive from stored maximum plus one.

## Time Implementation

Required event fields:

```text
Order.SubmittedAt
Order.EligibleAt
Order.StatusAt
Fill.ExecutionAt
BBO.ExchangeTime
```

Submission, eligibility, status, cancellation, and Fill execution use Nuubot Clock time.

BBO exchange timestamp remains separate market evidence.

Current Backtest ordering is wrong for direct `Clock.NowMS()` inside the BBO callback:

```text
current:
  publish BBO
  Simulator callback
  advance TickClock
```

The implementation must admit replay time before callbacks but fire Controller timers afterward:

```text
target:
  admit TickClock time
  publish BBO
  Simulator callback
  fire due timers
```

Do not simply call current `Advance` first.

Current `Advance` also fires Controller timers, which would run before the new BBO.

The Clock change therefore needs an explicit two-phase TickClock operation.

## Hyperliquid Wire Implementation

`internal/hyperliquid` becomes the single wire-shape owner.

Add exact encoders for:

- submit response;
- cancel response;
- open Orders;
- exact Order status;
- Fill history; and
- clearinghouse state.

Simulator converts private state into those wire values.

Exact status includes:

- remaining `sz`;
- original `origSz`;
- OID and CLOID;
- submitted and status timestamps;
- limit and trigger prices;
- trigger condition;
- children;
- reduce-only;
- Order type;
- TIF;
- position-TP/SL flag; and
- exact status.

Frozen official fixtures prove field names, nesting, omission rules, ordering, and values.

## Supported Capability Header

Add one durable comment near the top of `simulator.go`:

```go
// Simulator capability limits:
//
// Supported:
//   - GTC and IOC Orders.
//   - na and normalTpsl grouping.
//   - Non-market TP/SL triggers.
//   - One-price MarketData matching.
//
// Not supported:
//   - ALO Orders.
//   - positionTpsl grouping.
//   - Trigger-market Orders.
//   - Order-book depth.
//   - Queue position.
//   - Available-liquidity sizing.
//   - Maker simulation.
//   - Partial fills caused by limited market liquidity.
//
// Unsupported behavior fails loudly.
// Update this list whenever Simulator capability changes.
```

These are verified limitations, not guesses.

## Exact Production File Changes

### `internal/simulator/simulator.go`

- Add SweepID, BotID, Network, AccountName, Clock, and required database path to Config.
- Remove stored `PersistMode`.
- Split timestamps and price meanings.
- Replace active-map sorting with ordered open and closed collections.
- Add per-symbol positions and dirty-record tracking.
- Export `Persist(mode string) error`.
- Delete full-state stage, commit, storedState, and position replay.
- Delete `SetFillFeeAvailableForTest`.
- Use canonical Hyperliquid encoders.

### `internal/simulator/store.go`

- Hardcut the JSON table.
- Create normalized Simulator tables.
- Add scoped load queries.
- Add dirty-row transaction writes.
- Use generated database IDs.
- Trust committed database rows.
- Add no migration or repair code.

### `internal/account/account.go`

- Add `Persist(string) error` to private Venue contract.
- Add one Account persistence method.
- Preserve detached `[]byte` Venue responses.
- Keep Account as Simulator owner.
- Remove test-only venue mutation access.

### `internal/account/recon.go`

- Pass SweepID, BotID, Network, Clock, and path into Simulator Init.
- Stop passing PersistMode into Simulator Config.
- Keep Recon behavior outside this Simulator change.

### `internal/executor`, `internal/botcycle`, and `internal/controller`

- Add the smallest loop-end Account persistence path.
- Persist each active Account once per Bot loop.
- Join Controller and persistence errors.
- Do not add queues, workers, locks, or generic repositories.

### `internal/backtest/backtest.go` and Clock

- Split TickClock time admission from timer firing.
- Publish BBO after Clock admission.
- Fire Controller timers after Simulator callbacks.
- Call loop-end persistence.

### `internal/live`, MarketData, and WebSocket

- Preserve one Bot event loop.
- Queue future WebSocket events into that loop.
- Do not call Account or Simulator concurrently.
- Add proof only when simnet paper/test execution exists.

### `internal/hyperliquid`

- Own exact official response encoders.
- Reuse the same field definitions for Simulator output and Account decoding.
- Keep normalized Account-facing results separate from wire structs.

### Tests

- Delete `SetFillFeeAvailableForTest`.
- Delete fabricated Simulator delayed-fee mutation tests.
- Rewrite full-JSON persistence tests.
- Rewrite persistence-failure tests for dirty retry.
- Rewrite tests that drop the old JSON table.
- Update every test Venue for `Persist`.

## Required New Proof

- Account-owned Simulator state carries across two Bot cycles.
- `none` temporarily writes Simulator state and reloads it.
- Balance, positions, OID, TID, Orders, Fills, and database IDs continue.
- 8,000 clean Orders plus three changes writes exactly three Orders.
- Failed persistence retains dirty records.
- Retry writes each retained dirty record once.
- Restart preserves open and closed partitioning.
- Batch cancellation returns ordered mixed results.
- Bracket parent cancellation updates children without extra response rows.
- Rejected valid Order consumes CLOID and Venue OID.
- Blank or malformed CLOID fails loudly.
- TickClock and WallClock produce exact Fill times.
- Fill memory retains 10,000 and response returns at most 2,000.
- Unsupported capability inputs fail loudly.
- Official Hyperliquid fixtures match every supported response.
- Simulator has no test-only production surface.

Final implementation proof:

```text
focused Go tests with -tags noasm
full Go tests with -tags noasm
full go vet with -tags noasm
git diff --check
two complete rebuild.sh rounds
Observer, Trade, and Grid sweeps created from templates
Observer, Trade, and Grid executions pass both rounds
second round proves no stale database dependency
```

## Implementation Order

1. Add failing official wire and dirty-persistence tests.
2. Hardcut Simulator SQLite schema.
3. Replace JSON restore with scoped row loading.
4. Replace full-state staging with dirty-record tracking.
5. Add loop-end `Persist(mode)` propagation.
6. Split Order timestamps and price meanings.
7. Implement open/closed Order movement and Fill limits.
8. Centralize Hyperliquid response encoders.
9. Delete the test-only production method and fabricated tests.
10. Fix TickClock event ordering.
11. Add unsupported capability header and fail-loud tests.
12. Run focused and full proof.
13. Create `rebuild.sh`.
14. Run two clean Observer, Trade, and Grid rebuild rounds.
15. Update canonical wiki pages and HANDOFF proof with implemented facts.

## Deferred Decisions

- Exact Simulator connection return semantics remain implementation-tested against the live Exchange contract.
- Closed-Order retention has no approved cap.
- Duplicate CLOIDs are admitted. OID remains canonical; the CLOID lookup may point to the latest matching OID.
- No Simulator Venue-rejection rule is invented until a required rule is identified.
- True memory-only Backtest persistence requires the future Account/Simulator lifetime redesign.

## Current State Shape

`Simulator` retains:

```text
orders             []*simOrder                 all accepted Orders
latestOIDByCLOID   map[string]uint64            latest matching Venue OID only
activeOrders       map[uint64]*simOrder         all nonterminal Orders
fills              []simFill                    all Fills
currentPosition    position                     maintained finance
store              *simulatorStore              max only
subscription       *market.Subscription         one exact BBO stream
lastPrice          decimal.Decimal              transient
lastPriceKey       comparisonKey                transient
lastTimestampMS    uint64                       transient
observedMS         uint64                       durable Venue event time
```

Every Order appears once as a private pointer.

Both indexes point to the same Order.

Every Fill appears once by value.

The current position is maintained during execution and independently rebuilt from Fills during max-mode recovery.

## Confirmed Matching Behavior

- The first BBO warms state and performs no matching.
- Later BBOs cannot move backward in time.
- Same-timestamp BBOs are accepted.
- Orders cannot match a BBO whose timestamp equals their current matching timestamp.
- Submission may immediately fill IOC and marketable GTC Orders against the latest BBO state.
- IOC crosses regardless of submitted price.
- Limit, TP, and SL threshold comparisons use exact positive decimal keys.
- Fill price uses submitted limit price or trigger price, then configured adverse slippage.
- No order-book depth, queue position, liquidity quantity, maker behavior, or partial market liquidity exists.
- Reduce-only execution cannot open or reverse exposure.
- One private batch fills at most one leg during one BBO.
- A filled entry arms trigger children.
- A filled TP or SL cancels trigger siblings.
- Canceling a limit Order cancels its trigger children.
- Canceling one trigger Order does not cancel its sibling.
- Waiting trigger children remain official open Orders.
- All generated Fills report `crossed=true`, `liquidity=taker`, and `feeToken=USDC`.

## Confirmed Finance Behavior

- Position size is signed.
- Same-side Fills calculate weighted entry price.
- Opposite-side Fills realize PnL only on the closed quantity.
- A position flip sets entry price to the flip Fill price.
- Fees accumulate once per Fill.
- Clearinghouse equity is initial equity plus realized and unrealized PnL minus fees.
- Position notional uses absolute size times latest price.
- Margin uses fixed 5x leverage.
- Maintenance margin and cumulative funding are always zero.
- Liquidation price is always absent.
- Withdrawable equals account value minus current margin.

These are Simulator model facts.

They are not proven Hyperliquid financial parity.

## External Compatibility Assessment

### Preserve

- Account remains the only production Simulator owner.
- Simulator remains hidden behind Account's Venue boundary.
- CLOID remains Account-created, mandatory, opaque, and unchanged.
- Venue OID and Venue TID remain Simulator-assigned.
- Submit and cancel responses remain Hyperliquid-shaped envelopes.
- Open Orders, Fills, exact Order status, and Account state remain detached JSON.
- Trigger waiting/arming remains private under Nuubot's approved official-status contract.
- Fill JSON may omit CLOID; Account resolves Venue OID.
- Current exact one-price crossing and accepted Trade/Grid finance remain until a separate bid/ask behavior change is approved and rebaselined.
- Reduce-only cannot open or reverse exposure.
- Current entry Fill arms children; TP/SL Fill cancels siblings.
- Current Nuubot fee and PnL separation remains.
- Unsupported behavior fails loudly.

### Modify for compatibility

- Submission before any market observation currently rests silently; Nuutrader6 fails loudly.
- Immediate filled submit JSON currently omits fee; Nuutrader6 includes it.
- Unknown cancel currently returns a Go error before an official item status; Nuutrader6 returns an error status.
- Exact Order status currently conflates remaining and original size.
- Exact Order status omits supported official fields already represented by the submitted Order.
- One timestamp currently conflates submission, arming, and terminal status time.
- `positionTpsl` and trigger-market inputs are admitted without distinct implemented semantics.
- Submitted limit price and trigger threshold are conflated by `orderPrice`.

### Reference differences that must not overwrite approved Nuubot behavior

- Nuutrader6 bid/ask matching cannot be copied into a one-price MarketData contract.
- Nuutrader6 exposes waiting statuses; Nuubot deliberately does not.
- Nuutrader6 string Fill keys do not replace Nuubot numeric Venue TIDs.
- Nuutrader6 includes fee inside realized PnL while separately accumulating total fees; Nuubot's accepted finance keeps them separate.
- Nuutrader6 persists the last BBO; Nuubot recovery deliberately requires fresh market truth.
- Nuutrader6 stores domain fields such as position ID, level, purpose, and parent CLOID inside Simulator state; Nuubot deliberately excludes Account/Ledger domain identity.

## Internal Implementation Assessment

External semantics do not require current internal structure.

Confirmed internal changes are justified:

- eliminate full-history clone staging;
- eliminate repeated active-map slice allocation and sorting;
- reduce per-BBO full active-set crossing work;
- separate submission, eligibility/arming, and status timestamps;
- separate submitted limit price, trigger threshold, and Fill basis;
- retain exact comparison-key behavior;
- trust scoped database rows during recovery without Fill replay repair;
- accept persistence mode only through `Persist(mode)`;
- temporarily persist dirty rows for both `none` and `max`;
- avoid full-history rewrite per durable mutation;
- add database-assigned row identity for every physical row;
- keep recovery validation limited to schema, decode, partition, and required indexes; and
- remove the test-only production mutation.

No interface, factory, event bus, generic repository, compatibility bridge, or duplicate persistence flag is justified.

## Current Persistence and Recovery

`persist_mode=none` opens no Simulator database.

`persist_mode=max` stores one schema-version-3 JSON snapshot row.

The row key is official Account plus symbol.

The payload contains identity, policy, counters, event time, every Order, and every Fill.

The latest BBO is intentionally not stored.

Every accepted max-mode mutation persists before the staged memory becomes visible.

Recovery:

1. loads one complete payload;
2. rejects identity, policy, or schema mismatch;
3. rebuilds CLOID and active indexes;
4. validates sequential OIDs and counters;
5. restores Fills;
6. validates sequential TIDs while replaying them; and
7. reconstructs position and finance.

A recovered open position cannot return Account state until one fresh BBO arrives.

### Current Persistence Flag Flow

`Config.PersistMode` currently duplicates Account's copied canonical value inside Simulator.

Current shaping is:

```text
BotConfig TOML persist_mode
  -> botspec trade/grid config validation
  -> botspec.ExecutorSpec.PersistMode
  -> Trade/Grid Executor Account Config
  -> Account Config.PersistMode
  -> Ledger Config.PersistMode
  -> Simulator Config.PersistMode
```

Exact production assignments:

- `internal/botspec/config.go` decodes and validates `none|max`.
- `internal/executor/trade.go` passes `ctx.Spec.PersistMode`.
- `internal/executor/grid.go` passes `ctx.Spec.PersistMode`.
- `internal/account/recon.go` passes the value unchanged to Ledger and Simulator.
- `internal/simulator/simulator.go` selects no store for `none` and SQLite for `max`.

Current Backtest configs use `none`.

Current max-mode Account tests pass `max`.

Approved target:

- Account retains the copied canonical mode.
- Simulator Config stores no mode.
- Caller invokes `Simulator.Persist(mode)`.
- Simulator validates `none|max`.
- First implementation writes dirty rows for both values.
- Future architecture restores `none` as a no-write mode.
- Simulator never infers Backtest, paper, test, or Live.

### Database Row Identity

User rule:

```text
database assigns each row's unique DB ID at row creation
```

Current `simulator_venue_state` has no database-assigned row ID.

Its primary key is the natural pair `(account_name, symbol)`.

That conflicts with the rule for the one physical Simulator row.

Orders and Fills inside `payload_json` are not database rows.

Their OID, TID, CLOID, and batch values are embedded external/private identities, not database row IDs.

Backtest `none` creates no runtime Simulator row.

Its OID, TID, batch, slice position, and map entries are transient runtime references only.

Target Simulator Orders and Fills are physical rows with database-assigned IDs.

CLOID, Venue OID, Venue TID, Account, symbol, and batch identity remain additional keys or indexes.

### Parent-Creates-Child Identity Rule

Account creates the local Order and its CLOID before Venue submission.

Simulator creates its own private Venue Order record and assigns Venue OID.

Simulator creates each private Venue Fill record and assigns Venue TID.

Simulator assigns its private batch ID.

These are child identities inside Simulator-owned Venue truth.

No current caller assigns Venue OID or Venue TID.

No current ownership violation was found.

Database row IDs are separate from those identities and are currently absent.

## Confirmed Performance Evidence

Retained function profiles are instrumented A/B/C evidence, not uninstrumented production wall-time attribution.

Grid profile:

```text
workspace/perf/fprofiles/s11-b15-20260728T034420Z/report.txt

77,745,401  crosses                         4.767s flat
 3,055,509  sortedActiveOrders             4.516s flat
   277,208  OpenOrders                     3.675s flat
77,745,401  compareComparisonKeys          2.436s flat
 2,775,570  match                          2.257s flat, 13.484s cumulative
 2,775,620  onBBO                          1.535s flat, 15.791s cumulative
   277,208  AccountState                   1.073s flat
```

Trade profile:

```text
workspace/perf/fprofiles/s9-b13-20260728T033505Z/report.txt

2,774,333  onBBO                           1.453s flat
  277,847  AccountState                  968.304ms flat
3,052,279  sortedActiveOrders            447.273ms flat
2,774,140  match                         357.177ms flat
5,543,392  crosses                       331.130ms flat
2,774,959  newComparisonKey              327.335ms flat
  277,847  OpenOrders                    257.439ms flat
5,543,392  compareComparisonKeys         181.949ms flat
```

Confirmed source causes:

- `match` calls `sortedActiveOrders` inside its outer loop.
- `sortedActiveOrders` allocates and sorts.
- `match` linearly tests active Orders for each admitted BBO.
- `OpenOrders` separately allocates and sorts the same active set.
- `Fills` linearly scans all retained Fill history for every range query.
- max-mode `stage` clones all Order and Fill history before every BBO match attempt.
- max-mode persistence marshals and rewrites all Order and Fill history after every changed mutation.

The retained profiles exercise Backtest `persist_mode=none`.

There is no retained max-mode or long-running Live profile.

## Confirmed Issues

### BLOCKER — current `none` loses Simulator continuity between Bot cycles

Current Account and Simulator objects are recreated for every Executor-cycle.

Controller carries ending equity forward.

It does not carry OID/TID counters, Order history, Fill history, or private Simulator state.

Current `none` opens no database and therefore cannot reload those values.

Approved first implementation temporarily persists dirty Simulator rows for both `none` and `max`.

HANDOFF records restoration of true memory-only Backtest behavior as future architecture work.

### BLOCKER — physical Simulator row has no database-assigned ID

`simulator_venue_state` uses `(account_name, symbol)` as its primary key.

The database does not assign a unique row ID at creation.

OID, TID, CLOID, and private batch identity are not substitutes for the database row ID.

Embedded Orders and Fills are JSON objects, not physical database rows.

### BLOCKER — max-mode work grows with complete history

`stage`, `storedState`, JSON encoding, and one-row upsert all scale with retained Orders and Fills.

`onBBO` calls `stage` before knowing whether a match changes Venue truth.

In max mode, every later BBO therefore clones complete history even when it creates no Fill.

No retained profile measures this path.

The effect on a long-running Live simnet process is unproven.

### BLOCKER — Live serialization is not implemented or proven

Simulator has no mutex.

Account has no mutex around Venue calls or its dirty callback.

Backtest invokes BBO, Account, and Controller work synchronously.

Current Live Run fails before live execution and WebSocket publication remains unimplemented.

Canonical architecture says Runner serializes external events and clock events, but no running Live path currently proves that contract.

Concurrent future BBO publication and Controller Account calls would access Simulator and Account mutable state without local synchronization.

This is a future-path risk, not a reproduced current Backtest race.

### HIGH — failed Init can leave retained state

When subscription or initial-BBO processing fails, `Init` clears `started` and closes the store.

It does not clear Config, Orders, Fills, position, counters, event time, or BBO fields.

A later `Init` resets maps and counters but does not explicitly clear every retained slice and value.

No current test covers retry after partial Init failure.

### HIGH — failed Stop leaves admission state inconsistent

`Stop` removes and nils the subscription before persistence and store close.

If either later operation fails, `started` remains true and `stopped` remains false.

Mutation and query guards therefore still admit calls, but Simulator no longer receives BBO updates.

No current test injects Stop persistence or close failure.

### HIGH — one Order timestamp owns three meanings

`simOrder.timestampMS` begins as submission time.

Arming overwrites it with arming time.

Fill and cancellation overwrite it with terminal time.

The same field controls BBO eligibility, open-Order timestamp, nested Order timestamp, and status timestamp.

Current JSON therefore cannot preserve all three event times independently.

### HIGH — exact Order status differs from current official shape

Current exact status returns `sz` from original requested quantity even for terminal Orders.

It also omits current official fields such as trigger condition, trigger price, children, position-TP/SL flag, reduce-only, Order type, and TIF.

Current Account validation requires `sz` to remain positive.

Official examples show terminal remaining `sz` may be zero while `origSz` retains submitted size.

The current Simulator and Account decoder agree with each other, but external shape parity is incomplete.

Nuutrader6 keeps submitted Order values and returned Fill/status values distinct.

### HIGH — Fill history limits differ from official API

Current `Fills` returns every retained row inside the inclusive range.

It enforces neither a per-response cap nor a retained-history cap.

Current Hyperliquid documentation states `userFillsByTime` returns at most 2,000 Fills and exposes only the latest 10,000.

Account's Live design already describes those limits.

### MEDIUM — accepted official inputs exceed implemented semantics

`positionTpsl` grouping is accepted but has no distinct behavior.

Trigger `isMarket` is accepted but ignored.

Trigger request `p` and `triggerPx` are both stored, but matching and Fill basis use `triggerPx`.

`Alo` is official but rejected.

The current application sends mandatory CLOID, `Gtc`/`Ioc`, `normalTpsl` or `na`, and non-market triggers.

No current production caller requires the broader cases.

### MEDIUM — current recovery performs rejected repair-style validation

Recovery validates core identity, counters, positive Order price/quantity, basic kind/status, Fill sequence, and derived position facts.

The approved target trusts scoped committed rows.

It performs schema, SQL, and decoding checks only.

It rebuilds memory indexes without replaying or repairing stored finance.

### MEDIUM — official wire shapes are duplicated

Simulator declares private response DTOs for submit, cancel, open Orders, exact status, Fills, and clearinghouse state.

`internal/hyperliquid` separately declares corresponding decode DTOs and normalized results.

Both packages therefore encode knowledge of the same external JSON fields.

This is confirmed duplication.

No current test compares every Simulator byte shape against frozen official fixtures.

### LOW — production code contains a test-only mutation

`SetFillFeeAvailableForTest` has no production caller.

It bypasses max-mode staging and persistence.

Only Account tests use it to simulate delayed fee evidence.

### LOW — Simulator has no lifecycle logging

Simulator Config carries no logger.

Init and Stop emit no Simulator lifecycle entry, completion, or terminal statistics.

Account emits its own Stop summary.

Assessment: no Simulator logger is required.

Account remains the lifecycle log owner.

Make no logging change.

## Backtest and Live Facts

### Backtest

- MarketData publication, Simulator callback, Clock advance, Controller, Account, and Ledger work are synchronous.
- Existing Trade and Grid system proof exercises the real Simulator path.
- Backtest uses `persist_mode=none`.
- Interrupted Backtest does not recover Simulator state.
- ResultPublisher publishes reconciled Ledger evidence, not Simulator private state.

### Live

- Live Run currently cannot execute trading.
- WebSocket publication into MarketData is pending.
- Hyperliquid credentialed Exchange is pending.
- simnet Live would select Simulator under the approved common startup contract.
- max-mode Simulator recovery exists independently.
- complete Live recovery also requires Runner, Controller, Executor, Account, and Ledger continuation.
- no current Live test proves Simulator startup, serialization, persistence volume, recovery, or shutdown failure.

## Current Tests

Direct Simulator tests prove:

- one canonical Order record with shared indexes;
- detached official JSON;
- ordered OID assignment;
- bracket Fill and sibling cancellation;
- terminal no-rematch;
- one schema-v3 persistence round trip;
- max-mode mutation failure atomicity;
- opaque shape-valid CLOID admission;
- exact decimal-key ordering;
- crossing parity against decimal comparisons;
- zero allocation inside `crosses`;
- fuzzed positive-decimal comparison; and
- one comparison benchmark.

Account integration tests additionally prove:

- reconciled bracket behavior;
- dirty Recon and clean sweep cadence;
- max persistence recovery of dirty Venue state;
- submit-persistence failure recovery;
- accepted uncertain submission is not retried;
- immediate Fill survives Account submit-persistence failure;
- missing submitting Order repair;
- delayed fee repair; and
- final finance stability while fee evidence remains pending.

Current focused proof run:

```text
C:\Users\PC\.local\go1.26.5\go\bin\go.exe test -tags noasm \
  ./internal/simulator ./internal/account ./internal/hyperliquid ./internal/market

PASS
ok nuubot/internal/simulator
ok nuubot/internal/account
ok nuubot/internal/hyperliquid
ok nuubot/internal/market
```

## Proof Checked

- Complete current `simulator.go` and `store.go`.
- Complete direct Simulator test file.
- Every current production function declaration.
- Account's private Venue interface and every concrete call.
- Account reconciliation queries and evidence translation.
- MarketData subscription and synchronous notification.
- Backtest and Live lifecycle design.
- Simulator, MarketData, Account, Startup, parity, and trading-schema design pages.
- Retained Trade and Grid function-profile reports.
- Current official Hyperliquid Exchange and Info documentation.
- Complete Nuutrader6 `simulator.py` and `exchange.py`.
- Nuutrader6 Account Simulator construction and call path.
- Nuutrader6 audited Account/Exchange/Simulator parity plan and recorded proof.
- Focused existing tests with native Go 1.26.5 and `-tags noasm`.

Official references checked:

- `https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/exchange-endpoint`
- `https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint`
- `https://hyperliquid.gitbook.io/hyperliquid-docs/trading/take-profit-and-stop-loss-orders-tp-sl`

## Proof Missing

- Frozen official JSON fixtures covering every supported response.
- Controlled Simulator-versus-testnet mutation parity.
- Executable simnet parity-probe path.
- Official limit-trigger behavior when `p` differs from `triggerPx`.
- `positionTpsl`, trigger-market, and unsupported-TIF behavior proof.
- 2,000-response and 10,000-retention Fill behavior.
- Partial Init retry proof.
- Stop persistence and close-failure proof.
- Scoped normalized recovery proof.
- Dirty-row persistence write-count proof.
- Temporary `none` persistence and cross-cycle reload proof.
- long-running memory-retention profile.
- two-phase TickClock event-time proof.
- Live serialization and race proof.
- Live simnet startup, recovery, and shutdown proof.

## Assumptions

- Current source, not an in-progress Account target, defines implemented behavior.
- Retained system counts prove accepted Backtest behavior, not external exchange parity.
- Function-profile call counts and instrumented time identify hot paths, not exact production speedup potential.
- No current Live execution means Live concurrency observations remain risk analysis.

## Open Questions

- Exact connection return behavior remains implementation-tested against the live Exchange contract.
- Closed-Order memory retention has no approved cap.
- Duplicate CLOIDs are accepted without recovery validation; OID remains the canonical Order identity.
- No Venue-rejection rule is implemented until one is required.
- Live simnet event serialization remains blocked by unimplemented WebSocket publication.

## Numbered Issue Inventory

Discussion completed:

1. **Persistence:** `Persist(mode)` writes dirty rows for both modes temporarily.
2. **Identity:** one account row per SweepID, BotID, Network, and AccountName.
3. **Init:** open database, load scoped rows, then publish initialized state.
4. **Stop:** connection behavior remains implementation-tested; persistence is explicit.
5. **Durability:** upsert only dirty records in one transaction.
6. **Matching:** scan ordered open Orders only.
7. **Time:** separate submission, eligibility, status, Fill, and BBO times.
8. **Price:** separate limit, trigger, and Fill execution prices.
9. **Submit:** valid Venue rejection consumes CLOID/OID and creates a closed record.
10. **Cancel:** ordered per-item results; only bracket parent cancellation cascades.
11. **Status:** return complete official Hyperliquid fields.
12. **Fill history:** database retains all; memory exposes 10,000; response caps at 2,000.
13. **Unsupported inputs:** fail loudly and document verified limitations at file top.
14. **Serialization:** one Bot loop; no Simulator mutex or goroutine.
15. **Recovery:** trust scoped database rows; no repair or replay engine.
16. **Wire ownership:** `internal/hyperliquid` owns exact official response shapes.
17. **Test-only API:** delete production mutation and fabricated tests.
18. **Logging:** Account owns lifecycle logs; Simulator adds none.

Bloat check: no fake Simulator or unused runtime was found. One dead test-only API, duplicated wire DTOs, full-history staging, repeated sorting, linear scans, incomplete Live wiring, lifecycle risks, and unproven max-mode scaling were found.
