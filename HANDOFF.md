# Handoff

Last updated: 2026-07-24

## Focus

Rework Clock from the Nautilus multi-timer design.

## Current status

- Clock hardcut is complete.
- Program Flow wording completed a repo-wide verb-first hardcut.
- `STYLE.md` and `RULES.md` now require verb-first action steps and source comments.
- Every implemented source action comment starts with a verb.
- Every implemented source action comment has an exact `wiki/design/**` match.
- Full-worktree commit and push are authorized.
- `clock.Create(kind)` returns TickClock or WallClock through one Clock contract.
- Both clocks implement Init, Start, Stop, Err, NowMS, RegisterTimer, Advance,
  NextFireMS, and CancelTimer.
- TickClock advances from admitted replay ticks.
- WallClock starts its own wall-time loop and advances at its next timer.
- Shared Clock state contains only common lifecycle and timer mechanics.
- Shared `clockState` is defined in `clock.go`, its logical owner.
- `timer.go` contains only timer types and timer mechanics.
- The generic Clock `Duration` helper and one-function file are removed.
- BotCycle and Observer calculate their own proof duration without depending on Clock.
- WallClock owns its stop channel, completion channel, loop, waits, and wall-time reads.
- WallClock self-advancement passed ten consecutive focused test runs.
- Both clocks share named multi-timer checking and synchronous callbacks.
- Due callbacks order by scheduled timestamp, then timer name.
- Timer callbacks receive scheduled fire time.
- Optional timer stop time is inclusive.
- BtRunner uses the common Clock contract and one named `runtime` timer.
- Clock rework passed full `go test -tags noasm ./...`.
- Clock rework passed fresh 2/2 and 20/20 replay stability tests.
- Latest 2x log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260724T035208Z.log`.
- Latest 20x log: `workspace/logs/nuubot5-rtest-s6-b9-20-20260724T035217Z.log`.
- Replay `Create` is non-canonical; Replay is one concrete lifecycle owner and must use `Init`.
- Replay keeps `Next` as its precise streaming operation.
- BotCycle `Create` is non-canonical; BotCycle is one concrete lifecycle owner and must use `Init`.
- Executor `Create` remains the configured implementation factory inside BotCycle initialization.
- Replay and BotCycle source and owning design now use `Init`.
- Focused and full `go test -tags noasm ./...` passed.
- `rtest.sh` keeps requested suite runs separate from parsed Runtime runs.
- Replay, Signaler, and BotCycle initialize their owned fields directly.
- Replay's unused failure field is removed.
- Risk uses `AssessStop`, not generic `Run`, for stop-specific assessment.
- Reader exhaustion owns replay completion; Runtime does not stop from replay dates.
- Runtime shutdown closes any active cycle after clearing `r.cycle`.
- Reader-owned replay completion passed fresh `1x` and `20x` proof.
- Signaler is one concrete lifecycle owner; Macross and RSI are internal calculators.
- Signaler `Init` must select its calculator, resolve requirements, load OHLCV,
  calculate and validate Signals, then commit initialized state.
- Runtime must only initialize Signaler and must not orchestrate Signaler internals.
- Runtime child-tree cleanup is approved for Replay, Signaler, BotCycle, Risk,
  and Executor. Command and BtRunner are excluded.
- The `rtest.sh` counter collision and dead Replay failure state are approved fixes.
- Empty `runtime.Create` is removed; BtRunner owns one Runtime value and calls Init directly.
- Every remaining exported and private `Create` path performs real construction,
  validation, selection, resource opening, or required logger binding.
- No remaining `Create` action should move into `Init`.
- Multiline paste into Codex CLI strips leading tabs and doubles newlines.
- `Ctrl+V` and VS Code terminal right-click paste reproduce the same corruption.
- `ss/paste.txt` and `ss/paste.go` contain the same five lines and exact tabs.
- Raw bytes prove one leading tab on outer lines and two on the nested return.
- Active path was Codex CLI 0.144.4 inside VS Code integrated PowerShell.
- Global config had no paste, keymap, newline, or burst-paste override.
- The current Codex manual documents top-level `disable_paste_burst`.
- `C:\Users\PC\.codex\config.toml` now sets `disable_paste_burst = true`.
- Codex must restart before the setting can be tested.
- BtRunner Program Flow comments use verb-first operation wording.
- Comments with error paths match their error operation text.
- User selected PocketBase-owned SQLite instead of PostgreSQL for live, paper,
  and simulator persistence.
- PocketBase will be embedded in `nuubot-server`.
- PocketBase will own HTTP, authentication, administration, API, realtime,
  migrations, SQLite connections, and physical write serialization.
- Nuubot will own domain transactions, constraints, generations, and valid
  state transitions.
- One Server-owned PocketBase process will write its SQLite database.
- Runners will not open the PocketBase database directly.
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
- Executor and Risk are permanent factories for many concrete implementations.
- Signaler is one concrete lifecycle owner with selected calculator implementations.
- Approved baseline remains Macross, Observer stop-loss, and BalancedRisk.
- Setup owns configuration and Bot admission.
- Signaler initialization owns its OHLCV requirements, loading, calculation, and validation.
- Runtime returns stop intent; BtRunner owns final cascading teardown.
- `internal` contains 21 Go packages.
- `internal/toolkit` groups `clock` and `logging`.
- `internal/toolkit` is not a Go package.
- `internal/common`, `internal/clock`, and `internal/logging` are removed.
- `account`, `fill`, `ledger`, `meta`, `order`, `simulator`, and `trade` reserve approved package names only.
- `wiki/design/packages` contains exactly one page per Go package.
- `wiki/design/concepts` contains 28 process, program, venue, type, and cross-package pages.
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

- `wiki/coding/STYLE.md` and `RULES.md`: require verb-first action wording.
- Implemented Go source and owning designs: converted Program Flow wording to verb-first.
- `internal/toolkit/clock/clock.go`: Clock contract and implementation factory.
- `internal/toolkit/clock/tickclock.go`: replay-time Clock implementation.
- `internal/toolkit/clock/wallclock.go`: UTC wall-time Clock implementation.
- `internal/toolkit/clock/timer.go`: shared named multi-timer mechanics.
- `internal/toolkit/clock/clock_test.go`: TickClock and WallClock contract proof.
- `internal/btrunner/btrunner.go`: common Clock creation, initialization,
  registration, start, and advancement.
- Clock, WallClock, Runner, Replay, BtRunner, Project, and Architecture pages:
  aligned the implemented multi-timer contract.
- `internal/replay/replay.go`: replaced `Create` with `Reader.Init`.
- `internal/replay/replay.go`: initializes owned state directly and removes dead failure state.
- `internal/btrunner/btrunner.go`: owns Replay Reader by value and initializes it directly.
- `internal/botcycle/botcycle.go`: replaced `Create` with `Control.Init`.
- `internal/botcycle/botcycle.go`: initializes owned state directly.
- `internal/signaler/signaler.go`: initializes owned rows and Signals directly.
- `rtest.sh`: separates requested suite runs from parsed Runtime runs.
- `internal/risk`, Runtime, and the Risk design: renamed Risk `Run` to `AssessStop`.
- Runtime, replay proof, and owning designs: removed date-driven Runtime stopping.
- `internal/runtime/runtime.go`: initializes one local BotCycle before owning it.
- Replay, BotCycle, Runtime, BtRunner, and architecture pages: aligned lifecycle flow.
- `internal/runtime/runtime.go`: removed the empty Runtime constructor.
- `internal/btrunner/btrunner.go`: owns Runtime by value and initializes it directly.
- `wiki/design/packages/runtime.md`: removed the empty create phase.
- `wiki/design/packages/btrunner.md`: removed the empty Runtime create step.
- `audits/07-24-create-functions-v1.md`: recorded the complete constructor audit.
- `HANDOFF.md`: recorded the active review task and proof.
- `C:\Users\PC\.codex\config.toml`: disabled TUI burst-paste detection.
- `HANDOFF.md`: exact paste diagnosis, fix, restart requirement, and proof step.
- `internal/btrunner/btrunner.go`: verb-first Program Flow comments.
- `wiki/design/packages/btrunner.md`: exact mirrored Program Flow wording.
- `wiki/design/concepts/pocketbase.md`: PocketBase ownership, concurrency,
  boundaries, responsibility split, and proof contract.
- `wiki/PROJECT.md`: PocketBase replaces PostgreSQL for live, paper, and
  simulator persistence.
- `wiki/ARCHITECTURE.md`: Server-owned PocketBase web, API, realtime, and
  writable SQLite boundary.
- `wiki/DESIGN.md`: PocketBase concept index entry.
- `HANDOFF.md`: PocketBase decision, scope, and proof.
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
- Executor and Risk factories use the hardcut constructor name `Create`.

## Proof

- Verb-first audit found zero non-verb source action comments.
- Source/design alignment audit found zero missing exact action matches.
- Verb-first hardcut passed full tests and `go vet -tags noasm ./...`.
- Verb-first hardcut passed `gofmt -d` and `git diff --check`.
- `clockState` exists only in `internal/toolkit/clock/clock.go`.
- `duration.go`, `clock.Duration`, and Clock-only BotCycle/Executor imports are absent.
- Focused Clock, Executor, and BotCycle tests passed with `-tags noasm`.
- Full tests and `go vet -tags noasm ./...` passed.
- `gofmt -d` and `git diff --check` passed.
- WallClock self-advancement test passed 10/10 focused runs.
- Full `go test -tags noasm ./...` passed after WallClock loop ownership.
- Fresh `./rtest.sh 2 6 9` passed 2/2 in 4,633 ms.
- 2x process averaged 2,067 ms; replay averaged 1,427 ms.
- Fresh `./rtest.sh 20 6 9` passed 20/20 in 35,788 ms.
- 20x process averaged 1,542 ms; replay averaged 1,456 ms.
- Every fresh run served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Every fresh run produced 55 Signals, 18 closed cycles, and 17 stop-loss exits.
- Focused Clock and BtRunner tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed after the Clock hardcut.
- Fresh `./rtest.sh 1 6 9` passed 1/1.
- Fresh Clock replay served 7,948,800 ticks and triggered 794,880 Runtime runs.
- It produced 55 Signals, 18 closed cycles, and 17 stop-loss exits.
- Suite duration was 2,695 ms; process duration was 2,441 ms.
- Replay duration was 1,275 ms.
- Result log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T022339Z.log`.
- Clock comment alignment passed focused and full `noasm` tests.
- Fresh `./rtest.sh 2 6 9` passed 2/2.
- The 2x suite took 4,574 ms.
- Process averaged 2,024 ms; range was 1,390-2,659 ms.
- Replay averaged 1,327 ms; range was 1,302-1,352 ms.
- 2x log: `workspace/logs/nuubot5-rtest-s6-b9-2-20260724T033156Z.log`.
- Fresh `./rtest.sh 20 6 9` passed 20/20.
- The 20x suite took 34,197 ms.
- Process averaged 1,452 ms; range was 1,336-2,207 ms.
- Replay averaged 1,362 ms; range was 1,259-2,080 ms.
- Every stability run served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Every stability run produced 55 Signals, 18 closed cycles, and 17 stop-loss exits.
- 20x log: `workspace/logs/nuubot5-rtest-s6-b9-20-20260724T033207Z.log`.
- `git diff --check` passed.
- No old Clock constructor, single-timer Run path, or stale design wording remains.
- Full `go test -tags noasm ./...` passed after removing date-driven Runtime stopping.
- Fresh `./rtest.sh 1 6 9` passed 1/1.
- Fresh 1x suite took 1,434 ms; process took 1,187 ms; replay took 1,108 ms.
- Fresh 1x log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T013013Z.log`.
- `./rtest.sh 20 6 9` passed 20/20 fresh processes.
- The 20x suite took 28,285 ms.
- Process averaged 1,180 ms; range was 1,162-1,197 ms.
- Replay averaged 1,101 ms; range was 1,089-1,119 ms.
- Every run served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Every run produced 55 Signals, 18 closed cycles, and 17 stop-loss exits.
- Every final active cycle closed during Runtime shutdown with `parent_stop`.
- 20x log: `workspace/logs/nuubot5-rtest-s6-b9-20-20260724T013022Z.log`.
- Focused Risk, Runtime, and BtRunner tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed after the `AssessStop` rename.
- No stale Risk `Run()` caller or implementation remains.
- `gofmt -d` and `git diff --check` passed.
- Focused Replay, BotCycle, Signaler, and Runtime tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed after direct-owner cleanup.
- `gofmt -d`, `bash -n rtest.sh`, and `git diff --check` passed.
- `./rtest.sh 1 6 9` passed 1/1 without changing the requested run count.
- Suite duration was 2,597 ms; process duration was 2,341 ms.
- Replay duration was 1,116 ms.
- Result log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T010318Z.log`.
- Focused Replay, BtRunner, BotCycle, and Runtime tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed after both lifecycle conversions.
- `gofmt -d` returned no output for all four changed Go files.
- `bash -n rtest.sh` passed.
- The attempted 1x suite completed 82 exact successful replays before tool timeout.
- Those 82 runs averaged 1,196.30 ms process and 1,102.32 ms replay time.
- Replay range was 1,070-1,330 ms.
- Every completed run produced 7,948,800 ticks, 794,880 Runtime runs, 55 Signals,
  18 cycles, 17 stop-loss exits, and one end-date exit.
- Partial suite log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T004156Z.log`.
- No Replay or BotCycle `Create` call remains in source.
- Focused Runtime and BtRunner tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed.
- `gofmt -d` returned no output.
- `git diff --check` passed.
- No `runtime.Create`, `create runtime`, or `runtime create` reference remains.
- Every remaining exported and private `Create` function and caller was inspected.
- `codex --strict-config doctor --summary` loaded the updated config.
- Codex Doctor reported 17 OK, zero warnings, and zero failures.
- Live paste proof remains pending because Codex must restart first.
- All 22 BtRunner Program Flow steps match source and owning design.
- All 12 error-bearing steps match their error operation text.
- BtRunner comment cleanup passed `go test -tags noasm ./internal/btrunner`.
- BtRunner comment cleanup passed `gofmt -d` and `git diff --check`.
- PocketBase documentation change passed `git diff --check`.
- All `wiki/DESIGN.md` Markdown targets resolve.
- The PocketBase page has the required Status, Covers, and Purpose header.
- No PostgreSQL persistence approval remains.
- The responsibility table separates user and PocketBase ownership.
- No source or dependency changed.
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

- PocketBase replaces PostgreSQL for live, paper, and simulator persistence.
- One embedded PocketBase application in `nuubot-server` owns the writable
  SQLite database.
- PocketBase provides the web server, API framework, authentication,
  administration, realtime, migrations, and physical write serialization.
- Nuubot provides trading behavior, domain DML, transactions, constraints,
  trading UI, analytics, and reports.
- Runners and Bots never open the PocketBase database directly.
- `Run` executes one operation; `Loop` owns repeated iteration until its stop condition.
- Components create ordinary errors directly; custom error-construction helpers are prohibited.
- Package pages live in `wiki/design/packages`.
- Concept pages live in `wiki/design/concepts`.
- Toolkit is a grouping directory, not a package.
- Toolkit children remain independent and domain-neutral.
- Explicitly approved reservation files contain only package documentation and declaration.
- Components construct ordinary errors directly with `fmt.Errorf`.
- Statistics owners calculate their own millisecond durations.
- `main.go` has exactly three responsibilities: parse, run, and log the timed result.
- Log fields exist only when removing them would hide useful information.
- `duration` remains on BtRunner Loop results.
- `error` remains on failures.
- Boundary messages name `parseInput()`, `logging.OpenBot()`, or the failed BtRunner lifecycle phase.
- New and changed Go prefers `var` declarations and later `=` assignments.
- `:=` remains allowed only when `var` or `=` is impossible or materially less clear.
- Setup, BtRunner, and Runtime form one cascading lifecycle.
- Executor and Risk remain configuration-selected factories.
- Signaler remains one concrete lifecycle owner initialized directly by Runtime.
- Runtime owns Signaler preparation; BtRunner does not know OHLCV or Signaler.
- Observer retains canonical one-percent adverse-move exit behavior.

## Not run

- Replay was not run because `AssessStop` is a compile-time-only method rename.
- BtRunner replay was not run because Runtime construction behavior did not change.
- BtRunner replay was not run because only comments changed.
- Go tests were not run because only Markdown changed.
- Runtime `Run` hardcut did not run a real replay because behavior did not change.

## Next action

User continues the code review. Root stands by.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
