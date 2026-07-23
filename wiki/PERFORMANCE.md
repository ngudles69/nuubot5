# Performance History

Each row records one fresh-process `rtest.sh` suite.

## Runtime

| Commit | Change | Runs | Passed | Suite ms | Process avg [min-max] ms | Replay avg [min-max] ms | Log |
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
| Uncommitted | Simple BtRunner logs | 2 | 2 | 4,011 | 1,754 [1,186-2,322] | 1,112 [1,109-1,116] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T154011Z.log` |
| Uncommitted | BtRunner review cleanup | 2 | 2 | 4,025 | 1,771 [1,191-2,351] | 1,111 [1,110-1,113] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T154808Z.log` |
| Uncommitted | Exact-format Logger | 2 | 2 | 4,049 | 1,776 [1,204-2,349] | 1,117 [1,114-1,120] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T161651Z.log` |
| Uncommitted | BtRunner Loop and direct errors | 2 | 2 | 4,010 | 1,762 [1,194-2,331] | 1,115 [1,114-1,117] | `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T163450Z.log` |

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
| Uncommitted | Simple BtRunner logs | 2 | 35.913 [35.641-36.185] | 975.692 [975.657-975.726] | 49.000 [49-49] | 3.399 [3.014-3.783] |
| Uncommitted | BtRunner review cleanup | 2 | 35.833 [33.424-38.243] | 975.703 [975.679-975.726] | 49.500 [49-50] | 3.013 [3.005-3.020] |
| Uncommitted | Exact-format Logger | 2 | 34.548 [34.383-34.712] | 975.697 [975.675-975.719] | 50.000 [50-50] | 7.098 [4.099-10.097] |
| Uncommitted | BtRunner Loop and direct errors | 2 | 40.306 [33.423-47.189] | 975.742 [975.687-975.796] | 49.500 [49-50] | 3.170 [3.116-3.224] |

## Comparison

Seven-column Stream versus seven-column Load:

- Replay improved 19.2 percent.
- Process time improved 18.4 percent.
- Total allocation fell 63.5 percent.
- Heap fell 75.7 percent.
- Garbage collections fell 4.2 percent.

Seven-column Stream remains 3.31 times slower than the two-column stream baseline.

Six-column Stream versus seven-column Stream:

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
