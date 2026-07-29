# Account Stack Coherency Adversarial Audit V2

Date: 2026-07-29
Round: 1
Reviewer status: FAIL

## Objective

Audit the flat Account stack after every finding in the first coherency audit
was marked fixed.

External callers, tests, Recon, Store, Venue implementations, executors, and
result publication were excluded.

## Independent Reviewer Report

### 1. BLOCKER — Fill identity conflicts can be hidden

Reachability:

`AddFill` receives both CLOID and Venue OID.

Evidence:

`ledger.go:537-549` accepts one recognized identity when the other is unknown.

`ledger.go:284` replaces incoming CLOID with the parent CLOID.

Impact:

A conflicting Exchange identity can be misattributed and normalized away.

Invariant:

Every supplied identity must resolve to the same Order.

Smallest fix:

Validate each supplied identity independently. Require equality when both
exist.

### 2. BLOCKER — Terminal Trades can mutate before rejection

Reachability:

Account submits later Orders using an already terminal TradeID.

A late Fill targets a canceled Order.

Evidence:

`ledger.go:160-178` stores Orders before `refreshTrade`.

`ledger.go:289-304` stores Fill and mutates Order before `refreshTrade`.

`trade.go:104-106` rejects changed terminal Trade values afterward.

Impact:

The operation returns an error but leaves Ledger maps, indexes, and totals
changed.

Invariant:

Validation must precede mutation. Terminal Trades cannot gain records.

Smallest fix:

Reject terminal parent Trades and Orders before storing anything.

### 3. MAJOR — New-Trade Order quantities are forced symmetrical

Reachability:

A new Trade submits DCA entries, partial exits, or unequal protection
quantities.

Evidence:

`account.go:600-623` finds one maximum quantity and assigns it to every Order.

Impact:

Submitted business intent changes. DCA and staged-exit quantities are
destroyed.

Invariant:

A Trade allows arbitrary Order counts, roles, and quantities without
symmetry.

Smallest fix:

Apply rounding and minimum notional independently to each Order.

### 4. MAJOR — All-canceled zero-size Trades are not Closed

Reachability:

Every Order cancels before any Fill.

Evidence:

`trade.go:258-265` assigns `Canceled`.

`ledger.go:618-623` counts only exact `Closed`.

Impact:

The Trade is terminal but excluded from completed Trade counts.

Invariant:

Zero size plus every Order closed means Trade Closed.

Smallest fix:

Assign `Closed` when no exposure exists and every Order is closed.

### 5. MAJOR — Closure ownership remains split

Reachability:

A canceled, rejected, expired, or error Order has a Fill awaiting fee evidence.

Evidence:

`order.go:164-174` reports those Orders closed regardless `PendingFeeCount`.

`trade.go:198-210` and `trade.go:262` separately block closure using Fill fee
state.

Impact:

All Orders report closed, size is zero, yet Trade remains Closing.

`ActiveOrders` can report zero while closure still waits.

Invariant:

`Order.IsClosed` owns detailed Order completion. Trade only consumes that
result.

Smallest fix:

Make `IsClosed` include pending-fee completion for every terminal status.

Remove Trade's separate pending-Fill closure rule.

### 6. MAJOR — Existing Fill conflicts are ignored

Reachability:

The same VenueTID arrives with changed parent, quantity, price, side, or
timestamp.

Evidence:

`ledger.go:251-268` routes directly to `Fill.Update`.

`fill.go:57-75` checks only fee conflict and raw JSON.

Impact:

Conflicting Exchange execution evidence is silently accepted or fee-enriched.

Invariant:

VenueTID identifies one immutable execution.

Smallest fix:

Validate immutable incoming Fill evidence before applying fee or raw updates.

### 7. MAJOR — Synthetic identity leftovers remain

Evidence:

`trade.go:41` and `trade.go:147-150` retain `TradeNumber` and its old 21-bit
limit.

`ledger.go:93`, `ledger.go:126`, `ledger.go:150`, and `ledger.go:372` allocate
that number.

`fill.go:16` retains unused `ID` while VenueTID is the canonical Fill key.

Impact:

The hardcut still exposes redundant identities and an artificial Trade limit.

Invariant:

TradeID, OrderID, and VenueTID are canonical keys.

Smallest fix:

Remove `TradeNumber`, `nextTradeNo`, the 21-bit cap, and `Fill.ID`.

## Proof Checked

- Read every scoped source and design file.
- Traced Order planning, creation, updates, Fill admission, Trade calculation,
  and CLOID encoding.
- Confirmed flat Ledger ownership and CLOID layout.
- Confirmed OID conflict prevalidation in `UpdateOrders`.
- Confirmed no Batch identity remains.
- Full `git diff --check` passes.

## Proof Missing

- No compilation.
- No external caller, Recon, Store, Venue, executor, or publisher review.

## Bloat and Duplication Verdict

FAIL.

`TradeNumber` and `Fill.ID` are synthetic identity leftovers.

No compatibility bridge, clone path, nested ownership, or persistence plumbing
remains.

## Root Triage

### Finding 1 — ACCEPTED — FIXED

One recognized identity currently hides another unknown supplied identity.

This violates the Exchange-evidence direction agreed with the user.

Remediation:

Ledger now requires every supplied CLOID and OID to resolve exactly.

It preserves Exchange CLOID, OID, TID, symbol, side, and execution evidence.

Unknown, missing, or conflicting Order identity fails before mutation.

### Finding 2 — ACCEPTED — FIXED WITH CORRECTED INVARIANT

Ledger mutation occurs before terminal-Trade rejection.

Returned errors can leave admitted Orders, Fills, indexes, and totals behind.

The reviewer's proposed terminal-parent rejection was incorrect.

Ledger, Trade, Order, and Fill represent current Exchange snapshots.

Remediation:

Removed Trade's terminal-state mutation lock.

Later synchronized evidence may revise a previously closed Trade snapshot.

Removed Fill aggregation's inferred `Filled` and `PartiallyFilled` Order status
changes.

Only synchronized Exchange Order evidence changes `Order.Status`.

### Finding 3 — ACCEPTED — FIXED

The quantity normalization directly violates arbitrary DCA and staged-exit
quantities.

Remediation:

Removed shared maximum-quantity normalization.

Account now rounds and validates each Order quantity independently.

An Order below minimum notional is rejected instead of silently resized.

### Finding 4 — REJECTED

The current `Canceled` Trade state contradicts the canonical zero-size and
all-Orders-closed rule.

Rejection reason:

The finding conflated detailed outcome with closed meta state.

An all-canceled zero-size Trade has `Status == Canceled` and
`IsClosed() == true`.

`Canceled` preserves that no execution completed.

`Closed` preserves that executed exposure returned to zero.

Canceled Trades must not count as completed round trips.

### Finding 5 — ACCEPTED — FIXED

Order and Trade currently split ownership of closure completeness.

`Order.IsClosed` must be the sole detailed Order completion contract.

Remediation:

`Order.IsClosed` now rejects every Order with a pending Fill fee.

Filled Orders additionally require full submitted quantity.

Trade no longer interprets Fill fee completeness.

Fee enrichment refreshes the Ledger active-Order index before Trade refresh.

### Finding 6 — REJECTED

VenueTID must reject changed immutable execution evidence before enrichment.

Rejection reason:

A valid Exchange cannot reuse one VenueTID for a different execution.

The reviewer supplied no reachable producer for that impossible identity
rewrite.

Account already rejects changed same-TID execution evidence while merging
Exchange rows.

Adding the same validation inside trusted Fill state would duplicate the
boundary contract.

### Finding 7 — ACCEPTED — FIXED WITH CORRECTED FILL IDENTITY

`TradeNumber`, `nextTradeNo`, and `Fill.ID` contradict the canonical key
hardcut.

Remediation:

Removed `TradeNumber`, `nextTradeNo`, and the old 21-bit limit.

Kept memory-only `nextTradeID`.

Renamed `Fill.ID` to `FillID`.

Added memory-only `nextFillID`.

Ledger now stores Fills by `FillID`, indexes VenueTID to FillID, and links
Orders to FillIDs.

The reviewer's proposed removal of local Fill identity was incorrect.

## Root Verdict

FAIL.

Five findings are accepted.

Findings 4 and 6 are rejected.

All five accepted findings are fixed.

Findings 4 and 6 remain rejected.

No production fixes were applied during this read-only audit.
