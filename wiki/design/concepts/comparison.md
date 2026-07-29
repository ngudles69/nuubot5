# Comparison

User-facing backtest results use fixed-width comparison blocks.

Every current run shows:

```text
Executor               Baseline 1      Baseline 2      Current Run     Diff vs Baseline 2
                                        Post ALTOFRVS
```

ALTOFRVS means Account, Ledger, Trade, Order, Fill, Recon, Venue, and Simulator.

Use these metric groups in this order:

1. Suite, BtBot, and replay-loop time.
2. Ticks, Cycles, Trades, Orders, Fills, Cancellations, and Stop Orders.
3. Gross PnL, Fees, Net PnL, Ending Equity, and Maximum Drawdown.
4. Heap, Allocations, and GC Runs.

Use `same` when values match.

Show percentage differences for timing and resource metrics.

Show absolute differences for execution counts and financial values.

Completed round trips are not a baseline comparison metric.

Baseline 1 and Baseline 2 remain fixed until the user explicitly replaces them.

## Saved Grid Baselines

```text
Grid                   Baseline 1      Baseline 2       Change
                                       Post ALTOFRVS

Suite                    73.784s         38.828s        -47.4%
BtBot                    68.756s         36.028s        -47.6%
Replay loop              64.846s         32.089s        -50.5%

Ticks                  7,948,800       7,948,800        same
Cycles                        50              50        same
Trades                     1,982           1,980        -2
Orders                     4,697           4,693        -4
Fills                      2,636           2,632        -4
Cancellations              2,061           2,061        same
Stop orders                  733             733        same

Gross PnL              -14.987750      -15.217000       -0.229250
Fees                    42.290182       42.223214       -0.066969
Net PnL                -57.277932      -57.440213       -0.162281
Ending equity          942.722068      942.559787       -0.162281
Maximum drawdown        75.655128       75.537687       -0.117441

Heap                       708.1MB         440.9MB       -37.7%
Allocations             55,255.7MB      24,832.1MB      -55.1%
GC runs                      384             166        -56.8%
```
