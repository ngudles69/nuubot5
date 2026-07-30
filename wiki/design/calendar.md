# Calendar

Status: Period resolution implemented. Named sessions deferred.
Covers: `internal/toolkit/calendar/**`
Purpose: Resolve reusable calendar labels without inspecting market data.

## Period Labels

```text
YYYY        2024
YYYY-Hn     2024-H1
YYYY-Qn     2024-Q1
YYYY-Mmm    2024-M01
YYYY-Www    2024-W34
YYYY-MM-DD  2024-01-03
```

`ResolvePeriod(label)` returns UTC `start`, exclusive UTC `end`, and `error`.

ISO weeks start Monday. Invalid dates, unavailable ISO weeks, and unknown
labels fail.

Resolution never checks whether market data exists.

## Sweep Use

Each entry uses one mode:

```toml
periods = [
    { start = "2024-01-01", end = "2024-10-01" },
    { label = "2025" },
    { label = "2022-Q1" },
    { label = "2022-Q4" },
]
```

An explicit range requires both dates. A labelled range requires only `label`.

Each entry is one atomic Sweep dimension value. Sweep never permutes `start`
and `end` independently.

## Named Sessions

Future fixed sessions include `Asia`, `London`, and `NewYork`.

The Session API will provide timestamp membership and elapsed `time.Duration`.

Canonical hours and timezone behavior must be defined before implementation.

## Does Not

- Inspect replay files.
- Clip ranges to available data.
- Tag market regimes.
- Analyze results.

Analysis and regime tagging remain separate from Sweep date selection.
