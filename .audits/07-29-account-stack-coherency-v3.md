# Account Stack Coherency Adversarial Audit V3

Date: 2026-07-29
Round: 2
Reviewer status: FAIL

## Objective

Re-audit the flat Account stack after all accepted V2 fixes and the complete
identity hierarchy hardcut.

External callers, Recon, Store, Venue implementations, executors, and result
publication were excluded.

## Independent Reviewer Report

### 1. BLOCKER — Account invents Exchange Order status

Reachability:

Every Order placement uses these paths.

Evidence:

`order.go:113` assigns `Created` locally.

`account.go:294` assigns `Error` from local submission failure.

`account.go:386` assigns `Submitted` from placement acknowledgements.

`trade.md` says only synchronized Exchange evidence changes `Order.Status`.

Impact:

Order snapshots contain application lifecycle guesses instead of Exchange
status.

Active Order and Trade closure decisions consume those invented values.

Owning invariant:

Never invent or backfill Exchange Order status.

Smallest fix:

Keep submission lifecycle outside `Order.Status`.

Apply only exact synchronized Exchange statuses.

Remove locally admitted Orders when non-submission is proven.

### 2. BLOCKER — Reversing Fill errors after Ledger mutation

Reachability:

Current exposure is one.

An exit Order permits quantity two.

Exchange supplies a quantity 1.5 Fill.

The Fill fits its Order but reverses Trade exposure.

Evidence:

`ledger.go:307-315` stores the Fill, indexes it, advances identity, and updates
Order totals.

`ledger.go:316` refreshes the Trade afterward.

`trade.go:223-225` then rejects the reversing Fill.

Impact:

`AddFill` returns an error after Ledger records, indexes, counters, and Order
totals changed.

Owning invariant:

Rejected evidence must not mutate flat records or indexes.

Smallest fix:

Preflight prospective Trade calculation before committing Fill and Order
changes.

## Proof Checked

- Read every scoped source and design file.
- Traced identity allocation, Order submission, Fill admission, Trade finance,
  closure, and CLOID generation.
- Confirmed complete parent identity on Trade, Order, and Fill.
- Confirmed memory counters for TradeID, OrderID, and FillID.
- Confirmed VenueTID deduplication.
- Confirmed independent Order quantities.
- Confirmed Order-owned fee closure.
- Confirmed no Batch or TradeNumber remains.
- Confirmed `gofmt` and `git diff --check`.

## Proof Missing

- No compilation.
- No external caller, Recon, Store, Venue, executor, or publisher review.

## Bloat and Duplication Verdict

PASS.

No compatibility records, clones, persistence plumbing, Batch identity, or
material duplicate ownership remains.

## Root Triage

### Finding 1 — ACCEPTED

Local `Created` and submission-failure `Error` states contradict Exchange
snapshot authority.

Placement acknowledgement must preserve exact Exchange evidence.

### Finding 2 — REJECTED

Executor owns Order quantity and reduce-only decisions.

Current production Executors make every TP, SL, and closure Order reduce-only.

They size initial exits from entry quantity and later closures from open
exposure.

An Exchange Fill cannot exceed its submitted Order quantity.

Hyperliquid cannot reverse exposure through a reduce-only Order.

The finding supplied no reachable production producer for reversing evidence.

Account and Ledger must not repair or hide an Executor business-logic defect.

## Root Verdict

FAIL.

Finding 1 remains accepted.

No production fixes were applied during this read-only audit.
