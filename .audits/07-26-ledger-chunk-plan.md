# Ledger Reconciliation Chunk Plan

Date: 2026-07-26

## Result

This plan contains sixteen executable pre-replay chunks, three stopped chunks,
and four terminal proof chunks.

Chunk 16 is the only production reconciliation cutover.

Every chunk ends with zero unexpected test failures and one complete production authority.

Future Runner heartbeat, cleanup, telemetry writing, grace, and third-failure stoppage remain non-executable.

## Recon-Only Scope

The user limited production behavior changes to reconciliation.

- `CreateTrade`, `AddOrders`, and `RecordSubmit` retain current behavior and complete-store routing.
- Chunks 11–13 are stopped and non-executable.
- Non-Recon source may change only when strictly required for Recon safety.
- Any such support change must preserve exact non-Recon behavior and proof.
- A material cross-area behavior or ownership change stops execution for user review.
- Broad static, replay, stability, and performance commands remain required verification.

## Canonical Owners and Resolved Conflicts

- Source and runtime evidence own implemented truth.
- `wiki/PROJECT.md` owns accepted terminal proof.
- `wiki/baselines/macross-grid-bot.md` owns dated Grid baseline history.
- `rtest.sh` owns exact Observer control counts and semantic gates.
- `trtest.sh` and `grtest.sh` own relational and semantic gates, not exact finance or complete domain literals.
- Exact Trade and Grid literals come from canonical docs and retained successful database, log, and profile evidence.
- `wiki/PERFORMANCE.md` owns accepted performance history and timer meanings.
- `HANDOFF.md` owns restart state. It does not override accepted baseline owners.
- Trade maximum drawdown release expectation is `4.200462813402` USDC from `wiki/PROJECT.md`.
- `HANDOFF.md` value `4.21244716452` USDC is a recorded conflict, not the release gate.
- Any new replay producing either unexplained value fails before performance interpretation.

## Terminal Baseline Manifest

### Observer — Sweep 6 Bot 9

Owners: `rtest.sh`, `wiki/PROJECT.md`, and successful result database.

- Ticks: `7,948,800`.
- Controller runs: `794,880`.
- Signal packages: `2,208`, owned by current `rtest.sh` runtime gate.
- Skipped starts: `724`.
- Cycles: `64` started, zero rejected, `64` closed.
- Stop-loss exits: `17`.
- Telemetry samples: `794,881`.
- Stop reason: `parent_stop`.
- Semantic completion: process zero, nonempty report JSON, and exact controller stop line.
- Account cursor: not applicable because Observer owns no Account or Ledger.

`wiki/PROJECT.md` still states `2,207` signal packages. Current executable gate requires `2,208`; runtime gate owns this current implemented count.

### Trade — Sweep 9 Bot 13

Literal owners: `wiki/PROJECT.md` and retained successful result evidence. `trtest.sh` owns relational and semantic gates only.

- Ticks: `7,948,800`.
- Controller runs: `794,880`.
- Cycles: `193`.
- Trades: `193`.
- Orders: `626`.
- Fills: `386`.
- Stop Orders: `47`.
- Retries: `0`.
- Capital: `1,000` USDC.
- Net PnL: `-3.90459332761` USDC.
- Ending equity: `996.09540667239` USDC.
- Maximum drawdown: `4.200462813402` USDC.
- Telemetry samples: `794,881`.
- Cursor: every terminal `account_ledger.fills_through_ms` equals that Account's final successful reconciliation timestamp; the last cycle equals `backtest_result.last_ms`.
- Semantic completion: completed flag one, exact Config match, cycle-to-cycle equity carry, terminal equity match, zero false-equity samples, nondecreasing maximum drawdown, zero legacy `close` Orders, no `.partial`, integrity `ok`, and no foreign-key rows.

### Grid — Sweep 10 Bot 14

Literal owners: the current section of `wiki/baselines/macross-grid-bot.md` and retained successful result evidence. `grtest.sh` owns relational and semantic gates only. `HANDOFF.md` is corroborating restart state.

- Ticks: `7,948,800`.
- Controller runs: `794,880`.
- Signal packages: `2,208`.
- Cycles: `50`.
- Trades: `1,982`.
- Orders: `4,697`.
- Fills: `2,636`.
- Cancellations: `2,061`.
- Stop Orders: `733`.
- Retries: `0`.
- Round trips: `585`.
- Net PnL: `-57.420074089999999993851` USDC.
- Ending equity: `942.579925910000000006149` USDC.
- Maximum drawdown: `75.791979199999999992245` USDC.
- Telemetry samples: `794,881`.
- Cursor: every terminal `account_ledger.fills_through_ms` equals that Account's final successful reconciliation timestamp; each terminal cycle cursor equals its terminal Account observation.
- Semantic completion: 1,500 Grid levels, 100 boundaries, 1,400 initialized active levels, zero active Orders, zero nonflat Accounts, exact Config match, no legacy `close`, no `.partial`, integrity `ok`, and no foreign-key rows.

## Global Green Contract

Every chunk must:

- start from all earlier chunks green;
- run its focused tests with `CGO_ENABLED=0` and `-tags noasm`;
- rerun frozen characterization oracles;
- leave zero expected or unexpected failures;
- preserve one production reconciliation authority;
- limit production behavior changes to reconciliation;
- preserve current CreateTrade, AddOrders, and RecordSubmit behavior and routing;
- preserve exact Trade, Order, Fill, cursor, raw-state, Snapshot, and Result behavior;
- preserve complete public active `order.Snapshot` values with nested Fills;
- preserve BtRunner and Sweep first-error behavior;
- avoid new dependencies, CGO, `unsafe`, native code, assembly, or generic repositories; and
- avoid replay until terminal Chunk 21.

## Worktree and ResultPublisher Isolation

Current worktree contains unrelated `runreport` to `report` rename edits.

Ledger work uses this isolation strategy:

- Freeze `internal/resultpublisher/resultpublisher.go`.
- Freeze `internal/resultpublisher/resultpublisher_test.go`.
- Freeze `internal/btrunner/btrunner.go` and command/report files throughout Ledger work.
- BtRunner first-error proof must use only new `internal/btrunner/btrunner_test.go` against existing public behavior.
- No BtRunner production seam is permitted by this plan.
- Add Ledger publication characterization only in new `internal/resultpublisher/ledger_characterization_test.go`.
- That new file imports the current package name present when its chunk starts.
- No Ledger chunk completes rename work or absorbs rename diff.
- Before each chunk, record `git diff --` for its exact write set.
- After each chunk, prove no frozen file changed from its pre-chunk bytes.
- Chunk 1 inventories compile state but has no compile gate.
- Known in-scope `internal/ledger/ledger_test.go` placeholder failure belongs to Chunk 2.
- Known in-scope `internal/ledger/store.go` partial hook formatting belongs to Chunk 5.
- Unrelated frozen rename compilation failures remain hard external blockers and cannot be repaired under Ledger scope.
- The first mandatory fully green baseline occurs after Chunks 2–5 complete their in-scope repairs and proof.

## Candidate Generation Contract

Ledger owns one monotonic in-memory generation token.

- Initial successful `Init` or reload establishes generation `1`.
- Successful CreateTrade advances generation once.
- Successful AddOrders advances generation once.
- Successful RecordSubmit advances generation once.
- Successful Recon publication advances generation once.
- Failed operations do not advance generation.
- Candidate build captures Ledger identity and current generation.
- Candidate commit compares both immediately before any SQL statement.
- Mismatch rejects before SQL.
- Discarded, failed-validation, committed, or reused candidates reject before SQL.
- Generation is synchronous and package-local. It creates no asynchronous lifetime or generic transaction machinery.
- Non-Recon generation updates are internal invalidation only. They cannot change persistence routing or published behavior.

## Index Collision Contract

Rebuild and admission have separate behavior:

- Reload derives and validates complete indexes before loaded state is published.
- Every mutation stages and validates its complete resulting index set before any SQL.
- Rebuild or staged-index collision is an invariant error. It never silently keeps one map entry.
- Candidate admission checks new or enriched identities and the complete staged indexes before SQL.
- Candidate collision rejects only that candidate and leaves published indexes unchanged.
- SQL failure publishes neither tree nor indexes.
- Successful SQL is followed only by non-failing tree/index assignment.
- `none` follows the same validated publication order without SQL.
- Zero Venue Order ID is absent from its index.
- Nonzero Venue Order ID may enrich once and cannot map to another Order.
- Venue TID identifies one immutable execution; identical repeat is idempotent, changed repeat fails.

## Exact SQL Measurement Contract

Store tests and benchmarks report these distinct metrics:

- `statements_attempted`: every `Exec` or prepared `Stmt.Exec` entered inside the measured transaction.
- `statements_succeeded`: attempted statements returning nil before commit.
- `target_rows`: exact primary-key identities the operation intends to insert or update, grouped by table.
- `rows_affected`: driver `RowsAffected` sum, reported but never treated as material-change proof.
- `material_rows_changed`: before/after SQL row values differ for the named primary keys.
- `rows_inserted`: target primary keys absent before and present after.
- `rows_deleted`: target primary keys present before and absent after.
- `commit_attempted`: transaction reached commit admission.
- `commit_succeeded`: real `Commit` returned nil.

Instrumentation begins after `Begin` succeeds and ends when `Commit` or rollback returns.

Each store test names exact expected statements and primary keys. Untouched-row proof compares complete before/after values for sentinel rows.

## Retained Grid Performance Baseline and Release Gate

The retained pre-redesign profile is:

`workspace/perf/profiles/pptest-s10-b14-20260725T152717Z/`

Required readable artifacts are exactly:

- `run-001.cpu.pprof`;
- `run-001.trace`;
- `run-001.heap.pprof`;
- `run-001.allocs.pprof`;
- `run-001.block.pprof`; and
- `run-001.mutex.pprof`.

Accepted baseline values:

- allocation profile sample type: cumulative `alloc_space`;
- total cumulative allocation: `144,495.70 MB`;
- `Account.Reconcile` cumulative allocation: `112,609.46 MB`;
- `Ledger.Recon` cumulative allocation: `49,622.44 MB`;
- CPU profile GC CPU: `53.19 seconds`;
- historical-data loop duration: `76,356 ms`;
- complete profiled process duration: `81,129 ms`.

Timer boundaries:

- historical-data loop starts immediately before BtRunner historical replay iteration and stops after exact replay verification;
- complete process starts at command entry and stops after terminal publication, report output, and profile shutdown;
- allocation values come from one complete Grid run's allocation profile using cumulative `alloc_space`;
- GC CPU comes from the same run's CPU profile;
- benchmark candidate timers remain separate and cannot substitute for profile proof.

Hard correctness gates remain exact finance, counts, cursor rules, integrity, foreign keys, and semantic completion.

Hard performance release criterion:

- equivalent profiling settings and one complete Grid run must produce `Ledger.Recon` cumulative `alloc_space` below `24,811.22 MB`, strictly less than 50 percent of `49,622.44 MB`;
- total cumulative allocation and `Account.Reconcile` cumulative allocation must both be below their retained baselines;
- no hard duration reduction is required because process and loop durations are noisy observations;
- report duration deltas without calling them regressions unless correctness or allocation gates fail.

If any retained artifact is missing or unreadable, measured profile reduction is unproven. Restore or identify equivalent retained evidence before Chunk 23; benchmark scaling alone cannot satisfy this gate.

## Chunk 1 — Inventory Inputs and Freeze Collision Boundaries

### Objective

Create a read-only execution inventory before Ledger repairs.

### Boundary

This chunk resolves ownership, baseline, and write collisions only. It changes no source.

### Prerequisites

Current worktree and this plan.

### Exact Write Set

- `.audits/07-26-ledger-implementation-manifest.md`

### Allowed / Forbidden

Allowed: record hashes, diffs, baseline owners, frozen files, and commands. Forbidden: source, wiki, handoff, replay, or rename edits.

### Steps

1. Record current hashes for all frozen ResultPublisher, BtRunner, command, and report files.
2. Record `internal/ledger/ledger_test.go` invalid `TO_CAPTURE` placeholder as a known in-scope Chunk 2 compile blocker.
3. Record `internal/ledger/store.go` partial hook formatting as known in-scope Chunk 5 repair work.
4. Record any unrelated frozen rename compile failure separately as an external blocker.
5. Record the terminal manifest above and its canonical owners.
6. Record exact write sets for Chunks 2–19.
7. Do not require green compilation in this inventory chunk.

### Commands

```sh
git --no-pager diff -- internal/ledger internal/account internal/resultpublisher internal/btrunner cmd
```

### Acceptance

Inventory is reviewable. Frozen hashes exist. Known in-scope Ledger blockers and unrelated frozen blockers are recorded separately. No source changed. Compilation is not an acceptance condition.

### Risks / Handoff

Wrong ownership contaminates later diffs. Chunk 2 starts after inventory completion and owns the first in-scope compile repair.

## Chunk 2 — Freeze Reviewable Ledger Oracles

### Objective

Capture complete successful Ledger behavior after each mutation class.

### Boundary

Tests only. Ledger is the single owner.

### Prerequisites

Chunk 1 inventory complete.

### Exact Write Set

- `internal/ledger/ledger_test.go`
- `internal/ledger/testdata/characterization/*.json`

### Allowed / Forbidden

Allowed: committed readable canonical JSON and typed field coverage. Forbidden: hash-only oracle, production edits, capture placeholders.

### Steps

1. Repair `internal/ledger/ledger_test.go` syntax and remove the invalid `TO_CAPTURE` placeholder before running characterization proof.
2. Use at least two Trades and multiple batches.
3. Capture separate readable JSON after CreateTrade, AddOrders, RecordSubmit, partial Recon, enrichment Recon, complete Recon, duplicate Recon, and `max` reload.
4. Cover partial Fill, complete Fill, metadata enrichment, equal-timestamp TID ordering, open exposure, terminal exposure, cursor, raw state, and alias independence.
5. Add an explicit field-coverage test listing every field of Ledger Result, Trade Snapshot, Order Snapshot, and Fill Snapshot.
6. Permit SHA-256 only as extra convenience beside readable JSON.
7. Remove every `TO_CAPTURE` or equivalent placeholder.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run 'TestLedgerCharacterization|TestLedgerOracleFieldCoverage' -count=10
```

### Acceptance

`internal/ledger/ledger_test.go` compiles. Every oracle is readable and separately reviewable. Exact decimals remain text. All Ledger tests are green.

### Risks / Handoff

An omitted field invalidates downstream proof. Chunk 3 extends oracles upward without editing Ledger fixtures.

## Chunk 3 — Freeze Account and Publication Oracles

### Objective

Capture successful Account Snapshot, terminal Result, public ActiveOrders, and detached publication behavior.

### Boundary

Tests only. ResultPublisher collision uses one new file.

### Prerequisites

Chunk 2 green and frozen files unchanged.

### Exact Write Set

- `internal/account/account_test.go`
- `internal/account/testdata/characterization/*.json`
- `internal/resultpublisher/ledger_characterization_test.go`
- `internal/resultpublisher/testdata/ledger_characterization/*.json`

### Allowed / Forbidden

Allowed: readable complete JSON and typed SQL row expectations. Forbidden: edits to existing ResultPublisher files or unrelated rename work.

### Steps

1. Capture Account successful Recon Snapshot and telemetry observation.
2. Capture terminal Account Result and alias independence.
3. Capture partially filled public ActiveOrders with nested Fill evidence.
4. Capture `none` and `max` equality.
5. In the new ResultPublisher test, publish detached Account children and compare exact Ledger SQL rows.
6. Add field-coverage tests for Account Snapshot and Result.
7. Verify frozen file hashes from Chunk 1.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/account -run 'TestAccountCharacterization|TestAccountOracleFieldCoverage' -count=10
CGO_ENABLED=0 go test -tags noasm ./internal/resultpublisher -run 'TestLedgerPublicationCharacterization' -count=10
```

### Acceptance

Readable complete oracles pass. Existing ResultPublisher files remain byte-identical. All tests green.

### Risks / Handoff

Package rename drift can break imports. Stop rather than editing frozen files. Chunk 4 adds green failure characterization.

## Chunk 4 — Add Green Failure and First-Error Characterization

### Objective

Characterize current failures, including current split generation, without leaving red tests.

### Boundary

Each owner proves its own stop boundary in an exact file.

### Prerequisites

Chunks 2–3 green.

### Exact Write Set

- `internal/ledger/ledger_test.go`
- `internal/account/account_test.go`
- `internal/executor/trade_test.go`
- `internal/executor/grid_test.go`
- `internal/botcycle/botcycle_test.go`
- `internal/controller/controller_test.go`
- `internal/btrunner/btrunner_test.go`

### Allowed / Forbidden

Allowed: test fixtures using existing public behavior only. Forbidden: production seams and intentionally failing checked-in tests.

### Steps

1. Ledger proves contradictory Order, changed TID, overflow, reversal, incomplete filled Order, and terminal change publish nothing.
2. Account test asserts the current split-generation observation as expected pre-cutover behavior and remains green.
3. Trade and Grid Executor tests prove Account Recon error stops Executor policy.
4. BotCycle test proves the first Executor Recon error prevents later Executor Recon and `OnRecon`.
5. Controller test proves BotCycle Recon error prevents Risk assessment.
6. New BtRunner test proves Controller error creates no successful telemetry sample and no completed publication through existing public lifecycle and observable result-path behavior.
7. BtRunner proof edits only `internal/btrunner/btrunner_test.go` and uses existing public lifecycle, telemetry result, result-path, and `.partial` observations.
8. No production seam is allowed in this chunk. Failure to prove the assertion black-box blocks Chunk 4; it is not deferred.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account ./internal/executor ./internal/botcycle ./internal/controller ./internal/btrunner -run 'Test.*Recon.*Error|Test.*FailureCharacterization' -count=3
```

### Acceptance

All tests green. Current defect is explicitly characterized, not left red. Each higher boundary has a named owner and file.

### Risks / Handoff

A broad fake can test itself. Use narrow counters and call order. Chunk 5 adds real SQL faults.

## Chunk 5 — Prove Real SQL Fault Boundaries

### Objective

Add disabled package-local statement and commit-admission faults around real SQLite transactions.

### Boundary

Fault infrastructure and rollback proof only.

### Prerequisites

Chunk 4 green.

### Exact Write Set

- `internal/ledger/store.go`
- `internal/ledger/ledger_test.go`

### Allowed / Forbidden

Allowed: real trigger failures and package-local hooks. Forbidden: fake stores, repository interfaces, production behavior changes.

### Steps

1. Repair `internal/ledger/store.go` partial hook formatting and restore `gofmt` before adding or running SQL fault proof.
2. Confirm the formatting repair changes no normal store behavior.
3. Name fault points by operation and statement purpose, not ordinal alone.
4. Exercise real statements, trigger rollback, pre-commit admission failure, and reload.
5. Record SQL metrics under the exact measurement contract.
6. Prove hooks are disabled in ordinary construction.
7. Run the complete Chunk 2–5 package baseline, including frozen ResultPublisher and black-box BtRunner tests.

### Commands

```sh
CGO_ENABLED=0 gofmt -w internal/ledger/store.go internal/ledger/ledger_test.go
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run 'TestLedgerSQL.*Failure|TestLedgerSQLMetrics' -count=3
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account ./internal/resultpublisher ./internal/executor ./internal/botcycle ./internal/controller ./internal/btrunner -count=1
```

### Acceptance

Memory and reloaded SQL remain exact after every fault. Chunks 2–5 establish a fully green Ledger, Account, ResultPublisher, Executor, BotCycle, Controller, and BtRunner baseline.

### Risks / Handoff

Commit admission is not driver commit failure. Name it exactly. No production implementation starts unless the complete Chunk 2–5 baseline is green. Any unrelated frozen rename failure blocks here without authorizing frozen-file edits. Chunk 6 then defines candidates.

## Chunk 6 — Define Opaque Candidates and Generation Invalidation

### Objective

Build and validate opaque candidates off-path with exact generation rejection.

### Boundary

No commit, SQL, publication, or production routing.

### Prerequisites

Chunks 2–5 green.

### Exact Write Set

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account.go`
- `internal/account/account_test.go`

### Allowed / Forbidden

Allowed: package-local candidate token and immutable aggregates. Forbidden: mutable objects crossing Ledger ownership or asynchronous lifetimes.

### Steps

1. Implement the Candidate Generation Contract exactly.
2. Build candidates without published mutation.
3. Return immutable aggregates needed by Account.
4. Derive and validate Account candidate Snapshot before SQL.
5. In this pre-commit phase, reject foreign, stale, discarded, and failed-validation candidates before SQL.
6. Test invalidation after successful CreateTrade, AddOrders, RecordSubmit, reload, and current production Recon.
7. Defer real committed, finished, and reused rejection to Chunk 15's actual persistence/publication entry point.
8. Keep production Recon on the clone path.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account -run 'Test.*Candidate|Test.*Generation' -count=3
```

### Acceptance

Every invalid candidate rejects before SQL. Production authority remains clone Recon. All tests green.

### Risks / Handoff

Generation must advance once, not per internal assignment. Chunk 7 shares calculations.

## Chunk 7 — Share Exact Trade and Order Calculations

### Objective

Reuse exact financial and Fill-total algorithms in published and candidate paths.

### Boundary

Mechanical extraction only.

### Prerequisites

Chunk 6 green and reviewable oracles frozen.

### Exact Write Set

- `internal/trade/trade.go`
- `internal/trade/trade_test.go`
- `internal/order/order.go`
- `internal/order/order_test.go`

### Allowed / Forbidden

Allowed: pure calculation inputs and outputs. Forbidden: arithmetic reordering, rounding, float conversion, or lifecycle change.

### Steps

1. Preserve `(timestamp_ms, venue_tid)` sorting.
2. Preserve decimal statement order, fees, average entry, realized PnL, status, and timestamps.
3. Characterize Order Fill totals across repeated fresh test processes before selecting an order; current map iteration has no preservable order.
4. Adopt deterministic Order Fill accumulation by `(timestamp_ms, venue_tid)` as an explicit behavior decision.
5. Capture reviewable before-decision observations and the selected deterministic expected decimals.
6. Apply calculated values only after successful validation.
7. Compare candidate and current deterministic outputs field-for-field.
8. Require terminal replay in Chunk 21 to prove accepted finance after this decision.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/trade ./internal/order -count=10
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run 'TestLedgerCharacterization' -count=10
```

### Acceptance

Readable oracles remain exact. One algorithm serves both paths. All tests green.

### Risks / Handoff

Deterministic ordering can expose prior process variability. Any finance change fails readable fixtures and later terminal proof. Chunk 8 stages indexes.

## Chunk 8 — Stage Observational Indexes and Collision Proof

### Objective

Add capacity reserves and observational indexes without switching reads.

### Boundary

Indexes are staged and validated before persistence. Operational reads still ignore them.

### Prerequisites

Chunk 7 green.

### Exact Write Set

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Allowed / Forbidden

Allowed: reserved maps, pre-SQL staged index validation, and non-failing paired publication. Forbidden: read cutover, post-commit rebuild, or silent collision overwrite.

### Steps

1. Add Trade ID, Order ID, CLOID, nonzero Venue Order ID, Venue TID, and active identity indexes.
2. Reserve 1,000 Trades, 2,000 Orders, 2,000 Fills, and evidence buffers without hard limits.
3. Implement reload collision handling separately from candidate admission.
4. Build and validate reload indexes before loaded state publication.
5. For CreateTrade, AddOrders, RecordSubmit, and current Recon, stage complete resulting indexes before SQL.
6. Make candidate collision reject candidate before SQL.
7. On SQL failure, publish neither tree nor indexes.
8. After SQL success, publish tree and staged indexes through non-failing assignments only.
9. Use the same staged publication order for `none` without SQL.
10. Compare tree and indexes after every mutation and reload.
11. Leave all reads on tree scans.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run 'TestLedger.*Index|TestLedger.*Collision|TestLedger.*Capacity' -count=3
```

### Acceptance

Reload and candidate-admission collision tests are distinct. Every collision check precedes SQL. Post-SQL publication cannot fail. All tests green.

### Risks / Handoff

Reload failure is intentional corruption rejection, not ordinary behavior drift. Chunk 9 builds deltas.

## Chunk 9 — Build Exact Deltas Off-Path

### Objective

Build touched Order, Fill, Trade, metadata, diagnostic, and dirty identity deltas.

### Boundary

No SQL, publication, or production routing.

### Prerequisites

Chunks 6–8 green.

### Exact Write Set

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/order/order.go`
- `internal/order/order_test.go`
- `internal/fill/fill.go`
- `internal/fill/fill_test.go`

### Allowed / Forbidden

Allowed: composition of already proven Order, Fill, Trade, index, and candidate mechanics plus bounded diagnostics and counters. Forbidden: new public or package domain contracts, live unresolved behavior, or dual runtime comparison.

### Steps

1. Match through candidate-safe indexes.
2. Stage changed Orders and new or enriched Fills only.
3. Recalculate each touched Trade once.
4. Validate completeness and cursor monotonicity.
5. Keep terminal history outside routine work.
6. Keep unknown CLOIDs nonfatal, unadopted, and diagnostic-only.
7. Compare candidate output with clone output in tests only.
8. Stop rule: if Order or Fill needs any new public or package contract beyond Chunk 7 calculations, stop Chunk 9 green before that work and insert a separate prerequisite chunk with exact files and proof.
9. Chunk 9 may compose proven mechanics; it must not invent multiple domain mutation protocols.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/order ./internal/fill -run 'Test.*Delta|Test.*Unknown|Test.*Touched' -count=3
```

### Acceptance

Candidate output equals readable oracles. Visited counters match exact touched identities. All tests green.

### Risks / Handoff

Missing parent dirtiness diverges finance. Chunk 10 isolates complete store.

## Chunk 10 — Isolate Complete Store Ownership

### Objective

Create one explicit complete replacement operation for empty initialization and terminal publication.

### Boundary

Complete store only. No incremental operation.

### Prerequisites

Chunks 2–9 green.

### Exact Write Set

- `internal/ledger/store.go`
- `internal/ledger/publish.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/resultpublisher/ledger_characterization_test.go`

### Allowed / Forbidden

Allowed: explicit complete replacement and stale-row removal. Forbidden: existing ResultPublisher file edits or routine dirty writes.

### Steps

1. Rename complete replacement explicitly.
2. Preserve identity and Trade-before-Order-before-Fill order.
3. Keep all deletion inside complete replacement.
4. Route empty `max` Init and `ledger.Publish` only.
5. Prove stale-row removal, exact round trip, SQL metrics, and rollback.
6. Run collision-free ResultPublisher characterization.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run 'TestLedgerCompleteStore|TestLedger.*Publish' -count=3
CGO_ENABLED=0 go test -tags noasm ./internal/resultpublisher -run 'TestLedgerPublicationCharacterization' -count=3
```

### Acceptance

Complete replacement has one owner. Frozen ResultPublisher files remain unchanged. All tests green.

### Risks / Handoff

Upsert-only complete publication can leave stale rows. Non-Recon mutations retain this complete-store owner. Chunk 14 adds only Recon dirty storage.

## Chunk 11 — Cut CreateTrade Incremental Store

Status: STOPPED — violates the user’s Recon-only scope.

`CreateTrade` retains current behavior and complete-store routing.

## Chunk 12 — Cut AddOrders Incremental Store

Status: STOPPED — violates the user’s Recon-only scope.

`AddOrders` retains current behavior and complete-store routing.

## Chunk 13 — Cut RecordSubmit Incremental Store

Status: STOPPED — violates the user’s Recon-only scope.

`RecordSubmit` retains current behavior and complete-store routing.

## Chunk 14 — Prove Recon Dirty Store Off-Path

### Objective

Implement exact dirty Recon persistence without production activation.

### Boundary

Disposable store tests only. Clone Recon remains production authority.

### Prerequisites

Chunks 9–10 green. Chunks 11–13 remain stopped.

### Exact Write Set

- `internal/ledger/store.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account_test.go`

### Allowed / Forbidden

Allowed: dirty Trades, Orders, new/enriched Fills, cursor, metadata, and raw state. Forbidden: production Recon routing.

### Steps

1. Name exact target primary keys and expected statements.
2. Persist one already validated candidate to a disposable store.
3. Prove untouched rows survive.
4. Prove exact SQL metrics and reload equality.
5. Prove Snapshot validation failure executes zero SQL statements.
6. Prove statement and commit-admission faults publish no memory.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account -run 'Test.*ReconDirtyStore|Test.*SnapshotValidation.*SQL' -count=5
```

### Acceptance

Dirty Recon store is callable and proven but unreachable from production Recon. All tests green.

### Risks / Handoff

Dirty persistence proves SQL mechanics but does not publish memory or finish candidate lifecycle. Chunk 15 prepares exact commit and publication off-path.

## Chunk 15 — Prepare Off-Path Candidate Commit and Memory Publication

### Objective

Define and prove the exact Ledger-owned operation that consumes one validated candidate, admits persistence, and publishes memory non-fallibly.

### Boundary

The operation remains unreachable from production Recon. This chunk creates publication mechanics only.

### Prerequisites

Chunks 6, 8, 9, and 14 green. Candidate deltas, staged indexes, dirty SQL, aggregates, diagnostics, and generation rules already exist.

### Exact Write Set

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account_test.go`

### Allowed / Forbidden

Allowed: one off-path commit entry point and assignment-only memory publication. Forbidden: production Recon routing, new calculations, new SQL, index rebuild, collision checks, or allocation-dependent validation after commit.

### Steps

1. Accept only one validated candidate carrying current Ledger identity and generation.
2. Recheck identity, generation, lifecycle, and already validated staged indexes immediately before SQL.
3. Under `max`, consume persistence admission before the first SQL statement so repeated entry cannot execute SQL twice.
4. On SQL failure, keep published memory unchanged and finish the failed candidate as non-reusable.
5. After successful SQL, assign touched domain objects, complete staged indexes, cursor, reconciliation time, raw account state, diagnostics, and next generation.
6. Under `none`, execute the same pre-publication checks and identical memory publication operation without SQL.
7. Make the publication sequence assignment-only, preallocated, and non-failing.
8. Transition the candidate to committed, then finished, without any fallible call.
9. Prove committed, finished, and reused candidates reject through the real entry point before SQL.
10. Prove foreign, stale, discarded, and failed-validation candidates also reject before SQL.
11. Prove `none` and `max` publish canonically identical memory.
12. Prove the operation is unreachable from production Recon.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account -run 'Test.*CandidateCommit|Test.*MemoryPublication|Test.*CandidateLifecycle' -count=10
```

### Acceptance

Exact candidate-to-live mapping is proven. Index validation precedes SQL. Post-commit publication cannot fail. Every lifecycle test is green.

### Risks / Handoff

Any allocation or validation after commit recreates split generations. Chunk 16 performs routing and deletion only.

## Chunk 16 — Atomic Recon Routing Cutover

### Objective

Route production Recon to prebuilt candidate mechanisms and delete old Recon routing.

### Boundary

This is the sole atomic cutover. Maximum production edit shape is routing calls, deleting clone-path calls, and changing the green pre-cutover defect expectation.

### Prerequisites

Chunks 2–15 green. Every calculator, SQL statement, index rule, Snapshot formula, lifecycle transition, generation rule, and publication assignment already exists and has direct tests.

### Exact Write Set

- `internal/account/account.go`
- `internal/account/account_test.go`
- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Allowed / Forbidden

Allowed: call existing gather, build, validate, commit, and publish operations; delete old clone and terminal Result routing. Forbidden: new algorithms, SQL, formulas, index rules, or diagnostics.

### Steps

1. Route Account through gather, candidate build, Snapshot validation, exact commit, Ledger publication, then Account publication.
2. Route Ledger commit and publication through Chunk 15's proven entry point.
3. Do not add or alter tree/index assignments.
4. Delete Recon `cloneTrades` and terminal `Ledger.Result` use.
5. Change Chunk 4 Account test from expected current split generation to expected failure-before-publication.
6. Keep all higher-level first-error tests unchanged and green.
7. If any new mechanism is needed, stop and return it to Chunks 6–15.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account -count=1
CGO_ENABLED=0 go test -tags noasm ./internal/executor ./internal/botcycle ./internal/controller ./internal/btrunner -run 'Test.*Recon.*Error|Test.*FailureCharacterization' -count=3
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account -run 'Test.*Characterization|Test.*Candidate|Test.*SQL' -count=10
```

### Acceptance

One delta Recon authority exists. No fallible work follows commit. All tests green. Rollback is whole-chunk revert.

### Risks / Handoff

Partial routing creates dual authority. Chunk 17 switches Ledger reads only.

## Chunk 17 — Switch Ledger Identity Reads

### Objective

Switch Ledger internal identity lookups to proven indexes.

### Boundary

Ledger reads only. Public ActiveOrders stays complete. Account stays unchanged.

### Prerequisites

Chunk 16 green.

### Exact Write Set

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`

### Allowed / Forbidden

Allowed: Order ID, CLOID, Venue Order ID, Venue TID, and active identity lookups. Forbidden: Account narrow values or benchmarks.

### Steps

1. Switch one lookup family at a time.
2. Compare tree and indexed answers after every transition.
3. Preserve complete public snapshots and ordering.
4. Prove partially filled public ActiveOrders retains nested Fills.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run 'TestLedgerIndexedRead|TestLedger.*Active|TestLedgerCharacterization' -count=10
```

### Acceptance

Indexed and tree identities match exactly. All tests green.

### Risks / Handoff

Missing transitions can skip repair. Chunk 18 narrows Account missing-status reads only.

## Chunk 18 — Switch Account Missing-Status Reads

### Objective

Add one narrow internal active-Order value for Account reconciliation.

### Boundary

Only missing-status reconciliation changes. Cancellation remains unchanged.

### Prerequisites

Chunk 17 green.

### Exact Write Set

- `internal/ledger/ledger.go`
- `internal/ledger/ledger_test.go`
- `internal/account/account.go`
- `internal/account/account_test.go`

### Allowed / Forbidden

Allowed: package-internal identity/status value. Forbidden: changing `Account.ActiveOrders()`, `Ledger.ActiveOrders()`, cancellation validation, or public snapshots.

### Steps

1. Define the narrow value with only required identity, status, active, and ownership fields.
2. Use it for missing exact Order status checks.
3. Prove public and narrow sets contain identical active identities.
4. Prove the narrow type contains no Fill tree.
5. Prove Grid and Trade shutdown still receive complete public snapshots.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account ./internal/executor -run 'Test.*NarrowActive|Test.*ActiveOrders|Test.*Stop' -count=10
```

### Acceptance

Missing-status reads are narrow. Public and cancellation behavior is unchanged. All tests green.

### Risks / Handoff

A narrow type escaping outward drops evidence. Chunk 19 adds measurement only.

## Chunk 19 — Add Exact Scaling Benchmarks

### Objective

Measure candidate Recon scaling without changing production behavior.

### Boundary

Test and benchmark files only.

### Prerequisites

Chunks 15–17 green.

### Exact Write Set

- `internal/ledger/recon_benchmark_test.go`
- `internal/ledger/ledger_test.go`

### Allowed / Forbidden

Allowed: deterministic fixtures, counters, and SQL metrics. Forbidden: production tuning or wall-clock release gates.

### Steps

1. Use retained terminal history cases `0`, `1,000`, and `10,000` Orders with active Orders `10` and incoming Fills `10`.
2. Use active Order cases `1`, `10`, `100`, and `1,000` with retained Orders `10,000` and incoming Fills `10`.
3. Use incoming Fill cases `0`, `1`, `10`, `100`, and `1,000` with retained Orders `10,000` and active Orders `1,000`.
4. Run every case in `none` and `max`.
5. Build fixtures before timing.
6. Start timer immediately before candidate build.
7. Stop timer immediately after candidate validation for `none`.
8. For `max`, separately time candidate build/validation and transaction Begin-through-Commit.
9. Report allocations, `visited_active_orders`, `incoming_order_evidence`, `incoming_fill_evidence`, `new_fills`, `enriched_fills`, `touched_orders`, `touched_trades`, and exact SQL metrics.
10. Measure public ActiveOrders snapshot construction separately.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run 'TestReconBenchmarkFixtures|TestReconWorkCounters' -count=3
CGO_ENABLED=0 go test -tags noasm ./internal/ledger -run '^$' -bench 'BenchmarkLedgerReconScaling|BenchmarkLedgerPublicActiveOrders' -benchmem -count=5
```

### Acceptance

Retained history does not change visited routine work when active and incoming cases stay fixed. All tests green.

### Risks / Handoff

Fixture setup can pollute timing. Chunk 20 runs full static proof.

## Chunk 20 — Terminal Full Static Proof

### Objective

Prove the complete implementation before replay.

### Boundary

Commands only. No implementation edits during proof.

### Prerequisites

Chunks 1–19 green.

### Exact Write Set

None. Accepted failures return to their owning chunk.

### Allowed / Forbidden

Allowed: tests, vet, diff checks, benchmark report. Forbidden: replay or opportunistic fixes.

### Steps

1. Run focused domain packages.
2. Run full tests and vet.
3. Run benchmark matrix.
4. Verify frozen ResultPublisher hashes and exact Ledger write sets.
5. Verify no deferred Runner behavior entered source.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/fill ./internal/order ./internal/trade ./internal/ledger ./internal/account ./internal/executor ./internal/botcycle ./internal/controller ./internal/btrunner ./internal/resultpublisher -count=1
CGO_ENABLED=0 go test -tags noasm ./... -count=1
CGO_ENABLED=0 go vet -tags noasm ./...
git --no-pager diff --check
```

### Acceptance

All commands pass. No source changes occur during proof.

### Risks / Handoff

Static proof cannot establish full-run finance. Chunk 21 runs exact replay proof.

## Chunk 21 — Terminal Exact Replay Proof

### Objective

Prove Observer, Trade, and Grid against the Terminal Baseline Manifest.

### Boundary

One replay per program first. Performance interpretation is forbidden here.

### Prerequisites

Chunk 20 green.

### Exact Write Set

Generated ignored `workspace/**` evidence only.

### Allowed / Forbidden

Allowed: approved scripts and database inspection. Forbidden: baseline rewriting or performance claims.

### Steps

1. Verify scripts build with `CGO_ENABLED=0` and `-tags noasm`.
2. Run Observer, Trade, and Grid once.
3. Compare every manifest literal and relational cursor rule.
4. Record report, log, and database paths.
5. Fail immediately on unexplained mismatch, including Trade drawdown conflict.

### Commands

```sh
./rtest.sh 1 6 9
./trtest.sh 1 9 13
./grtest.sh 1 10 14
```

### Acceptance

All manifest values and semantic completion rules pass. Any mismatch blocks Chunk 22.

### Risks / Handoff

A script's broad relational gate can pass financial drift. Query exact report values separately. Chunk 22 runs stability and profiles.

## Chunk 22 — Terminal Stability and Performance Proof

### Objective

Prove fresh-process stability and measured allocation reduction.

### Boundary

Operator proof only.

### Prerequisites

Chunk 21 exact replay proof green.

### Exact Write Set

Generated ignored `workspace/**` evidence only.

### Allowed / Forbidden

Allowed: stability and profiling scripts. Forbidden: code tuning during measurement or live latency claims.

### Steps

1. Run ten fresh processes per program.
2. For any user-requested `Nx`, use `./rtest.sh N 6 9` exactly.
3. Run Grid profiling.
4. Validate all six profile artifacts.
5. Report attempted runs, pass/fail, total and average duration, replay timing, and paths.
6. Compare the new profile directly with `workspace/perf/profiles/pptest-s10-b14-20260725T152717Z/` under the Retained Grid Performance Baseline contract.
7. Verify equivalent profile settings, artifact types, sample types, and timer boundaries.
8. Report total `alloc_space`, `Account.Reconcile` cumulative `alloc_space`, `Ledger.Recon` cumulative `alloc_space`, GC CPU, loop duration, process duration, visited work, statements, and target rows.
9. Apply the hard criterion: `Ledger.Recon` must be below `24,811.22 MB`; total and Account allocation must each be below baseline.
10. Treat duration as observation only.
11. If any retained artifact is missing or unreadable, report measured reduction as unproven and fail this chunk without substituting benchmark evidence.

### Commands

```sh
./rtest.sh 10 6 9
./trtest.sh 10 9 13
./grtest.sh 10 10 14
./pptest.sh 10 14
```

### Acceptance

All attempts pass exact semantic checks. Both retained and new profile artifacts are readable. `Ledger.Recon` is below `24,811.22 MB`; total allocation is below `144,495.70 MB`; Account allocation is below `112,609.46 MB`. Finance remains exact.

### Risks / Handoff

Machine noise affects duration. Allocation and visited-work evidence carries the primary redesign claim. Chunk 23 audits all work.

## Chunk 23 — Overall Adversarial Implementation Audit

### Objective

Audit ownership, behavior, atomicity, persistence, scope, and proof.

### Boundary

Integrated audit after all terminal evidence exists.

### Prerequisites

Chunks 1–22 green.

### Exact Write Set

- One new dated `.audits/` implementation review
- Accepted fixes only in the exact owning chunk write set

### Allowed / Forbidden

Allowed: at most three audit/fix rounds. Forbidden: reopening deferred Runner work or accepting unexplained baseline drift.

### Steps

1. Trace the real Account-to-Ledger path.
2. Verify one production Recon authority.
3. Verify generation comparison occurs immediately before SQL.
4. Verify Account validation precedes SQL.
5. Verify post-commit work is assignment-only and non-failing.
6. Verify rebuild and candidate collision behavior differ exactly as specified.
7. Verify complete and four incremental store owners are explicit.
8. Verify SQL metrics and target identities are exact.
9. Verify public snapshots remain complete.
10. Verify every failure class leaves memory and SQL unchanged.
11. Verify terminal manifest and performance evidence.
12. Verify frozen ResultPublisher files were not absorbed.
13. Verify heartbeat, cleanup, telemetry writer, grace, and third-failure behavior remain absent.

### Commands

```sh
CGO_ENABLED=0 go test -tags noasm ./... -count=1
CGO_ENABLED=0 go vet -tags noasm ./...
git --no-pager diff --check
git --no-pager diff -- internal/account internal/ledger internal/trade internal/order internal/fill internal/executor internal/botcycle internal/controller internal/btrunner internal/resultpublisher
```

### Acceptance

Audit verdict is PASS with no unresolved blocker or material finding. Every accepted fix receives focused and proportional proof.

### Risks / Handoff

Tests alone can miss dual authority. Audit source flow and SQL evidence directly. Work completes only after PASS.

## Dependency Summary

```text
1 ownership and collision manifest
  -> 2 reviewable Ledger oracles
  -> 3 Account and publication oracles
  -> 4 green failure and first-error characterization
  -> 5 real SQL failure proof
  -> 6 candidate generation protocol
  -> 7 exact calculation reuse
  -> 8 observational indexes and collisions
  -> 9 off-path deltas
  -> 10 complete store
  -> 11 STOPPED CreateTrade store
  -> 12 STOPPED AddOrders store
  -> 13 STOPPED RecordSubmit store
  -> 14 off-path Recon store
  -> 15 off-path candidate commit and memory publication
  -> 16 atomic routing-only Recon cutover
  -> 17 Ledger indexed reads
  -> 18 Account narrow reads
  -> 19 scaling benchmarks
  -> 20 full static proof
  -> 21 exact replay proof
  -> 22 stability and performance proof
  -> 23 overall audit
```

## Rollback Safety

- Chunks 2–5 are proof infrastructure.
- Chunks 6–9 remain off the production Recon path.
- Chunk 10 owns complete storage without rerouting non-Recon mutations.
- Chunks 11–13 are stopped by the Recon-only scope.
- Chunk 14 owns only the off-path Recon dirty-store operation.
- Chunk 15 is off-path publication preparation.
- Chunk 16 reverts as one routing/deletion unit.
- Chunks 17 and 18 revert independently to prior reads.
- Chunk 19 is test-only.
- No failed chunk hands off red tests.
- No baseline changes to accommodate unexplained output.

## Explicitly Deferred and Non-Executable

- Runner implementation and heartbeat.
- Live Hyperliquid pagination and completeness.
- Unresolved-Order quarantine and cleanup.
- Cleanup cadence and safety policy.
- Runner telemetry writer and durability.
- CreateTrade incremental persistence.
- AddOrders incremental persistence.
- RecordSubmit incremental persistence.
- First/second failure grace.
- Third-consecutive-failure stoppage.
- Live stopping gate and notifications.
- Balance/equity observability cadence changes.
- Telemetry retention and downsampling.
- Terminal Result clone reduction.
