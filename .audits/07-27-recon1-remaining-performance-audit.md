# Recon1 Remaining Performance Audit

Date: 2026-07-27

Profile:

```text
workspace/perf/profiles/stest-s11-b15-20260726T184654Z/
```

## Summary

Snapshot removal preserved exact finance and reduced cost materially.

```text
Metric                 Original Recon1   Snapshot-free   Change
BtBot                    77.118 s          50.857 s      -34.1%
Historical loop          71.698 s          46.689 s      -34.9%
Total allocation        108,227 MB        46,730 MB      -56.8%
GC runs                     752               320        -57.4%
```

Exact proof remains:

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

The remaining cost is repeated full Recon and Simulator matching. It is not database waiting.

## Clone-Removal Follow-up

```text
Profile             workspace/perf/profiles/stest-s11-b15-20260727T033944Z/
BtBot               49.514 s
Historical loop     45.366 s
Total allocation    46,354.773 MB
GC runs             318
cloneTrades samples 0
```

Compared with the immediate snapshot-free profile, allocation fell `375.563 MB` and profiled runtime fell `1.343s`.

Exact domain and finance results remained unchanged. Five fresh processes also passed identical stability proof.

## Ranked Findings

### 1. Scheduled Recon remains the largest cost

Recon consumed `26.27s` cumulative CPU and `30.83 GB` cumulative allocation.

Every scheduled polling Recon must still request Venue truth. Local state cannot prove that exchange-side Orders or Fills remained unchanged.

Recon-frequency optimization is deferred. Future WebSocket evidence may provide an exchange-dirty hint, but it cannot change polling behavior without separate proof.

Current tuning should reduce repeated work after each response, not skip scheduled Recon.

### 2. Decimal work dominates allocation

```text
math/big.nat.make              21.71 GB flat
Decimal.rescale                17.61 GB cumulative
Decimal.Add                    13.78 GB cumulative
Decimal.Cmp                    13.71 GB cumulative
```

Most work comes from repeating finance, Account summary, and Simulator price comparisons.

Do not replace decimal arithmetic blindly. First remove repeated calculations after unchanged Venue evidence is compared.

### 3. Account summary repeats full Trade aggregation

`Ledger.ReconSummary` allocated `8.21 GB` cumulatively and consumed `5.98s` CPU.

Every full Recon loops all Trades and adds decimal totals.

Recommended change:

- maintain Ledger finance totals when touched Trades change;
- refresh current marks only for open Trades;
- keep closed Trade totals static.

Linear aggregation remains acceptable as a fallback. It is expensive only because it runs 277,208 times.

### 4. Trade finance recalculates every active Trade

`UpdateReconTrades` allocated `5.79 GB`; `Trade.RefreshRecon` owned `4.91 GB`.

Recommended change:

- recalculate Trades touched by changed Order or Fill evidence;
- refresh current marks only for open Trades;
- keep closed Trades unchanged;
- avoid repeated before-and-after decimal work when evidence is identical.

### 5. Order evidence is rebuilt every full Recon

```text
Download Order evidence        5.51 GB cumulative
Simulator.OpenOrders           3.11 GB cumulative
Account evidence slices        1.49 GB flat
Ledger.ActiveOrders            0.92 GB cumulative
```

`Simulator.OpenOrders` allocates and sorts a new slice every call.

Recommended Simulator-only change:

- preserve exchange-equivalent polling behavior;
- avoid rebuilding and sorting unchanged internal open-order representations;
- invalidate cached representations on place, cancel, or match;
- use direct active-order iteration where ordering is unnecessary.

### 6. AccountState is recalculated every full Recon

`downloadAccountState` allocated `7.42 GB`; Simulator AccountState owned `6.89 GB`.

`Simulator.AccountState` calls `position()`. That function replays every historical Fill each time.

Recommended change:

- maintain Simulator position and realized totals when Fills are accepted;
- return current Account state without replaying complete Fill history;
- derive the compact Account Snapshot from maintained Ledger totals and the current mark;
- retain exact Account Snapshot freshness telemetry.

### 7. Simulator matching remains independent work

Simulator BBO ingestion consumed `13.09s` CPU. Matching consumed `12.04s`; price crossing consumed `7.58s`.

This is not a Recon snapshot problem.

Recommended later review:

- inspect repeated decimal rescaling in `simulator.crosses`;
- normalize comparable price scales at Order admission;
- avoid changing arithmetic until exact matching parity tests exist.

### 8. Residual canonical Ledger cloning is removed

`CreateTrade`, `AddOrders`, and `RecordSubmit` now mutate only touched owned records.

Under `none`, they perform no persistence preparation. Under `max`, one transaction writes only touched identity, Trade, and Order rows.

The follow-up allocation profile contains zero `cloneTrades` samples. Frozen Recon2 retains its control-path clone.

### 9. Smaller allocation owners

```text
Order.copyInput                 1.39 GB
BtBot telemetry                 1.63 GB cumulative
Terminal result publication    1.71 GB cumulative
Parquet allocation             0.80 GB
```

Recommended order:

1. maintain incremental Ledger totals;
2. recalculate only touched and open Trades after polling;
3. maintain Simulator position totals incrementally;
4. avoid rebuilding unchanged Simulator open-order representations;
5. narrow Order Recon state to remove pointer copying;
6. review telemetry retention;
7. inspect Simulator decimal scale normalization.

## Trace Result

Scheduler delay totaled only `395.18ms`.

Application syscall delay was about `1.37s`. SQLite statement execution contributed about `0.27s`. Mutex delay totaled about `351ms`.

Block profiles are dominated by profiling runtime goroutines, not application waits.

No lock, channel, SQLite, or syscall wait explains the remaining 50.9-second runtime.

The run remains CPU and allocation bound.

## Next Proof

After the next approved tuning change, run:

```text
./stest.sh -bot 15
./stest.sh -bot 15 -pp
```

Required proof:

- exact Trade, Order, Fill, PnL, equity, and drawdown parity;
- unchanged scheduled Recon count and exact polling behavior;
- lower `UpdateReconTrades`, `ReconSummary`, `downloadAccountState`, and `downloadOrderEvidence` cost;
- profile path recorded under `workspace/perf/profiles/`.
