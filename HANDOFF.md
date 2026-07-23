# Handoff

Last updated: 2026-07-23

## Focus

Finish the clean Go BtRunner baseline without changing replay behavior.

## Current status

- Go BtRunner replays 7,948,800 Parquet ticks through the full runtime path.
- Canonical builds use `-tags noasm`; the optimized decoder produced one corrupt timestamp at run 183.
- `rtest.sh` starts a fresh OS process per run with no inter-run delay.
- Every successful run reports 55 signals, 18 cycles, 17 stop-loss exits, and 1 end-date exit.
- Git baseline `e1a83f7` is pushed.
- The existing replay path now uses injected standard `log/slog` with structured fields.
- The custom `common.Logger` was removed.
- Runtime, BotCycle, and Executor now use `Pass`. `MainLoop` has no compatibility alias.
- Ownership-boundary errors use lowercase context and `%w`.
- Arrow and Parquet errors are translated where no `errors.Is` or `errors.As` contract exists.
- Sweep and Bot identity is bound once at the command logger boundary and inherited by every child.
- Touched source follows mandatory lifecycle ordering and source sections.
- Coding rules now describe current source truth instead of completed migration work.
- `wiki/design/**` is excluded.

## Files changed

- `rtest.sh`: removed the obsolete one-second delay.
- `rtest.sh`: parses structured `slog` fields and validates exact replay statistics.
- `cmd/nuubot-btrunner/main.go`: creates one logger and logs terminal errors once.
- `internal/logging/logging.go`: owns standard `slog` construction.
- `internal/common/common.go`: retains only shared state-error and duration mechanics.
- `internal/datastore/sweep.go`: uses lowercase errors and the mandatory source layout.
- Replay components now receive `*slog.Logger` and emit structured terminal proof.
- Runtime, BotCycle, and Executor use `Pass`.
- `wiki/coding/STYLE.md` and `wiki/coding/RULES.md`: removed completed migration drift.
- `HANDOFF.md`: records cleanup state and proof.

## Proof

- 2/2 passed: process average 476 ms; replay average 377 ms.
- 200/200 passed: process average 457 ms; replay average 382 ms.
- 1000/1000 passed: process average 445 ms (435-638); replay average 371 ms (358-565).
- The 1000x audit found zero failure markers, zero incorrect statistics, and zero stderr.
- Full log: `workspace/logs/nuubot5-rtest-s6-b9-1000-20260723T041701Z.log`.
- `gofmt` completed with Go 1.26.5.
- `go test -tags noasm ./...` passed.
- `go vet -tags noasm ./...` passed.
- `go build -tags noasm ./cmd/nuubot-btrunner` passed.
- One direct Sweep 6 Bot 9 smoke replay exited zero.
- Smoke replay: 7,948,800 ticks, 794,880 passes, 55 signals, 18 cycles, 17 stop-loss exits, and 1 end-date exit.
- Smoke replay time: 375 ms.
- Searches found no `common.Logger`, `NewLogger`, custom `Logger.Info`, or `MainLoop`.
- Searches found no uppercase errors in replay or Sweep loading.
- Final focused search found no Arrow or Parquet `%w` wrapping in bars or replay.
- Smoke logs prove inherited `sweep_id=6` and `bot_id=9` on every child completion line.
- `bash -n rtest.sh` passed.
- Canonical `bash rtest.sh 1 6 9` passed 1/1.
- Final script proof: process 1,629 ms; replay 377 ms; timing summary recorded.
- Final proof log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260723T070604Z.log`.
- The ignored root `nuubot-btrunner.exe` artifact was removed.

## Decisions

- Prefer the stable `noasm` build. It remains faster than the accepted Rust replay average.
- Keep timestamp validation; it prevented corrupt decoder output from entering Runtime.

## Not run

- DuckDB has only verified the source timestamp. It has not been benchmarked as a loader.
- The cleanup did not rerun 200x or 1000x stability tests.

## Next action

None. The authorized baseline cleanup is complete.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
