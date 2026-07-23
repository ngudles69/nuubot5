# Handoff

Last updated: 2026-07-23

## Focus

Evaluate whether the Go BtRunner is simple, stable, and within 2x of Rust.

## Current status

- Go BtRunner replays 7,948,800 Parquet ticks through the full runtime path.
- Canonical builds use `-tags noasm`; the optimized decoder produced one corrupt timestamp at run 183.
- `rtest.sh` starts a fresh OS process per run with no inter-run delay.
- Every successful run reports 55 signals, 18 cycles, 17 stop-loss exits, and 1 end-date exit.
- No active agents. No blockers.

## Files changed

- `rtest.sh`: removed the obsolete one-second delay.
- `HANDOFF.md`: created this restart state.

## Proof

- 2/2 passed: process average 476 ms; replay average 377 ms.
- 200/200 passed: process average 457 ms; replay average 382 ms.
- 1000/1000 passed: process average 445 ms (435-638); replay average 371 ms (358-565).
- The 1000x audit found zero failure markers, zero incorrect statistics, and zero stderr.
- Full log: `workspace/logs/nuubot5-rtest-s6-b9-1000-20260723T041701Z.log`.

## Decisions

- Prefer the stable `noasm` build. It remains faster than the accepted Rust replay average.
- Keep timestamp validation; it prevented corrupt decoder output from entering Runtime.

## Not run

- DuckDB has only verified the source timestamp. It has not been benchmarked as a loader.

## Next action

Benchmark DuckDB against the current Arrow `noasm` loader on the same files and full BtRunner path before considering any loader change.
