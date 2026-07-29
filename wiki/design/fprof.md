# Function Profiler

Status: Implemented and proven with Observer Bot 9.
Purpose: Measure exact function calls and instrumented function time without modifying tracked application source.

## Outcome

`nuubot-fprof` answers two questions:

```text
How many times did each function run?
Where did the instrumented runtime go?
```

The report prioritizes:

```text
calls
flat
flat%
cum
cum%
average flat per call
average cumulative per call
```

## Command

```text
cmd/nuubot-fprof
internal/fprof
```

The first system proof runs Observer Bot 9 through A, B, and C.

## Output

Every profiling session owns one directory below:

```text
workspace/perf/fprofiles/
```

Recommended shape:

```text
workspace/perf/fprofiles/<session>/
  overlays/
  _generated/
  binaries/
  runs/
  profiles/
  report.txt
  report.json
```

No generated file belongs outside the session directory.

## Source Protection

The profiler must not edit tracked application source.

It parses source through standard Go AST packages.

It writes instrumented copies below the session directory.

Go build overlays compile those copies in place of tracked files.

The Git working tree must remain unchanged except for profiler implementation and approved documentation.

Temporary generated source must never be committed.

## Build Variants

```text
A original
  normal source
  no function instrumentation
  pprof disabled

B structural
  exact function entry counting
  entry and exit structure
  no function timestamps
  pprof disabled

C timed
  exact function entry counting
  entry and exit timing
  flat and cumulative aggregation
  pprof disabled
```

A uses the normal build.

B and C use generated source through `go build -overlay`.

All builds use the same Go toolchain, build tags, environment, command arguments, and source revision.

Canonical builds use `-tags noasm`.

## Overhead

The report calculates:

```text
B - A  structural instrumentation overhead
C - B  timing and aggregation overhead
C - A  total observed profiler overhead
```

Each difference includes milliseconds and percentage of its comparison baseline.

Observed overhead may include secondary GC, scheduling, locking, allocation, and cache effects.

The profiler must not claim that overhead affects every function proportionally.

## Function Metrics

`calls` counts exact instrumented function entries.

`cum` is elapsed function time including instrumented descendants.

`flat` is cumulative time minus instrumented descendant time.

`flat%` is function flat time divided by total instrumented root time.

`cum%` is function cumulative time divided by total instrumented root time.

Average values divide each function total by exact call count.

Recursive calls count independently.

The first row totals flat and cumulative time across every instrumented function, including functions hidden by `-top`.

Total calls and averages stay blank because aggregating unrelated functions makes those values misleading.

Total cumulative time and percentage may exceed wall time and 100% because parent and child cumulative times overlap.

The report sorts function rows by flat time descending by default.

## Timing Meaning

Function timing describes the instrumented C run.

It does not reconstruct exact uninstrumented A function time.

Wall-clock runtime includes operating-system, scheduler, cache, and garbage-collector variation.

Exact calls and report arithmetic are deterministic for one deterministic execution path.

## Call Stack

Flat time requires one correctly nested instrumented call stack.

The initial implementation targets Nuubot's synchronous BtBot execution path.

A detected non-LIFO function exit fails the profile instead of publishing corrupt flat time.

General overlapping instrumented goroutines are unsupported and require a separately approved goroutine-safe stack design.

Runtime internals and goroutine-ID extraction are prohibited.

## Existing pprof

Existing `nuubot-backtest -pp` support remains unchanged.

A, B, and C never enable `-pp`.

`pprof` remains the sampled CPU, heap, allocation, block, mutex, and trace tool.

`nuubot-fprof` owns exact calls and deterministic instrumented timing.

The tools are complementary.

## Garbage Collection

A, B, and C capture comparable GC observations.

The report includes:

```text
GC runs
cumulative GC pause
heap allocation
cumulative allocation
```

Nuubot's terminal RunReport is the primary source when available.

Optional `GODEBUG=gctrace=1` raw output may be retained for diagnosis.

GC collection must use the same settings across A, B, and C.

## Behavior Proof

Every variant must exit successfully.

The profiler compares stable RunReport fields across A, B, and C.

Stable proof includes:

```text
identity
status
tick count
Controller runs
Signal packages
cycle counts
Trade, Order, and Fill counts
financial results
telemetry sample count
```

Timing and memory fields are expected to differ.

Behavior mismatch fails the session.

## Final Report

The text report is fixed-width and directly readable.

The JSON report preserves exact machine-readable values.

Example:

```text
Runtime
Variant  Elapsed       Difference   Percent
A         20.000s                -         -
B         40.000s          20.000s   100.00%
C         90.000s          70.000s   350.00%

Overhead
Comparison  Elapsed   Percent
B - A       20.000s   100.00%
C - B       50.000s   125.00%
C - A       70.000s   350.00%

Functions
   Calls       Flat (Flat%)         Cum (Cum%)  Avg Flat   Avg Cum  Function
────────────────────────────────────────────────────────────────────────────
            90.000s (100.00%)  540.000s (600.00%)                       Total
────────────────────────────────────────────────────────────────────────────
 7948800     31.500s (35.00%)     72.000s (80.00%)   3.963µs   9.058µs  simulator.match
  794880      9.000s (10.00%)     18.000s (20.00%)  11.322µs  22.644µs  account.Reconcile
```

## Must Do

- Keep tracked application source unchanged during profiling.
- Use standard Go AST and overlay support.
- Count exact function entries.
- Report flat and cumulative instrumented time.
- Measure A, B, and C independently.
- Verify stable behavior across variants.
- Preserve raw variant output and profile JSON.
- Fail loudly on incomplete, mismatched, or corrupt profiles.

## Must Not Do

- Do not replace or remove existing pprof support.
- Do not enable pprof during A, B, or C.
- Do not use CGO, `unsafe`, assembly, runtime internals, or `go:linkname`.
- Do not rewrite tracked source in place.
- Do not silently ignore AST, build, run, profile, or behavior failures.
- Do not report sampled timing as deterministic timing.
- Do not claim C timing equals A timing.
- Do not publish flat time from an invalid call stack.
- Do not commit generated overlays, binaries, raw runs, or reports.

## Initial Proof

Observer Sweep 6 Bot 9 passed A, B, and C with matching stable RunReport behavior.

```text
Instrumented source files: 53
Instrumented functions: 544

A original:       6.594s
B structural:     9.035s
C timed:         16.307s

B - A:            2.441s   37.02%
C - B:            7.272s   80.48%
C - A:            9.713s  147.30%
```

The timed report records exact high-frequency calls, including:

```text
MarketData.IngestBBO     7,948,800
Controller.onBBO         7,948,800
Replay.Reader.Next       7,948,801
TickClock.Advance        7,948,800
Controller.Run             794,880
```

Proof output:

```text
workspace/perf/fprofiles/s6-b9-20260728T030335Z/report.txt
workspace/perf/fprofiles/s6-b9-20260728T030335Z/report.json
```

The profiling run changed no tracked application source.
