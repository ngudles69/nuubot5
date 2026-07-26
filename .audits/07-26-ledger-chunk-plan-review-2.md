# Ledger Reconciliation Chunk Plan Review 2

Date: 2026-07-26

Scope: adversarial re-audit of `.audits/07-26-ledger-chunk-plan.md` against review 1, canonical docs, current source/tests, scripts, and worktree. No source, wiki, or `HANDOFF.md` changed. No replay ran.

## Verdict

**FAIL — four BLOCKERS remain.**

The revision resolves all six round-one blocker subjects and all five round-one material subjects at the stated design level.

The chunk sequence is much smaller, greener, and more file-specific.

It still omits the mechanism that publishes a committed delta candidate.

Observational index collision handling can fail after durable SQL commit.

Frozen BtRunner ownership conflicts with Chunk 4’s required proof.

Performance reduction has no exact comparison baseline or release criterion.

## Round-one BLOCKER status

### BLOCKER 1 — Red test across multiple chunks

**RESOLVED.**

Chunk 4 keeps the current split-generation characterization green.

Chunk 15 changes the same expectation during the atomic cutover.

The global contract requires every handoff to have zero expected or unexpected failures.

### BLOCKER 2 — Combined incremental stores

**RESOLVED.**

Chunks 11–14 isolate CreateTrade, AddOrders, RecordSubmit, and Recon persistence.

Each has its own write set, activation boundary, rollback proof, reload proof, and acceptance.

Chunk 15 remains the sole production Recon cutover.

### BLOCKER 3 — Read cutover mixed with benchmarks

**RESOLVED.**

Chunk 16 switches Ledger identity reads.

Chunk 17 switches only Account missing-status reads.

Cancellation remains unchanged.

Chunk 18 owns measurement only and names fixture sizes, timers, counters, modes, and SQL metrics.

### BLOCKER 4 — Opaque characterization hash

**RESOLVED.**

Chunks 2–3 require reviewable JSON, separate mutation checkpoints, field-coverage tests, multiple Trades, exact decimal text, and alias proof.

A hash is permitted only as additional evidence.

Capture placeholders block acceptance.

### BLOCKER 5 — Missing exact terminal baselines

**RESOLVED FOR CORRECTNESS PROOF.**

The plan names exact Observer, Trade, and Grid manifests.

It identifies canonical owners and records source conflicts.

Current `rtest.sh` confirms `2,208` signal packages and the listed Observer control counts.

Trade and Grid scripts provide relational gates. Chunk 20 separately requires exact manifest queries.

The Trade drawdown release gate is explicitly `4.200462813402` USDC from `wiki/PROJECT.md`.

### BLOCKER 6 — ResultPublisher collision

**RESOLVED AS AN ISOLATION STRATEGY, but affected by remaining BLOCKER 3.**

Existing ResultPublisher files are frozen.

New publication characterization uses `internal/resultpublisher/ledger_characterization_test.go`.

The plan requires pre-chunk hashes, exact diffs, and byte-identity proof.

It correctly blocks rather than absorbing unrelated rename work.

## Round-one MATERIAL status

### MATERIAL 1 — Candidate generation invalidation

**RESOLVED.**

The plan defines Ledger ownership, generation initialization, advancing operations, identity capture, pre-SQL comparison, failure behavior, and single-use state.

### MATERIAL 2 — Index collision rollout

**RESOLVED AS POLICY, but affected by remaining BLOCKER 2.**

Rebuild corruption and candidate admission now have separate rules.

Reload corruption fails before publication. Candidate collision rejects only the candidate.

The remaining problem is operation ordering, not collision classification.

### MATERIAL 3 — Ambiguous SQL row metrics

**RESOLVED.**

The plan separates attempts, successes, targets, driver rows affected, material changes, inserts, deletes, and commit state.

It defines exact timer boundaries and sentinel-row comparison.

### MATERIAL 4 — Unowned higher-level error proof

**RESOLVED AS TEST OWNERSHIP, but affected by remaining BLOCKER 3.**

Chunk 4 names exact Executor, BotCycle, Controller, and BtRunner test files and separate assertions.

The frozen BtRunner production seam remains contradictory.

### MATERIAL 5 — Oversized atomic cutover

**RESOLVED AS A BOUNDARY, but affected by remaining BLOCKER 1.**

Chunk 15 limits production work to routing, deletion, and expectation change.

It forbids new algorithms, SQL, formulas, index rules, and diagnostics.

The prerequisite chunks do not yet create every claimed publication mechanism.

## BLOCKER 1 — No chunk implements delta memory publication

**Plan sections:** Candidate Generation Contract; Chunks 6, 8, 9, 14, and 15.

Chunk 6 builds candidates but forbids commit, SQL, and publication.

Chunk 8 creates observational indexes rebuilt from current authoritative state. It does not create delta tree/index publication.

Chunk 9 builds deltas but forbids publication.

Chunk 14 persists validated Recon candidates but forbids production activation and does not define memory publication.

Chunk 15 says every publication assignment already exists and forbids new mechanisms.

It then requires Ledger to publish tree, indexes, cursor, raw state, diagnostics, and generation through existing non-failing assignments.

No prior chunk owns those assignments or proves their exact candidate-to-live mapping.

This is a hidden cross-chunk half-state.

An isolated Chunk 15 coder must invent the publication API, candidate lifecycle transition, index replacement, and generation update during a routing-only cutover.

**Required correction:**

Add one off-path publication-preparation chunk before Chunk 15.

It must:

1. Define the exact Ledger-owned operation that applies one already validated candidate to memory.
2. Publish touched domain objects, indexes, cursor, reconciliation time, raw account state, diagnostics, and generation as one non-failing assignment sequence.
3. Define candidate transition from validated to committed or finished.
4. Prove the operation cannot run before successful persistence admission under `max`.
5. Prove `none` and `max` use the same memory publication operation.
6. Prove stale, foreign, discarded, failed, committed, and reused candidates reject before SQL.
7. Prove no fallible calculation, allocation-dependent validation, map rebuild, or index collision check occurs after SQL commit.
8. Keep the operation unreachable from production Recon until the routing chunk.

Chunk 15 may then route only existing gather, build, validate, persist, and publish operations.

## BLOCKER 2 — Observational index rebuild can fail after SQL commit

**Plan sections:** Index Collision Contract lines 137–147; Chunk 8 steps 3–7; current mutation flow in `internal/ledger/ledger.go`.

The plan says normal rebuild collision fails the operation before publication.

Current CreateTrade, AddOrders, RecordSubmit, and Recon persist first under `max`, then publish memory.

If Chunk 8 rebuilds indexes only after existing successful persistence, collision validation can fail after durable SQL commit.

Returning failure then leaves SQL advanced while memory and indexes remain old.

Rebuilding after memory publication is also unsafe because a collision error arrives after visible mutation.

Calling these indexes “observational” does not solve the ordering.

**Required correction:**

Chunk 8 must stage and validate complete candidate indexes before any SQL for every mutation class.

After SQL commit, tree and observational indexes must publish through non-failing assignments together.

Alternatively, Chunk 8 must avoid operation-failing rebuilds entirely and defer all candidate index validation/publication to the later operation-specific store chunks.

The plan must choose one order explicitly.

It must prove:

1. Collision detection occurs before SQL.
2. SQL failure publishes neither tree nor indexes.
3. Successful SQL is followed only by non-failing tree/index assignment.
4. `none` follows the same validated publication order without SQL.
5. Reload builds and validates indexes before publishing loaded state.

## BLOCKER 3 — Chunk 4’s exact write set conflicts with frozen BtRunner ownership

**Plan sections:** Worktree and ResultPublisher Isolation; Chunk 1 acceptance; Chunk 4 exact write set and step 7.

The plan freezes `internal/btrunner/btrunner.go` until the unrelated report rename owner finishes.

Chunk 4 requires a BtRunner test proving Controller error creates no successful telemetry sample and no publication call.

Chunk 4 then permits a package-private seam in `internal/btrunner/btrunner.go` if unavoidable.

That production file is absent from Chunk 4’s exact write set and explicitly frozen.

The plan says the seam is blocked until rename ownership clears, but no chunk clears that dependency.

Chunk 1 acceptance also permits “compile check is green or unrelated blockage is explicit.”

That is not a green implementation handoff. Recording a blockage does not make Chunk 4 executable.

**Required correction:**

Choose one executable path before Chunk 2:

1. Make completion of the unrelated report rename a hard external prerequisite; or
2. Prove the BtRunner assertion using only `internal/btrunner/btrunner_test.go`; or
3. Remove BtRunner proof from Chunk 4 and place it after the frozen owner completes.

If a production seam is authorized, add its exact production file to the chunk write set and remove its frozen status first.

Chunk 1 acceptance must be green, not “green or blocked.” A blocked compile check prevents downstream execution.

## BLOCKER 4 — Performance reduction lacks an exact baseline and pass criterion

**Plan sections:** Chunk 18 acceptance; Chunk 21 objective, steps 6–7, and acceptance.

Chunk 21 promises “measured allocation reduction.”

It says to compare equivalent profile timer boundaries with `wiki/PERFORMANCE.md`.

`wiki/PERFORMANCE.md` records current Grid total allocation and runtime history, but it does not own the retained Ledger-focused profile values.

The reassessment identifies the actual pre-redesign profile:

`workspace/perf/profiles/pptest-s10-b14-20260725T152717Z/`

It also records Ledger and Account reconciliation allocation attribution.

The revised plan names neither that profile path nor its exact baseline metrics.

It defines no numeric or relational criterion for “allocation reduction.”

“Active/touched scaling is demonstrated” proves asymptotic work shape in benchmarks. It does not prove the real Grid profile improved.

An isolated operator cannot know which artifact is comparable or what result passes.

**Required correction:**

1. Name the exact pre-redesign profile directory and artifacts.
2. List the accepted baseline values and units for total allocation, Account.Reconcile cumulative allocation, Ledger.Recon cumulative allocation, GC CPU, loop duration, and process duration.
3. Define each timer boundary and profile sample type.
4. Require equivalent profiling settings and one complete Grid run.
5. Define the release criterion for allocation reduction.
6. Separate hard correctness gates from noisy duration observations.
7. State how missing or unreadable baseline artifacts are handled.
8. Do not claim measured reduction when only benchmark scaling is proven.

## MATERIAL 1 — Candidate lifecycle proof is placed before commit exists

**Plan sections:** Candidate Generation Contract; Chunk 6 steps 5–6; Chunk 14 risks.

Chunk 6 requires committed, reused, and finished candidate rejection before any commit operation exists.

It can test state-machine helpers, but not the real pre-SQL commit boundary.

Chunk 14 says off-path persistence must not mutate candidate ownership, which conflicts with a committed single-use candidate.

**Required correction:** split lifecycle proof by phase.

Chunk 6 proves foreign, stale, discarded, and failed-validation states.

The off-path persistence/publication-preparation chunk proves committed, finished, and reused rejection through the real commit entry point.

Persistence admission must mark or consume the exact candidate once without permitting repeated SQL.

## MATERIAL 2 — Order Fill accumulation order is not currently deterministic

**Plan section:** Chunk 7 step 3; current `internal/order/order.go` `refreshFills`.

Current Fill totals iterate a Go map.

There is no stable current map iteration order to preserve.

Trade executions are explicitly sorted by `(timestamp_ms, venue_tid)`, but Order Fill quantity, notional, and fee accumulation are not sorted.

The plan cannot require exact operation-order preservation without naming that order.

**Required correction:**

1. Characterize whether current decimal outputs vary across repeated fresh processes.
2. Select one explicit deterministic Fill accumulation order if variability exists.
3. Treat that selection as a behavior decision, not mechanical extraction.
4. Prove exact accepted finance through readable fixtures and terminal replay.

Do not claim preservation of an undefined map order.

## MATERIAL 3 — Chunk 9 remains broad across three domain packages

**Plan section:** Chunk 9.

The chunk builds Order lifecycle deltas, Fill enrichment deltas, Trade recalculation, diagnostics, dirty identities, counters, and clone-path equivalence.

It touches Ledger, Order, and Fill production and tests.

This may still fit one coherent candidate algorithm, but it is the largest non-cutover implementation chunk.

**Required correction:** add an explicit stop rule.

If Order or Fill requires a new public or package contract beyond Chunk 7 calculations, split that contract into a prerequisite chunk.

Chunk 9 should compose already proven mechanics, not invent multiple domain mutation protocols.

## MATERIAL 4 — Manifest ownership overstates script exactness

**Plan sections:** Canonical Owners and Resolved Conflicts; Terminal Baseline Manifest.

`rtest.sh` owns exact Observer counts.

`trtest.sh` and `grtest.sh` mainly enforce relational checks. They do not enforce exact cycles, finance, drawdown, or all listed domain counts.

The plan later compensates by requiring exact database queries in Chunk 20.

**Required correction:** state that exact Trade and Grid literals come from canonical docs and retained successful evidence, while scripts own relational and semantic gates.

Do not describe scripts as owners of exact values they do not test.

## NOTE 1 — Public snapshot preservation remains correct

Chunks 3, 16, and 17 preserve complete detached public `order.Snapshot` values with nested Fills.

The narrow Account reconciliation value cannot replace public or cancellation results.

## NOTE 2 — Future Runner scope remains correctly excluded

Heartbeat, live pagination, unresolved cleanup, telemetry persistence, grace, and third-failure stoppage remain explicitly non-executable.

No current chunk requires Runner implementation.

## NOTE 3 — Store chunks are now appropriately isolated

Complete replacement, CreateTrade, AddOrders, RecordSubmit, and Recon dirty persistence have separate boundaries.

Their exact files and proof are clear.

## Holistic execution assessment

The plan is holistic in scope and mostly bite-sized.

Chunks 2–7, 10–14, and 16–20 are individually understandable.

Chunks 8, 9, 15, and 21 still require correction or tighter prerequisites.

The plan is not yet executable without reconstructing candidate publication and index ordering.

## Required corrected order

Keep the revised order with these corrections:

1. Make Chunk 1 a genuinely green prerequisite.
2. Resolve frozen BtRunner proof ownership before Chunk 4.
3. Stage and validate indexes before SQL, not after durable commit.
4. Add off-path non-failing candidate publication preparation before Recon cutover.
5. Move real committed/reused candidate proof to that commit entry point.
6. Keep Chunk 15 routing-only.
7. Define deterministic Order Fill accumulation behavior.
8. Name exact performance baseline artifacts and pass criteria before operator proof.

## Final assessment

The revision fixes round one’s structure and most executability defects.

It still asks Chunk 15 to route a publication mechanism no earlier chunk creates.

It also permits collision failure after durable persistence, carries an unresolved frozen-file dependency, and cannot prove measured reduction against an unnamed baseline.

**Result: FAIL. Do not state PASS until candidate publication, pre-SQL index validation, frozen-file ownership, and exact performance comparison are fully executable.**
