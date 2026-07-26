# Ledger Reconciliation Chunk Plan Review 1

Date: 2026-07-26

Scope: adversarial audit of `.audits/07-26-ledger-chunk-plan.md` against canonical docs, reassessments, reviews 1–3, current source/tests, worktree state, and user intent. No source, wiki, or `HANDOFF.md` changed. No replay ran.

## Verdict

**FAIL — BLOCKERS remain.**

The plan has the correct target architecture, one reconciliation cutover, complete public snapshots, dirty-store ordering, and explicit future Runner exclusion.

It is not executable chunk-by-chunk without reconstructing proof and sequencing choices.

Six blockers remain.

## BLOCKER 1 — Chunk 2 leaves the required branch red through Chunk 9

**Plan sections:** Chunk 2 steps 7–8, acceptance evidence, handoff; Chunks 3–9 prerequisites and focused commands.

Chunk 2 requires one focused Account test to expose the current split-generation defect.

It explicitly leaves that test failing until Chunk 10.

Chunks 3–9 then say prior chunks pass. Their package commands include `internal/account` or broader suites that must encounter the intentional failure.

This creates eight isolated coder sessions with no valid green handoff.

A known red test also hides regressions and makes “all prerequisites pass” false.

**Required correction:**

1. Keep the defect characterization checked-in and green before cutover.
2. Assert the current split generation as the expected pre-cutover observation.
3. In Chunk 10, change that same test expectation to failure-before-publication.
4. Require every chunk handoff to have zero unexpected failures.
5. If a temporarily failing test is unavoidable, create and resolve it inside Chunk 10 only.

## BLOCKER 2 — Chunk 9 is not bite-sized

**Plan section:** Chunk 9.

Chunk 9 designs and implements four distinct SQL transaction classes:

- CreateTrade;
- AddOrders;
- RecordSubmit; and
- Recon.

It also cuts three production mutation paths from complete replacement to incremental persistence.

Each path has different counters, parent dirtiness, row sets, foreign-key ordering, reload behavior, and rollback proof.

A failure can leave production persistence partly migrated while reconciliation remains on the old path.

The stated “one at a time” routing does not create isolated chunk acceptance or rollback boundaries.

**Required correction:**

Split Chunk 9 into isolated store chunks:

1. Add shared transaction primitives and measurement only if one concrete operation requires them.
2. Implement and cut CreateTrade incremental persistence.
3. Implement and cut AddOrders incremental persistence.
4. Implement and cut RecordSubmit incremental persistence.
5. Implement Recon dirty persistence off-path only.

Each chunk must prove exact memory, exact rows, untouched-row survival, reload equality, statement failure, commit-admission failure, and canonical fixture preservation.

Chunk 10 must still remain the sole production Recon cutover.

## BLOCKER 3 — Chunk 11 combines correctness cutover with the complete performance program

**Plan section:** Chunk 11.

Chunk 11 changes Account and Ledger operational reads, introduces a new internal value, changes cancellation checks, proves public snapshot preservation, builds two benchmarks, and executes a six-dimensional matrix.

Those are separate correctness and measurement outcomes.

A benchmark defect must not block rollback or diagnosis of indexed-read behavior.

The benchmark matrix also lacks named fixture sizes and exact measured timer boundaries.

“Vary” does not tell an isolated coder which cases prove constant retained-history cost.

**Required correction:**

1. Make one chunk switch Ledger identity lookups with direct tree/index equivalence proof.
2. Make one chunk add the narrow active-Order value and switch Account missing-status checks.
3. Keep cancellation callers unchanged unless current flow proves they need the narrow value.
4. Make a later test-only chunk add the benchmark matrix.
5. Name exact retained-history, active-Order, and incoming-Fill case sizes.
6. Define timer start and stop around candidate reconciliation only.
7. Define exact visited-work counters and SQL statement/row accounting.

Public `Account.ActiveOrders()` and `Ledger.ActiveOrders()` must remain complete throughout every intermediate chunk.

## BLOCKER 4 — Characterization acceptance permits an opaque hash

**Plan sections:** Chunk 1 steps 5–6, acceptance evidence, risk; current `internal/ledger/ledger_test.go`.

The plan says every exported nested field must be encoded.

It then permits canonical encoded bytes without requiring those bytes to remain reviewable.

Current partial work returns only SHA-256 and compares it with `characterizedLedgerJSON`.

That constant is not JSON and currently contains an invalid `TO_CAPTURE` suffix.

A hash cannot show omitted fields, wrong normalization, changed decimal text, or accidental fixture drift.

The plan’s risk says to keep encoding reviewable or pair the hash with typed assertions, but acceptance does not require either.

**Required correction:**

1. Require committed reviewable canonical JSON or complete typed expectations as the primary oracle.
2. Permit a hash only as an additional convenience.
3. Require explicit field-coverage review for every Ledger, Trade, Order, Fill, Account Snapshot, and terminal Result field.
4. Capture expectations separately after CreateTrade, AddOrders, RecordSubmit, each Recon class, and reload.
5. Require at least two Trades, as the plan already states.
6. Require partial Fill, complete Fill, metadata enrichment, equal-timestamp ordering, open exposure, terminal exposure, cursor, raw state, and alias independence.
7. Remove every capture placeholder before Chunk 1 acceptance.

Without this, later exact-finance claims can pass against an incomplete oracle.

## BLOCKER 5 — Terminal finance proof does not identify one exact Trade baseline

**Plan sections:** Chunk 12 steps 3–4 and acceptance evidence.

Grid values are listed exactly. Trade values are not.

Canonical project evidence records Trade maximum drawdown as `4.200462813402` USDC.

Current `HANDOFF.md` records `4.21244716452` USDC.

The plan says “compare every accepted count and financial result” without resolving this conflict or naming one exact expected set.

An isolated executor cannot know which value blocks release.

Observer expectations are also incomplete. The plan names no exact signal, cycle, stop-loss, telemetry, cursor, or semantic completion values.

**Required correction:**

1. Resolve the Trade drawdown conflict before implementation execution.
2. Name the canonical owner for every terminal baseline.
3. List exact Trade cycles, Trades, Orders, Fills, stop Orders, retries, PnL, equity, drawdown, cursor, and semantic completion values.
4. List exact Observer signals, skipped starts, cycles, exits, telemetry, cursor, and semantic completion values.
5. Keep Grid’s complete exact set.
6. State that any unexplained mismatch fails before performance interpretation.

## BLOCKER 6 — ResultPublisher write-set collision has no executable isolation rule

**Plan sections:** Current Worktree Constraint; Chunk 1 write set; Chunk 8 write set and risks.

The worktree currently contains an unrelated `runreport` to `report` rename touching:

- `internal/resultpublisher/resultpublisher.go`;
- `internal/resultpublisher/resultpublisher_test.go`;
- `internal/btrunner/btrunner.go`; and
- command/report files.

Chunk 1 requires ResultPublisher characterization edits before that unrelated work is settled.

Chunk 8 again requires ResultPublisher proof while changing Ledger complete-store ownership.

“Preserve unrelated changes” is not an isolation procedure.

A coder cannot distinguish Ledger-required edits from rename completion, nor prove its write set without absorbing unstable unrelated code.

**Required correction:**

Choose one explicit strategy before Chunk 1:

1. Finish and verify the report rename as a separate prerequisite; or
2. Freeze ResultPublisher files and move Ledger publication characterization to a new collision-free test file; or
3. Assign one owner to integrate both changes with a named diff boundary and proof.

Every chunk must list exact files it may edit. “Existing test file if seams require it” is not sufficient for isolated sessions.

## MATERIAL 1 — Candidate generation invalidation is underspecified

**Plan sections:** Chunk 4 steps 1–3; Chunk 6; Chunk 10 step 11.

The plan requires foreign, stale, reused, discarded, and finished-candidate rejection.

It does not define which successful operations advance the generation.

CreateTrade, AddOrders, RecordSubmit, reload, and successful Recon can all invalidate an earlier candidate or its index assumptions.

**Required correction:** define the generation token owner, every advancing operation, comparison point, and pre-SQL rejection proof.

Do not add asynchronous lifetime or generic transaction machinery.

## MATERIAL 2 — Index collision behavior during observational rollout is unclear

**Plan section:** Chunk 6 steps 2–7.

Chunk 6 rebuilds indexes from the current authoritative tree while production reads remain unchanged.

It also says collisions reject before future candidate acceptance.

The plan does not state what Init or rebuild does if persisted current state contains a duplicate Venue Order ID, CLOID, or TID.

Silently keeping one map entry would create false equivalence. Failing Init changes current reload behavior.

**Required correction:** define rebuild collision handling separately from new candidate admission and test both.

## MATERIAL 3 — Store proof counts rows ambiguously

**Plan sections:** Chunk 9 steps 12 and risks; Chunk 11 performance proof.

SQLite upsert execution, `RowsAffected`, statements attempted, rows matched, and rows materially changed are different metrics.

The plan says “rows written” and later warns that upserts obscure unchanged writes.

**Required correction:** define each reported metric and its instrumentation boundary. Require exact expected statements and targeted row identities per operation.

## MATERIAL 4 — Higher-level first-error proof has no owned test seam

**Plan section:** Chunk 2 write set and steps 9–10.

The write set permits an unspecified BotCycle or Controller test file “if current seams require it.”

Current higher-level tests do not expose a named reconciliation-failure seam covering later Executors, Risk, `OnRecon`, telemetry, and publication.

That proof may require coordinated fixtures across BotCycle, Controller, and BtRunner.

**Required correction:** investigate and name the exact owner and files before assigning Chunk 2. Split package-local proof if one fixture cannot prove all boundaries cleanly.

## MATERIAL 5 — Chunk 10 remains large but may stay atomic only with mechanical prerequisites

**Plan section:** Chunk 10.

The one cutover boundary is correct.

However, Chunk 10 also removes clone aggregation, publishes indexes, changes Account snapshot construction, changes dirty success, and proves orchestration.

It is acceptable only if Chunks 4–9 leave complete callable mechanisms and Chunk 10 contains routing plus deletion, not new algorithms.

**Required correction:** state a maximum production edit shape for Chunk 10. Any new calculator, SQL statement, index rule, or snapshot formula must return to its prerequisite chunk.

## NOTE 1 — Future Runner scope is correctly excluded

The plan consistently forbids heartbeat, cleanup, unresolved quarantine, live telemetry writing, retries, grace, and third-failure stoppage.

Chunk 13 correctly audits their absence.

## NOTE 2 — Public snapshot preservation is correctly stated

The plan preserves complete detached `order.Snapshot` values with nested Fills.

The narrow internal value remains separate.

The corrected boundary from reassessment review 2 survives.

## NOTE 3 — Store and cutover ordering is directionally correct

Complete replacement precedes incremental operations.

Dirty Recon persistence precedes production Recon activation.

Tree and indexes publish together at one cutover.

Operational reads switch only afterward.

## Required revised dependency order

A safe plan should use this order:

1. Resolve worktree ownership and baseline conflicts.
2. Capture reviewable exact Ledger characterization.
3. Capture Account and ResultPublisher characterization.
4. Add green pre-cutover failure characterization.
5. Add real SQL failure seams and rollback proof.
6. Define opaque candidate construction and generation invalidation.
7. Extract exact Trade and Order calculations.
8. Stage observational indexes and collision proof.
9. Build exact deltas off-path.
10. Separate complete-store ownership.
11. Cut CreateTrade to its incremental store.
12. Cut AddOrders to its incremental store.
13. Cut RecordSubmit to its incremental store.
14. Prove Recon dirty persistence off-path.
15. Perform the sole atomic Recon cutover.
16. Switch Ledger indexed reads.
17. Switch Account narrow active reads.
18. Add exact scaling benchmarks.
19. Run full static proof.
20. Run exact Observer, Trade, and Grid replay proof.
21. Run fresh-process and profile proof.
22. Perform the overall adversarial implementation audit.

Every numbered implementation chunk must finish green and leave one complete production authority.

## Final assessment

The plan understands the approved architecture.

It fails isolated execution because it carries a known red test across eight chunks, combines multiple production cutovers, permits opaque characterization, and leaves baseline conflict unresolved.

**Result: FAIL. Do not state PASS until every chunk is green, bite-sized, collision-safe, and executable from explicit inputs and exact acceptance values.**
