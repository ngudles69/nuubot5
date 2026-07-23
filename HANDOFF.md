# Handoff

Last updated: 2026-07-24

## Focus

BtRunner review complete.

## Current status

- BtRunner stops Clock, Reader, then Runtime.
- BtRunner Runtime callback is named `runtimeRun`.
- TickClock is constructed without timer policy.
- BtRunner registers the configured 10-second Runtime callback.
- TickClock invokes the callback synchronously from replay time.
- Callback errors return through TickClock to BtRunner.
- Runtime stop intent sets BtRunner-owned `stopRequested`.
- The old exposed `due` branch and `passes_due` clock field are removed.
- Runtime uses `Run` for one timer-driven control operation.
- TickClock invokes BtRunner's registered Runtime `Run` callback.
- The 10-second gate currently runs Risk and BotCycle logic; reconciliation remains unimplemented.
- `ohlcv.Load` retains complete-range materialization.
- `ohlcv.Open` streams six validated columns.
- `ohlcv.Load` consumes `ohlcv.Open`; both use one decoder.
- Replay streams through `ohlcv.Open`.
- `rtest.sh` aggregates heap, total allocation, GC runs, and GC pause.
- `close_time_us` and stored `EndMS` are removed.
- The next `StartMS` proves the previous bar closed.
- Six-column Stream passed full tests, 2x, and 500x.
- Arrow batch size matches the standard 122,880-row Parquet group.
- `wiki/design/packages/ohlcv.md` records the accepted 500x result and tradeoff.
- Confirmed implementation facts now update their owning design page.
- BtRunner messages no longer duplicate component, event, or status fields.
- BtRunner logger parameters use the established `log` name.
- Every Go `*logging.Logger` parameter uses the established `log` name.
- BtRunner logs initialization only after every owned child and replay proof succeeds.
- BtRunner keeps replay range selection, validation, and proof conversion together.
- Command calls BtRunner Init, Start, Loop, and Stop explicitly.
- Package-level lifecycle wrapper is removed.
- Main owns one BtRunner value; its Init method fills that value and returns only an error.
- BtRunner Init binds its logger before fallible initialization.
- BtRunner uses Loop because the method owns repeated replay iteration.
- Components construct lifecycle errors directly with `fmt.Errorf`.
- `internal/toolkit/errors` and the one-line `StateError` wrapper are removed.
- Logging inlines the Bot filename inside `OpenBot`; the one-use `BotLog` wrapper is removed.
- Command embeds elapsed duration in its completed message.
- Command and BtRunner construct complete strings before calling `log`.
- Logging writes `YYYY-MMM-DD HH:MM:SS [LEVEL] message`.
- Logging right-aligns levels to a minimum width of five.
- Logging supports Debug, Info, Warning, Error, and Critical.
- Senders own complete message content and value formatting.
- Logging owns destination, timestamp, level, record format, and line writing.
- Bot filenames own identity; Logger binds no hidden fields.
- `rtest.sh` identifies BtRunner completion from its message.
- Simple BtRunner logs passed full tests and 2/2 fresh processes.
- Final BtRunner review cleanup passed full tests and 2/2 fresh processes.
- `internal/ohlcv` is the only Parquet decoder.
- `wiki/PERFORMANCE.md` records commit-linked benchmark history.
- Nuubot4 core ownership is ported into compact Go designs and one cascading implementation.
- Signaler, Executor, and Risk are permanent factories for many concrete implementations.
- Approved baseline remains Macross, Observer stop-loss, and BalancedRisk.
- Setup owns configuration and Bot admission.
- Runtime initialization owns Signaler OHLCV loading and preparation.
- Runtime returns stop intent; BtRunner owns final cascading teardown.
- `internal` contains 21 Go packages.
- `internal/toolkit` groups `clock` and `logging`.
- `internal/toolkit` is not a Go package.
- `internal/common`, `internal/clock`, and `internal/logging` are removed.
- `account`, `fill`, `ledger`, `meta`, `order`, `simulator`, and `trade` reserve approved package names only.
- `wiki/design/packages` contains exactly one page per Go package.
- `wiki/design/concepts` contains 27 process, program, venue, type, and cross-package pages.
- `wiki/DESIGN.md` indexes both categories.
- `main.go` now only parses input, runs BtRunner, and logs the final result with duration.
- The command owns BtRunner construction, lifecycle, and shutdown.
- `internal/toolkit/logging` owns append-only Server and Bot log opening.
- Pre-identity failures use `server.log`.
- Identified Bot failures use only `bot_<sweep_id>_<bot_id>.log`.
- Console output occurs only when `server.log` cannot open.
- All 36 Go files retain all three numbered section comments.
- Go module and internal imports use `nuubot`, not repository-version name `nuubot5`.
- `rtest.sh` reads each run from the append-only Bot log and reports automatic timing.
- All 49 `wiki/design/**` pages use the standard Status, Covers, Purpose header.
- `wiki/design/concepts/nuubot-btrunner.md` owns the command design.
- `wiki/design/packages/btrunner.md` covers only `internal/btrunner/btrunner.go`.
- BtRunner review commit and push are authorized.

## Active agents

- Root only.

## Blockers

- None.

## Files changed

- `internal/**`: package reservations and toolkit reorganization.
- `cmd/nuubot-btrunner/main.go`: updated logging import.
- `wiki/design/packages/**`: one page per Go package.
- `wiki/design/concepts/**`: non-package design pages.
- `wiki/DESIGN.md`: package and concept index.
- `wiki/ARCHITECTURE.md`, `wiki/coding/**`, and `wiki/logic/btrunner.md`: aligned paths.
- `AGENTS.md`: approved package-reservation exception.
- `go.mod` and Go imports: module renamed from `nuubot5` to `nuubot`.
- `rtest.sh`: Bot-log validation plus suite and average timing.
- `AGENTS.md`: general `Nx` test command and reporting rule.
- `wiki/coding/STYLE.md`: explicit `var` and `=` declaration standard.
- `wiki/design/**`: standardized headers and command/package ownership split.
- `wiki/DESIGN.md`: current-source and Nuubot4 canonical-source distinction.
- Eight core package designs: compact Nuubot4 intent and lifecycle flows.
- Six trading-domain designs: compact Nuubot3 and Nuutrader6 intent.
- `internal/setup`, `internal/btrunner`, and `internal/runtime`: cascading initialization and teardown.
- Signaler, Executor, and Risk factories: hardcut constructor name `Create`.

## Proof

- Final BtRunner closeout passed `go test -tags noasm ./...`.
- Final BtRunner closeout passed 2/2 fresh processes.
- Both runs served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Both runs produced 55 Signals, 18 cycles, 17 stop-loss exits, and one end-date exit.
- Suite duration was 4,037 ms; process average was 1,769 ms.
- Replay averaged 1,115 ms; range was 1,115-1,116 ms.
- Bot log proves Clock, Reader, Runtime, then BtRunner stop order.
- Final result log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T180537Z.log`.
- Final BtRunner closeout passed `git diff --check`.
- `runtimeRun` callback rename passed `go test -tags noasm ./...`.
- `runtimeRun` callback rename passed `git diff --check`.
- BtRunner source and owning design contain zero stale `runRuntime` or `run_runtime` references.
- Registered timer callback passed its focused cadence and error-propagation test.
- Registered timer callback passed `go test -tags noasm ./...`.
- Registered timer callback passed 2/2 fresh processes.
- Both runs served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Both runs produced 55 Signals, 18 cycles, 17 stop-loss exits, and one end-date exit.
- Replay averaged 1,115 ms; range was 1,112-1,118 ms.
- Result log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T172003Z.log`.
- No manual `due`, interval-bearing `clock.New`, or `passes_due` path remains.
- Registered timer callback passed `git diff --check`.
- Runtime `Run` hardcut passed focused Runtime and BtRunner tests.
- Runtime `Run` hardcut passed `go test -tags noasm ./...`.
- Runtime `Run` hardcut passed `git diff --check`.
- `Runtime.Pass`, `runtime.Pass`, and Runtime-specific pass error text have zero stale references.
- Direct-error cleanup passed `go test -tags noasm ./...`.
- Direct-error cleanup passed 2/2 fresh processes.
- Replay averaged 1,115 ms; range was 1,114-1,117 ms.
- Total allocation averaged 975.742 MB; GC averaged 49.5 runs.
- Result log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T163450Z.log`.
- Internal packages and package pages both total 21.
- `StateError`, `nuuerrors`, `toolkit/errors`, and `BotLog` have zero stale references.
- BtRunner Loop rename passed `go test -tags noasm ./...`.
- BtRunner Loop rename passed `git diff --check`.
- BtRunner logger-first change passed `go test -tags noasm ./internal/btrunner`.
- BtRunner logger-first change passed `git diff --check`.
- Exact-format Logger passed its focused format test.
- Exact-format Logger passed full `go test -tags noasm ./...`.
- Exact-format Logger passed 2/2 with exact replay proof.
- Replay averaged 1,117 ms; range was 1,114-1,120 ms.
- Completion message included elapsed duration.
- Result log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T161651Z.log`.
- Simple BtRunner logs passed 2/2 with exact replay proof.
- Replay averaged 1,112 ms; range was 1,109-1,116 ms.
- Result log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T154011Z.log`.
- Final review cleanup replay averaged 1,111 ms; range was 1,110-1,113 ms.
- Final result log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260723T154808Z.log`.
- Final `go test -tags noasm ./...` and `bash -n rtest.sh` passed.
- Row-group-sized batch passed 2/2 and 500/500 fresh processes.
- Batch 500x suite duration was 728,463 ms.
- Process average was 1,219 ms; range was 1,177-1,530 ms.
- Replay average was 1,134 ms; range was 1,098-1,445 ms.
- Heap averaged 31.792 MB; total allocation averaged 975.697 MB.
- GC averaged 49.880 runs and 5.011 ms pause.
- Every batch run produced exact Reader, Runtime, and BtRunner proof.
- The batch log contains zero failures, ERROR levels, or error fields.
- Batch stability log: `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T145016Z.log`.
- Larger batches cut allocation 26.1 percent and GC runs 21.7 percent.
- Replay slowed 0.9 percent; process time slowed 1.2 percent.
- Final `go test -tags noasm ./...` passed.
- `bash -n rtest.sh` passed.
- All 36 Go files retain all three section comments.
- All 21 internal packages match 21 package design pages.
- No stale `EndMS` or `close_time_us` references remain.
- `git diff --check` passed.
- Six-column Stream passed 2/2 and 500/500 fresh processes.
- Six-column 500x suite duration was 706,950 ms.
- Process average was 1,204 ms; range was 1,165-1,475 ms.
- Replay average was 1,124 ms; range was 1,090-1,338 ms.
- Heap averaged 28.604 MB; total allocation averaged 1,321.159 MB.
- GC averaged 63.722 runs and 5.090 ms pause.
- Every run produced exact Reader, Runtime, and BtRunner proof.
- The six-column log contains zero failures, ERROR levels, or error fields.
- Stability log: `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T143429Z.log`.
- Six columns improved replay 11.1 percent and allocation 14.7 percent.
- Six columns remain 2.94 times slower than the two-column baseline.
- Seven-column Load passed 2/2 and 500/500 fresh processes.
- Seven-column Stream passed full `go test -tags noasm ./...`.
- Seven-column Stream passed 2/2 and 500/500 fresh processes.
- Stream 500x suite duration was 766,287 ms.
- Stream process average was 1,345 ms; range was 1,329-1,383 ms.
- Stream replay average was 1,265 ms; range was 1,245-1,300 ms.
- Stream heap averaged 30.733 MB; total allocation averaged 1,549.676 MB.
- Stream GC averaged 66.202 runs and 5.072 ms pause.
- Stream improved replay 19.2 percent and allocation 63.5 percent against Load.
- Every Stream run produced exact ticks, passes, signals, cycles, and exits.
- The Stream log contains zero failures, ERROR levels, or error fields.
- Stream stability log: `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T124647Z.log`.
- Stream remains 3.31 times slower than the two-column baseline.
- `go list -tags noasm ./...` passed.
- `go test -tags noasm ./...` passed.
- Focused BtRunner command tests passed.
- Two invalid starts appended exactly two `server.log` entries.
- An identified failure appended zero Server entries and one Bot entry.
- Invalid and identified failures emitted zero console lines after logger creation.
- All 21 Go package names match 21 package page names.
- All 48 prior and new design pages are present after reorganization.
- All local wiki links resolve.
- No stale `internal/common`, `internal/clock`, or `internal/logging` references remain.
- `git diff --check` passed.
- `./rtest.sh 10 6 9` passed 10/10 fresh processes.
- Suite duration was 5,676 ms; average process duration was 459 ms.
- Average replay duration was 379 ms.
- Result log: `workspace/logs/nuubot5-rtest-s6-b9-10-20260723T095922Z.log`.
- All 49 design pages have exactly one ordered Status, Covers, Purpose header.
- All concrete Covers targets resolve.
- All 21 package page names match all 21 Go package names.
- No design page uses YAML metadata.
- No Nuubot5 path remains mislabeled as a Canonical Source.
- Fresh full `go test -tags noasm ./...` passed.
- `./rtest.sh 2 6 9` passed 2/2 with exact replay proof.
- Latest `./rtest.sh 10 6 9` passed 10/10 with exact replay proof.
- Latest 10x suite duration was 5,662 ms; average process duration was 454 ms.
- Average replay duration was 375 ms; range was 371-382 ms.
- Every run produced 7,948,800 ticks, 794,880 passes, 55 Signals, 18 cycles, 17 stop-loss exits, and one end-date exit.
- Latest result log: `workspace/logs/nuubot5-rtest-s6-b9-10-20260723T105957Z.log`.
- `./rtest.sh 500 6 9` passed 500/500 fresh processes.
- 500x suite duration was 291,614 ms; average process duration was 463 ms.
- Average replay duration was 382 ms; process range was 444-531 ms.
- Result scan found zero failures, ERROR levels, or error fields.
- All 500 runs recorded exact successful Parquet Reader, Runtime, and BtRunner proof.
- Stability log: `workspace/logs/nuubot5-rtest-s6-b9-500-20260723T110542Z.log`.

## Decisions

- `Run` executes one operation; `Loop` owns repeated iteration until its stop condition.
- Components create ordinary errors directly; custom error-construction helpers are prohibited.
- Package pages live in `wiki/design/packages`.
- Concept pages live in `wiki/design/concepts`.
- Toolkit is a grouping directory, not a package.
- Toolkit children remain independent and domain-neutral.
- Explicitly approved reservation files contain only package documentation and declaration.
- Components construct ordinary errors directly with `fmt.Errorf`.
- Millisecond `Duration` belongs to `internal/toolkit/clock`.
- `main.go` has exactly three responsibilities: parse, run, and log the timed result.
- Log fields exist only when removing them would hide useful information.
- `duration` remains on BtRunner Loop results.
- `error` remains on failures.
- Boundary messages name `parseInput()`, `logging.OpenBot()`, or the failed BtRunner lifecycle phase.
- New and changed Go prefers `var` declarations and later `=` assignments.
- `:=` remains allowed only when `var` or `=` is impossible or materially less clear.
- Setup, BtRunner, and Runtime form one cascading lifecycle.
- Signaler, Executor, and Risk remain configuration-selected factories.
- Runtime owns Signaler preparation; BtRunner does not know OHLCV or Signaler.
- Observer retains canonical one-percent adverse-move exit behavior.

## Not run

- Registered timer callback did not run a larger stability suite.
- Runtime `Run` hardcut did not run a real replay because behavior did not change.

## Next action

Review `internal/runtime/runtime.go`.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
