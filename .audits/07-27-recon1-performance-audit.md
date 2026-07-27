# Recon1 Performance Audit

Date: 2026-07-27

Scope: Read-only analysis of Sweep 11 Bot 15 Recon1.

Profile:

```text
workspace/perf/profiles/stest-s11-b15-20260726T172917Z/
```

## Summary

Recon1 is correct and faster than Recon2, but it still performs full reconciliation for nearly every active-cycle mark update.

```text
Metric                    Recon1          Recon2          Change
BtBot                     77.118 s        91.058 s        -15.3%
Historical loop           71.698 s        86.395 s        -17.0%
Total allocation          108,227 MB      146,027 MB      -25.9%
GC runs                   752             1,091           -31.1%
Recon calls               277,704         277,704         same
Clean skips               496             496             same
Executed Recon            277,208         277,208         same
Failed Recon              0               0               same
```

Exact Trade, Order, Fill, finance, equity, drawdown, BotCycle timing, and Recon counts match.

The primary remaining fault is not a missing skip assignment.

The current dirty rule intentionally makes every open-position BBO require Recon:

```go
if changed || !a.lastSnapshot.PositionQuantity.IsZero() {
    a.dirty = true
}
```

Grid normally carries open exposure. Price-only BBOs therefore execute all ten Recon steps.

Most attempts download and rebuild the same Order and Fill evidence. Only marked unrealized PnL and Account equity changed.

The highest-value change is to separate Venue reconciliation from mark refresh.

```text
Venue dirty  -> full ten-step Recon
Mark dirty   -> lightweight Trade and Account mark refresh
Clean        -> skip
```

A 10–15 second total run is not proven achievable from Recon changes alone.

The first changes should remove most of Recon's 42.42 cumulative CPU seconds and large snapshot allocations.

Simulator matching still consumes 13.15 cumulative CPU seconds. Replay, telemetry, and terminal publication also remain.

## Accepted Correctness Proof

```text
Sweep 11 Bot 15  Recon1
Sweep 12 Bot 16  Recon2
```

Both produced:

```text
Trades             1,982
Orders             4,697
Fills               2,636
Gross PnL           -15.13202
Fees                42.28805409
Net PnL             -57.420074089999999993851
Ending equity       942.579925910000000006149
Maximum drawdown    75.791979199999999992245
```

Database comparison found zero Trade, Order, or Fill differences.

## Cadence Evidence

Both paths recorded:

```text
Controller runs         794,880
BotCycles                    50
Time inside BotCycles    32d 02:59:30
Time outside BotCycles   59d 21:00:29
Recon calls             277,704
Clean skips                 496
Executed Recon          277,208
Succeeded Recon         277,208
Failed Recon                  0
```

Executed Recon rate:

```text
277,208 / 277,704 = 99.82%
```

Ordinary resting Grid Orders do not set reconciliation pending.

`pendingOrders` contains only unresolved reconciliation or missing-fee evidence.

Open position marking is the reason almost every active-cycle call executes.

## CPU Evidence

Profile wall duration was 77.93 seconds. CPU samples totaled 124.70 seconds because GC workers run concurrently.

```text
Owner                                      Cumulative CPU
Account/BotCycle Recon                     42.42 s
Update Trade records                       12.04 s
Grid OnRecon and Trade lookup               8.74 s
Update Account Snapshot                     7.35 s
Update Order records                        6.11 s
Download Account state                      5.65 s
Persist and publish attempt                 5.33 s
Download Order evidence                     3.66 s
Simulator matching                         13.15 s
```

Runtime allocation and GC dominate CPU:

```text
GC background workers                      42.71 s
mallocgc                                   26.33 s
Trade.snapshot                             14.12 s
Decimal rescaling                          16.03 s
Order.Snapshot                              6.61 s
```

Recon optimization must reduce allocation, not only shorten loops.

## Allocation Evidence

The allocations profile measured 107.60 GB.

```text
Owner                                      Cumulative allocation
Trade.snapshot                             44.55 GB
Trade.State calls                          31.26 GB
UpdateReconTrades                          25.93 GB
Stage touched Order                        11.29 GB
UpdateReconOrders                          12.92 GB
Download Order evidence                     9.40 GB
Download Account state                      7.29 GB
Validate Recon indexes                      4.81 GB
Ledger.ActiveOrders                         4.89 GB
```

`Trade.snapshot` is the largest single owner:

```text
Allocate Order snapshot slice              22.70 GB
Snapshot owned Orders                       5.00 GB
Sort Order snapshots                       13.41 GB cumulative
Mark finance                                3.44 GB cumulative
```

`UpdateReconTrades` deep-snapshots each Trade twice:

```text
previous Trade.State                        9.77 GB
RefreshRecon                                5.46 GB
new Trade.State                            10.59 GB
```

This is repeated for every active Trade on every executed Recon.

## Trace Evidence

The run is CPU and allocation bound.

Application syscall delay was about 2.28 seconds.

Per-Recon path logging contributed about 1.31 seconds of syscall delay.

SQLite execution contributed below one second.

No material application lock or channel wait explains the 78-second run.

Trace-generated profiler goroutines dominate the sync profile. They are measurement infrastructure, not application blocking.

## Recommendation 1 — Separate Dirty Reasons

Add exact Account facts:

```text
venueDirty
markDirty
```

Set `venueDirty` when:

- placing or canceling Orders;
- Simulator reports matching changes;
- pending fee or Order repair exists;
- Recon fails;
- forced full Recon is requested.

Set `markDirty` when:

- an open position exists; and
- the mark changed after the last trusted Snapshot.

Step 1 chooses:

```text
venueDirty  -> full Recon
markDirty   -> mark refresh
neither     -> skip
```

Mark refresh should:

1. read the current mark;
2. update stored unrealized PnL for open Trades;
3. aggregate stored Trade finance;
4. update Account equity and drawdown;
5. publish one trusted Snapshot;
6. emit mark-refresh telemetry.

Mark refresh must not:

- download Order history;
- download Fill history;
- clone Orders or Fills;
- rebuild execution PnL;
- validate identity indexes;
- advance Fill cursors;
- write unchanged domain rows.

Expected impact: largest available reduction.

Exact full-Recon reduction is unknown because current telemetry does not count Simulator `changed=true` observations.

Add counters before implementation proof:

```text
full_recon
mark_refresh
clean_skip
venue_changed
```

## Recommendation 2 — Stop Deep Trade Snapshots During Recon

`UpdateReconTrades` needs compact finance and status before-and-after values.

It does not need complete Order and Fill snapshots for equality comparison.

Use one allocation-light Trade state containing:

```text
status
side
open quantity
average entry
realized PnL
unrealized PnL
gross PnL
fees
net PnL
timestamps
```

Keep complete `Trade.Snapshot` for terminal result and explicit external reads only.

Expected impact: remove a large part of the 25.93 GB allocated by `UpdateReconTrades`.

## Recommendation 3 — Give Grid a Compact Trade Status Read

`GridExecutor.OnRecon` reads each active level's Trade through `Account.Trade`.

That call builds and sorts a complete Trade, Order, and Fill snapshot.

Grid uses only:

```text
TradeNo
Status
```

Add a narrow immutable status read for this decision.

Do not build complete evidence trees for level-state checks.

Expected impact: remove much of the 13.6 GB allocated through marked `Trade.Snapshot` calls.

## Recommendation 4 — Refresh Only Touched Execution Trades

Current `UpdateReconTrades` clones every active Trade before refreshing marked finance.

After Recommendation 1:

- full Recon refreshes only Trades touched by changed Order or Fill evidence;
- mark refresh updates only open Trade finance;
- closed Trades remain static.

Do not combine execution refresh and mark refresh into one complete-Trade clone.

## Recommendation 5 — Avoid Full Order Snapshots for Evidence Matching

`downloadOrderEvidence` allocates 9.40 GB cumulatively.

It builds:

- Simulator open-Order snapshots;
- an open-CLOID map;
- Ledger active-Order snapshots containing unnecessary fields and Fills.

The matching step needs compact values:

```text
OrderID
CLOID
VenueOrderID
Status
```

Use stable Ledger indexes and compact Venue evidence.

Do not copy owned Fill slices for active-Order membership checks.

## Recommendation 6 — Validate Only Changed Identities

`validateReconIndexes` allocates 4.81 GB cumulatively.

Mark refresh changes no identity and should perform no index validation.

Full Recon should validate only staged identity additions or transitions.

Complete index rebuild remains an Init and focused-test operation.

## Recommendation 7 — Reduce Touched-Tree Cloning

`stageOrder` cumulatively allocates 11.29 GB.

It clones the complete owning Trade, including every Order and Fill, when one Order changes.

After compact state changes are proven, evaluate copy-on-write staging for touched Orders only.

This is higher risk than Recommendations 1–3. Implement it later.

## Recommendation 8 — Measure Decimal Work After Structural Fixes

Decimal and `math/big` work remain large:

```text
math/big.nat.make flat                    23.16 GB
Decimal rescale cumulative                18.92 GB
Decimal Add cumulative                    14.77 GB
```

Do not replace `decimal` now.

First remove repeated calculations and snapshots.

Then profile remaining decimal operations and normalize exponents only where exact parity proves safe.

## Recommendation 9 — Remove Temporary Per-Recon Path Logging Later

The path log proves Recon selection, but writes 277,208 lines per run.

Measured syscall delay is about 1.31 seconds.

This is not the primary bottleneck.

Remove or aggregate it after Recon1 cutover proof.

## Expected Performance Range

Confirmed current result:

```text
77–78 seconds
108 GB allocation
```

Reasonable first target after Recommendations 1–3:

```text
20–40 seconds
materially below 50 GB allocation
```

This range is an inference, not proof.

A 10–15 second run probably also requires:

- Simulator matching allocation reduction;
- reduced decimal rescaling;
- lighter periodic telemetry collection;
- faster terminal result publication.

Profile after each coherent change. Do not stack unmeasured optimizations.

## Implementation Order

```text
1. Add full_recon, mark_refresh, clean_skip, and venue_changed telemetry.
2. Split venueDirty from markDirty.
3. Implement lightweight mark refresh.
4. Replace Recon Trade.State comparisons with compact state.
5. Replace Grid complete Trade reads with compact status reads.
6. Re-run exact Recon1/Recon2 parity.
7. Profile and reassess.
8. Optimize Order evidence and touched-tree cloning only if still material.
```

## Required Proof

Every optimization round must prove:

```text
zero Trade differences
zero Order differences
zero Fill differences
exact finance equality
exact equity equality
exact drawdown equality
zero Recon failures
same BotCycle start/end/duration
same accepted execution decisions
```

Report:

```text
full Recon count
mark refresh count
clean skip count
Venue changed count
BtBot duration
historical loop duration
total allocation
GC runs
CPU top
allocation top
trace syscall and scheduler delay
```
