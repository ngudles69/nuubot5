# Function Profiler Package

Status: Implemented.
Covers: `internal/fprof/**`
Purpose: Generate temporary function instrumentation, execute A/B/C variants, verify behavior, and render exact function profiles.

## Ownership

`cmd/nuubot-fprof` owns one profiling session.

The root `internal/fprof` package owns orchestration, AST transformation, overlay generation, execution, comparison, and reporting.

`internal/fprof/runtime` owns dependency-free instrumentation used by generated application source.

## Program Flow

```text
Run
  validate profile request
  create isolated profile session
  generate instrumented build overlay
  build A, B, and C binaries
  run A, B, and C sequentially
  verify equivalent behavior
  calculate runtime overhead and function profile
  write final JSON and text reports
```

## Instrumentation

Generated source adds one deferred exact entry and exit handler to each selected function.

Tracked application source remains unchanged.

The generated main function also defers final profile publication.

B records exact calls without timestamps.

C records exact calls with nested flat and cumulative timing.

B and C call counts must match exactly.

## Report

The report contains:

```text
calls
flat
flat%
cum
cum%
average flat
average cumulative
A, B, and C runtime
B-A, C-B, and C-A overhead
GC and allocation observations
```

The text report sorts functions by flat time descending.

The JSON report retains every observed function.

## Boundaries

- Standard Go AST, formatter, overlay, JSON, process, runtime, and timing packages only.
- Canonical `-tags noasm` builds.
- Existing pprof remains disabled during A, B, and C.
- No CGO, unsafe code, assembly, runtime internals, or source rewriting.
- General overlapping instrumented goroutines remain unsupported.

## Proof

Observer Sweep 6 Bot 9 passed A, B, and C with matching stable behavior.

The proof instrumented 53 source files and 544 functions.

See [Function Profiler](../fprof.md) for permanent semantics and exact proof.
