# Ledger Clone Audit

Date: 2026-07-25

Scope: read-only source, wiki, tests, and existing Grid profile evidence. No replay ran.

## 1. Executive verdict

**Verdict: the user hypothesis is materially correct, but incomplete.**

Confirmed:

- Ledger owns the mutable `Trade -> Order -> Fill` graph.
- Every dirty reconciliation deep-clones the entire graph before validating venue evidence.
- The same reconciliation then snapshots the graph five full times.
- Existing Grid evidence attributes 112,609.46MB cumulative allocation to `Account.Reconcile`.
- `Ledger.Recon` owns 49,622.44MB cumulative allocation.
- `cloneTrades` owns 18.14GB on its focused path, almost entirely recursive object cloning.
- These copies occur inside one synchronous owner chain. They do not cross an ownership boundary.
- Terminal results require detached copies. Reconciliation staging requires rollback semantics, not necessarily a whole-Ledger deep clone.

Adversarial qualification:

- Removing cloning without replacing transaction-like rollback would break correctness.
- The clone is currently the mechanism proving failed reconciliation leaves the Ledger unchanged.
- The evidence supports narrowing the atomic staging boundary. It does not prove direct in-place mutation is safe.
- Snapshot amplification is at least as important as `cloneTrades`. Removing only the clone leaves repeated full scans and nested Fill copies.

## 2. Confirmed call graph

### Reconciliation path

```text
Controller.Run
  -> BotCycle.Reconcile
     internal/botcycle/botcycle.go:198-218
  -> gridExecutor.Reconcile / tradeExecutor.Reconcile
     internal/executor/grid.go:347-353
     internal/executor/trade.go:262-268
  -> Account.Reconcile
     internal/account/account.go:467-570
       -> Simulator.OpenOrders
          internal/account/account.go:485-492
       -> Simulator.Fills
          internal/account/account.go:494-510
       -> Ledger.ActiveOrders [full graph snapshot traversal #1]
          internal/account/account.go:512-532
          internal/ledger/ledger.go:499-513
       -> Simulator.AccountState
          internal/account/account.go:534-545
       -> Ledger.Recon
          internal/account/account.go:547-557
          internal/ledger/ledger.go:329-426
            -> cloneTrades [full graph deep clone]
               internal/ledger/ledger.go:339
               internal/ledger/ledger.go:604-610
              -> Trade.Clone
                 internal/trade/trade.go:209-217
                -> Order.Clone
                   internal/order/order.go:262-271
                  -> Fill.Clone
                     internal/fill/fill.go:143-147
            -> indexCLOIDs [full graph snapshot traversal #2]
               internal/ledger/ledger.go:340
               internal/ledger/ledger.go:633-642
              -> Trade.Orders
                 internal/trade/trade.go:225-232
                -> Order.Snapshot
                   internal/order/order.go:237-260
                  -> Fill.Snapshot
                     internal/fill/fill.go:119-141
            -> Order.Snapshot per incoming Fill
               internal/ledger/ledger.go:360-384
            -> Trade.Refresh for every Trade [full graph snapshot traversal #3]
               internal/ledger/ledger.go:389-393
               internal/trade/trade.go:147-196
              -> Order.Snapshot -> Fill.Snapshot
                 internal/trade/trade.go:153-175
            -> Trade.Orders validation [full graph snapshot traversal #4]
               internal/ledger/ledger.go:394-402
            -> persistCandidate
               internal/ledger/ledger.go:405-418
            -> publish staged graph
               internal/ledger/ledger.go:420-425
       -> Ledger.Result [full graph snapshot traversal #5]
          internal/account/account.go:559-565
          internal/ledger/ledger.go:428-455
         -> Trade.Snapshot
            internal/trade/trade.go:198-201,236-280
           -> Order.Snapshot -> Fill.Snapshot
       -> accountSnapshot scans detached result
          internal/account/account.go:565
          internal/account/account.go:782-829
  -> Controller builds immutable Risk input
     internal/controller/controller.go:468-511
  -> Executor.OnRecon
     internal/botcycle/botcycle.go:221-237
     Grid reads up to each active level through Account.Trade
     internal/executor/grid.go:355-383
     -> Ledger.Trade -> Trade.Snapshot
        internal/ledger/ledger.go:515-526
```

### Mutation and index flow

- `Ledger.CreateTrade` stages one new Trade, but `withTrade` clones every prior Trade: `internal/ledger/ledger.go:170-213,612-620`.
- `Ledger.AddOrders` clones every Trade although one Trade changes: `internal/ledger/ledger.go:216-256`.
- `Ledger.RecordSubmit` clones every Trade although one submission batch changes: `internal/ledger/ledger.go:258-327`.
- `Ledger.Recon` clones every Trade regardless of evidence count or change: `internal/ledger/ledger.go:329-426`.
- `indexOrders` snapshots every Order to recover IDs, then retrieves mutable pointers: `internal/ledger/ledger.go:622-631`.
- `indexCLOIDs` snapshots every Order to recover CLOIDs, then retrieves mutable pointers: `internal/ledger/ledger.go:633-642`.
- No durable Ledger identity map exists for Order ID or CLOID. Indexes are rebuilt per operation.

### Active-order queries

Production classifications:

- `Account.Reconcile -> Ledger.ActiveOrders`: internal same-owner reconciliation; `internal/account/account.go:512-532`.
- `Account.CancelOrders -> Ledger.ActiveOrders`: internal ownership validation before venue mutation; `internal/account/account.go:404-435`.
- `Account.ActiveOrders -> Ledger.ActiveOrders`: cross-owner immutable return to Executor; `internal/account/account.go:652-655`.
- Grid shutdown calls `Account.ActiveOrders`: cross-owner immutable query; `internal/executor/grid.go:195-211,253-265`.
- Trade shutdown calls `Account.ActiveOrders`: cross-owner immutable query; `internal/executor/trade.go:161-176,201-213`.
- Trade completion calls `Account.ActiveOrders`: cross-owner immutable query; `internal/executor/trade.go:330-339`.

`Ledger.ActiveOrders` currently snapshots every Order and every nested Fill, including inactive Orders discarded afterward: `internal/ledger/ledger.go:499-513`.

### `Trade.Orders`, snapshots, clones, and helpers

Production call classification:

- `Trade.Orders` in `Ledger.ActiveOrders`: cross-owner immutable return, but over-broad implementation.
- `Trade.Orders` in `Ledger.Recon` validation: internal same-owner reconciliation.
- `Trade.Orders` in `indexOrders`: internal same-owner indexing.
- `Trade.Orders` in `indexCLOIDs`: internal same-owner indexing.
- `Trade.Snapshot` in `Ledger.Result`: terminal/cross-owner immutable return.
- `Trade.Snapshot` in `Ledger.Trade`: cross-owner immutable return to Account/Executor.
- `Trade.State` in `withTrade`: internal persistence/staging identity lookup; full snapshot is over-broad.
- `Trade.Clone` only through `cloneTrades`: internal same-owner atomic staging.
- `Order.Snapshot` in constructors and ownership checks: other/internal validation.
- `Order.Snapshot` in `Trade.Refresh`: internal same-owner derivation.
- `Order.Snapshot` in Ledger indexes and reconciliation: internal same-owner reconciliation/indexing.
- `Order.Snapshot` in `Trade.Snapshot` and `Trade.Orders`: cross-owner immutable return or nested internal over-copy.
- `Order.Clone` only through `Trade.Clone`: internal same-owner atomic staging.
- `Fill.Clone` only through `Order.Clone`: internal same-owner atomic staging.
- `Fill.Snapshot` supplies immutable values and internal calculations. It allocates no nested collection itself.

### Telemetry

- Account telemetry returns only the last small `account.Snapshot`: `internal/account/account.go:592-598`.
- Executor telemetry carries that value: `internal/executor/grid.go:396-408`, `internal/executor/trade.go:353-365`.
- Controller telemetry reads those values: `internal/controller/controller.go:365-413`.
- Telemetry does not clone Ledger, Trade, Order, or Fill graphs.
- Existing profile shows `BtRunner.collectTelemetry` at 1,610.72MB cumulative, separate from Ledger graph cloning.

### Terminal result and publication

```text
Executor.OnStop
  -> Account.Result
     internal/account/account.go:572-590
    -> Ledger.Result [required detached graph]
    -> Simulator.Result [required detached slices]
  -> result.Clone cache
     grid: internal/executor/grid.go:271-289
     trade: internal/executor/trade.go:215-232
Executor.Result
  -> account.Result.Clone
     grid: internal/executor/grid.go:411-432
     trade: internal/executor/trade.go:368-385
BotCycle.Stop
  -> Executor.Result
  -> Executor.Result.Clone again
     internal/botcycle/botcycle.go:146-166
BotCycle.Result
  -> Executor.Result.Clone again
     internal/botcycle/botcycle.go:239-246
Controller.closeCycle stores BotCycle.Result
  internal/controller/controller.go:436-455
Controller.Result
  -> Executor.Result.Clone again for every stored cycle
     internal/controller/controller.go:340-363
BtRunner.Stop
  -> Controller.Result
  -> RunReport.Build
  -> ResultPublisher.Publish
     internal/btrunner/btrunner.go:281-317
ResultPublisher.Publish
  -> ledger.Publish reconstructs mutable trees from immutable results
     internal/resultpublisher/resultpublisher.go:19-72
     internal/ledger/publish.go:15-94
```

Classification:

- `Ledger.Result` from `Account.Reconcile`: internal same-owner aggregate input, not terminal despite its contract name.
- `Ledger.Result` from `Account.Result`: terminal immutable return.
- `account.Result.Clone`, `executor.Result.Clone`, `BotCycle.Result`, and `Controller.Result`: cross-owner immutable returns and terminal retention.
- Their repeated nesting is terminal-result over-copy, not reconciliation amplification.
- `ledger.Publish`: persistence boundary. Reconstruction validates terminal values before storage.
- `max` persistence uses `persistCandidate` during each mutation: `internal/ledger/ledger.go:573-578`.
- `none` persistence skips mutation-time storage and publishes terminal results: `wiki/ARCHITECTURE.md:295-312`.

## 3. Allocation amplification mechanism and multiplicity

### Confirmed mechanism

For one successful dirty reconciliation over `T` Trades, `O` Orders, and `F` Fills:

1. `Account.Reconcile` calls `ActiveOrders`: snapshots approximately `O + F` objects.
2. `Ledger.Recon` calls `cloneTrades`: allocates one map, `T` Trades, `T` Order maps, `O` Orders, `O` Fill maps, and `F` Fills.
3. `indexCLOIDs` calls `Trade.Orders`: snapshots approximately `O + F` objects.
4. Every `Trade.Refresh` snapshots approximately `O + F` objects and allocates execution slices.
5. Filled-order validation calls `Trade.Orders`: snapshots approximately `O + F` objects.
6. `Ledger.Result` calls `Trade.Snapshot`: snapshots approximately `T + O + F` objects.
7. `accountSnapshot` scans the detached graph without another graph copy.
8. Grid `OnRecon` then calls `Account.Trade` for each active level, copying that Trade's complete Orders and Fills.

Minimum whole-graph multiplicity per successful dirty reconciliation:

```text
1 deep mutable clone
5 full immutable snapshot traversals
+ one Trade subtree snapshot per active Grid level
+ one Order snapshot per incoming Fill
```

This is structural multiplicity. Allocation counts are higher because `decimal.Decimal` operations allocate `math/big` backing values.

### Existing Grid profile proof

Source: `workspace/perf/profiles/pptest-s10-b14-20260725T152717Z/run-001.allocs.pprof`.

Confirmed `alloc_space`:

```text
Total                                         144,495.70 MB
Account.Reconcile cumulative                  112,609.46 MB
Ledger.Recon cumulative                        49,622.44 MB
Trade.Orders flat                              24,999.31 MB
math/big.nat.make flat                         22,887.53 MB
Trade.snapshot flat                            14,361.20 MB
Ledger.ActiveOrders cumulative                 23,926.13 MB
Order.Snapshot flat                            11,128.71 MB
Order.Clone flat                               10,768.78 MB
Trade.Clone cumulative                         18,231.62 MB
Ledger.Result cumulative                       23,927.94 MB
indexCLOIDs cumulative                         12,742.31 MB
cloneTrades focused cumulative                     18.14 GB
cloneTrades inside Account.Reconcile                17.92 GB
```

Confirmed `alloc_objects`:

```text
Total                                      1,558,156,772
Account.Reconcile cumulative                 976,302,175
Trade.Orders cumulative                      163,977,096
Order.Snapshot cumulative                    246,061,360
Trade.Clone cumulative                       134,605,129
cloneTrades focused cumulative               135,546,406
```

Inference:

- Using the supplied approximately 90,000 successful reconciliations, clone staging averages roughly 0.20MB per reconciliation.
- All `Account.Reconcile` work averages roughly 1.25MB per reconciliation.
- Exact averages remain estimates because the report does not contain the exact reconciliation count for this profile.
- Graph cost grows with accumulated evidence, not current active evidence. Late reconciliations are therefore more expensive than early ones.
- Approximately linear graph growth plus repeated full scans creates approximately quadratic cumulative work across a long cycle.

## 4. Required clones versus redundant clones

### Required by present contracts

- `Ledger.Result` at the Account terminal boundary must detach slices and maps: `wiki/design/packages/ledger.md:191-198`.
- `Account.Result` must alias no mutable child state: `wiki/design/packages/account.md:239-255`.
- Cross-owner `Account.Trade` and `Account.ActiveOrders` must not expose mutable domain pointers.
- `Simulator.Result` must detach its history slices: `internal/simulator/simulator.go:409-421`.
- Controller and ResultPublisher need retained terminal values after Account, Ledger, and Simulator teardown.
- `max` persistence needs atomic storage and publication ordering.
- Failed reconciliation must leave domain state and cursor unchanged: `wiki/design/packages/account.md:164-166,267-278`; `wiki/design/packages/ledger.md:113-123,209-224`.

### Redundant or over-applied boundaries

Confirmed:

- `cloneTrades` is same-owner staging. It is not justified by cross-owner immutability.
- It clones unaffected terminal Trades, inactive Orders, and immutable execution identity.
- `indexCLOIDs` snapshots every nested Fill merely to read `OrderID` and `CLOID`.
- `indexOrders` snapshots every nested Fill merely to read `OrderID`.
- `ActiveOrders` snapshots inactive Orders and their Fills before filtering.
- `Trade.Refresh` snapshots complete Orders and Fills although Ledger already owns mutable access through `Trade.Order`.
- Reconciliation's completeness check snapshots all Orders and Fills again.
- `Account.Reconcile` builds a complete `ledger.Result` only to aggregate a small `account.Snapshot`.
- `Trade.State` in `withTrade` constructs a complete Trade/Order/Fill snapshot merely to read `TradeID`.
- Terminal result propagation repeatedly deep-clones already detached immutable values at Executor, BotCycle, and Controller boundaries.

Judgment:

- Same-owner atomic staging is required.
- Whole-graph recursive cloning is an over-application of that requirement.
- Rebuilt indexes and full snapshots inside same-owner logic violate the spirit of allocate-once ownership.
- Terminal clone repetition is redundant by value semantics, but lower impact than reconciliation because it occurs near cycle close.

## 5. Ownership and design conflicts

### Source versus wiki

Confirmed conflict:

- Wiki says Ledger owns Trades, Trade owns Orders, and Order owns Fills: `wiki/design/packages/ledger.md:12-22`.
- Source replaces the complete owned graph on every `AddOrders`, `RecordSubmit`, and `Recon`: `internal/ledger/ledger.go:223-255,268-326,339-425`.
- Object identity is therefore not stable across ordinary accepted mutations.

Confirmed tension:

- Architecture says each mutable object has one direct owner and values cross ownership boundaries: `wiki/ARCHITECTURE.md:11-17`.
- Same-owner Ledger mechanics repeatedly convert children to detached snapshots, then use those snapshots for indexing and validation.
- This preserves encapsulation mechanically but treats internal owner access like a cross-owner boundary.

Confirmed design mismatch:

- Ledger flow says `index active local Orders`: `wiki/design/packages/ledger.md:67-72`.
- Source rebuilds a CLOID index for every Order, active or terminal: `internal/ledger/ledger.go:633-642`.
- Source has no retained identity map despite Ledger owning stable Order and CLOID identity.

Confirmed naming mismatch:

- `Ledger.Result` is documented as terminal: `wiki/design/packages/ledger.md:191-198`.
- `Account.Reconcile` invokes it on every successful dirty reconciliation: `internal/account/account.go:559-565`.
- A terminal export contract is being used as an internal aggregation helper.

No conflict:

- Telemetry follows the documented small immutable snapshot path.
- `none` and `max` persistence behavior matches architecture.
- Terminal publication correctly crosses ownership boundaries using detached values.

## 6. Correctness risks of removing clones

High risks:

- A contradictory later Order or Fill could partially mutate earlier Orders before failure.
- A `Trade.Refresh` error could leave Order state advanced but Trade metrics stale.
- A `max` persistence failure could leave memory ahead of durable state.
- Cursor and snapshot publication could advance without complete domain acceptance.
- Existing idempotence and terminal-transition guarantees could fail.
- Reused mutable objects could leak through snapshots if slice, map, pointer, decimal, or raw-data ownership is mishandled.

Specific aliasing hazards:

- `order.Input` contains `*decimal.Decimal` request and trigger prices: `internal/order/order.go:47-67`.
- `Order.Clone` and `Order.Snapshot` rely on `copyInput`: `internal/order/order.go:237-270`.
- Fill input carries an optional fee pointer, intentionally normalized during `Fill.New`: `internal/fill/fill.go:14-32,66-85`.
- Removing copies around these fields without preserving ownership can expose mutable pointer aliases.

Test evidence:

- `TestLedgerReconcilesAtomicallyAndIdempotently` proves one contradictory Fill leaves observed economics unchanged: `internal/ledger/ledger_test.go:14-119`.
- `TestLedgerMaxPersistenceRestoresEvidence` proves persisted tree reconstruction: `internal/ledger/ledger_test.go:121-217`.
- Order and Fill tests prove transition, aggregation, duplicate, and enrichment behavior.

Test gaps:

- No test mutates a returned `ledger.Result`, `account.Result`, `Trade.Snapshot`, `Order.Snapshot`, or clone and proves the owner remains unchanged.
- No test proves `Trade.Clone`, `Order.Clone`, or `Fill.Clone` are independently mutable.
- No test proves failed reconciliation after multiple accepted mutations rolls back every touched Trade, Order, Fill, index, cursor, and snapshot.
- No test distinguishes `none` atomicity from `max` transaction failure atomicity across multiple Trades.
- No allocation or stable-object-identity contract exists.

## 7. Ranked findings

### F1 — Critical — High confidence

**Every successful dirty reconciliation deep-clones the complete historical Ledger graph and snapshots it five more times.**

Evidence: `internal/account/account.go:512-565`; `internal/ledger/ledger.go:329-455,499-513,604-642`.

Impact: dominant Grid allocation and GC pressure; cost rises with historical evidence.

Profile: `Account.Reconcile` 112.6GB cumulative; `Ledger.Recon` 49.6GB; clone path 17.92GB inside reconciliation.

### F2 — High — High confidence

**Identity indexes are rebuilt through deep snapshots instead of retained or built from narrow owner access.**

Evidence: `internal/ledger/ledger.go:622-642`; `internal/trade/trade.go:225-232`; `internal/order/order.go:237-260`.

Impact: `Trade.Orders` is the largest application flat allocator at 24,999.31MB.

### F3 — High — High confidence

**`Ledger.Result`, designed as terminal export, runs after every reconciliation.**

Evidence: `internal/account/account.go:559-565`; `wiki/design/packages/ledger.md:191-198`.

Impact: 23,927.94MB cumulative allocation to produce a small Account snapshot.

### F4 — High — High confidence

**Graph identity is unstable despite explicit Ledger ownership.**

Evidence: complete map replacement in `AddOrders`, `RecordSubmit`, and `Recon`.

Impact: contradicts allocate-once intent and prevents durable indexes from naturally following object identity.

Qualification: source does not promise pointer identity externally. This is a design conflict, not observed incorrect output.

### F5 — Medium — High confidence

**Active-order queries copy inactive Orders and all nested Fills before filtering.**

Evidence: `internal/ledger/ledger.go:499-513`.

Impact: 23,926.13MB cumulative in `Ledger.ActiveOrders` during Grid profiling.

### F6 — Medium — High confidence

**Terminal immutable results are repeatedly deep-cloned across Executor, BotCycle, Controller, and BtRunner.**

Evidence: `internal/executor/executor.go:92-101`; `internal/botcycle/botcycle.go:156-165,239-246`; `internal/controller/controller.go:340-363`.

Impact: avoidable cycle-close allocation and conceptual ownership noise.

Qualification: existing profile shows publication below reconciliation cost. This is not the primary Grid bottleneck.

### F7 — Medium — High confidence

**Tests prove selected rollback outcomes, not the advertised aliasing and immutable-return contracts.**

Evidence: repository test search and `internal/ledger/ledger_test.go:14-217`.

Impact: clone removal or narrowing could silently break ownership guarantees.

### F8 — Low — Medium confidence

**Full-history reconciliation work is approximately quadratic across long-lived Grid cycles.**

Confirmed basis: graph grows while reconciliation repeatedly scans all retained evidence.

Inference: exact complexity depends on Trade growth timing, dirty cadence, and terminal Trade retention.

## 8. Smallest safe change boundary

Evidence supports only this boundary:

- Keep public immutable snapshots and terminal result contracts unchanged.
- Keep failed-reconciliation atomicity unchanged in both persistence modes.
- Confine change to Ledger-owned staging, narrow owner access, and Ledger-to-Account aggregation.
- Treat retained Order ID and CLOID identity maps as Ledger-owned state if introduced.
- Prove behavior for contradictory multi-object evidence, persistence failure, duplicate evidence, and returned-value aliasing.

Evidence does **not** support:

- Removing all cloning.
- Mutating the live graph before complete validation without rollback.
- Exposing `*Trade`, `*Order`, or `*Fill` outside Ledger ownership.
- Weakening terminal result detachment.
- Combining Ledger and Simulator persistence.
- Changing reconciliation order or venue-authority rules.

The narrowest proven target is the internal reconciliation representation. Terminal publication and telemetry should remain separate.

## 9. Open questions

Local evidence cannot answer:

- Exact successful reconciliation count in the profiled Grid process. The approximately 90,000 figure was supplied, not encoded in the profile.
- Allocation distribution by early versus late cycle or graph size.
- Whether future live Account reconciliation will be concurrent. Current BtRunner is synchronous.
- Whether stable pointer identity is a required contract or only stable domain identity.
- Whether `max` must persist unchanged rows on every reconciliation or only affected trees.
- Whether retained identity indexes must survive reload as persisted data or may be rebuilt once after load.
- Whether future consumers require independently mutable terminal copies at every hierarchy layer.
- The minimum acceptable rollback mechanism. Source proves requirements, not one replacement design.

## Final challenge

The hypothesis survives adversarial review.

Ledger-owned objects are not allocated once. They are recursively replaced on routine reconciliation and submission changes. Internal owner logic repeatedly pays immutable-boundary costs.

However, “remove clones” is not a safe conclusion. The proven defect is **whole-graph copy-on-reconcile plus repeated full snapshots**, while atomic acceptance and detached outward values remain mandatory.
