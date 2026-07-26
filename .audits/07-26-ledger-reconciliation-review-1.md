# Ledger Reconciliation Reassessment Review 1

Date: 2026-07-26

Scope: adversarial review of `.audits/07-26-ledger-reconciliation-reassessment.md` against the five updated canonical design pages, current source/tests, and both prior Ledger audits. No source, wiki, or `HANDOFF.md` changed. No replay ran.

## Verdict

**FAIL — BLOCKERS remain.**

Changes #1 through #6 are directionally correct but not yet safely executable as one ordered tranche.

The report preserves the right target: validate first, commit dirty SQL second, publish Ledger and Account memory last without fallible work.

It does not yet define enough proof or ownership mechanics to guarantee exact financial and Trade/Order/Fill preservation.

Changes #7 through #10 remain future Runner design. They are not executable implementation recommendations under current ownership and dependencies.

## BLOCKER 1 — Exact successful-result baseline is missing

**Report sections:** Recommendation / Change #1, especially “Completion proof”; Recommendation / Change #6; Final assessment.

The report requires byte-equivalent fixtures but identifies no canonical fixtures, serialization, or captured pre-change outputs.

Current tests assert selected fields and counts. They do not preserve every successful Trade, Order, Fill, cursor, snapshot, raw value, timestamp, decimal, and terminal Result.

A failing rollback test proves failure atomicity. It does not prove successful-path equivalence after candidate calculations replace current mutations.

Existing accepted Grid and Trade replay results are stronger evidence, but the requested tranche forbids assuming a replay will be authorized.

“Byte-equivalent” is also undefined for Go values containing `decimal.Decimal`, maps, pointers, and detached nested slices.

**Required correction:**

1. In Change #1, create deterministic characterization fixtures before production edits.
2. Cover CreateTrade, AddOrders, RecordSubmit, Recon, reload, terminal Result, and ResultPublisher round trip.
3. Canonically encode all nested results with stable ordering and exact decimal text.
4. Record pre-change expected bytes or exact typed values in tests.
5. Require every successful fixture to remain exact after each production change.
6. Separate unit-fixture equivalence from optional replay baseline proof.
7. Do not claim accepted full-run financial preservation without replay or equivalent captured full-run evidence.

## BLOCKER 2 — The Account/Ledger publication protocol is not executable

**Report sections:** Recommendation / Change #3 “Exact files and behavior”; Recommendation / Change #4 “Exact files and behavior”; Change #4 “Completion proof”.

The required ordering needs Account to validate its candidate snapshot before Ledger persists or publishes.

Current `Ledger.Recon` owns validation, SQL, and memory publication in one call. Current `Account.Reconcile` receives no candidate handle or commit control.

The report says Ledger will “expose narrow candidate aggregate values,” then Account derives a snapshot, then SQL and memory publish.

It does not define who owns the candidate, who invokes persistence, who invokes publication, or how stale/reused candidates are rejected.

Without that protocol, an implementation can accidentally commit before Account validation, expose candidate internals, or add a generic transaction abstraction.

**Required correction:**

Define one narrow package-local protocol in Change #4. It must state:

1. Account gathers normalized Venue values.
2. Ledger builds one opaque candidate without published mutation.
3. Ledger returns only immutable aggregate values needed by Account.
4. Account derives and validates its candidate Snapshot.
5. Account asks Ledger to commit that exact candidate.
6. Ledger performs dirty SQL, then non-failing Ledger publication.
7. Account immediately performs non-failing Snapshot/stat publication.
8. Candidate ownership is single-use and synchronous.
9. No candidate pointer or mutable domain object crosses outside Account-owned Ledger coordination.
10. No fallible call remains after SQL commit.

The exact API names may remain implementation choices. The ownership and call order may not.

## BLOCKER 3 — Dirty persistence is mixed with full-store persistence

**Report sections:** Recommendation / Change #4, especially `internal/ledger/store.go`; Recommendation / Change #6 terminal publication proof.

Current `ledgerStore.save` serves four distinct paths:

- empty `max` initialization;
- every accepted mutation;
- reconciliation;
- terminal `ledger.Publish` reconstruction.

The report says to replace reconciliation deletion/rewrite with dirty-only upserts, but does not require separate store operations.

Changing shared `save` into dirty-only behavior can omit untouched rows during terminal publication, initialization, CreateTrade, AddOrders, or RecordSubmit.

Keeping one polymorphic save operation encourages hidden mode flags and accidental full-tree writes in the hot path.

This is an executability and data-loss blocker.

**Required correction:**

1. Split complete replacement/publication from incremental mutation persistence.
2. Keep terminal `ledger.Publish` able to write one complete detached Result.
3. Give reconciliation one explicit dirty-write operation.
4. Define dirty write sets for CreateTrade, AddOrders, RecordSubmit, and Recon separately.
5. Preserve Trade-before-Order-before-Fill foreign-key ordering.
6. Prove untouched rows survive every incremental transaction.
7. Prove stale rows are handled only by complete publication, never routine reconciliation.
8. Prove reload equals published memory after each mutation class.

## BLOCKER 4 — Future telemetry ordering has no safe writer contract

**Report sections:** Recommendation / Change #9 “Exact files and behavior” and “Risks”; Blockers; Final assessment.

The report requires exactly one row per heartbeat, including failures, while keeping telemetry persistence off the Controller decision path.

It does not define synchronous versus queued publication, queue capacity, backpressure, shutdown drain, write failure, or row ordering.

A synchronous write can block decisions. An asynchronous write can lose the required row, reorder sequence, or publish after later domain state.

The listed Runner database ownership blocker does not cover this publication contract.

**Required correction:**

Move Change #9 behind an explicit Runner telemetry-writer dependency.

That design must define one writer, bounded admission, sequence assignment, failure behavior, shutdown drain, database transaction boundary, and relation to domain-generation identifiers.

Do not promise both nonblocking decisions and exactly one durable row per heartbeat until that tradeoff is resolved.

## MATERIAL 1 — Candidate financial calculations need exact algorithm reuse

**Report sections:** Recommendation / Change #3; Recommendation / Change #6.

Current Trade economics sort executions by `(timestamp_ms, venue_tid)` and apply exact `decimal` operations in that order.

A new candidate calculator can produce different decimal coefficients or timestamps despite mathematically equivalent arithmetic.

“Touched Trades” and fixture equality do not require reuse of the current calculation function or exact execution ordering.

Required correction: extract or adapt one shared pure calculation path. Preserve ordering, fee inclusion, status derivation, terminal checks, and decimal operation order.

## MATERIAL 2 — Change #2 broadens beyond reconciliation

**Report section:** Recommendation / Change #2.

Stable indexes are justified. Updating CreateTrade, AddOrders, RecordSubmit, reload, Recon, and active transitions before candidate publication exists enlarges the first production step.

An index failure or partial update can change submission behavior before the atomic reconciliation path is ready.

Required correction: introduce indexes with characterization tests, rebuild from authoritative state after existing successful publications, then switch mutations incrementally after candidate publication exists.

Do not require stable pointer identity. Require stable domain identity and index/tree equivalence.

## MATERIAL 3 — Duplicate Venue identity policy is underspecified

**Report sections:** Recommendation / Change #2 and Change #3.

“Detect duplicate identities before publication” combines unlike identities.

Trade ID, Order ID, CLOID, Venue Order ID, and Venue TID have different presence and ownership rules.

Zero Venue Order ID means absent. Venue TID identifies one immutable execution. CLOID is locally deterministic. Venue Order ID arrives later.

Required correction: define uniqueness scope, zero handling, enrichment behavior, and error class for each index.

Do not let a later valid Venue Order ID enrichment collide silently or mutate indexes before candidate acceptance.

## MATERIAL 4 — Unknown Venue activity changes current behavior

**Report section:** Recommendation / Change #3.

Current `Ledger.Recon` silently skips unknown CLOIDs. Canonical design now requires a diagnostic without adoption.

Adding a diagnostic is an approved target, but it is still an observable logic change.

Required correction: define the diagnostic owner and output. Prove it neither fails current BtRunner nor changes Trade/Order/Fill Results, cursor, Account Snapshot, or telemetry unless separately authorized.

## MATERIAL 5 — Hyperliquid pagination is not implementable from a timestamp cursor alone

**Report sections:** Recommendation / Change #8; Blockers.

The report correctly records 2,000-row responses, latest-10,000 Fill retention, inclusive boundaries, and TID deduplication.

It does not define no-progress detection, page advancement, response ordering, or the case where more than 2,000 Fills share one timestamp.

“Cleanup reads historical Orders and Fills” also fails to name the Fill endpoint. `historicalOrders` is an Order endpoint, not a combined history source.

Required correction: keep Change #8 blocked on an exact Hyperliquid adapter contract. Name each endpoint and prove pagination completeness, inclusive-boundary progress, cap handling, and truncation classification.

Never advance the committed cursor when completeness is unproven.

## MATERIAL 6 — Future Runner work is presented as ordered implementation

**Report sections:** Recommendations / Changes #7 through #10; Final assessment.

The report labels Runner authority and owners unresolved, yet specifies files, capacities, timers, telemetry schema, cleanup, and failure policy in one numbered implementation sequence.

This mixes current BtRunner execution with future Runner design and makes #7 appear ready after #6.

Required correction: split the report into an executable current tranche and a blocked future dependency graph.

Changes #7 through #10 should be design prerequisites, not implementation recommendations, until Runner package, transport, persistence, safety, and fatal-error policy exist.

## MATERIAL 7 — SQL failure injection proof is not concrete enough

**Report section:** Recommendation / Change #1.

Statement failure can use SQLite triggers. Reliable commit failure is different and may require a narrow store seam.

“Test-only SQL statement and commit failure injection” does not state how production signatures remain clean or how the real transaction path stays exercised.

Required correction: require package-local injection around transaction operations, disabled by default, with proof that tests execute real SQL statements and real rollback behavior.

Do not introduce a generic repository or store interface.

## NOTE 1 — Capacity claims are hypotheses

**Report sections:** Recommendation / Change #2 and Change #7.

The fixed capacities are canonical approved reserves, not measured optimal sizes.

The report correctly calls them growth hints. It should not imply they independently produce meaningful speed gains.

## NOTE 2 — Performance proof is proportional but incomplete

**Report section:** Recommendation / Change #6.

Allocation trends and visited-identity counts are better gates than machine-time thresholds.

A constant-touched benchmark must also vary active Orders separately from retained terminal history.

Measure `none` and `max` separately. Report SQL rows and statements for `max`.

Do not infer future live latency from BtRunner benchmarks.

## NOTE 3 — Terminal clone reduction remains correctly deferred

The report avoids mixing terminal Result optimization into reconciliation work.

That preserves detached Result ownership and keeps the primary allocation cause in scope.

## Ordered sequence assessment

The current sequence is not safe as written.

Use this corrected order:

1. Add deterministic successful-result characterization fixtures.
2. Add Account post-publication defect proof and multi-object failure proof.
3. Add real statement, rollback, and commit-failure proof seams.
4. Define the opaque Account/Ledger candidate protocol.
5. Extract one exact shared Trade calculation path.
6. Add authoritative identity indexes without switching mutation ownership prematurely.
7. Build and validate exact reconciliation deltas.
8. Derive the Account candidate Snapshot before persistence.
9. Split complete-store and dirty-store operations.
10. Persist the exact dirty candidate transactionally.
11. Publish Ledger and Account through only non-failing assignments.
12. Switch active-order and identity queries to indexes.
13. Remove reconciliation `cloneTrades` and terminal `Ledger.Result` use.
14. Prove exact fixtures after every successful mutation class.
15. Run focused tests, full tests, vet, and proportional benchmarks with `CGO_ENABLED=0` and `-tags noasm`.
16. Run replay or fresh-process proof only under separate authority.
17. Treat Runner heartbeat, cleanup, telemetry persistence, and grace policy as blocked future design.

## Final assessment

The reassessment has the correct architecture and rejects unsafe memory rollback.

It does not yet prove exact successful financial preservation or define the API and store boundaries needed for safe publication.

**Result: FAIL. Four BLOCKER findings require correction before Changes #1 through #6 are implementation-ready.**
