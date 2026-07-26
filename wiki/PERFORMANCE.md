# Performance History

Each row records one fresh-process system-test suite.

## Performance Profiles

`stest.sh` owns fresh-process system, stress, and profiling proof:

```text
./stest.sh -bot <bot_id>
./stest.sh -sweep <sweep_id>
./stest.sh -bot <bot_id> -runs N
./stest.sh -bot <bot_id> -pp
```

`-pp` requires one run.

One session writes under:

```text
workspace/perf/profiles/stest-s<sweep>-b<bot>-<UTC timestamp>/
```

The run writes `run-001.cpu.pprof`, `.trace`, `.heap.pprof`,
`.allocs.pprof`, `.block.pprof`, and `.mutex.pprof`.

Inspect CPU samples by flat or cumulative cost:

```text
go tool pprof -top workspace/perf/profiles/<session>/run-001.cpu.pprof
go tool pprof -top -cum workspace/perf/profiles/<session>/run-001.cpu.pprof
go tool pprof -http=:8080 workspace/perf/profiles/<session>/run-001.cpu.pprof
```

Inspect the execution timeline:

```text
go tool trace workspace/perf/profiles/<session>/run-001.trace
```

Use the same `go tool pprof` commands with `.heap.pprof`, `.allocs.pprof`,
`.block.pprof`, or `.mutex.pprof`.

Profiling adds CPU, memory, synchronization, trace, GC, and file-write overhead.
Profiled timing is diagnostic, not a normal performance baseline.

## Escalating Test Ladder

```text
Level  Command                       Purpose
1      ./stest.sh -bot 9             Basic runtime, Parquet, loop, telemetry, Report, publication
2      ./stest.sh -bot 9 -pp         Performance-profile pipeline
3      ./stest.sh -bot 13            Account, Ledger, Simulator, Orders, Fills, PnL
4      ./stest.sh -bot 15            Planned one-month Grid smoke and state-growth stress
5      ./stest.sh -bot 14            Three-month Grid baseline proof
6      ./stest.sh -sweep 10 -runs 10 Fresh-process deterministic stability
7      Planned six-month Grid Extended state-growth stability
8      Planned 2025 full year Long-run lifecycle and strategy proof
```

Run Levels 1 through 3 after ordinary trading changes.
Run Level 4 after its Sweep exists.
Run expensive stability levels only after lower levels pass.

### Expected Baselines

```text
Level  BtBot ms  Replay ms  Expected outcome
1      5,604        2,152      7,948,800 ticks; 794,880 runs; 64 cycles; report complete
2      7,111        2,190      Level 1 plus six readable performance artifacts
3      14,250       7,972      193 cycles; 193 Trades; 626 Orders; 386 Fills
4      Pending      Pending    Establish after Sweep 11 Bot 15 exists
5      76,688 avg   72,237 avg 50 cycles; 1,982 Trades; 4,697 Orders; 2,636 Fills
```

Deterministic semantic results must match their accepted baseline.
Timing is observational but must remain explainable.
A material speedup or slowdown requires investigation.
Replace a timing baseline only after the changed cost and result remain proven.

### Grid Baseline 1 — Post-Rollback

The first post-Chunk rollback baseline used unchanged reconciliation logic.

```text
Command             ./stest.sh -sweep 10
BtBot                80,225 ms
Historical loop      75,049 ms
Total allocation     142,471.675 MB
Trades               1,982
Orders               4,697
Fills                2,636
Net PnL              -57.420074089999999993851 USDC
Ending equity        942.579925910000000006149 USDC
Maximum drawdown     75.791979199999999992245 USDC
Result log           workspace/logs/nuubot5-stest-s10-b14-1-20260726T104521Z.log
Suite report         workspace/logs/nuubot5-stest-s10-b14-1-20260726T104521Z.json
```

Profiled proof preserves the same semantic result:

```text
Command             ./stest.sh -bot 14 -pp
BtBot                82,880 ms
Historical loop      77,930 ms
Total allocation     142,514.548286438 MB
Trades               1,982
Orders               4,697
Fills                 2,636
Net PnL              -57.420074089999999993851 USDC
Ending equity        942.579925910000000006149 USDC
Maximum drawdown     75.791979199999999992245 USDC
Result log           workspace/logs/nuubot5-stest-s10-b14-1-20260726T104650Z.log
Suite report         workspace/logs/nuubot5-stest-s10-b14-1-20260726T104650Z.json
Profile directory    workspace/perf/profiles/stest-s10-b14-20260726T104650Z/
```

Profiled timing is not the normal timing baseline.

## Controller Replay

| Commit | Change | Runs | Passed | Suite ms | BtBot avg [min-max] ms | Historical-data loop avg [min-max] ms | Log |
|---|---|---:|---:|---:|---:|---:|---|
| `b088e98` | Two-column stream baseline | 10 | 10 | 5,662 | 454 [447-464] | 375 [371-382] | `workspace/logs/nuubot5-rtest-s6-b9-10-20260723T105957Z.log` |
| `b088e98` | Two-column stream stability | 500 | 500 | 291,614 | 463 [444-531] | 382 [364-442] | `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T110542Z.log` |
| Uncommitted | Seven-column Load | 2 | 2 | 4,787 | 2,260 [1,639-2,881] | 1,590 [1,559-1,621] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T112959Z.log` |
| Uncommitted | Seven-column Load stability | 500 | 500 | 893,221 | 1,649 [1,629-1,701] | 1,566 [1,548-1,626] | `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T113055Z.log` |
| Uncommitted | Seven-column Stream | 2 | 2 | 4,199 | 1,917 [1,349-2,486] | 1,264 [1,261-1,268] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T124625Z.log` |
| Uncommitted | Seven-column Stream stability | 500 | 500 | 766,287 | 1,345 [1,329-1,383] | 1,265 [1,245-1,300] | `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T124647Z.log` |
| Uncommitted | Six-column Stream | 2 | 2 | 3,957 | 1,772 [1,204-2,340] | 1,125 [1,124-1,127] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T143417Z.log` |
| Uncommitted | Six-column Stream stability | 500 | 500 | 706,950 | 1,204 [1,165-1,475] | 1,124 [1,090-1,338] | `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T143429Z.log` |
| Uncommitted | Six-column Stream, 122,880 batch | 2 | 2 | 3,994 | 1,773 [1,189-2,358] | 1,110 [1,110-1,111] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T144936Z.log` |
| Uncommitted | Six-column Stream, 122,880 batch stability | 500 | 500 | 728,463 | 1,219 [1,177-1,530] | 1,134 [1,098-1,445] | `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T145016Z.log` |
| Uncommitted | Simple BtBot logs | 2 | 2 | 4,011 | 1,754 [1,186-2,322] | 1,112 [1,109-1,116] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T154011Z.log` |
| Uncommitted | BtBot review cleanup | 2 | 2 | 4,025 | 1,771 [1,191-2,351] | 1,111 [1,110-1,113] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T154808Z.log` |
| Uncommitted | Exact-format Logger | 2 | 2 | 4,049 | 1,776 [1,204-2,349] | 1,117 [1,114-1,120] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T161651Z.log` |
| Uncommitted | BtBot Loop and direct errors | 2 | 2 | 4,010 | 1,762 [1,194-2,331] | 1,115 [1,114-1,117] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T163450Z.log` |
| `5011d91` | Release-driven Signaler | 20 | 20 | 35,788 | 1,542 [1,488-1,626] | 1,456 [1,410-1,542] | `workspace/logs/nuubot5-rtest-s6-b9-20-20260724T035217Z.log` |
| Uncommitted | Flat-map Signal packages | 2 | 2 | 4,731 | 2,112 [1,562-2,663] | 1,483 [1,482-1,484] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260724T082230Z.log` |
| Uncommitted | Typed Signal packages | 2 | 2 | 4,641 | 2,050 [1,491-2,610] | 1,413 [1,406-1,421] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260724T083008Z.log` |
| Uncommitted | Typed Signal package stability | 20 | 20 | 34,499 | 1,478 [1,463-1,509] | 1,397 [1,377-1,426] | `workspace/logs/nuubot5-rtest-s6-b9-20-20260724T083014Z.log` |
| Uncommitted | Direct history index experiment | 2 | 2 | 4,625 | 2,061 [1,504-2,618] | 1,420 [1,415-1,426] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260724T082609Z.log` |
| Uncommitted | BotSpec and Controller proof | 2 | 2 | 4,550 | 1,968 [1,801-2,136] | 1,859 [1,700-2,019] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260724T202905Z.log` |
| Uncommitted | BotSpec and Controller stability | 10 | 10 | 31,059 | 2,665 [1,809-2,928] | 2,475 [1,707-2,757] | `workspace/logs/nuubot5-rtest-s6-b9-10-20260724T202915Z.log` |
| Uncommitted | Post-audit Controller proof | 1 | 1 | 2,064 | 1,788 [1,788-1,788] | 1,694 [1,694-1,694] | `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T204033Z.log` |
| Uncommitted | Telemetry and RunReport proof | 1 | 1 | 7,541 | 7,177 [7,177-7,177] | 2,351 [2,351-2,351] | `workspace/logs/nuubot5-rtest-s6-b9-1-20260725T091159Z.json` |

## TradeBot Replay

Each row runs Sweep 9 Bot 13 through Account, Ledger, Simulator, and result
publication.

| Commit | Change | Runs | Passed | Suite ms | BtBot avg ms | Historical-data loop avg ms | Log |
|---|---|---:|---:|---:|---:|---:|---|
| Uncommitted | Exact Config and equity proof | 2 | 2 | 14,737 | 6,560 | 4,248 | `workspace/logs/nuubot5-trtest-s9-b13-2-20260724T202731Z.log` |
| Uncommitted | Exact Config and equity stability | 10 | 10 | 67,743 | 5,986 | 4,261 | `workspace/logs/nuubot5-trtest-s9-b13-10-20260724T202751Z.log` |
| Uncommitted | Post-audit complete result proof | 1 | 1 | 10,246 | 9,393 | 4,248 | `workspace/logs/nuubot5-trtest-s9-b13-1-20260724T204016Z.log` |
| Uncommitted | Telemetry and RunReport proof | 1 | 1 | 14,955 | 14,250 | 7,972 | `workspace/logs/nuubot5-trtest-s9-b13-1-20260725T092825Z.json` |

## GridBot Replay

Each row runs Sweep 10 Bot 14 through GridExecutor, Account, Ledger, Simulator, and result publication.

| Commit | Change | Runs | Passed | Suite ms | BtBot avg ms | Historical-data loop avg ms | Log |
|---|---|---:|---:|---:|---:|---:|---|
| `ff63826` + Grid | INVALID: initial Grid sizing defect | 1 | 1 | 27,006 | 25,848 | 23,908 | `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T064809Z.log` |
| `ff63826` + Grid | INVALID: initial Grid sizing defect | 2 | 2 | 51,072 | 24,381 | 23,622 | `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T065218Z.log` |
| `ff63826` + Grid | INVALID: initial Grid sizing defect | 10 | 10 | 256,561 | 24,550 | 23,787 | `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T065316Z.log` |
| `ff63826` + Grid | INVALID: corrected round-trip metric defect | 1 | 1 | 27,199 | 26,007 | 24,066 | `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T070934Z.log` |
| `ff63826` + Grid | INVALID: corrected round-trip metric defect | 2 | 2 | 51,805 | 24,754 | 24,006 | `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T071011Z.log` |
| `ff63826` + Grid | INVALID: corrected round-trip metric defect | 10 | 10 | 260,764 | 24,856 | 24,099 | `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T071111Z.log` |
| `ff63826` + Grid | INVALID: GTC matching and stale Account snapshots | 1 | 1 | 26,514 | 25,397 | 23,533 | `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T072506Z.log` |
| `ff63826` + Grid | INVALID: GTC matching and stale Account snapshots | 2 | 2 | 52,776 | 25,314 | 24,588 | `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T072539Z.log` |
| `ff63826` + Grid | INVALID: GTC matching and stale Account snapshots | 10 | 10 | 268,307 | 25,781 | 24,220 | `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T072637Z.log` |
| `ff63826` + Grid | Current telemetry and stop-role proof | 2 | 2 | 155,660 | 76,524 | 71,970 | `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T093231Z.json` |
| `ff63826` + Grid | Current telemetry and stop-role stability | 10 | 10 | 780,162 | 76,688 | 72,237 | `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T093515Z.json` |

## Memory

| Commit | Change | Runs | Heap avg [min-max] MB | Total allocation avg [min-max] MB | GC runs avg [min-max] | GC pause avg [min-max] ms |
|---|---|---:|---:|---:|---:|---:|
| `b088e98` | Two-column stream baseline | 10 | 16.280 [10.590-18.240] | 452.788 [452.707-452.855] | 43.200 [43-44] | 4.179 [2.006-5.987] |
| `b088e98` | Two-column stream stability | 500 | 15.147 [7.129-22.185] | 452.787 [452.650-452.923] | 43.234 [42-45] | 3.530 [0.000-16.977] |
| Uncommitted | Seven-column Load | 2 | 121.925 [113.692-130.157] | 4,251.356 [4,251.274-4,251.438] | 69.500 [68-71] | 5.157 [4.542-5.772] |
| Uncommitted | Seven-column Load stability | 500 | 126.560 [89.374-161.525] | 4,251.283 [4,251.077-4,251.579] | 69.086 [67-73] | 5.860 [0.000-21.317] |
| Uncommitted | Seven-column Stream | 2 | 31.367 [25.255-37.479] | 1,549.660 [1,549.618-1,549.703] | 65.000 [65-65] | 8.348 [5.209-11.487] |
| Uncommitted | Seven-column Stream stability | 500 | 30.733 [14.757-48.164] | 1,549.676 [1,549.499-1,549.874] | 66.202 [64-69] | 5.072 [0.000-18.556] |
| Uncommitted | Six-column Stream | 2 | 26.979 [21.933-32.025] | 1,321.129 [1,321.122-1,321.135] | 64.000 [64-64] | 7.268 [4.999-9.536] |
| Uncommitted | Six-column Stream stability | 500 | 28.604 [13.189-41.045] | 1,321.159 [1,321.016-1,321.325] | 63.722 [62-66] | 5.090 [0.000-18.856] |
| Uncommitted | Six-column Stream, 122,880 batch | 2 | 33.604 [33.420-33.789] | 975.720 [975.673-975.766] | 50.000 [50-50] | 2.877 [2.629-3.126] |
| Uncommitted | Six-column Stream, 122,880 batch stability | 500 | 31.792 [13.197-47.537] | 975.697 [975.524-975.912] | 49.880 [48-52] | 5.011 [0.000-15.789] |
| Uncommitted | Simple BtBot logs | 2 | 35.913 [35.641-36.185] | 975.692 [975.657-975.726] | 49.000 [49-49] | 3.399 [3.014-3.783] |
| Uncommitted | BtBot review cleanup | 2 | 35.833 [33.424-38.243] | 975.703 [975.679-975.726] | 49.500 [49-50] | 3.013 [3.005-3.020] |
| Uncommitted | Exact-format Logger | 2 | 34.548 [34.383-34.712] | 975.697 [975.675-975.719] | 50.000 [50-50] | 7.098 [4.099-10.097] |
| Uncommitted | BtBot Loop and direct errors | 2 | 40.306 [33.423-47.189] | 975.742 [975.687-975.796] | 49.500 [49-50] | 3.170 [3.116-3.224] |
| `5011d91` | Release-driven Signaler | 20 | 34.064 [21.146-47.183] | 975.698 [975.599-975.872] | 49.750 [48-51] | 5.634 [1.000-18.778] |
| Uncommitted | Flat-map Signal packages | 2 | 37.873 [27.085-48.662] | 977.862 [977.840-977.884] | 47.500 [47-48] | 7.651 [1.706-13.596] |
| Uncommitted | Typed Signal packages | 2 | 28.375 [20.967-35.782] | 976.632 [976.593-976.670] | 48.000 [48-48] | 2.038 [1.050-3.026] |
| Uncommitted | Typed Signal package stability | 20 | 29.939 [18.070-48.305] | 976.613 [976.513-976.724] | 48.550 [47-50] | 4.242 [0.000-23.410] |
| Uncommitted | Direct history index experiment | 2 | 34.493 [34.463-34.523] | 976.631 [976.624-976.637] | 48.500 [48-49] | 3.318 [1.730-4.906] |
| Uncommitted | BotSpec and Controller stability | 10 | 17.852 [4.806-27.879] | 1,215.230 [1,215.112-1,215.332] | 64.500 [61-66] | 9.278 [3.131-17.918] |
| `ff63826` + Grid | Current Grid telemetry stability | 10 | 690.165 [457.830-790.377] | 142,471.599 [142,471.403-142,471.729] | 1,055.600 [1,053-1,059] | 104.072 [91.321-134.493] |

## Comparison

Seven-column Stream versus seven-column Load:

- Replay improved 19.2 percent.
- Process time improved 18.4 percent.
- Total allocation fell 63.5 percent.
- Heap fell 75.7 percent.
- Garbage collections fell 4.2 percent.

Seven-column Stream remains 3.31 times slower than the two-column stream baseline.

Six-column Stream versus seven-column Stream:

Passive Signaler hardcut versus release-driven Signaler:

- Typed packages improved 20x historical-data-loop average 4.1 percent.
- Typed packages improved 20x BtBot average 4.2 percent.
- Total allocation increased 0.094 percent.
- Direct regular-history indexing showed no improvement and was removed.

- Replay improved 11.1 percent.
- Process time improved 10.5 percent.
- Total allocation fell 14.7 percent.
- Heap fell 6.9 percent.
- Garbage collections fell 3.7 percent.

Six-column Stream remains 2.94 times slower than the two-column stream baseline.

122,880 batch versus 65,536 batch:

- Replay slowed 0.9 percent.
- Process time slowed 1.2 percent.
- Total allocation fell 26.1 percent.
- Heap rose 11.1 percent.
- Garbage collections fell 21.7 percent.

The larger batch reduces allocation and garbage collections without improving speed.
