# Ledger Reconciliation Reassessment Review 2

Date: 2026-07-26

Scope: re-review of `.audits/07-26-ledger-reconciliation-reassessment.md` against review 1, five canonical design pages, and current source/tests. No source, wiki, or `HANDOFF.md` changed. No replay ran.

## Verdict

**FAIL — three BLOCKERS remain.**

The revision resolves all four round-one blocker subjects and all seven material subjects at the design level.

The current executable tranche is still not safe as ordered.

Change #3 requires dirty SQL before Change #7 creates dirty store operations.

Changes #6 through #8 conflict over when reconciliation stops using `cloneTrades`.

Change #8 removes nested Fills from active `order.Snapshot` values, changing current outward behavior without proof or authority.

## Prior BLOCKER status

### Round-one BLOCKER 1 — Exact successful-result baseline

**RESOLVED.**

Change #1 now requires pre-edit deterministic characterization fixtures.

It covers CreateTrade, AddOrders, RecordSubmit, Recon, `max` reload, Account Snapshot, terminal Result, ResultPublisher round trip, and alias independence.

It defines stable ordering, exact decimal text, optional presence, raw values, timestamps, and canonical bytes.

It correctly limits claims to covered fixtures and forbids full-run preservation claims without replay-equivalent evidence.

### Round-one BLOCKER 2 — Account/Ledger publication protocol

**RESOLVED AS A PROTOCOL, but affected by remaining BLOCKER 1 below.**

Change #3 now defines Account coordination, opaque Ledger candidate ownership, immutable aggregates, pre-SQL Account validation, single-use commit, stale-candidate rejection, and non-failing publication.

The ownership and publication order match canonical Account, Ledger, and reconciliation contracts.

### Round-one BLOCKER 3 — Dirty and complete persistence mixing

**RESOLVED AS A STORE DESIGN, but affected by remaining BLOCKER 1 below.**

Change #7 now separates complete-store replacement from explicit incremental operations.

It defines dirty sets for CreateTrade, AddOrders, RecordSubmit, and Recon.

It preserves complete terminal publication, untouched rows, stale-row removal ownership, foreign-key ordering, reload proof, and real transaction failures.

### Round-one BLOCKER 4 — Future telemetry writer

**RESOLVED.**

Future telemetry is no longer an executable recommendation.

The blocked dependency graph names writer ownership, synchronization choice, capacity, sequence, generation ordering, backpressure, failures, shutdown, transaction, and durability.

It correctly rejects simultaneous promises of nonblocking decisions and guaranteed durable heartbeat rows before design resolution.

## Prior MATERIAL status

### Round-one MATERIAL 1 — Exact financial algorithm reuse

**RESOLVED.**

Change #4 requires one shared calculation path.

It preserves `(timestamp_ms, venue_tid)` ordering, decimal operation order, fees, average entry, realized PnL, status, timestamps, terminal checks, and Fill-total order.

### Round-one MATERIAL 2 — Premature index rollout

**RESOLVED.**

Change #5 separates authoritative rebuild proof from later publication updates and final operational read switching.

It requires stable domain identity, not pointer identity.

### Round-one MATERIAL 3 — Duplicate identity policy

**RESOLVED.**

Change #5 defines scope and zero/enrichment behavior for Trade ID, Order ID, CLOID, Venue Order ID, Venue TID, and active state.

Collision failures occur before candidate publication.

### Round-one MATERIAL 4 — Unknown Venue activity

**RESOLVED.**

Change #6 gives Ledger one bounded operation diagnostic.

Unknown CLOIDs remain non-fatal, are never adopted, and cannot change Results, cursor, Account Snapshot, or current telemetry.

### Round-one MATERIAL 5 — Hyperliquid pagination

**RESOLVED.**

Live pagination and cleanup moved behind an exact Hyperliquid adapter prerequisite.

The prerequisite includes ordering, boundaries, caps, no-progress detection, same-timestamp overflow, retention truncation, endpoint separation, and completeness classification.

It forbids cursor advancement when completeness is unproven.

### Round-one MATERIAL 6 — Runner scope mixing

**RESOLVED.**

Future Runner work is now an explicitly non-executable dependency graph.

The current tranche contains only BtRunner-owned work.

### Round-one MATERIAL 7 — SQL failure injection

**RESOLVED.**

Change #2 defines package-local disabled hooks surrounding real transaction operations.

SQLite triggers prove statement behavior. Real rollback and reload remain exercised.

Generic repository and fake-store abstractions are prohibited.

## Remaining BLOCKER 1 — Change #3 depends on Change #7

**Report sections:** Recommendation / Change #3, lines 130–169; Recommendation / Change #7, lines 341–396; Final assessment.

Change #3 requires Ledger to write dirty SQL before publishing the candidate.

Change #7 later introduces the explicit incremental dirty-store operations required to perform that write.

Current source has only `ledgerStore.save`, which deletes and rewrites the complete Ledger tree.

Therefore Change #3 cannot satisfy its own successful-commit and call-order proof when implemented in the stated order.

Using current `save` would violate Change #3’s dirty-SQL requirement.

Adding provisional dirty SQL in Change #3 would duplicate Change #7 and recreate the store-boundary ambiguity review 1 rejected.

**Required correction:**

Choose one executable order:

1. Move store-operation separation before successful candidate commit; or
2. Split Change #3 into protocol types/build/validation first and commit activation after Change #7.

The preferred order is:

1. characterize;
2. prove failures;
3. define candidate construction and Account validation without activating commit;
4. extract exact arithmetic;
5. stage indexes;
6. build exact deltas;
7. split complete and incremental stores;
8. activate candidate commit and atomic Ledger/Account publication;
9. switch reads and remove broad work.

No step may claim successful candidate commit before its exact persistence operation exists.

## Remaining BLOCKER 2 — `cloneTrades` removal order is contradictory

**Report sections:** Recommendation / Change #6, especially lines 292–303 and 331–335; Recommendation / Change #7, lines 365–391; Recommendation / Change #8, lines 410–455.

Change #6 requires candidate work to visit only active Orders, incoming Fills, and touched Trades.

The same change says to keep `cloneTrades` in reconciliation until Change #8.

Change #7 then publishes exact candidate deltas and claims the split-generation defect fixed.

Change #8 finally removes reconciliation `cloneTrades`.

These statements cannot all describe one active path.

If reconciliation still calls `cloneTrades`, Change #6’s visited-work proof is false.

If Change #7 commits and publishes deltas while the clone path remains active, two reconciliation representations coexist.

That creates duplicate validation, uncertain authority, and accidental behavior divergence.

**Required correction:**

Define one cutover point.

Recommended:

1. Build and test delta candidates without production routing.
2. Complete dirty-store persistence.
3. Atomically switch `Ledger.Recon` to the candidate path.
4. Remove reconciliation `cloneTrades` in the same cutover change.
5. Keep `cloneTrades` only for unrelated mutation paths until separately removed, if still needed.
6. Run canonical fixture and failure proof immediately after cutover.

Do not leave old and new reconciliation mutation paths active together.

## Remaining BLOCKER 3 — Active Order snapshots lose Fill results

**Report section:** Recommendation / Change #8, lines 410–416 and Completion proof.

Change #8 says active Orders crossing ownership boundaries must not copy nested historical Fills.

Current `Ledger.ActiveOrders` returns complete `order.Snapshot` values from `Trade.Orders()`.

Each snapshot includes its `Fills` slice.

Partially filled active Orders can therefore carry accepted Fill evidence today.

Removing nested Fills changes the outward `order.Snapshot` result even when trading behavior and terminal Results remain unchanged.

This conflicts with the tranche promise to preserve exact successful Order and Fill behavior.

Canonical Account design requires immutable active-Order snapshots. It does not authorize a reduced snapshot shape.

**Required correction:**

Preserve complete `order.Snapshot` values for public `Account.ActiveOrders` and every existing outward caller.

Use a separate narrow internal identity/status value for Account reconciliation and cancellation validation.

The internal value may omit Fills because it does not cross as the existing public result contract.

Add a characterized partially-filled active Order proving public `ActiveOrders` retains exact nested Fill evidence.

## Ordered sequence assessment

The revised sequence has the right components but remains non-executable at Changes #3, #6, and #7.

Safe corrected sequence:

1. Capture exact successful fixtures.
2. Add failure-before-publication and real SQL failure proof.
3. Define opaque candidate construction and immutable Account aggregates.
4. Extract one exact shared Trade calculation path.
5. Stage authoritative indexes without switching reads.
6. Build and test exact delta candidates off the production path.
7. Split complete and incremental store operations.
8. Cut reconciliation to delta commit and remove reconciliation `cloneTrades` together.
9. Publish Ledger and Account through non-failing assignments.
10. Switch internal identity and active-state reads to indexes.
11. Preserve complete public active `order.Snapshot` values.
12. Run focused tests, full tests, vet, and the performance matrix with `CGO_ENABLED=0` and `-tags noasm`.
13. Make no full-run financial claim without separately authorized replay-equivalent proof.

## Final assessment

The revision fully addresses review 1’s subjects.

Three integration defects remain in the revised executable order and behavior boundary.

**Result: FAIL. Do not state PASS until dirty-store availability precedes commit, reconciliation has one cutover path, and public active Order Fill evidence remains exact.**
