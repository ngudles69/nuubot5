**Reconciliation**

Status: Implementable reorganization and optimization of working Simulator/BtBot Recon.
Covers: Account, Ledger, Simulator evidence, domain updates, persistence, Snapshot publication, outcomes, telemetry, parity, and performance.
Purpose: Convert one Recon's Venue evidence into one validated Snapshot generation while preserving current behavior and finance.

The current Simulator and BtBot behavior is canonical.
Recon changes organization, indexing, selected work, and allocation behavior only.
Recon adds no new execution policy, lifecycle, scheduler, or persistence mode.

**Core invariant**

One successful Recon publishes one validated Snapshot generation.
No consumer decides from partial Venue evidence, partially updated records, or a failed attempt.
Exchange changes observed after a download wait for the next Recon.
Recon is one synchronous Account operation.
Account calls it; Recon owns no loop.
Runner or BtBot owns invocation policy.

**Source layout**

`internal/account/recon.go` Section 1 is the lean program flow.
It shows these exact comments and one clear call for each step:

```text
// Step 1 - Prepare Attempt
// Step 2 - Download Current Order Evidence
// Step 3 - Download Fill History
// Step 4 - Download Current Account State
// Step 5 - Update Fill Records
// Step 6 - Update Order Records
// Step 7 - Update Trade Records
// Step 8 - Update Account Snapshot
// Step 9 - Persist and Publish
// Step 10 - Finalize Recon Outcome and Return
```

Section 2 contains the detailed domain implementation.
Detailed blocks may use ordered comments such as `Step 1.1` and `Step 1.2`.
Section 1 must not contain mechanical loops, mapping, or mutation details.

**Init**

Account.Init calls Ledger.Init.
Init allocates empty dynamic maps and sets.
There is no fixed preallocation, buffer-size requirement, reusable buffer requirement, or object pool.
Persistence mode `none` opens no database.
Persistence mode `max` loads persisted rows when present; otherwise it starts empty.
After loading, Ledger rebuilds all locators, active sets, pending sets, and derived indexes.
Init validates every record, relationship, identity, index, set, and persisted finance value before use.
Full Ledger validation belongs to Init and focused tests, not routine Recon.

Ledger is the sole owner of its domain records and indexes:

```text
TradeID -> Trade
TradeID -> OrderID -> Order
TradeID -> OrderID -> VenueTID -> Fill

OrderID -> {TradeID, OrderID}
CLOID -> {TradeID, OrderID}
nonzero VenueOID -> {TradeID, OrderID}
VenueTID -> {TradeID, OrderID, VenueTID}

active Trade IDs
active Order IDs
pending Order IDs
pending Fill IDs
```

Locator maps contain stable IDs, never duplicate domain objects.
Each Trade, Order, and Fill exists once under Ledger ownership.
Record updates maintain locators and set membership in the same operation.
Maps and sets grow dynamically.
Stable locators make ordinary identity lookup O(1).
Selected-ID sets avoid scanning inactive history.
Routine Recon updates touched records only.
It performs no full Ledger clone or full Ledger traversal.
It does not repeatedly allocate replacement objects for unchanged records.

**Persistence**

`none` keeps current behavior: no Recon database work.
`max` keeps current behavior using existing `account_ledger`, Trade, Order, and Fill storage.
Only changed rows are written.
Nullable Fill fee remains durable fee-presence evidence.
`NULL` means missing; zero and negative decimal values mean present.
No speculative schema is part of this design.
There is no memory rollback path.
Persistence failure blocks Snapshot publication, decisions, and Sweep success.

**Telemetry**

Configuration selects one Recon telemetry record per invocation.
Every record contains `recon_kind=standard|sweep|recovery|startup`.
Every record contains `outcome=skipped|succeeded|failed`.
Each physical Fill pull records its requested range, returned rows, duration, and error.
Fill additions and missing-fee enrichments produce identity-bearing logs.
Telemetry reports observed work; it does not alter domain results.

## Step 1 Prepare Attempt
**What**
Create one attempt outcome with `failed` as the default.
Capture the trusted Snapshot, dirty fact, pending counts, forced request, Fill cursor, and current failure count.
Skip only when all are true:
- Account is clean.
- No Order or Fill work is pending.
- The request is not forced.
- A trusted Snapshot exists.
A valid skip returns that trusted Snapshot without creating a new generation.
Otherwise, select active and pending stable IDs for this attempt.
**Why**
Fail-default handling prevents an early return from being reported as success.
A strict skip gate prevents decisions from using unverified state.
**Things to watch**
Dirty alone does not define required work.
A prior failure or missing trusted Snapshot cannot produce a skip.

## Step 2 Download Current Order Evidence
**What**
Download current open Orders from Simulator.
For each selected active local Order absent from that response, download its exact current status.
Validate all returned identities before record mutation.
**Why**
Open Orders cover current working evidence.
Exact selected status checks resolve local active Orders no longer returned as open.
**Things to watch**
Absence invents nothing.
It does not prove cancellation, rejection, fill, completion, or deletion.
Unexpected or conflicting Simulator evidence fails the attempt.

## Step 3 Download Fill History
**What**
Download new Fills from the inclusive committed Fill cursor through the attempt boundary.
Recheck each pending missing fee during the next Recon using a bounded timestamp window around its known Fill time.
Merge physical pull results and deduplicate them by Venue TID.
Validate identity, Order ownership, quantity, price, timestamp, and fee presence.
**Why**
The inclusive cursor prevents a boundary Fill from being skipped.
The bounded repair window finds delayed fees after the normal cursor has advanced.
Venue TID deduplication admits one execution fact once.
**Things to watch**
A missing fee remains pending.
A zero fee and a negative rebate are present fees.
No Fill is created without a Venue TID.

## Step 4 Download Current Account State
**What**
Download only the current Simulator AccountState through the existing Venue Account-state operation.
Validate resource identity, observation time, position, prices, Account value, withdrawable value, margin values, and PnL inputs.
Keep the validated response attempt-local until publication.
**Why**
One current Account observation supplies Venue-owned finance and position facts for the same Recon attempt.
**Things to watch**
Do not combine Account-state values from different Recon attempts.
Do not publish this evidence before all later updates validate.

## Step 5 Update Fill Records
**What**
Resolve each Fill by Venue TID.
Add a new Fill once, or enrich an existing Fill when previously missing fee evidence becomes present.
Preserve immutable execution identity, quantity, price, side, and timestamp.
Mark a missing fee pending.
Touch the owning Order and Trade IDs.
Update Fill locators and pending membership with the record.
**Why**
One owned Fill object preserves exact execution evidence without duplicate allocation.
Touch propagation limits later calculations to affected parents.
**Things to watch**
Never overwrite a present fee with a conflicting value.
Zero and negative fees remain valid present values.
An unchanged duplicate performs no domain update.

## Step 6 Update Order Records
**What**
Update touched Orders from exact Order evidence and their owned Fills.
Recalculate status, filled quantity, remaining quantity, average fill price, and fees.
Update active and pending membership and all Order locators with each record.
**Why**
Order values must agree with admitted Fill evidence and current Simulator status.
**Things to watch**
Do not mark local completion until Fill identity, total quantity, and every Fill fee are complete.
A Venue-filled acknowledgement can remain locally pending.
Absence from open Orders never supplies a status.

## Step 7 Update Trade Records
**What**
Update touched Trades only.
Recalculate status, quantities, entry and exit prices, realized PnL, unrealized PnL, gross PnL, net PnL, fees, and timestamps.
Update active Trade membership and cached Ledger totals from each old-to-new record delta.
**Why**
Touched-only updates remove complete-history work while preserving exact Trade and Ledger finance.
Cached deltas avoid repeated Ledger traversal.
**Things to watch**
Late fee enrichment must update its closed Trade's fees and dependent PnL once.
Trade identity and established execution facts remain immutable.

## Step 8 Update Account Snapshot
**What**
After all touched Trades validate, calculate Account-level state:
- Trade, Order, Fill, active, and pending counts.
- Position and entry price.
- Account value, balance, equity, withdrawable value, and margin values.
- Realized, unrealized, gross, and net PnL.
- Fees.
- Account peak, current drawdown, and maximum drawdown.
Build one immutable Snapshot with the next generation.
Risk, Controller, and BotCycle consume only that Snapshot.
**Why**
Account finance depends on final touched Trade values and the current AccountState.
One immutable generation gives every consumer the same validated truth.
**Things to watch**
Account drawdown and Controller Bot drawdown are separate values with separate owners.
Preserve existing finance equations and decimal behavior exactly.

## Step 9 Persist and Publish
**What**
For `none`, perform no database work.
For `max`, write only changed `account_ledger`, Trade, Order, and Fill rows through existing storage.
Persist nullable fee presence with the existing Fill representation.
After persistence succeeds, apply validated touched updates to Ledger records, locators, sets, cached totals, cursor, and Account state.
Publish the immutable Snapshot and generation last.
**Why**
Dirty-row persistence removes full rewrites.
Publishing last prevents decisions from observing partial truth.
**Things to watch**
There is no full clone and no memory rollback.
Any persistence or publication error fails Recon and blocks decisions.
A failed Recon makes Sweep exit unsuccessful.

## Step 10 Finalize Recon Outcome and Return
**What**
Account stores latest telemetry and owns failure count, dirty state, and last trusted Snapshot.
On success:
- store the new trusted Snapshot;
- clear dirty state;
- reset the failure count;
- return `succeeded` with the Snapshot.
On valid skip:
- retain the trusted Snapshot;
- leave the failure count unchanged;
- return `skipped` with that Snapshot.
On failure:
- retain dirty state;
- increment the failure count exactly once;
- store the error;
- return `failed` without a decision Snapshot.
**Why**
One terminal path gives each invocation one exact outcome.
Runner or BtBot owns response policy.
Recon owns no repeated invocation.
**Things to watch**
Controller and BotCycle remain unchanged except that failed Recon blocks decisions.
A failed attempt never falls back to stale truth for decisions.

**Exact implementation impact**

- `internal/account/account.go`: trusted Snapshot, dirty state, failure count, and latest Recon telemetry.
- `internal/account/recon.go`: exact ten-step flow, selected work, publication, and outcome.
- `internal/ledger/ledger.go`: sole-owned records, stable locators, active and pending sets, and cached totals.
- `internal/ledger/recon.go`: touched Fill, Order, and Trade updates with index maintenance.
- `internal/ledger/store.go`: existing `none` behavior, `max` load, validation, and changed-row writes.
- `internal/simulator/simulator.go`: current Order, Fill, and AccountState evidence plus deterministic delayed-fee proof.
- Focused tests beside these owners prove indexes, pending work, delayed fees, persistence, failures, parity, and performance.
No Controller, BotCycle, execution-policy, or unrelated persistence redesign is required.

**Cutover**

Recon2 remains the current control.
First record exact current Recon and Recon2 results under identical Grid inputs.
Then change canonical Recon only.
Compare exact Trades, Orders, Fills, statuses, quantities, timestamps, finance, Account values, drawdown, and terminal result.
Profile canonical Recon after parity passes.
Recon2 deletion requires later approval.

**Proof**

Focused proof covers locator correctness, dynamic growth, active and pending membership, selected-only updates, and Init rebuild validation.
Delayed-fee proof advances the inclusive cursor, then enriches the same Venue TID exactly once through its bounded repair window.
Persistence proof covers `none` with no database and `max` load plus changed-row writes.
Failure proof covers each download, validation, and persistence boundary with one increment and no decision Snapshot.
Grid proof requires exact domain and finance parity against the recorded control.
Performance proof reports Recon runtime, allocated bytes, allocation count, selected records, changed rows, and total Ledger size.
Proof must show no routine full clone, full traversal, or repeated unchanged-object allocation.
All Go proof and profiling uses `-tags noasm`.
