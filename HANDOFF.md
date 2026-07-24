# Handoff

Last updated: 2026-07-24

## Focus

Prepare the next approved Nuubot5 task.

## Current status

- Stock `github.com/sonirico/go-hyperliquid` failed the Nuubot5 SDK audit.
- The audited clone is `D:\rust\references\go-hyperliquid` at `6cb9ba8`.
- The working fork at `D:\rust\go-hyperliquid` remains reference-only.
- Fork baseline setup is complete.
- The module is `github.com/ngudles69/go-hyperliquid`.
- `origin` owns the fork; `upstream` tracks Sonirico.
- Explicit perpetual Meta retrieval is already implemented as `Info.Meta`.
- Live mainnet perpetual Meta retrieval is proven.
- Hyperliquid design-tree documentation is complete.
- `wiki/design/hyperliquid.md` owns the boundary.
- REST, WebSocket, and Meta details have focused child pages.
- `wiki/DESIGN.md` explains what, where, why, how, in, out, and status.
- Official Hyperliquid API documentation is authoritative.
- Sonirico's Go client is the secondary implementation reference.
- Python `async_hyperliquid` is the third known-working reference.
- Nuubot rewrites from the official API and targets no library parity.
- Hyperliquid decision-record documentation is complete.
- Full-worktree verification passed.
- Full-worktree publication is authorized.
- The stock SDK contains unsafe error handling, WebSocket deadlock paths,
  missing TP/SL grouping, insufficient signing proof, and dependency bloat.
- The audit rejects direct stock-module adoption.
- Setup contains the approved deferred Meta-admission comment.
- Meta refresh and minimum-notional design is recorded.
- Meta belongs inside the future NuubotDB.
- Every Setup caller will check Meta freshness.
- Empty or 24-hour-old Meta triggers caller-driven refresh.
- Hyperliquid's minimum order notional is USDC 10.
- Nuubot will configure USDC 11 to buffer price and size rounding.
- Setup rewrite is complete and partially user reviewed.
- Setup exposes one coordinator function returning one Context.
- Config and credentials load from separate owning files.
- Existing config validation remains; new detailed validation is deferred.
- Credentials will receive TOML decoding only; account validation is deferred.
- Datastore behavior and shape remain unchanged.
- Shared WebSocket ownership remains TBD.
- `workspace/` is the only approved mutable filesystem root and future Docker mount.
- Shared config, databases, logs, and runtime data belong under `workspace/`.
- Main datastore files belong directly under `workspace/db/`.
- Per-Bot Sweep result databases belong under
  `workspace/db/sweeps/sweep_<sweep_id>/bot_<bot_id>.db`.
- Sweep workers write only their owned result database.
- Main datastore technology, schema, and filename remain unresolved.
- `workspace/config/config.toml` is approved for Git and must contain no secrets.
- `workspace/config/credentials.toml` contains secrets and must remain ignored.
- PocketBase remains unresolved and must not shape the finalized datastore yet.
- `HANDOFF.md` owns current DONE, TODO, and deferred work buckets.
- Tranche 1 is the active user-review TODO.
- Tranches 2 and 3 are deferred implementation buckets.
- `wiki/DESIGN.md` owns the user review tree and To Code checklist.
- Checklist states use DONE, PARTIAL, NOT REVIEWED, and TO CODE.
- DONE and PARTIAL rows record `2026-07-24 12:55:22 +08:00`.
- The To Code list is workflow state, not package implementation status.
- Observer increments IngestBBO and OnBBO counters silently.
- Observer logs both summarized counts only during Stop.
- Runtime, BotCycle, and Executor implement the approved IngestBBO route.
- Runtime's BBO method is named `IngestBBO`.
- Observer retains its existing stop-loss termination behavior.
- The rejected 20-BBO completion rule is not part of the design.
- Runtime, BotCycle, Executor, Account, Venue, and Simulator designs link the
  canonical IngestBBO concept.
- Executor design owns the factory, common concrete structure, current Observer
  template, future TradeExecutor template, and complex-Executor page rule.
- `IngestBBO` is approved only for driving Simulator matching and fills.
- `OnBBO` remains the separate Executor BBO policy callback.
- Live Venue `IngestBBO` is a no-op.
- Simulator outcomes reach Executor state only through reconciliation.
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

## Work tracker

Last updated: 2026-07-24 13:33:22 +08:00

### DONE

- nuubot-btrunner review.
- BtRunner review.
- Runtime review.
- WallClock implementation.
- Setup coordinator rewrite.

### TODO — Tranche 1

1. Review Signaler, Macross, and RSI.
2. Review Risk and BalancedRisk.
3. Continue Setup review: validation, datastore shape, and shared WebSocket ownership.
4. Select the SDK, including Hyperliquid SDK selection.

### DEFERRED — Tranche 2

- Account.
- Ledger.
- Trade.
- Order.
- Fill.
- Simulator.
- TradeExecutor.

### DEFERRED — Tranche 3

- Runner.
- Detailed Risk implementation.

### DEFERRED — Unassigned

- Simulator parity.
- Meta.
- PocketBase consideration: do not add yet.

## Active agents

- Root only.

## Blockers

- None.

## Files changed

- `internal/setup/setup.go`: marks the future Meta-admission location without
  adding disabled code or fake types.
- `wiki/design/packages/setup.md`: records the deferred Meta integration.
- `workspace/config/config.toml`: configures Hyperliquid minimum order notional
  at USDC 11.
- `internal/config/config.go`: admits the Hyperliquid policy section.
- `internal/config/config_test.go`: proves the configured USDC 11 floor.
- `wiki/design/packages/meta.md`: NuubotDB ownership, Setup-driven 24-hour
  refresh, normalized constraints, retirement, concurrency, and rounding buffer.
- `wiki/design/packages/config.md`: records Hyperliquid policy ownership.
- `.gitignore`: tracks shared workspace config while retaining credential ignores.
- `workspace/config/config.toml`: prominent no-secrets warning.
- `internal/config/config.go`: shared config types, loading, existing validation,
  and shared-data path admission.
- `internal/config/credentials.go`: typed credentials TOML decoding without
  semantic account validation.
- `internal/config/config_test.go`: idempotent loading and malformed credentials proof.
- `internal/setup/setup.go`: one Setup coordinator returning one Context.
- `internal/btrunner/btrunner.go`: calls `setup.Setup`.
- `wiki/design/packages/setup.md`: partially reviewed Setup contract.
- `wiki/design/packages/config.md`: config and credentials ownership.
- `wiki/design/concepts/filesystem.md`: mutable root, directory ownership,
  Docker mount, database layout, secret rules, and current drift.
- `wiki/design/packages/datastore.md`: main-store and per-Bot result-store expectations.
- `wiki/DESIGN.md`: Filesystem concept index entry.
- `HANDOFF.md`: filesystem decision and direct proof.
- `wiki/DESIGN.md`: canonical user review tree, timestamps, and To Code checklist.
- `HANDOFF.md`: review-tracker task and proof.
- `internal/btrunner/btrunner.go`: calls `Runtime.IngestBBO`.
- `internal/runtime/runtime.go`: routes BBO through BotCycle IngestBBO before OnBBO.
- `internal/botcycle/botcycle.go`: routes IngestBBO through active Executors.
- `internal/executor/executor.go`: requires Executor IngestBBO.
- `internal/executor/observer.go`: counts IngestBBO and OnBBO calls and reports
  both counts during Stop.
- `internal/executor/observer_test.go`: proves counters, retained stop loss, and
  summarized stop logging.
- `wiki/design/concepts/ingestbbo.md`: Simulator-only matching route and hard
  boundary from Executor `OnBBO`.
- Runtime, BotCycle, Executor, Account, Venue, and Simulator designs: added their
  direct IngestBBO responsibility and canonical concept link.
- `wiki/DESIGN.md`: IngestBBO concept index entry.
- `HANDOFF.md`: active IngestBBO design task and proof.
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

- Full uncached tests passed across 25 packages with zero failures.
- Eight packages ran tests; 17 packages reported no test files.
- Full tests used `CGO_ENABLED=0`, `-tags noasm`, and took 3.165 seconds.
- Full `go vet -tags noasm ./...` passed in 0.486 seconds.
- Modified Go files passed `gofmt`.
- No credential file is tracked.
- Full-worktree `git diff --check` passed.
- Hyperliquid local Markdown links passed validation.
- No stale Hyperliquid concept link remains.
- Hyperliquid design pages contain no port or selective-port wording.
- The recorded local Python `async_hyperliquid` reference exists.
- Hyperliquid decision terms passed direct `rg` verification.
- Hyperliquid decision-record prose contains no line above 25 words.
- Hyperliquid documentation passed `git diff --check`.
- Live perpetual Meta returned 232 assets and seven margin tables.
- Live BTC Meta reported five size decimals and 40 maximum leverage.
- Live ETH Meta reported four size decimals and 25 maximum leverage.
- Focused `TestMeta` passed with `CGO_ENABLED=0` and `-tags noasm`.
- Working fork compilation passed with `CGO_ENABLED=0` and `-tags noasm`.
- Working fork root tests passed with `CGO_ENABLED=0` and `-tags noasm`.
- Working fork `go vet -tags noasm ./...` passed.
- Working fork `git diff --check` passed.
- Hyperliquid focused tests passed uncached with `CGO_ENABLED=0` and `-tags noasm`.
- Hyperliquid root tests and root `go vet -tags noasm .` passed.
- Hyperliquid examples compile, but full `go test -tags noasm ./...` timed out
  because live WebSocket examples use unbounded contexts.
- The stock SDK compiles 101 non-standard packages and 19,263 generated lines.
- `-tags noasm` still leaves third-party assembly through the dependency graph.
- Audit report:
  `D:\rust\references\go-hyperliquid\audits\07-24-go-hyperliquid-v1.md`.
- Setup contains the approved four-step Meta-admission comment.
- Setup design records the same deferred behavior.
- `git diff --check` passed after the comment-only update.
- Tests and replay were not run because runtime behavior did not change.
- Focused Config tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed after the Hyperliquid config addition.
- `git diff --check` passed.
- Config proof reads `hyperliquid.min_order_notional_usdc = 11`.
- Meta design distinguishes Nuutrader6's empty-only load from Nuubot5's
  caller-driven 24-hour refresh.
- Replay was not run because no runtime consumes the new Meta policy yet.
- Setup has exactly one function: `Setup`.
- Root `config.toml` no longer exists.
- Setup loads `workspace/config/config.toml` and
  `workspace/config/credentials.toml`.
- Datastore source and behavior were unchanged.
- Config and credentials idempotence tests passed.
- Malformed credentials TOML test passed.
- Focused Config, Setup, and BtRunner tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed.
- Full `go vet -tags noasm ./...` passed.
- `gofmt -d` and `git diff --check` passed.
- Fresh `./rtest.sh 1 6 9` passed 1/1.
- Fresh suite took 3,142 ms; process took 2,867 ms; replay took 1,566 ms.
- Replay served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Replay produced 55 Signals, 18 cycles, and 17 stop-loss exits.
- Replay log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T053308Z.log`.
- Setup review is PARTIAL at `2026-07-24 13:33:22 +08:00`.
- `workspace/config/credentials.toml` is ignored and untracked.
- `workspace/config/config.toml` is no longer ignored and is visible to Git.
- Shared config starts with `NO SECRETS ALLOWED IN THIS FILE`.
- Filesystem and Datastore Markdown links resolve.
- Filesystem documentation passed `git diff --check`.
- `wiki/DESIGN.md` contains all 14 review-tree components and 14 TO CODE rows.
- DONE and PARTIAL rows carry `2026-07-24 12:55:22 +08:00`.
- NOT REVIEWED and TO CODE rows carry no review timestamp.
- PocketBase remains marked as a new consideration that must not be added yet.
- Review-tracker documentation passed `git diff --check`.
- Focused Executor, BotCycle, Runtime, and BtRunner tests passed with `-tags noasm`.
- Full `go test -tags noasm ./...` passed.
- Full `go vet -tags noasm ./...` passed.
- `./rtest.sh 1 6 9` passed 1/1 in 2,918 ms.
- The replay process took 2,650 ms and replay took 1,444 ms.
- Replay served 7,948,800 ticks and triggered 794,880 Runtime runs.
- Replay produced 55 Signals, 18 cycles, and 17 stop-loss exits.
- All 18 latest Observer stop lines contain equal IngestBBO and OnBBO counts.
- Cycle 1 reported `ingest_bbo_count=22272 on_bbo_count=22272`.
- Cycle 18 reported `ingest_bbo_count=4676400 on_bbo_count=4676400`.
- Observer summaries are in `workspace/logs/bot_6_9.log`.
- Replay result: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T044530Z.log`.
- `wiki/design/concepts/ingestbbo.md` exists and its `wiki/DESIGN.md` link resolves.
- The page defines non-overlapping `IngestBBO` and `OnBBO` responsibilities.
- Every IngestBBO owner design contains its direct call and canonical concept link.
- All Markdown file links in the changed design pages resolve.
- IngestBBO documentation passed `git diff --check`.
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

- Go tests and replay were not run because this change only adds documentation,
  a Git tracking rule, and TOML comments.
- Replay was not run because `AssessStop` is a compile-time-only method rename.
- BtRunner replay was not run because Runtime construction behavior did not change.
- BtRunner replay was not run because only comments changed.
- Go tests were not run because only Markdown changed.
- Runtime `Run` hardcut did not run a real replay because behavior did not change.

## Next action

Clone the Python reference or implement internal Meta only after the user selects the next scope.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
