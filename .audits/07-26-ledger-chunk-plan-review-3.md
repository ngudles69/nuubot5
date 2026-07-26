# Ledger Reconciliation Chunk Plan Review 3

Date: 2026-07-26

Scope: final permitted audit of `.audits/07-26-ledger-chunk-plan.md` against reviews 1–2, canonical docs, current source/tests, scripts, and worktree. No source, wiki, or `HANDOFF.md` changed. No replay ran.

## Verdict

**FAIL — one BLOCKER remains.**

All substantive candidate, index, store, publication, deterministic-finance, baseline, and Runner-scope findings are resolved in the revised design.

The execution sequence still cannot start green.

Chunk 1 requires the current Ledger packages to compile before Chunk 2 may repair known invalid partial Ledger work.

## Review 1 blocker verification

### Red test across multiple chunks

**RESOLVED.**

Chunk 4 characterizes current split-generation behavior with a green expectation.

Chunk 16 changes that expectation during the sole Recon cutover.

No planned handoff carries an intentional red test.

### Combined incremental stores

**RESOLVED.**

Chunks 10–14 separate complete replacement, CreateTrade, AddOrders, RecordSubmit, and Recon dirty persistence.

Each operation has one owner, exact files, rollback proof, reload proof, and SQL identity proof.

### Read cutover mixed with performance work

**RESOLVED.**

Chunk 17 switches Ledger identity reads.

Chunk 18 switches only Account missing-status reads.

Chunk 19 owns benchmarks only.

Cancellation and public snapshots remain unchanged.

### Opaque characterization hash

**RESOLVED.**

Chunks 2–3 require readable JSON, separate mutation checkpoints, explicit field coverage, multiple Trades, exact decimals, nested Fills, and alias proof.

Hash-only proof and capture placeholders block acceptance.

### Missing terminal baselines

**RESOLVED.**

Observer, Trade, and Grid manifests name exact counts, finance, telemetry, cursor rules, and semantic completion.

Script ownership is now accurate.

The Trade drawdown conflict is explicit and has one release owner.

### ResultPublisher collision

**RESOLVED AS A FILE BOUNDARY.**

Existing ResultPublisher and BtRunner production files remain frozen.

Ledger publication proof uses one new test file.

Hash and diff checks prevent accidental rename absorption.

## Review 1 material verification

### Candidate generation invalidation

**RESOLVED.**

The plan defines identity, generation, every advancing operation, immediate pre-SQL comparison, failure behavior, and single-use lifecycle.

### Index collision behavior

**RESOLVED.**

Reload and candidate admission have separate policies.

Every complete resulting index set validates before SQL.

SQL failure publishes neither tree nor indexes.

Post-SQL tree/index publication is assignment-only.

### SQL row metrics

**RESOLVED.**

Attempts, successes, targets, affected rows, material changes, inserts, deletes, and commit state are separately defined.

### Higher-level first-error ownership

**RESOLVED.**

Chunk 4 names exact Executor, BotCycle, Controller, and BtRunner test files.

BtRunner proof is black-box through existing public lifecycle and result-path observations.

No frozen production seam is allowed.

### Atomic cutover size

**RESOLVED.**

Chunk 15 prepares candidate commit and memory publication off-path.

Chunk 16 performs routing, clone-path deletion, and expectation change only.

Any new mechanism returns to its prerequisite owner.

## Review 2 blocker verification

### Missing delta memory publication

**RESOLVED.**

Chunk 15 owns one off-path Ledger commit entry point.

It proves exact candidate-to-live mapping, persistence admission, lifecycle consumption, identical `none` and `max` publication, and assignment-only post-commit work.

Chunk 16 only activates that proven operation.

### Index failure after durable commit

**RESOLVED.**

Chunk 8 stages and validates complete resulting indexes before SQL for every current mutation.

Reload validates before publishing loaded state.

`none` follows the same validated publication order without SQL.

### Frozen BtRunner conflict

**RESOLVED AS A CHUNK CONTRACT.**

Chunk 4 permits only `internal/btrunner/btrunner_test.go`.

A black-box proof failure blocks the chunk instead of authorizing a production seam.

### Missing performance baseline and release criterion

**RESOLVED.**

The plan names the retained profile directory, six artifacts, sample type, exact allocations, GC CPU, timer boundaries, and durations.

Chunk 22 requires equivalent profiling and exact allocation gates.

Missing retained evidence fails proof rather than allowing benchmark substitution.

## Review 2 material verification

### Candidate lifecycle proof timing

**RESOLVED.**

Chunk 6 owns pre-commit invalid states.

Chunk 15 owns committed, finished, and reused rejection through the real entry point.

### Nondeterministic Order Fill accumulation

**RESOLVED AS AN EXPLICIT PROVEN DECISION.**

Chunk 7 first characterizes current process variability.

It selects `(timestamp_ms, venue_tid)` ordering explicitly.

Readable oracles and terminal replay reject any accepted finance drift.

### Broad delta chunk

**RESOLVED WITH A STOP RULE.**

Chunk 9 may compose proven mechanics only.

Any new Order or Fill contract requires a separate prerequisite chunk and proof.

### Script ownership

**RESOLVED.**

`rtest.sh` owns exact Observer gates.

Trade and Grid scripts own relational and semantic checks only.

Canonical docs and retained evidence own exact Trade and Grid literals.

## Remaining BLOCKER — Chunk 1 requires green compilation before known invalid Ledger repair

**Plan sections:** Worktree and ResultPublisher Isolation; Chunk 1 steps 2, 5–6, acceptance, and handoff; Chunk 2 steps 1 and 7; Chunk 5 step 1.

Chunk 1 changes only:

`.audits/07-26-ledger-implementation-manifest.md`

It forbids source edits.

Chunk 1 requires this command to compile:

```sh
CGO_ENABLED=0 go test -tags noasm ./internal/ledger ./internal/account ./internal/resultpublisher -run '^$'
```

Its acceptance says compile must be green. No downstream chunk starts otherwise.

Current `internal/ledger/ledger_test.go` does not compile.

It contains:

```go
const characterizedLedgerJSON = "096020a24a61486d24842ed5d166fa84ba11b622da1d08ed1bdb1c7d27322108"TO_CAPTURE"
```

Current `internal/ledger/store.go` also contains malformed partial hook formatting scheduled for repair in Chunk 5.

The plan itself records these files as unverified partial work.

Chunk 2 is the first owner allowed to remove `TO_CAPTURE` and repair characterization.

Chunk 5 is the owner assigned to repair fault hooks.

Neither chunk may start because Chunk 1 cannot pass.

This is a deterministic orchestration deadlock, not a possible external blockage.

An orchestrator following the plan exactly must stop before the first repair.

**Required correction:**

Choose one coherent startup order:

1. Make Chunk 1 a read-only inventory with no compile acceptance, then require Chunk 2 to restore Ledger test compilation before any production work; or
2. Move minimum mechanical repair of current malformed partial Ledger files into a new first source chunk with exact files and proof; or
3. Revert or otherwise resolve the unverified partial Ledger edits before starting this plan under separate authority.

The simplest plan correction is:

- Chunk 1 records hashes, diffs, frozen boundaries, and known compile blockers.
- Known Ledger partial compile failure is accepted only as inventory evidence.
- Chunk 2 repairs `internal/ledger/ledger_test.go`, removes placeholders, creates readable oracles, and must finish green.
- Chunk 5 repairs `internal/ledger/store.go` before its focused SQL proof.
- No production implementation chunk starts before Chunks 2–5 establish a fully green base.
- Unrelated frozen-file compile failures remain hard external blockers.

The plan must distinguish known in-scope partial Ledger repair from unrelated frozen rename failure.

## Holistic boundary assessment

Except for startup ordering, the plan is holistic and executable.

Chunks have coherent outcomes, exact files, direct proof, and green handoffs.

Store cutovers are bite-sized.

Candidate, index, persistence, and publication dependencies are correctly ordered.

Public snapshots remain complete.

Future Runner work remains excluded.

Chunk 8 is broad across current mutation classes, but one invariant requires all published trees and indexes to remain paired. Its atomic boundary is justified.

Chunk 9 is broad but has a valid stop rule against hidden contract growth.

## Performance-gate note

The 50-percent `Ledger.Recon` allocation threshold is precise and executable.

It is a plan-selected release gate, not a previously canonical measured requirement.

Failure blocks this plan’s performance acceptance. It must not be reinterpreted as finance or correctness failure.

## Final assessment

The revised plan resolves every substantive round-one and round-two architecture finding.

One startup dependency remains impossible against the current worktree.

**Result: FAIL. Correct Chunk 1 so known in-scope partial Ledger compilation is repaired before the first mandatory green compile gate.**
