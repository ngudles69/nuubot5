**Reconciliation**

Status: Implemented canonical Recon. Further measured optimization remains pending.
Covers: Account, Ledger, Simulator evidence, flat domain updates, persistence, Account Snapshot publication, outcomes, telemetry, parity, and performance.
Purpose: Convert one Recon's Venue evidence into one validated Account Snapshot generation while preserving current behavior and finance.

The current Venue, Ledger, and BtBot results are canonical.
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

Recon-owning Go files use Section 1 for the complete lean program flow.
Section 1 presents Init before Recon.
Init shows memory allocation, persisted-state loading, and Ledger validation.
Recon shows start conditions, skip conditions, these exact comments, and one clear call for each step:

```text
// Step 1 - Prepare Attempt
// Step 2 - Download Current Order Evidence
// Step 3 - Download Fill History
// Step 4 - Download Current Account State
// Step 5 - Update Fill Records
// Step 6 - Update Order Records
// Step 7 - Search Fills by Updated Order OIDs
// Step 8 - Update Trade Records
// Step 9 - Update Account Snapshot
// Step 10 - Persist and Publish
// Step 11 - Finalize Recon Outcome and Return
```

Section 2 contains the detailed domain implementation.
Detailed blocks may use ordered comments such as `Step 1.1` and `Step 1.2`.
Section 1 contains only flow decisions, step comments, clear calls, and the minimum flow-driving values needed for clarity.
It must not contain mechanical loops, mapping, mutation, or incidental values.

**Init**

Each Account-owning Executor creates one Account and Ledger for one BotCycle.
The next BotCycle creates new instances; it does not reuse the completed Ledger.
Account.Init calls Ledger.Init.
Init allocates empty dynamic maps and sets.
There is no fixed preallocation, buffer-size requirement, reusable buffer requirement, or object pool.
Persistence mode `none` opens no database.
Persistence mode `max` loads persisted rows when present; otherwise it starts empty.
After loading, Ledger rebuilds all locators, active sets, pending sets, derived indexes, and its cached Summary once.
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
Capture the trusted Snapshot, dirty fact, pending counts, forced request, Fill cursor, current failure count, and elapsed time since successful Recon.
Treat pending Order or Fill work as dirty for cadence selection.
Execute immediately when no successful Recon exists or the request is forced.
Before the Recon sweep interval, skip a clean Account.
Before the normal Recon interval, skip a dirty Account.
At the Recon sweep interval, execute the normal Recon pipeline regardless of dirty state.
A valid skip returns the trusted Snapshot without creating a new generation or advancing `lastReconMS`.
Otherwise, select active and pending stable IDs for this attempt.
**Why**
Fail-default handling prevents an early return from being reported as success.
A strict skip gate prevents decisions from using unverified state.
**Things to watch**
Dirty selects the shorter Recon cadence; it does not bypass that cadence.
Only a successfully published Recon advances `lastReconMS`.
A prior failure or missing trusted Snapshot cannot produce a skip.

## Step 2 Download Current Order Evidence
**What**
Call Venue for fresh detached Open Orders and Order History JSON.
Decode and validate both through `internal/hyperliquid`.
Use Order History for selected active local Orders absent from Open Orders.
For each selected active local Order absent from both responses, download its exact current status.
Treat each exact status download as exception handling and count every attempted request in Recon telemetry.
Validate all returned identities before record mutation.
**Why**
Open Orders cover current working evidence.
Order History covers recent terminal evidence.
Exact selected status checks resolve active local Orders absent from both bulk responses.
**Things to watch**
Absence invents nothing.
It does not prove cancellation, rejection, fill, completion, or deletion.
Unexpected or conflicting Venue evidence fails the attempt.
The exception count proves whether this gap occurs before any removal or optimization is considered.
Resolve official evidence by CLOID when present, then OID.
If both exist, they must identify the same Ledger Order.

## Step 3 Download Fill History
**What**
Call Venue for fresh detached Fill-history JSON from the inclusive committed cursor through the attempt boundary.
Recheck each pending missing fee during the next Recon using a bounded timestamp window around its known Fill time.
Merge physical pull results and deduplicate them by Venue TID.
Validate identity, Order ownership, quantity, price, timestamp, and fee presence.
**Why**
The inclusive cursor prevents a boundary Fill from being skipped.
The bounded repair window finds delayed fees after the normal cursor has advanced.
Venue TID deduplication admits one execution fact once.
Official Fill rows may omit CLOID, so OID resolves their owning Ledger Order.
**Things to watch**
A missing fee remains pending.
A zero fee and a negative rebate are present fees.
No Fill is created without a Venue TID.

## Step 4 Download Current Account State
**What**
Download one fresh detached official clearinghouse JSON response through Venue.
Decode and validate it through `internal/hyperliquid`.
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
Detect accepted semantic changes through the Order-owned transient mutation revision.
Update active and pending membership and all Order locators with each record.
**Why**
Order values must agree with admitted Fill evidence and current Simulator status.
**Things to watch**
Do not mark local completion until Fill identity, total quantity, and every Fill fee are complete.
A Venue-filled acknowledgement can remain locally pending.
Absence from open Orders never supplies a status.

## Step 7 Search Fills by Updated Order OIDs
**What**
Keep OID-only Fills unresolved during Step 5 when their Venue OID is not yet
indexed.
After Step 6, search those Fills against the updated Order OID index.
Always record matched Order and Fill counts in Recon telemetry.
Always emit one searchable log:

```text
Recon-OIDSearch found nothing
Recon-OIDSearch found orders=2 fills=3
```

**Why**
Order evidence may add a Venue OID after the first Fill pass.
The second search proves whether sequencing deferred valid Fill ownership.

**Things to watch**
The search never copies CLOID from an Order into a Fill.
Apply matched Fills immediately using the same Step 5 function.
Reapply the same Step 6 function to distinct owning Orders only.
Refresh only Trades touched by those Fill and Order updates.
Multiple matched Fills for one Order reapply that Order once.
After preserving all matched evidence, fail Recon because sequencing was abnormal.
Any still-unmatched OID-only Fill also fails Recon.
Never depend on a later Fill download to recover current evidence.

## Step 8 Update Trade Records
**What**
Structurally update touched Trades only.
After Fill and Order updates, recalculate their exposure, status, realized PnL, fees, quantities, prices, and timestamps from owned evidence.
Update every active Trade's current-mark finance from its stored exposure without rereading Orders or Fills.
Store realized PnL, unrealized PnL, gross PnL, net PnL, fees, quantities, prices, status, and timestamps.
When every execution and fee is complete, set unrealized PnL to zero, calculate final realized PnL and fees once, then mark the Trade closed.
Closed Trades retain static stored totals and are never recalculated.
Update active Trade membership after each changed Trade.
Apply each changed Trade's exact old-to-new summary delta to the Ledger-owned cached Summary.
**Why**
Touched-only in-memory Trade updates remove repeated historical recalculation while preserving exact finance.
**Things to watch**
A fee-incomplete Trade remains pending until a later Recon finalizes it once.
Trade identity and established execution facts remain immutable.

## Step 9 Update Account Snapshot
**What**
After all touched Trades validate, read the Ledger-owned cached Summary without traversing Trades.
The cache contains exact stored Trade finance and evidence totals; it does not recalculate Trade PnL.
Trades remain authoritative, and Init or persisted reload rebuilds the derived cache once.
Calculate Account-level state:
- Trade, Order, Fill, active, and pending counts.
- Position and entry price.
- Account value, balance, equity, withdrawable value, and margin values.
- Realized, unrealized, gross, and net PnL.
- Fees.
- Account peak, current drawdown, and maximum drawdown.
Build one immutable Snapshot with the next generation.
Risk, Controller, and BotCycle consume only that Snapshot.
**Why**
Account finance depends on final touched Trade values, their exact cached deltas, and the current AccountState.
One immutable generation gives every consumer the same validated truth.
**Things to watch**
Account drawdown and Controller Bot drawdown are separate values with separate owners.
Preserve existing finance equations and decimal behavior exactly.

## Step 10 Persist and Publish
**What**
For `none`, perform no Ledger, Trade, Order, Fill, or Simulator database work.
For `max`, write only changed `account_ledger`, Trade, Order, and Fill rows through existing storage.
Cached Summary is derived and not persisted; the schema remains unchanged.
Persist nullable fee presence with the existing Fill representation.
Refresh Ledger locators and sets after direct domain updates.
Publish the compact Account Snapshot and generation last.
**Why**
Dirty-row persistence removes full rewrites.
Publishing last prevents decisions from observing partial truth.
**Things to watch**
There is no full clone and no memory rollback.
Any persistence or publication error fails Recon and blocks decisions.
A failed Recon makes Sweep exit unsuccessful.
Persistence failure may leave directly mutated records and cached totals equally untrusted.

## Step 11 Finalize Recon Outcome and Return
**What**
Account stores latest telemetry and owns failure count, dirty state, and last trusted Snapshot.
On success:
- store the new trusted Snapshot;
- clear dirty state;
- reset the failure count;
- return `succeeded` with the Snapshot and zero consecutive failures.
On valid skip:
- retain the trusted Snapshot;
- leave the failure count unchanged;
- return `skipped` with that Snapshot and the unchanged count.
On failure:
- retain dirty state;
- increment the failure count exactly once;
- store the error;
- return `failed` without a decision Snapshot and with the consecutive count.
**Why**
One terminal path gives each invocation one exact outcome.
Runner or BtBot owns response policy.
Recon owns no repeated invocation.
**Things to watch**
BotCycle reconciles every capable running Executor before completing one barrier.
Any failure suppresses the complete Snapshot barrier and reports the maximum consecutive count.
Controller skips Risk, Executor `OnRecon`, and BotCycle `Run` after failures one and two.
Controller returns an error when any consecutive count reaches three.
A failed attempt never falls back to stale truth for decisions.
Persistence or execution failures outside Account Recon remain immediately fatal.

**Exact implementation impact**

- `internal/account/account.go`: trusted Snapshot, dirty state, failure count, and latest Recon telemetry.
- `internal/account/recon.go`: exact ten-step flow, selected work, publication, outcome, and explicit failure count.
- `internal/botcycle/botcycle.go`: complete multi-Executor barrier, failure fact, and maximum consecutive count.
- `internal/controller/controller.go`: first-two-pass skip and third-consecutive-failure Sweep error.
- `internal/account/ledger/ledger.go`: sole-owned records, stable locators, and active and pending sets.
- `internal/account/ledger/recon.go`: touched Fill, Order, OID-search, and Trade updates with index maintenance.
- `internal/account/ledger/store.go`: existing `none` behavior, `max` load, validation, and changed-row writes.
- `internal/simulator/simulator.go`: canonical private Venue truth and fresh official JSON responses.
- `internal/hyperliquid/protocol.go`: strict mutation and information response decoding.
- Focused tests beside these owners prove indexes, pending work, delayed fees, persistence, failures, parity, and performance.
No execution-policy, cadence, or unrelated persistence redesign is required.

**Cutover**

Canonical Recon is the only implementation.
Recon2 source, configuration, and Bot 16 are retired.
Change and profile canonical Recon only.
Compare exact Trades, Orders, Fills, statuses, quantities, timestamps, finance, Account values, drawdown, and terminal result against the accepted Bot 15 baseline.

**Proof**

Focused proof covers locator correctness, dynamic growth, active and pending membership, selected-only updates, and Init rebuild validation.
Delayed-fee proof advances the inclusive cursor, then enriches the same Venue TID exactly once through its bounded repair window.
OID-search proof records zero matches normally and fails before publication when updated Order OIDs reveal deferred Fills.
Persistence proof covers `none` with no database and `max` load plus changed-row writes.
Failure proof covers each download, validation, and persistence boundary with one increment and no decision Snapshot.
Grid proof requires exact domain and finance parity against the recorded control.
Performance proof reports Recon runtime, allocated bytes, allocation count, selected records, changed rows, and total Ledger size.
Proof must show no routine full clone, full traversal, or repeated unchanged-object allocation.
A test-only complete traversal oracle must equal cached Summary after mutations, Recon, failed validation, and maximum-persistence reload.
Focused proof should show `Summary` and `ReconSummary` reads allocate nothing.
All Go proof and profiling uses `-tags noasm`.
