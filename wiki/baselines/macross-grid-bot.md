# Macross GridBot Baselines

This page preserves dated Grid replay baselines.

Invalid baselines remain visible until Grid behavior is stable.

## 2026-07-25 Initial Baseline

Status: INVALID - ERROR found

Defect: quantity used only the initial entry price.

Higher long re-entry prices could exceed their equal capital slice.

Audit also found whole-batch retries could duplicate an accepted uncertain submission.

Both defects were fixed before the corrected baseline.

Code base: `ff638268b0f08c8d9a7603d061cbbf179726fc26` plus uncommitted Grid implementation.

BotSpec: `macross_grid_bot`

Config hash: `a8fbe0216fb67ab7008eaf27d034767ab9f7cd54771d86c6130fa065f55ab3e5`

Replay: Sweep 10, Bot 14, BTC, 2026-03-01 through 2026-06-01.

| Metric | Result |
|---|---:|
| BtBot elapsed time | 25,848 ms |
| BtBot historical-data loop elapsed time | 23,908 ms |
| Tick rows | 7,948,800 |
| Controller runs | 794,880 |
| Signal packages | 2,207 |
| BotCycles | 50 |
| Trades | 1,954 |
| Orders | 4,641 |
| Fills | 2,578 |
| Cancellations | 2,063 |
| Closure Orders | 733 |
| Submission retries | 0 |
| Net PnL | -70.864647459999999999278 USDC |
| Ending equity | 929.135352540000000000722 USDC |
| Maximum drawdown | 88.027421204999999999563 USDC |

Result database: `workspace/db/sweeps/sweep_10/bot_14.db`

Result log: `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T064809Z.log`

SQLite integrity, foreign keys, Level counts, Config identity, final Orders, and final positions passed.

Fresh-process stability passed 2 of 2 and 10 of 10.

The 10x suite produced identical counts, PnL, equity, and drawdown.

2x log: `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T065218Z.log`

10x log: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T065316Z.log`

## 2026-07-25 Corrected Baseline

Status: INVALID - ERROR found

Defect: boundary-tick TP fills were omitted from the `round_trips` metric.

Trading behavior, PnL, equity, and drawdown remained correct.

Correction: size each Level against the greater initial or re-entry price.

Correction: retry only a proven non-submission.

Code base: `ff638268b0f08c8d9a7603d061cbbf179726fc26` plus uncommitted corrected Grid implementation.

BotSpec: `macross_grid_bot`

Config hash: `a8fbe0216fb67ab7008eaf27d034767ab9f7cd54771d86c6130fa065f55ab3e5`

Replay: Sweep 10, Bot 14, BTC, 2026-03-01 through 2026-06-01.

| Metric | Result |
|---|---:|
| BtBot elapsed time | 26,007 ms |
| BtBot historical-data loop elapsed time | 24,066 ms |
| Tick rows | 7,948,800 |
| Controller runs | 794,880 |
| Signal packages | 2,207 |
| BotCycles | 50 |
| Trades | 1,954 |
| Orders | 4,641 |
| Fills | 2,578 |
| Cancellations | 2,063 |
| Closure Orders | 733 |
| Submission retries | 0 |
| Net PnL | -69.766463889999999999562 USDC |
| Ending equity | 930.233536110000000000438 USDC |
| Maximum drawdown | 86.609100424999999999246 USDC |

Result database: `workspace/db/sweeps/sweep_10/bot_14.db`

Result log: `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T070934Z.log`

SQLite integrity, foreign keys, Level counts, Config identity, final Orders, and final positions passed.

Fresh-process stability passed 2 of 2 and 10 of 10.

All stability runs produced identical domain and financial results.

2x suite: 51,805 ms; BtBot average 24,754 ms; historical-data-loop average 24,006 ms.

10x suite: 260,764 ms; BtBot average 24,856 ms; historical-data-loop average 24,099 ms.

2x log: `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T071011Z.log`

10x log: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T071111Z.log`

## 2026-07-25 Final Corrected Baseline

Status: INVALID - ERROR found

Defect: marketable Grid GTC Orders were not matched during submission.

Defect: Account equity and drawdown snapshots stayed stale between Fill events.

Correction: derive completed round trips from terminal filled TP Orders.

Code base: `ff638268b0f08c8d9a7603d061cbbf179726fc26` plus uncommitted final Grid implementation.

BotSpec: `macross_grid_bot`

Config hash: `a8fbe0216fb67ab7008eaf27d034767ab9f7cd54771d86c6130fa065f55ab3e5`

Replay: Sweep 10, Bot 14, BTC, 2026-03-01 through 2026-06-01.

| Metric | Result |
|---|---:|
| BtBot elapsed time | 25,397 ms |
| BtBot historical-data loop elapsed time | 23,533 ms |
| Tick rows | 7,948,800 |
| Controller runs | 794,880 |
| Signal packages | 2,207 |
| BotCycles | 50 |
| Trades | 1,954 |
| Orders | 4,641 |
| Fills | 2,578 |
| Cancellations | 2,063 |
| Closure Orders | 733 |
| Submission retries | 0 |
| Net PnL | -69.766463889999999999562 USDC |
| Ending equity | 930.233536110000000000438 USDC |
| Maximum drawdown | 86.609100424999999999246 USDC |

Result database: `workspace/db/sweeps/sweep_10/bot_14.db`

Result log: `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T072506Z.log`

SQLite integrity, foreign keys, Level counts, Config identity, final Orders, and final positions passed.

Fresh-process stability passed 2 of 2 and 10 of 10.

All stability runs produced identical domain and financial results.

2x suite: 52,776 ms; BtBot average 25,314 ms; historical-data-loop average 24,588 ms.

10x suite: 268,307 ms; BtBot average 25,781 ms; historical-data-loop average 24,220 ms.

2x log: `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T072539Z.log`

10x log: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T072637Z.log`

## 2026-07-25 Telemetry And Stop-Role Baseline

Status: SUPERSEDED

This baseline includes corrected marketable Grid matching, current Account
equity, telemetry, RunReport, and canonical shutdown `stop` Orders.

Code base: `ff638268b0f08c8d9a7603d061cbbf179726fc26` plus uncommitted current implementation.

BotSpec: `macross_grid_bot`

Config hash: `a8fbe0216fb67ab7008eaf27d034767ab9f7cd54771d86c6130fa065f55ab3e5`

Replay: Sweep 10, Bot 14, BTC, 2026-03-01 through 2026-06-01.

| Metric | Result |
|---|---:|
| Tick rows | 7,948,800 |
| Controller runs | 794,880 |
| Telemetry samples | 794,881 |
| Signal packages | 2,208 |
| BotCycles | 50 |
| Trades | 1,982 |
| Orders | 4,697 |
| Fills | 2,636 |
| Cancellations | 2,061 |
| Closure Orders | 733 |
| Canonical `stop` Orders | 733 |
| Legacy `close` Orders | 0 |
| Submission retries | 0 |
| Gross PnL | -15.13202 USDC |
| Fees | 42.28805409 USDC |
| Net PnL | -57.420074089999999993851 USDC |
| Ending equity | 942.579925910000000006149 USDC |
| Maximum drawdown | 75.791979199999999992245 USDC |
| Result database size | 178,806,784 bytes |

Telemetry has one terminal sample.

Active-cycle zero-equity defects: zero.

Maximum-drawdown decreases: zero.

SQLite integrity, foreign keys, Level counts, Config identity, final Orders,
final positions, telemetry, and Order taxonomy passed.

Fresh-process stability passed 2 of 2 and 10 of 10.

All stability runs produced identical execution and financial results.

2x suite: 155,660 ms.

2x BtBot average: 76,524 ms.

2x historical-data-loop average: 71,970 ms.

10x suite: 780,162 ms.

10x BtBot average: 76,688.2 ms.

10x BtBot range: 75,153 through 77,857 ms.

10x historical-data-loop average: 72,237.3 ms.

10x historical-data-loop range: 70,800 through 73,338 ms.

10x heap-before-publication average: 690.165 MB.

10x heap-before-publication range: 457.830 through 790.377 MB.

1x report: `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T092852Z.json`

2x report: `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T093231Z.json`

10x report: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T093515Z.json`

10x log: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T093515Z.log`

## Baseline 1

Status: BASELINE

Config hash: `b22af3910b57625c47faca4d5b77142a6d64fd96ae9c3ea285dc7fc3dc6f0cf4`

| Metric | Result |
|---|---:|
| Suite elapsed time | 73,784 ms |
| BtBot elapsed time | 68,756 ms |
| Historical-data loop elapsed time | 64,846 ms |
| Tick rows | 7,948,800 |
| BotCycles | 50 |
| Trades | 1,982 |
| Orders | 4,697 |
| Fills | 2,636 |
| Cancellations | 2,061 |
| Stop Orders | 733 |
| Gross PnL | -14.98775 USDC |
| Fees | 42.290182025 USDC |
| Net PnL | -57.277932024999999997236 USDC |
| Ending equity | 942.722067975000000002764 USDC |
| Maximum drawdown | 75.655127584999999996194 USDC |

Report: `workspace/logs/nuubot5-stest-s11-b15-1-20260728T162532Z.json`

## Baseline 2 - Post ALTOFRVS Change

Status: CURRENT

ALTOFRVS means Account, Ledger, Trade, Order, Fill, Recon, Venue, and Simulator.

Config hash: `b22af3910b57625c47faca4d5b77142a6d64fd96ae9c3ea285dc7fc3dc6f0cf4`

| Metric | Result |
|---|---:|
| Suite elapsed time | 38,828 ms |
| BtBot elapsed time | 36,028 ms |
| Historical-data loop elapsed time | 32,089 ms |
| Tick rows | 7,948,800 |
| BotCycles | 50 |
| Trades | 1,980 |
| Orders | 4,693 |
| Fills | 2,632 |
| Cancellations | 2,061 |
| Stop Orders | 733 |
| Gross PnL | -15.217 USDC |
| Fees | 42.2232135 USDC |
| Net PnL | -57.440213499999999998272 USDC |
| Ending equity | 942.559786500000000001728 USDC |
| Maximum drawdown | 75.537686959999999996921 USDC |

Report: `workspace/logs/nuubot5-stest-s11-b15-1-20260729T042021Z.json`

Canonical comparison:

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

## Invalidation Rule

If a behavior bug changes this result, mark it `INVALID - ERROR found`.

Record the defect and retain its evidence.

Add the corrected rerun as a new dated baseline.
