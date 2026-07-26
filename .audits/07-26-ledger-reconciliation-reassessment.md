# Ledger Reconciliation Reassessment

Date: 2026-07-26

Scope: five updated heartbeat/reconciliation pages, both prior Ledger audits, current source/tests, and review 1. No source, wiki, or `HANDOFF.md` changed. No replay ran.

## Verdict

Changes #1 through #9 form the executable current BtRunner tranche.

They preserve exact successful Trade, Order, Fill, cursor, snapshot, raw-state, and terminal Result behavior through deterministic characterization fixtures.

Future Runner heartbeat, cleanup, telemetry writing, and third-failure stoppage remain blocked design dependencies. They are not implementation recommendations.

# Recommendation / Change #1

**Purpose**

Capture deterministic successful-result behavior before production edits.

**Dependencies**

None.

**Exact files and behavior**

- `internal/ledger/ledger_test.go`: characterize CreateTrade, AddOrders, RecordSubmit, Recon, and `max` reload.
- `internal/account/account_test.go`: characterize successful Account reconciliation and terminal Account Result.
- Existing ResultPublisher tests: characterize complete detached publication and database round trip.
- Build one deterministic fixture containing multiple Trades, Order batches, lifecycle states, partial and complete Fills, Fill enrichment, open exposure, terminal exposure, cursor, timestamps, and raw values.
- Include one partially-filled active Order. Characterize public `Account.ActiveOrders()` with its complete `order.Snapshot`, including exact nested Fill evidence.
- Canonically encode fixture Results with stable Trade/Order/Fill ordering and exact decimal text.
- Normalize optional values explicitly. Never compare pointer addresses, map iteration order, or internal decimal representation.
- Store pre-change expected canonical bytes or complete exact typed expectations in tests.
- Run every successful mutation class through both `none` and `max` where applicable.
- Require each production change to preserve those expected bytes.

**Logic impact**

Tests only.

**Risks**

A weak encoder can hide changes. It must include every exported nested field, optional presence, cursor, raw value, timestamp, and decimal string.

**Focused tests**

- CreateTrade canonical result.
- AddOrders canonical result.
- RecordSubmit canonical result.
- Recon canonical result.
- `max` reload canonical result.
- Account Snapshot and terminal Result canonical result.
- ResultPublisher database round trip.
- Returned-result alias independence.
- Partially-filled public active Order with exact nested Fills.

**Completion proof**

- Expected fixtures are captured before production edits.
- Fixtures are deterministic across repeated test execution.
- All expected bytes remain exact after every later change.
- Unit fixtures prove only covered deterministic scenarios.
- No full-run financial-preservation claim is made without separately authorized replay or equivalent captured full-run evidence.

**Expected performance effect**

None in production.

# Recommendation / Change #2

**Purpose**

Prove all failure-before-publication boundaries using real SQL behavior.

**Dependencies**

Change #1 fixtures.

**Exact files and behavior**

- `internal/account/account_test.go`: expose the current marked-snapshot failure after otherwise valid Ledger evidence.
- `internal/ledger/ledger_test.go`: add multi-Order, multi-Fill, and multi-Trade candidate failures.
- Add package-local Ledger store fault hooks around transaction operations.
- Hooks remain nil or disabled by default and are inaccessible outside package tests.
- Hooks may fail before a selected statement, after selected statements before commit, or at commit admission.
- Statement failure also uses SQLite triggers where the real schema operation is the behavior under test.
- Tests execute real SQLite transactions, statements, constraints, rollback, and reload.
- Do not add a repository interface, generic transaction interface, or alternate fake store.
- Compare canonical fixture bytes, indexes, cursor, raw state, Account Snapshot, telemetry observation, and SQL rows before and after failure.
- Prove BotCycle and Controller stop at the first reconciliation error before Risk or `OnRecon`.

**Logic impact**

Test seams only. Production behavior remains unchanged until later fixes.

**Risks**

A hook placed above SQL can bypass the real transaction path. Hooks must surround, not replace, actual transaction operations.

**Focused tests**

- Contradictory later Order after accepted earlier candidate evidence.
- Changed same-TID Fill after accepted earlier evidence.
- Trade calculation and filled-Order completeness failures.
- Account mark-required failure after Ledger candidate validation.
- Triggered statement failure with real rollback.
- Injected pre-commit and commit failure with unchanged memory.
- First deterministic reconciliation error prevents later Executors, Risk, and `OnRecon`.

**Completion proof**

- Current post-publication defect is demonstrated before Change #7.
- Every other candidate and SQL failure leaves canonical published bytes unchanged.
- Real database reload proves rollback.
- Fault hooks are disabled in normal production construction.

**Expected performance effect**

None in normal production.

# Recommendation / Change #3

**Purpose**

Define opaque candidate construction and Account validation without activating commit.

**Dependencies**

Changes #1 and #2.

**Exact files and behavior**

- `internal/account/account.go` remains the reconciliation coordinator.
- `internal/ledger/ledger.go` remains candidate, persistence, and Ledger publication owner.
- Account gathers normalized Venue Orders, Fills, account state, cursor boundary, observation time, and mark input.
- Ledger builds one opaque package-local candidate without mutating published memory or indexes.
- Ledger returns only immutable aggregate values required to derive the Account candidate Snapshot.
- Account derives and validates that Snapshot before any persistence or publication.
- Define the later activation contract: Account will ask Ledger to commit that exact candidate only after Change #7 provides its dirty-store operation.
- Define candidate identity and state needed to reject foreign, stale, reused, discarded, or already-finished candidates before SQL.
- Candidate ownership is synchronous and single-use.
- Keep production `Ledger.Recon` routed through the current implementation during this change.
- Do not add provisional candidate commit, persistence, or publication.
- No candidate pointer, mutable Trade, mutable Order, mutable Fill, or mutable index crosses beyond Account-owned Ledger coordination.
- No fallible call remains after SQL commit.
- Exact API names remain implementation choices. This ownership and call order does not.

**Logic impact**

Defines and tests construction and validation only. It does not change production reconciliation or claim the publication defect fixed.

**Risks**

A reusable candidate permits stale commit. Exposing candidate internals permits external mutation. A fallible Account assignment after commit recreates split generations.

**Focused tests**

- Candidate construction changes no published state.
- Candidate aggregates are immutable values.
- Foreign, stale, reused, and discarded candidates fail before SQL.
- Candidate cannot be activated before the exact dirty-store operation exists.
- Account Snapshot validation failure leaves the candidate uncommittable and changes nothing.
- Protocol tests define that no fallible call may remain after the future SQL commit.

**Completion proof**

- Construction-order instrumentation proves gather, build, immutable aggregates, and Account validation.
- Production reconciliation still uses the characterized current path.
- No successful candidate commit claim exists before Change #8 activation.

**Expected performance effect**

Small candidate-control overhead. It enables safe delta and dirty-SQL optimization.

# Recommendation / Change #4

**Purpose**

Reuse the exact Trade calculation algorithm for candidate and published paths.

**Dependencies**

Change #3 protocol.

**Exact files and behavior**

- `internal/trade/trade.go`: extract or adapt one pure calculation path used by current refresh and candidate validation.
- Preserve execution ordering exactly by `(timestamp_ms, venue_tid)`.
- Preserve current decimal operation order, fee inclusion, average-entry calculations, realized PnL calculations, status derivation, opening/closing timestamps, and terminal-value checks.
- Candidate calculations receive the published executions plus exact staged Fill changes.
- They do not independently reimplement financial equations.
- Preserve current tie handling and stable Venue TID ordering.
- Preserve current Order Fill-total operation order where it affects decimal values.

**Logic impact**

Mechanical reuse only. No accepted financial or lifecycle result may change.

**Risks**

Mathematically equivalent arithmetic can produce different decimal coefficients or intermediate rounding behavior. Reordering equal-time executions can change realized PnL.

**Focused tests**

- Equal timestamps with reversed input order still calculate by Venue TID.
- Mixed entry and exit Fills preserve exact decimal strings.
- Fees and enrichment preserve exact totals.
- Candidate and current refresh outputs are canonically identical.
- Terminal Trade change remains rejected.

**Completion proof**

- Change #1 successful fixtures remain exact.
- One shared calculation path serves published refresh and candidate calculation.
- No duplicate candidate financial algorithm exists.

**Expected performance effect**

Neutral initially. Touched-Trade execution input reduces later work without changing arithmetic.

# Recommendation / Change #5

**Purpose**

Stage authoritative identity indexes before switching operational reads.

**Dependencies**

Changes #1 through #4.

**Exact files and behavior**

- `internal/ledger/ledger.go`: reserve capacities for 1,000 Trades, 2,000 Orders, and 2,000 Fills per BtRunner Account.
- Reserve reusable reconciliation evidence and dirty-identity containers. These are unmeasured growth hints, not speed claims or hard limits.
- Add authoritative-state-derived indexes for Trade ID, Order ID, CLOID, nonzero Venue Order ID, Venue TID, and active Order identity.
- First rollout phase rebuilds indexes after existing successful publication and reload.
- During this phase, current mutation ownership and current reads remain unchanged.
- Characterization tests compare index contents against the authoritative Trade tree after every operation.
- Second rollout phase updates indexes only through Change #8’s successful non-failing publication.
- Switch operational reads to indexes only in Change #9 after equivalence proof.
- Require stable domain identity and tree/index equivalence. Do not require stable pointer identity.

**Identity-specific rules**

- Trade ID: nonzero, unique within one Ledger, fixed at Trade admission.
- Order ID: nonzero, unique within one Ledger, fixed at Order admission, and mapped to exactly one Trade.
- CLOID: nonempty, unique within one Ledger, locally deterministic, fixed at Order admission.
- Venue Order ID: zero means absent and is not indexed; later nonzero enrichment is allowed once; one nonzero ID must not map to different Orders.
- Venue TID: nonzero, unique within one Ledger, and identifies one immutable execution; identical repeat is idempotent, changed execution is an invariant error.
- Active index: derived from canonical Order active state; it is not a new identity authority.
- Candidate identity collisions fail before candidate acceptance. Published indexes never mutate on rejection.

**Logic impact**

No operational read switch yet. Duplicate identity corruption becomes visible through tests before behavior depends on indexes.

**Risks**

Premature mutation switching can alter submissions. Rebuild-after-publication costs remain until Change #8.

**Focused tests**

- Tree/index equality after create, add, submit, reconcile, and reload.
- Zero Venue Order ID is absent.
- Valid Venue Order ID enrichment indexes once.
- Cross-Order Venue Order ID collision fails before publication.
- Same-TID identical repeat is idempotent; changed execution fails.
- Capacity grows beyond reserves.

**Completion proof**

- Phase-one indexes always reconstruct from authoritative state.
- Phase-two tree/index publication proof is deferred to Change #8.
- Operational indexed-read proof is deferred to Change #9.
- Change #1 canonical successful fixtures remain exact.

**Expected performance effect**

Phase one can add rebuild cost. Phase two prepares later scan removal. Fixed reserves alone have no claimed measured gain.

# Recommendation / Change #6

**Purpose**

Build exact active-Order and new-Fill reconciliation deltas.

**Dependencies**

Opaque protocol, shared Trade calculation, and staged indexes from Changes #3 through #5.

**Exact files and behavior**

- `internal/ledger/ledger.go`: candidate contains touched Order states, new or enriched Fills, touched Trade metrics, metadata, and exact dirty identity sets.
- `internal/order/order.go`: provide non-mutating candidate lifecycle and Fill-total validation using existing rules.
- `internal/fill/fill.go`: validate immutable execution identity and permitted metadata enrichment.
- Build and test this delta candidate off the production reconciliation path.
- Match known Orders through candidate-safe CLOID, Order ID, and nonzero Venue Order ID indexes.
- Match Fills through known Order ownership and Venue TID identity.
- Read untouched objects without mutation.
- Recalculate only Trades owning touched Orders or Fills, using Change #4’s exact calculation path.
- Validate filled-Order completeness for touched relevant Orders.
- Keep terminal Trades, inactive Orders, and old Fills out of routine candidate construction.
- Validate cursor monotonicity before commit.
- Do not mutate live objects and undo them.
- Keep production `Ledger.Recon` wholly on the existing clone path during this change.
- The off-path candidate must not persist, publish, or run beside the clone path as a second authority.

**Diagnostic behavior ownership**

- Current unknown CLOIDs remain non-fatal and are never adopted.
- Ledger owns one reconciliation diagnostic value containing primitive unknown Order/Fill counts and identities needed for debugging.
- Account may carry that diagnostic upward as operation evidence only.
- BtRunner executable boundary owns any eventual log emission and must log returned operational errors once.
- Diagnostic collection must not change Trade, Order, Fill, cursor, Account Snapshot, or current BtRunner telemetry.
- No new persisted diagnostic schema is part of this tranche.

**Logic impact**

No production reconciliation routing change. Off-path candidate and diagnostics are compared against characterized current behavior.

**Risks**

Dirty sets can omit parent Trades. Diagnostic growth can become unbounded; retain only bounded operation-local identities or counts.

**Focused tests**

- Active-Order status delta with no Fill.
- New Fill, identical repeat, and changed same-TID rejection.
- Multiple touched Orders refresh one Trade once.
- Reversing Fill, quantity overflow, and terminal change publish nothing.
- Unknown CLOID produces the owned diagnostic, adopts nothing, and remains non-fatal.
- Unknown activity changes no canonical Result, cursor, Account Snapshot, or telemetry.

**Completion proof**

- Off-path candidate work visits active Orders, incoming new Fills, and touched Trades only.
- Its canonical candidate outputs match current characterized outcomes.
- Production still has one active reconciliation path.
- Diagnostics have one owner and no domain side effect.

**Expected performance effect**

Major expected allocation reduction. Exact proof belongs to Change #8.

# Recommendation / Change #7

**Purpose**

Separate complete and incremental stores and prepare atomic publication.

**Dependencies**

Changes #3 through #6.

**Exact files and behavior**

- `internal/ledger/store.go`: replace polymorphic `save` ownership with explicit complete and incremental operations.
- Complete-store operation writes one complete detached Ledger tree and handles stale-row replacement.
- Keep complete-store behavior for empty `max` initialization where required and terminal `ledger.Publish` reconstruction.
- Incremental CreateTrade operation writes the Ledger identity counters, one new Trade, and its initial Orders.
- Incremental AddOrders operation writes the Ledger next-Order counter, owning dirty Trade when derived state changes, and new Orders.
- Incremental RecordSubmit operation writes only changed Orders and each touched Trade.
- Incremental Recon operation writes only dirty Trades, dirty Orders, new/enriched Fills, cursor, reconciliation metadata, and raw account state when changed.
- Every incremental transaction preserves untouched rows.
- Routine incremental operations never delete complete Ledger tables.
- Only complete-store publication handles stale-row removal.
- Preserve Trade-before-Order-before-Fill foreign-key ordering.
- Use real transactions and Change #2’s package-local failure seams.
- Implement and test candidate dirty persistence without activating candidate publication in production `Ledger.Recon`.
- Define the later post-commit Ledger and Account assignments as non-failing operations for Change #8.

**Logic impact**

Separates and proves store operations. It does not activate candidate commit or claim the split-generation defect fixed.

**Risks**

Sharing one hidden mode flag can recreate ambiguity. Missing a dirty parent can diverge SQL and memory. Complete publication must still remove stale rows.

**Focused tests**

- Complete-store replacement removes stale rows and round-trips every detached Result.
- Each mutation class writes its exact dirty set.
- Untouched rows survive every incremental transaction.
- Statement, rollback, and commit failure publish no memory.
- Reload equals published memory after CreateTrade, AddOrders, RecordSubmit, and Recon.
- Account Snapshot failure occurs before any SQL.
- `none` and `max` canonical domain Results remain equal.

**Completion proof**

- SQL trace identifies distinct complete and incremental operation paths.
- Complete and incremental store operations exist before candidate activation.
- Dirty candidate persistence is proven off-path.
- Existing production reconciliation remains characterized until Change #8.
- Change #1 successful fixtures remain exact.

**Expected performance effect**

Major `max` write reduction as retained history grows. `none` avoids complete hot-path Result construction through the candidate protocol.

# Recommendation / Change #8

**Purpose**

Perform one atomic reconciliation cutover to delta commit and non-failing Account/Ledger publication.

**Dependencies**

Changes #1 through #7. Candidate construction, exact arithmetic, delta validation, indexes, and dirty persistence must all exist first.

**Exact files and behavior**

- `internal/account/account.go`: route one reconciliation through gather, candidate build, Account Snapshot validation, exact-candidate commit, then Account publication.
- `internal/ledger/ledger.go`: atomically switch production `Ledger.Recon` from the clone path to the single-use delta candidate path.
- Activate exact dirty-store persistence from Change #7.
- After SQL commit, publish Ledger objects, indexes, cursor, raw state, and diagnostics through non-failing assignments.
- Account immediately publishes its already validated Snapshot, statistics, and dirty-state success through non-failing assignments.
- Remove reconciliation use of `cloneTrades` in this same change.
- Remove reconciliation terminal `Ledger.Result` aggregation in this same change.
- Do not leave old and new reconciliation mutation paths active together.
- Keep `cloneTrades` only for unrelated mutation paths still using it.
- Preserve current Simulator read order and proven missing created/submitted Order repair.
- Preserve BtRunner and Sweep first-error behavior. Add no live grace counters or retries.

**Logic impact**

This is the single production reconciliation cutover. It fixes the split-generation defect and activates delta persistence exactly once.

**Risks**

A partial cutover creates two authorities. Any fallible call after commit recreates mixed Account/Ledger generations.

**Focused tests**

- Full Change #1 characterization suite immediately after cutover.
- Full Change #2 failure suite on the activated path.
- Single-use, stale, foreign, reused, and discarded candidate rejection.
- `none` and `max` successful equality.
- First Executor error still prevents later reconciliation, Risk, `OnRecon`, successful telemetry, and completed publication.
- Source-level assertion or focused instrumentation proves only one production reconciliation path executes.

**Completion proof**

- Dirty-store availability precedes activation.
- Production reconciliation uses only the delta path.
- Reconciliation `cloneTrades` and terminal `Ledger.Result` calls are absent.
- Change #1 canonical bytes remain exact.
- Change #2 defect and SQL failure tests pass.

**Expected performance effect**

Activates touched-evidence memory and dirty-row SQL behavior. Measurement follows in Change #9.

# Recommendation / Change #9

**Purpose**

Switch narrow internal reads to indexes while preserving complete public active Order snapshots, then close proof.

**Dependencies**

Change #8 atomic cutover.

**Exact files and behavior**

- `internal/ledger/ledger.go`: add one narrow internal active-Order reconciliation value containing only required identity and status fields.
- The narrow value may include Order ID, Trade ID, CLOID, optional Venue Order ID, status, active state, and ownership needed for exact status lookup.
- It must not use or masquerade as public `order.Snapshot`.
- `internal/account/account.go`: use the narrow internal value for reconciliation missing-status checks and cancellation ownership validation where no public result is required.
- Switch internal Order ID, CLOID, Venue Order ID, Venue TID, and active-state reads to proven indexes.
- Preserve public `Account.ActiveOrders()` and `Ledger.ActiveOrders()` signatures and behavior.
- Public active Orders remain complete immutable `order.Snapshot` values, including nested Fills.
- Preserve exact partially-filled active Order Fill evidence characterized in Change #1.
- Keep terminal Ledger/Account Results and ResultPublisher complete and detached.
- Run focused tests, full tests, vet, and benchmarks with `CGO_ENABLED=0` and `-tags noasm`.
- Do not run replay, stress, or profiling without separate authority.

**Performance proof matrix**

- Vary retained terminal Trades/Orders/Fills while holding active Orders and incoming Fills constant.
- Vary active Orders while holding retained terminal history constant.
- Vary incoming new Fills while holding active and retained history constant.
- Run each matrix in `none` and `max` modes.
- Report allocations, visited Trades/Orders/Fills, and duration without brittle wall-clock gates.
- For `max`, report SQL statement count and rows written by table.
- Measure narrow internal reads separately from complete public snapshot construction.
- Prove retained terminal history does not drive routine reconciliation when active and incoming work stay fixed.
- Do not infer future live latency from BtRunner benchmarks.

**Logic impact**

Internal reconciliation reads narrow. Existing public active Order output does not change.

**Risks**

Using the narrow value in an outward caller would silently drop Fill evidence. A missing active index entry can skip authoritative repair.

**Focused tests**

- Partially-filled active public Order retains exact nested Fill snapshots.
- Narrow reconciliation identity contains no Fill tree and stays package-internal.
- Public and narrow active sets contain the same active Order identities.
- Active index covers every active and terminal transition.
- Grid and Trade shutdown still receive complete active snapshots.
- Inclusive duplicate Fills remain idempotent.
- Complete characterization fixture suite after indexed read switching.
- Performance matrix above.

**Completion proof**

- Change #1 canonical bytes remain exact for every successful mutation class.
- Public `Account.ActiveOrders()` behavior is unchanged.
- Internal reconciliation avoids complete nested Fill snapshots.
- Focused and full tests plus vet pass under canonical build settings.
- Allocation and visited-work results separate active work from retained history.
- No full-run financial claim is made because no replay ran.

**Expected performance effect**

Routine internal work should follow active Orders, new Fills, and touched Trades. Complete public snapshots retain their existing proportional copy cost.

## Current executable tranche

Execute only Recommendations / Changes #1 through #9, in order.

This tranche has current owners, concrete files, deterministic proof, and no dependency on Runner implementation.

## Blocked future Runner dependency graph

This section is non-executable. It contains prerequisites, not numbered implementation recommendations.

```text
Runner package and standalone persistence ownership
  -> exact live Hyperliquid adapter completeness contract
     -> heartbeat reconciliation and unresolved quarantine
        -> unresolved cleanup safety and cadence policy
  -> Runner telemetry-writer contract
     -> heartbeat JSON durability policy
  -> live error classification and stopping gate
     -> first/second transient-failure grace
     -> third-consecutive-failure stoppage
```

### Runner ownership prerequisite

Blocked on:

- Runner package owner and lifecycle implementation.
- Run database path and standalone status-write ownership.
- Event serialization and WallClock ownership.
- Direct operation while Server is absent.

Only after resolution may Runner own one drift-free ten-second heartbeat, one clock read, scheduled-boundary advancement, and per-Runner capacities.

### Hyperliquid adapter prerequisite

Blocked on an exact adapter contract for:

- `openOrders` ordering and completeness.
- `userFillsByTime` ordering, start/end semantics, 2,000-row cap, no-progress detection, and page advancement.
- More than 2,000 Fills sharing one timestamp.
- Latest-10,000 Fill retention detection.
- Inclusive-boundary Venue TID deduplication.
- Exact `orderStatus` conclusive versus inconclusive outcomes.
- `historicalOrders` latest-2,000 Order behavior.
- Cleanup Fill source, which must be named separately from `historicalOrders`.
- Truncation and incompleteness classification before Ledger candidate construction.

Never advance the committed Fill cursor when pagination completeness is unproven.

No live pagination, unresolved cleanup, or grace behavior is executable before this contract and focused adapter proof exist.

### Unresolved-Order safety prerequisite

Blocked on:

- Safety boundary for continuing other Grid levels while one level is quarantined.
- Cleanup interval default.
- Escalation age and attempt threshold.
- Exact cleanup endpoint and completeness behavior.

Approved target remains: inconclusive exact status quarantines the owning level without replacement, reuse, or assumed outcome. Sweep fails immediately.

### Runner telemetry-writer prerequisite

Blocked on an explicit writer design defining:

- one writer owner;
- synchronous versus queued publication;
- bounded admission or queue capacity;
- sequence assignment;
- ordering against domain-generation publication;
- domain-generation identifier relation;
- backpressure behavior;
- write-failure behavior;
- shutdown drain and timeout;
- database transaction boundary; and
- crash durability expectations.

Do not promise both nonblocking Controller decisions and exactly one durable JSON row per heartbeat until this tradeoff is resolved.

Only afterward may stable indexed identity, heartbeat time, sequence, and schema-version columns be combined with additive JSON reconciliation telemetry.

### Live third-failure prerequisite

Blocked on:

- Explicit transport/completeness error classification.
- Fatal policy for identity, invariant, contradictory evidence, and SQL failures.
- Runner stopping-gate ownership.
- Graceful-stop reconciliation permissions after the gate closes.
- Loud notification ownership.

Approved target remains: eligible failures one and two retain the prior generation; success resets; the third consecutive failure gates decisions and writes before stoppage.

This policy belongs in future Runner, never Ledger, Account, BtRunner, or Sweep.

## Deferred scope

- Runner implementation and every blocked dependency above.
- Live cross-process Account claims and WebSocket topology.
- Server monitoring and reconnection.
- Telemetry retention, downsampling, and historical range delivery.
- Configurable balance/equity cadence until proven observability-only.
- Equity/balance retention tiers and rollups.
- Terminal Result clone reduction.
- Replay, stress, and profile proof.

## Final assessment

The executable target is now exact:

1. characterize successful results, including partially-filled public active Orders;
2. prove failures;
3. define opaque candidate construction and Account validation without commit activation;
4. reuse exact Trade arithmetic;
5. stage indexes without switching reads;
6. build and test exact deltas off the production path;
7. split complete and dirty stores;
8. atomically activate delta commit, remove reconciliation cloning, and publish without failure; and
9. switch narrow internal reads while preserving complete public snapshots, then prove active-versus-retained scaling.

Future Runner work remains blocked and outside this implementation sequence.
