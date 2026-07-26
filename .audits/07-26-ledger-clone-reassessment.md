# Ledger Clone Reassessment

Date: 2026-07-26

Scope: current Ledger, reconciliation, persistence, tests, and retained allocation evidence. No source or wiki changed. No replay ran.

## 1. Verdict

The prior audit remains correct about allocation amplification.

Its rollback framing is now materially wrong for the clarified contract.

The required invariant is failure-before-publication, not mutation followed by memory rollback.

The current full-graph clone satisfies Ledger-local staging. It is not the required or minimum design.

The current Account path still violates the clarified publication boundary.

`Ledger.Recon` publishes before `Ledger.Result`, marking, and `accountSnapshot` complete.

A post-publication snapshot failure leaves mutated Ledger state observable while `Account.Reconcile` returns failure.

Current `max` persistence also rewrites every identity. The contract requires one transaction containing only dirty identities.

Live three-failure behavior does not exist. Current source implements synchronous Simulator-backed backtest behavior only.

## 2. Governing contract

Exchange truth is authoritative.

Sweep and backtest must stop loudly on the first reconciliation error.

They must not recover, retry reconciliation, or roll memory backward.

Future live success must publish current exchange truth and reset consecutive failures.

Live failures one and two may tolerate only transient transport or truncated-response faults.

Those failures must retain the last fully published Ledger and Account snapshot.

The third consecutive failure starts stoppage immediately.

Stoppage must block decisions and database writes, gracefully stop, and notify loudly.

No failed reconciliation may expose partial mutations during grace.

Validation must complete before publication through a candidate state or exact staged deltas.

Full-graph clone rollback is not the target.

`max` persistence must write only dirty identities in one SQL transaction after complete validation.

SQL rollback is valid. Memory rollback is not.

## 3. Confirmed current flow

```text
Controller.Run
  -> BotCycle.Reconcile
  -> Executor.Reconcile
  -> Account.Reconcile
       clear dirty
       read open Simulator Orders
       read bounded Simulator Fills
       read missing active Order statuses
       read Simulator account state
       marshal account state
       Ledger.Recon
         clone complete Trade/Order/Fill graph
         apply Order evidence to clone
         apply Fill evidence to clone
         refresh every cloned Trade
         validate filled Orders
         persist complete cloned graph when max
         publish cloned graph, cursor, timestamp, raw state
       Ledger.Result
         mark and copy complete published graph
       accountSnapshot
       publish lastSnapshot
       increment reconciles
       retain clean dirty flag
  -> Controller builds RiskInput
  -> Risk decides
  -> Executor.OnRecon
```

Sources:

- `internal/account/account.go:467-569`
- `internal/ledger/ledger.go:329-455`
- `internal/botcycle/botcycle.go:199-218`
- `internal/executor/grid.go:348-353`
- `internal/executor/trade.go:263-268`

## 4. Actual mutation and failure ordering

### 4.1 Account before Ledger admission

Confirmed facts:

1. `Account.Reconcile` sets `dirty=false` before any Venue read.
2. A deferred function restores `dirty=true` when the operation returns failure.
3. Simulator reads return values. They do not mutate Ledger.
4. Missing active status lookup may fail before Ledger admission.
5. Account-state acquisition may fail before Ledger admission.
6. JSON marshaling may fail before Ledger admission.

Sources: `internal/account/account.go:476-545`.

Current Simulator methods do not model transport truncation classification.

Future live transport failures remain unimplemented.

### 4.2 Ledger staging mutations

Confirmed facts:

1. `cloneTrades` recursively clones every Trade, Order, and Fill.
2. Order evidence mutates only cloned Orders.
3. Fill evidence mutates only cloned Orders and Fills.
4. `Trade.Refresh` mutates only cloned Trade metrics.
5. Filled-order completeness checks inspect cloned snapshots.
6. Any error before publication discards the local clone by return.
7. Live Ledger pointers remain unchanged before `persistCandidate` succeeds.

Sources:

- `internal/ledger/ledger.go:335-418,604-642`
- `internal/order/order.go:157-234,262-270`
- `internal/fill/fill.go:66-147`
- `internal/trade/trade.go:147-217`

The implementation uses discard-before-publication, not rollback of published memory.

That distinction is correct and must survive redesign.

### 4.3 Ledger publication

Confirmed facts:

After validation, `persistCandidate` runs first.

For `none`, it returns immediately.

For `max`, SQL commits before memory publication.

Only then does Ledger replace:

- `trades`;
- `fillsThroughMS`;
- `lastReconMS`; and
- `accountStateRaw`.

Sources: `internal/ledger/ledger.go:405-425,573-578`.

Ledger has no failing operation after these assignments inside `Ledger.Recon`.

Therefore, `Ledger.Recon` itself currently publishes atomically relative to its returned error.

### 4.4 Post-Ledger-publication failure

Confirmed critical fact:

`Account.Reconcile` calls `Ledger.Result` after successful Ledger publication.

`Ledger.Result` can fail while marking any Trade with open exposure.

`Trade.Snapshot` requires a positive mark price for positive open quantity.

Therefore this sequence is possible:

```text
Ledger validates candidate
SQL commits candidate when max
Ledger publishes candidate memory
Ledger.Result requests marked snapshots
mark is absent or non-positive
Ledger.Result returns error
Account.Reconcile returns failure
Account restores dirty=true
Account retains prior lastSnapshot
Ledger remains newly published
```

Sources:

- `internal/account/account.go:547-568`
- `internal/ledger/ledger.go:420-454`
- `internal/trade/trade.go:236-280`

This violates the clarified live grace contract.

During failure grace, callers can observe new Ledger state through `Account.Trade` and `Account.ActiveOrders`.

Telemetry still returns the prior `lastSnapshot`, creating two simultaneously visible generations.

In `max`, durable Ledger state also advances despite the reconciliation operation reporting failure.

No memory rollback occurs. The failure is publication ordering.

### 4.5 Account publication

Confirmed facts:

`accountSnapshot` returns no error.

After `Ledger.Result` succeeds, Account assigns `lastSnapshot`, increments statistics, and clears the failure guard.

The returned snapshot then reaches Controller before Risk and Executor policy.

Sources:

- `internal/account/account.go:559-569,782-829`
- `wiki/design/concepts/recon.md:8-25`

## 5. Post-mutation failure points

### 5.1 Staged-only mutations

These failures mutate only the cloned candidate:

- Later Order evidence contradicts an earlier accepted candidate transition.
- Later Fill evidence conflicts with an existing Venue TID.
- Fill quantity exceeds the requested quantity.
- Fill ownership or Venue identity mismatches.
- `Trade.Refresh` detects a reversing Fill.
- `Trade.Refresh` detects changed terminal Trade values.
- Filled Order lacks complete Fill evidence.
- SQL begin, delete, insert, or commit fails under `max`.

Sources:

- `internal/ledger/ledger.go:342-418`
- `internal/order/order.go:157-234`
- `internal/fill/fill.go:88-116`
- `internal/trade/trade.go:147-195,283-330`
- `internal/ledger/store.go:53-226`

These failures do not currently mutate published Ledger memory.

### 5.2 Published-memory mutation before Account success

One confirmed failure point exists:

- `Ledger.Result` can fail after Ledger memory and SQL publication.

This is the material contract defect.

### 5.3 Persistence mutation

SQL statements mutate the transaction before later SQL failures.

Deferred SQL rollback is valid under the clarified contract.

The transaction remains invisible as committed truth until `Commit` succeeds.

Source: `internal/ledger/store.go:53-226`.

## 6. Exchange authority assessment

Confirmed facts:

- Account gathers Simulator truth before calling Ledger.
- Ledger applies matching Venue Order and Fill evidence.
- Unknown CLOIDs are silently ignored.
- Absence from bounded Fill history deletes nothing.
- Missing active local Orders receive exact status lookup.
- Cursor advances only with an accepted reconciliation candidate.

Sources:

- `internal/account/account.go:485-554`
- `internal/ledger/ledger.go:335-425`

Adversarial finding:

The source does not prove one complete exchange generation.

Open Orders, Fills, status queries, and account state are separate reads without a shared exchange snapshot token.

For Simulator’s synchronous owner path, this is presently coherent by execution order.

For future live transport, coherence is unimplemented and must be designed explicitly.

Inference:

A live truncated response cannot be safely inferred from business validation failure.

Transport must classify incomplete responses before Ledger candidate construction.

Recommendation:

Treat only explicit transport or completeness faults as grace-eligible.

Identity conflicts, invalid transitions, contradictory evidence, invariant failures, and SQL failures are not transient reconciliation success.

The contract does not authorize tolerating those errors for two live cycles.

## 7. Sweep versus future live behavior

### 7.1 Current sweep and backtest

Confirmed facts:

- Current implementation is synchronous Simulator-backed replay.
- Reconciliation errors return through Executor and BotCycle toward Controller.
- Failed reconciliation prevents Risk and `OnRecon` for that control pass.
- No live retry counter or grace state exists.

Sources:

- `internal/botcycle/botcycle.go:199-237`
- `wiki/design/concepts/recon.md:8-25`
- `wiki/PROJECT.md:45-63`

Required behavior:

The first reconciliation error must terminate the run loudly.

No recovery, repeated reconciliation, or memory rollback belongs in sweep behavior.

The existing candidate discard still matters because failure must not publish partial evidence before process termination.

### 7.2 Future live Runner

Confirmed scope fact:

Live Runner, live Venue execution, recovery, and process control remain unimplemented.

Sources:

- `wiki/PROJECT.md:65-82`
- `wiki/ARCHITECTURE.md:224-253`

Required future state:

```text
successful complete reconciliation
  validate exchange generation
  validate Ledger candidate or exact deltas
  persist dirty identities transactionally
  publish Ledger and Account snapshot as one generation
  reset consecutive failures to zero

first or second eligible transient failure
  publish nothing
  retain prior Ledger and Account snapshot
  increment consecutive failures
  block this pass from Risk and execution policy
  notify at the required severity

third consecutive failure
  publish nothing
  set stopping gate before further decisions or DB writes
  request graceful stop
  notify loudly
```

No current source implements these live semantics.

## 8. Reassessment of prior conclusions

### Confirmed unchanged

The prior audit correctly found:

- Every dirty reconciliation deep-clones the complete historical graph.
- Reconciliation repeatedly snapshots complete or broad graph sections.
- Identity indexes rebuild from expensive snapshots.
- `Ledger.Result` is misused as a per-reconciliation aggregation input.
- Active-order queries copy inactive Orders and nested Fills before filtering.
- Terminal result propagation contains repeated deep copies.
- Allocation cost grows with retained historical evidence.

### Corrected conclusions

Prior statement:

> Reconciliation staging requires rollback semantics.

Corrected:

Reconciliation requires failure-before-publication semantics.

A fully validated candidate or exact staged deltas satisfy this without memory rollback.

Prior statement:

> The clone is currently the mechanism proving failed reconciliation leaves the Ledger unchanged.

Correction:

The clone proves only failures returned by `Ledger.Recon` leave published Ledger unchanged.

It does not protect the complete `Account.Reconcile` operation.

`Ledger.Result` can fail after Ledger publication.

Prior statement:

> Failed reconciliation must leave domain state and cursor unchanged.

Correction:

This is required, but current Account-level source does not guarantee it.

Prior statement:

> `max` persistence needs atomic storage and publication ordering.

Confirmed, with stronger requirement:

Only dirty identities may persist, in one transaction, after complete candidate validation.

Prior open question:

> The minimum acceptable rollback mechanism remained unresolved.

Resolved by the clarified contract:

Memory rollback is prohibited.

Use pre-publication candidate validation or exact staged deltas, plus SQL rollback inside one transaction.

### Rejected implications

Do not replace full cloning with direct in-place mutation plus undo records.

Do not treat the first two live failures as permission to expose new Ledger state.

Do not let SQL success precede possible Account-generation validation failures.

Do not use stale `lastSnapshot` beside newly published Ledger evidence.

Do not add live recovery behavior to sweep or BtRunner.

## 9. Allocation evidence reassessment

Retained profile source:

`workspace/perf/profiles/pptest-s10-b14-20260725T152717Z/run-001.allocs.pprof`

Prior reported `alloc_space`:

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

Confirmed interpretation:

The profile strongly proves full-graph cloning and broad snapshots are major allocation owners.

It also proves deleting only `cloneTrades` cannot remove most reconciliation allocation.

`Trade.Orders`, `ActiveOrders`, `Ledger.Result`, and decimal calculations remain material.

Adversarial limits:

- This task did not rerun or re-open the profile.
- Values are retained evidence from the prior audit.
- The profile proves allocation attribution, not correctness.
- The supplied approximate reconciliation count was not encoded in the profile.
- Per-reconciliation averages remain inference.
- The profile does not measure future live transport or persistence behavior.
- The profile does not prove exact dirty-identity savings.

Allocation conclusion:

The evidence supports removing whole-history work from the hot reconciliation path.

It does not select one exact delta representation.

Correctness and publication ordering must choose the representation first.

## 10. Minimum coherent redesign

### 10.1 Boundary

Keep public immutable terminal results unchanged.

Keep Account as reconciliation owner and Ledger as trading-tree owner.

Change only the reconciliation candidate, internal indexes, snapshot aggregation, and `max` persistence write-set.

Do not implement future live Runner behavior inside BtRunner.

### 10.2 Ledger candidate

Use exact staged deltas keyed by stable identities.

Recommended minimum delta contents:

- touched Order candidate states;
- new or enriched Fill candidate states;
- touched Trade derived metrics;
- candidate Fill cursor;
- candidate reconciliation timestamp;
- candidate account-state raw value; and
- dirty Trade, Order, and Fill identity sets.

Unchanged Trades, Orders, and Fills remain referenced read-only during validation.

No published object mutates during candidate construction.

Do not clone the complete graph.

### 10.3 Stable indexes

Ledger should own stable maps for:

- Trade ID to Trade;
- Order ID to owning Order and Trade;
- CLOID to owning Order and Trade; and
- Venue TID to existing Fill identity where useful.

Build indexes during admission and reload.

Update them only during successful publication.

Do not rebuild indexes through `Trade.Orders` snapshots.

### 10.4 Validation

Validate the complete normalized exchange generation before SQL or memory publication.

Validation must include:

- every matched identity;
- all Order transitions;
- Fill idempotence and enrichment;
- Fill quantity limits;
- complete filled-Order evidence;
- all touched Trade calculations;
- cursor monotonicity;
- account-state shape; and
- Account snapshot derivation using candidate Ledger values.

Candidate Account snapshot construction must complete before SQL commit.

This removes the current post-Ledger-publication failure.

### 10.5 Persistence

After complete validation, open one SQL transaction.

Persist only:

- dirty Trades;
- dirty Orders;
- dirty Fills;
- cursor and reconciliation metadata when changed; and
- account-state raw value when changed.

Use upserts for dirty identities.

Do not delete and rewrite complete Ledger tables.

On SQL failure, rollback the SQL transaction and publish no memory.

After SQL commit, publish exact staged memory deltas and Account snapshot without any remaining fallible work.

Publication assignments must be deterministic and non-failing.

### 10.6 Account publication generation

Account and Ledger need one accepted reconciliation generation boundary.

Minimum approach:

1. Account gathers and normalizes complete exchange evidence.
2. Ledger builds a validated candidate and candidate aggregate values.
3. Account derives and validates the candidate Account snapshot.
4. Ledger persists candidate dirty identities.
5. Ledger publishes exact deltas.
6. Account publishes `lastSnapshot`.
7. Account reports success.

Steps five and six must contain no fallible operations.

The synchronous owner chain makes adjacent non-failing assignments sufficient.

No general transaction framework is required.

### 10.7 Sweep failure policy

BtRunner should propagate the first reconciliation error to its executable boundary.

The executable must log once and exit nonzero after graceful child shutdown.

Do not count reconciliation failures or retry them in sweep code.

### 10.8 Future live failure policy

Future Runner owns consecutive reconciliation failure policy.

Account or transport must return an explicit error classification sufficient for Runner policy.

Do not infer transient status from error strings.

Runner must gate Controller decisions and writable operations while stopping.

Successful complete publication resets the live counter.

Failures one and two retain the prior generation.

Failure three sets stoppage before any later decision or DB write.

## 11. Owners and files

### Current implementation owners

`internal/account/account.go`

- Own complete exchange evidence gathering.
- Own candidate Account snapshot derivation.
- Remove post-Ledger-publication fallible work.

`internal/ledger/ledger.go`

- Own candidate deltas, validation, stable indexes, dirty identities, and non-failing publication.
- Remove full-graph reconciliation cloning.
- Stop using terminal `Result` for Account aggregation.

`internal/ledger/store.go`

- Own dirty-only transactional upserts.
- Remove complete-table deletion and rewrite from reconciliation persistence.

`internal/trade/trade.go`

- Provide narrow owner-only calculation over candidate Order and Fill values.
- Avoid complete Order snapshots for same-owner refresh.

`internal/order/order.go`

- Provide non-mutating candidate transition and Fill validation mechanics.
- Preserve immutable outward snapshots.

`internal/fill/fill.go`

- Preserve non-mutating execution and enrichment validation.

`internal/ledger/ledger_test.go`

- Own Ledger candidate, publication, identity, and SQL atomicity proof.

`internal/account/account_test.go`

- Own complete Account-generation publication proof.

### Future live owners

Future Runner owner file is not implemented and must be selected during Runner design.

That owner must hold failure counters, stopping gate, graceful-stop request, and loud notification policy.

Future live transport owner must classify transport and truncation faults before Account admission.

Do not place live failure counters in Ledger.

## 12. Sequencing

1. Add failing Account-level post-publication tests.
2. Add Ledger multi-identity candidate and SQL-failure tests.
3. Introduce stable Ledger indexes without changing behavior.
4. Introduce non-mutating Order, Fill, and Trade candidate calculations.
5. Build exact Ledger reconciliation deltas.
6. Derive candidate Account snapshot before persistence.
7. Replace full-tree SQL rewrite with dirty-only upserts.
8. Publish SQL, Ledger memory, and Account snapshot in the required order.
9. Remove reconciliation `cloneTrades` use.
10. Remove internal snapshot-based index and aggregation scans.
11. Run focused tests, full tests, vet, and allocation proof.
12. Design future live counters and stoppage only with Runner implementation.

This order keeps failing proof ahead of mechanics.

It also separates sweep correctness from unimplemented live orchestration.

## 13. Required tests

### Failure-before-publication

- Fail candidate Order transition after an earlier valid candidate transition.
- Fail candidate Fill after earlier valid Order and Fill candidates.
- Fail touched Trade calculation after candidate Order mutations.
- Fail filled-Order completeness after candidate Fill admission.
- Assert Ledger snapshots, cursor, raw state, and Account snapshot remain byte-equivalent.
- Assert `Account.Trade`, `ActiveOrders`, and telemetry expose only the prior generation.

### Current post-publication defect

- Create open exposure without an admissible mark.
- Supply otherwise valid new exchange evidence.
- Force Account snapshot derivation failure.
- Assert current test fails before redesign because Ledger advances.
- After redesign, assert Ledger, SQL, cursor, and Account snapshot do not advance.

### Dirty-only persistence

- Reconcile one Order under a Ledger containing multiple terminal Trades.
- Trace SQL or inspect row update markers.
- Prove only the Ledger row and touched identities write.
- Inject failure on one dirty Fill upsert.
- Prove SQL rollback and unchanged memory.
- Inject commit failure where practical.
- Prove no memory publication.

### Identity and idempotence

- Duplicate identical Order evidence changes nothing.
- Duplicate identical Fill evidence changes nothing.
- Changed Venue TID execution fails before publication.
- Stable Order ID and CLOID indexes remain correct after add, reload, and reconciliation.
- Unknown CLOID behavior remains explicit and tested.

### Sweep behavior

- One reconciliation error stops the control path immediately.
- Risk and `OnRecon` do not run after failure.
- Executable exits nonzero and logs loudly once.
- No second reconciliation attempt occurs.
- Grace counters remain absent from BtRunner.

### Future live behavior

These tests belong with future Runner implementation:

- First transient transport failure retains the prior generation.
- Second consecutive eligible failure retains the prior generation.
- Success after either failure publishes exchange truth and resets the counter.
- Third consecutive failure sets stopping before decisions and writes.
- Graceful stop starts and loud notification fires.
- Business invariant, identity, and SQL errors follow their separately approved fatal policy.
- Truncated responses never reach Ledger candidate construction.

### Allocation proof

- Benchmark reconciliation against increasing retained-history sizes.
- Hold touched identity count constant.
- Assert allocations scale with touched evidence, not complete history.
- Profile Grid only after correctness tests pass.

No replay was authorized or run for this reassessment.

## 14. Risks

### Correctness risks

- Candidate calculations may diverge from existing mutating domain methods.
- Dirty sets may omit a derived parent Trade.
- Stable indexes may publish before object deltas.
- SQL upserts may leave stale rows if deletion semantics later appear.
- Account and Ledger generations may diverge if fallible work remains after commit.
- Decimal pointer ownership may leak through candidate values.

### Operational risks

- Misclassifying invariant failures as transient could continue trading on invalid assumptions.
- Misclassifying transport truncation as complete truth could publish false absence.
- Third-failure gating could permit one final decision or write if set too late.
- Graceful stop may itself need reconciliation while the decision gate is closed.

### Performance risks

- Stable indexes improve scans but consume retained memory.
- Exact deltas can become complex if designed as a generic transaction engine.
- Dirty SQL upserts may still allocate heavily if built from complete snapshots.
- Trade recalculation may remain history-bound unless touched execution access narrows.

Mitigation:

Keep the delta model Ledger-specific.

Use typed, exact identity sets.

Keep outward snapshots unchanged.

Measure only after failure ordering is proven.

## 15. Confirmed facts, inference, and recommendations

### Confirmed facts

- Current reconciliation clones the complete Ledger graph.
- Candidate failures inside `Ledger.Recon` do not mutate published memory.
- `max` SQL commits before Ledger memory publication.
- `max` rewrites the complete Ledger tree.
- `Ledger.Result` can fail after Ledger and SQL publication.
- Account then returns failure while retaining the prior `lastSnapshot`.
- Current source has no live failure counter, grace, or third-failure stoppage.
- Existing allocation evidence identifies reconciliation cloning and snapshots as dominant costs.

### Inference

- Future live separate reads need an explicit completeness contract.
- Exact delta staging should scale with touched identities.
- Stable indexes should remove large internal snapshot scans.
- Synchronous non-failing adjacent publication can preserve one Account generation.

These claims require implementation proof.

### Recommendations

- Replace reconciliation full-graph cloning with exact staged deltas.
- Complete Account snapshot validation before SQL or memory publication.
- Persist only dirty identities in one SQL transaction.
- Publish no memory until SQL commit succeeds.
- Leave no fallible work after Ledger publication begins.
- Keep sweep fatal on the first reconciliation error.
- Put future live failure policy in Runner, not Ledger.
- Classify transient transport and truncation faults explicitly.
- Gate decisions and DB writes before third-failure graceful stoppage begins.

## 16. Final assessment

The previous performance diagnosis survives.

The previous safety conclusion does not survive unchanged.

Whole-graph cloning is not required for rollback.

Rollback is the wrong memory model.

The correct model is complete validation, dirty-only SQL transaction, then non-failing publication.

Current Ledger mostly stages before publication, but Account breaks the full boundary by performing fallible marked result construction afterward.

That ordering must change before live failure grace can safely retain the last published Ledger.

Sweep remains simpler: fail loudly on the first reconciliation error and stop without recovery.
