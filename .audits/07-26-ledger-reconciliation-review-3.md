# Ledger Reconciliation Reassessment Review 3

Date: 2026-07-26

Scope: final permitted plan review of the revised reassessment and reviews 1–2. No source, wiki, or `HANDOFF.md` changed. No replay ran.

## Verdict

**FAIL — one BLOCKER remains.**

All substantive round-one and round-two blockers are resolved.

The plan now has exact finance fixtures, deferred candidate activation, dirty-store-first ordering, one reconciliation cutover, complete public snapshots, and separated future Runner scope.

Change #5 still contains stale cross-references that contradict the final sequence.

## Round-two blocker verification

### Blocker 1 — Dirty-store availability before candidate commit

**RESOLVED.**

Change #3 now defines candidate construction and Account validation without commit activation.

Production `Ledger.Recon` remains on the current path.

Change #7 creates and proves complete and incremental store operations.

Change #8 activates candidate commit only after dirty persistence exists.

No provisional commit or duplicate store design is permitted.

### Blocker 2 — One reconciliation cutover

**RESOLVED.**

Change #6 builds and tests delta candidates off the production path.

The existing clone path remains the only production authority through Change #7.

Change #8 performs one cutover, activates dirty persistence, removes reconciliation `cloneTrades`, and removes terminal `Ledger.Result` aggregation together.

The plan explicitly forbids old and new reconciliation mutation paths running together.

### Blocker 3 — Public active Order snapshot preservation

**RESOLVED.**

Change #1 characterizes a partially filled active Order with complete nested Fill evidence.

Change #9 introduces a separate narrow internal value for reconciliation and ownership checks.

Public `Account.ActiveOrders()` and `Ledger.ActiveOrders()` retain complete immutable `order.Snapshot` values and nested Fills.

The narrow value cannot masquerade as the public snapshot type.

## Exact behavior preservation

**RESOLVED.**

Change #1 captures pre-edit canonical bytes or complete exact typed expectations.

Coverage includes:

- CreateTrade;
- AddOrders;
- RecordSubmit;
- Recon;
- `max` reload;
- Account Snapshot;
- terminal Account and Ledger Results;
- ResultPublisher round trip;
- returned-value alias independence; and
- partially filled public active Orders with nested Fills.

Canonical encoding includes stable Trade/Order/Fill ordering, exact decimal text, optional presence, cursor, raw values, timestamps, and every exported nested field.

Change #4 reuses one exact Trade calculation path.

It preserves `(timestamp_ms, venue_tid)` ordering, decimal operation order, fee treatment, average entry, realized PnL, lifecycle state, and timestamps.

Each later production change must preserve the captured canonical bytes.

The report correctly makes no full-run financial claim without separately authorized replay-equivalent evidence.

## Publication and persistence safety

**RESOLVED.**

Account remains reconciliation coordinator.

Ledger owns candidate construction, persistence, and Ledger publication.

Account validates its candidate Snapshot before SQL.

Change #7 separates complete replacement from explicit incremental operations.

Untouched rows survive incremental writes. Only complete publication removes stale rows.

Change #8 commits the exact single-use candidate, then performs only non-failing Ledger and Account assignments.

Failure tests cover candidate rejection, real SQLite statements, rollback, commit failure, reload, and first-error propagation.

## Current and future scope

**RESOLVED.**

Changes #1 through #9 contain only current BtRunner work.

BtRunner and Sweep retain deterministic first-error failure.

No retries, grace counters, heartbeat, cleanup, live telemetry writing, or third-failure policy enters current code.

Future Runner work remains a non-executable dependency graph.

It remains blocked on Runner ownership, Hyperliquid completeness, unresolved-Order safety, telemetry writing, and live error classification.

## Remaining BLOCKER — Change #5 names nonexistent activation steps

**Report section:** Recommendation / Change #5, especially lines 237–241 and Completion proof lines 271–275.

Change #5 says:

- phase-two indexes update through Change #3’s successful non-failing publication; and
- operational reads switch in Change #8.

The final plan now says:

- Change #3 performs no commit or publication;
- Change #8 activates candidate publication and index publication; and
- Change #9 switches operational reads.

Therefore Change #5 points to two wrong owners.

This is not cosmetic. Index mutation and read cutover are correctness-sensitive boundaries.

Following Change #5 literally would require nonexistent Change #3 publication and premature Change #8 read switching.

That conflicts with the approved one-cutover sequence and public snapshot preservation work in Change #9.

**Required correction:**

Replace Change #5’s stale references with:

1. Phase one rebuilds indexes from authoritative state while current reads remain unchanged.
2. Phase two updates tree and indexes together only during Change #8’s successful non-failing candidate publication.
3. Operational internal reads switch only in Change #9.
4. Change #5 completion proof defers phase-two publication proof to Change #8.
5. Change #5 completion proof defers read-equivalence proof to Change #9.

Also change Change #7’s purpose wording from “complete atomic publication” to store separation and publication preparation. Its body correctly defers activation to Change #8.

## Final dependency order

After the required cross-reference correction, the executable order is sound:

1. Capture exact successful fixtures.
2. Prove failure-before-publication and real SQL failures.
3. Define opaque candidate construction without commit.
4. Reuse exact Trade arithmetic.
5. Stage authoritative indexes without switching reads.
6. Build exact deltas off-path.
7. Split and prove complete and dirty stores.
8. Perform one candidate cutover and non-failing publication.
9. Switch narrow internal reads while preserving public snapshots.

## Final assessment

The revised design is substantively safe and narrowly scoped.

One correctness-sensitive dependency reference remains internally contradictory.

**Result: FAIL. Correct Change #5’s publication and read-switch owners before implementation.**
