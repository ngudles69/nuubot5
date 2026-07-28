# Handoff

Last updated: 2026-07-29

## Active Test Run

- DONE: Observer Bot 9 passed 1 of 1 through `./stest.sh -bot 9`.
- Observer suite total was 9,388 ms; BtBot average was 8,945 ms; replay was 5,406 ms.
- Observer result log is `workspace/logs/nuubot5-stest-s6-b9-1-20260728T162443Z.log`.
- DONE: Trade Bot 13 passed 1 of 1 through `./stest.sh -bot 13`.
- Trade suite total was 23,919 ms; BtBot average was 19,904 ms; replay was 16,407 ms.
- Trade result log is `workspace/logs/nuubot5-stest-s9-b13-1-20260728T162501Z.log`.
- DONE: Grid Bot 15 passed 1 of 1 through `./stest.sh -bot 15`.
- Grid suite total was 73,784 ms; BtBot average was 68,756 ms; replay was 64,846 ms.
- Grid result log is `workspace/logs/nuubot5-stest-s11-b15-1-20260728T162532Z.log`.
- All three runs completed without profiling.

## Commit Proof

- DONE: Committed Simulator audit preparation and test-run handoff as `0997cc2`.
- DONE: Pushed `0997cc2` to `origin/main`.
- DONE: Focused Simulator tests and whitespace checks passed.
- `workspace/backups/**` remains untracked and excluded.

## Active Fill Review

- Review starts with `.audits/07-28-fill-changes.md`.
- Read only the named file until the user identifies the first review topic.
- Make no Fill source, test, audit, or wiki change before explicit confirmation.

## Focus

Active design audit: make Account, flat Ledger/Trade/Order/Fill records, and Simulator faster, internally consistent, free of dead code, and still runnable.

Production implementation has not started. Another session owns the Ledger-stack redesign. This session owns Account and Simulator audit alignment and adversarial readiness proof.

## Active User Code Review

### DONE

- Renamed the local `admission` variable to `nuubotSetup` in `internal/backtest/backtest.go`.
- The former BtBot-specific Controller timer key was superseded by shared `[controller]` cadence configuration.
- BotSpec now contains BotSpecID plus typed Controller, Signaler, Risk, and Executor specifications.
- BotSpec decoding validates, applies defined defaults, and shapes exact BotConfig TOML.
- BotSpec no longer carries App Config, Meta, replay inputs, ResultPath, Bot identity, runtime objects, or runtime state.
- Controller now receives complete Setup plus typed BotSpec and constructs runtime Signaler and Risk objects.
- Minimum-notional checks now read Setup App Config instead of BotSpec or Executor specifications.
- Removed obsolete `bot.Definition`, Controller unit tests, and BotSpec unit tests.
- Added `wiki/testing.md` with unit, integration, system, and RTest boundaries.
- Full Go tests, full vet, and diff check passed after the architecture hardcut.
- Reordered BtBot `Init` into grouped dependency order.
- Added complete `Step N:` intent comments to BtBot `Init`, `Start`, `Loop`, and `Stop`.
- Added mandatory numbered multi-step flow rules to `wiki/coding/STYLE.md` and `wiki/coding/RULES.md`.
- Named `internal/backtest/backtest.go` as the canonical numbered-flow example.
- BtBot cleanup formatting, diagnostics, and scoped diff check pass.
- Aligned Setup, BotSpec, Controller, BtBot, architecture, project, design, and BotSpec concept documentation.
- Full Go tests and full vet pass after final source and documentation alignment.
- Fast Observer RTest passed 1 of 1 through `./stest.sh -bot 9`.
- Final Observer attempted one run; suite total was 5,966 ms and BtBot average was 5,644 ms.
- Final Observer historical replay timing was 2,190 ms.
- Final Observer result log is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T114142Z.log`.
- Final Observer suite report is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T114142Z.json`.
- Final Observer log proves `btbot stopped.` after successful results publication.
- Manual code review of `internal/backtest/backtest.go` is complete.
- User approved the final BtBot flow, intent comments, loop state, Timer callback comment, and Stop logging.
- User prohibited more agents during this review session.
- Added behavior-preserving code-reorganization rules to `wiki/coding/STYLE.md`.
- Created global Codex skill `C:\Users\PC\.codex\skills\code-reorg\SKILL.md`.
- Codex skill activates through `code reorg`, `code-reorg`, or `reorg <file>`.
- Reorganized Controller `Init`, `Start`, `Run`, and `Stop` with grouped code and contiguous `Step N:` intent comments.
- Controller reorganization changes only comments and whitespace; operation order and behavior remain unchanged.
- Aligned `wiki/design/packages/controller.md` with exact Controller step text.
- Full Go tests, full vet, formatting, diagnostics, and diff checks pass after Controller reorganization.
- Observer RTest passed 1 of 1 after Controller reorganization.
- Observer suite total was 5,979 ms; BtBot average was 5,642 ms; replay was 2,207 ms.
- Observer result log is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T115049Z.log`.
- Observer suite report is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T115049Z.json`.
- Replaced `setup.Infrastructure` with one shared `*setup.Nuubot` application harness.
- Setup now builds and validates BotSpec before returning Nuubot.
- Moved pure Executor specification and resource types into BotSpec, removing the Setup-BotSpec-Executor import cycle.
- BtBot, Controller, BotCycle, Executor, and Account now receive the same Nuubot pointer.
- Controller retains only `c.nuubot` plus the approved `c.log` convenience reference.
- Removed duplicate Controller Setup, BotSpec, and Bot identity fields.
- Renamed BotCycle Account reconciliation from `Reconcile` to `AcctRecon`.
- Reorganized BotCycle with grouped flow and exact `Step N:` intent comments.
- Renamed `botcycle.Control` to `botcycle.BotCycle`.
- Split BotCycle initialization and start into explicit `Init` and `Start` lifecycle methods.
- Added BotCycle init, init-completed, start, started, stop, stopped-stats, and repeated-stop logging.
- Aligned Nuubot, Setup, BotSpec, BtBot, Controller, BotCycle, Executor, and Account documentation.
- Full Go tests, full vet, diagnostics, formatting, and diff checks pass after the Nuubot hardcut.
- Focused BotCycle and Controller proof passes after BotCycle reorganization.
- Final full Go tests, full vet, diagnostics, formatting, stale-reference scan, and diff checks pass.
- Final Observer RTest passed 1 of 1 through `./stest.sh -bot 9`.
- Final Observer suite total was 5,917 ms; BtBot was 5,592 ms; replay was 2,213 ms.
- Final Observer result log is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T130730Z.log`.
- Final Observer suite report is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T130730Z.json`.
- User authorized the final commit and push after manual Controller review.
- User authorized commit and push of the testing-boundary cleanup.
- Reorganized approved Executor, Account, Ledger, Trade, Order, Fill, and Simulator production files without behavior changes.
- Added contiguous `Step N:` intent comments and grouped Section 2 concerns across all approved files.
- Aligned Executor, TradeExecutor, GridExecutor, Account, Ledger, Trade, Order, Fill, and Simulator Program Flow documentation.
- Formatting, project diagnostics, full Go tests, full vet, and diff checks pass.
- Grid Bot 15 passed 1 of 1 through `./stest.sh -bot 15` without profiling.
- Grid suite total was 68,763 ms; BtBot was 63,860 ms; replay was 59,905 ms.
- Grid result log is `workspace/logs/nuubot5-stest-s11-b15-1-20260727T143233Z.log`.
- Grid suite report is `workspace/logs/nuubot5-stest-s11-b15-1-20260727T143233Z.json`.
- Trade Bot 13 runtime completed 7,948,800 ticks, 794,880 runs, 193 cycles, 193 Trades, 626 Orders, and 386 Fills.
- Trade result database integrity passed and foreign-key check returned no rows.
- Fixed `stest.sh` Trade validation so `result_config_match` is evaluated for every persistence mode.
- Trade Bot 13 passed 1 of 1 through `./stest.sh -bot 13` without profiling.
- Trade suite total was 17,838 ms; BtBot was 14,073 ms; replay was 10,781 ms.
- Trade result log is `workspace/logs/nuubot5-stest-s9-b13-1-20260727T143744Z.log`.
- Trade suite report is `workspace/logs/nuubot5-stest-s9-b13-1-20260727T143744Z.json`.
- Committed the completed reorganization and proof-harness fix as `176ca90`.
- Pushed `176ca90` to `origin/main`.
- Added `wiki/design/entities.md` with backtest and live runtime ownership cardinality trees.
- Linked the entity ownership page from `wiki/DESIGN.md`.
- Strengthened `AGENTS.md`, `wiki/USER.md`, and `wiki/SOUL.md` against patronizing tone, invented proposals, and unsupported correction.
- Corrections now require verified evidence plus useful recommendations or options.
- Replaced the `nuubot-runner` print-only placeholder with a BtBot-shaped command and Runner lifecycle scaffold.
- Runner owns WallClock, shared public Info, shared WebSocket, Controller, and supervision lifecycle.
- Added public `hyperliquid.Info` with Meta and credential-free clearinghouse-state requests.
- Renamed Hyperliquid `client.go` and its test to `endpoint-info.go` and `endpoint-info_test.go`.
- Added the explicit unimplemented WebSocket lifecycle stub in `endpoint-ws.go`.
- Added the Account-owned credentialed Exchange reservation in `endpoint-exchange.go` without implementation.
- Nuubot now carries Runner-attached Clock, Info, and WebSocket shared infrastructure.
- Meta refresh retains its separate hardcoded-mainnet Info object.
- Added Runner package, process, entity, endpoint, Setup, binary, architecture, and project documentation.
- Moved the canonical Runner design to `wiki/design/runner.md` and aligned every reference.
- Runner remains non-runnable because Setup, Signaler construction, WebSocket transport, and live events remain replay-oriented or unimplemented.
- Gofmt, project diagnostics, stale-reference scans, and Git whitespace checks pass.
- `nuubot-runner` was not executed. No Runner tests or system tests were run by explicit instruction.
- Added `wiki/design/startup.md` as the canonical clean-start and crash-recovery design.
- Deleted the separate Recovery concept because Startup owns both paths.
- Linked Runner and the design index to the unified Startup contract.
- Scoped Startup recovery to Runner across `simnet`, `testnet`, and `mainnet`.
- Confirmed BtBot never restores runtime state; interrupted backtests rerun from the beginning.
- Added the explicit no-recovery crash model to BtBot package and command design pages.
- Datastore now loads and validates mutable Bot status from the existing Bot row.
- BtBot always resets loaded status to `configured` and runs `clearData` before Controller initialization.
- BtBot runtime persistence now uses a fresh `.partial` attempt database and preserves the prior completed result until publication.
- Runner retains loaded status and contains no data-clearing path.
- Runner rejects terminal `error` and `stopped` Bots before child construction.
- Terminal Runner reruns require cloning into a new Bot ID; failed evidence remains intact.
- Focused Datastore, Account, ResultPublisher, Executor, BotCycle, BtBot, and Runner tests pass.
- Observer Bot 9 passed 1 of 1 with its stored `running` status reset to `configured` by BtBot.
- Observer suite total was 6,198 ms; BtBot was 5,860 ms; replay was 2,309 ms.
- Observer result log is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T160925Z.log`.
- Observer suite report is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T160925Z.json`.
- Project diagnostics, stale destructive-Runner scan, formatting, and diff checks pass.
- Renamed `executor.Context` to `executor.BotCycleContext`; local variables remain `ctx`.
- Moved canonical Server design to `wiki/design/server.md`.
- Documented current `nuubot-*` development commands and the target single `nuubot` binary.
- Documented one Server process containing WebServer, API, BotManager, and SweepManager.
- Documented isolated Runner and BtBot child processes launched from the same target binary.
- Reorganized `internal/executor/observer.go` with exact numbered intent comments and no executable behavior changes.
- Aligned ObserverExecutor Program Flow documentation.
- Focused Executor tests, BotCycle compilation, formatting, and diff checks pass after Observer reorganization.
- Observer, Trade, and Grid now log Executor Init, Start, and Stop entry and completion.
- Every Executor lifecycle log includes cycle, Executor number, Executor ID, kind, and side.
- Observer and Trade now implement `StartHandler`; Init ends `Starting` and BotCycle Start marks them `Running`.
- High-frequency Run, reconciliation, Signal, and BBO paths gained no lifecycle logs.
- Observer Bot 9, Trade Bot 13, and Grid Bot 15 each passed 1 of 1 after lifecycle logging.
- Observer suite total was 8,687 ms; Trade was 23,017 ms; Grid was 74,122 ms.
- Lifecycle logs prove all six boundaries for cycle 1 Executor 1 in every Executor kind.
- Observer result log is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T165359Z.log`.
- Trade result log is `workspace/logs/nuubot5-stest-s9-b13-1-20260727T165359Z.log`.
- Grid result log is `workspace/logs/nuubot5-stest-s11-b15-1-20260727T165400Z.log`.
- Approved one root-owned MarketData object carried through Nuubot.
- MarketData owns BBO ingestion, latest buffers, subscriptions, and notifications.
- BtBot publishes Parquet BBO through MarketData; Runner WebSocket publication remains the next live implementation.
- Simulator subscribes directly and reads the latest buffer inside its callback.
- Executors pull latest BBO during Start and subscribe only for current strategy requirements.
- BBO leaves BotCycleContext and the Controller-BotCycle-Executor-Account-Venue forwarding chain is retired by the target design.
- Added permanent `wiki/design/marketdata.md` with reasoning, ownership, lifecycle, timing, recovery, prohibited regressions, cutover, and proof.
- Implemented shared exact-key MarketData buffering, subscriptions, and idempotent subscription shutdown.
- BtBot and Runner now create and attach one owned MarketData object using matching flow.
- Executor Start reads latest BBO; Observer and Grid subscribe only for required tick policy.
- Simulator now subscribes directly; the old Controller-to-Venue BBO forwarding chain is removed.
- Focused MarketData, Simulator, Account, Executor, BotCycle, Controller, BtBot, and Runner tests pass.
- Observer BotSpec now shapes complete Venue, network, and symbol market identity.
- Existing Observer configs receive the explicit `simulator/simnet` default when both fields are absent.
- Observer Bot 9 passed 1 of 1 with 63 cycles and 16 stop-loss exits after Start-time latest-BBO entry.
- Observer suite was 5,986 ms; BtBot was 5,660 ms; replay was 2,891 ms.
- Observer result log is `workspace/logs/nuubot5-stest-s6-b9-1-20260727T175352Z.log`.
- Trade Bot 13 passed 1 of 1 with unchanged 193 Trades, 626 Orders, and 386 Fills.
- Trade suite was 20,586 ms; BtBot was 16,589 ms; replay was 12,887 ms.
- Trade result log is `workspace/logs/nuubot5-stest-s9-b13-1-20260727T175406Z.log`.
- Grid Bot 15 passed 1 of 1 with unchanged 1,982 Trades, 4,697 Orders, 2,636 Fills, and 585 round trips.
- Grid suite was 73,723 ms; BtBot was 68,877 ms; replay was 64,938 ms.
- Grid result log is `workspace/logs/nuubot5-stest-s11-b15-1-20260727T175406Z.log`.
- MarketData implementation and BtBot integration are complete; Runner WebSocket publication remains Runner work.
- Executor `Create` centrally rejects terminal and unknown loaded status without advancing lifecycle state.
- Observer, Trade, and Grid `OnInit` explicitly retain loaded status without flattening `Running`, `Paused`, or `Stopping`.
- Every Executor `OnStart` logs first, then advances only `Configured` to `Starting`; recovered nonterminal status is preserved.
- Every Executor `OnStop` logs first, then advances eligible nonterminal status to `Stopping`; terminal status is preserved.
- Removed redundant terminal-status guards from Executor `OnStart`; `Create` owns loaded-status admission.
- Focused Executor, BotCycle, Controller, and BtBot tests pass after lifecycle alignment.
- The Observer integration run was user-stopped before execution and was not retried.
- Final full Go tests and full vet pass with canonical `-tags noasm`.
- Project diagnostics report no errors or warnings.
- Added `cmd/nuubot-fprof` and dependency-free `internal/fprof/runtime` instrumentation.
- Function profiling uses temporary AST-generated Go build overlays and changes no tracked application source.
- A records the original runtime; B records exact calls; C records exact calls plus flat and cumulative timing.
- Observer Bot 9 A/B/C behavior matched exactly across 7,948,800 ticks and 794,880 Controller runs.
- Observer A was 6.294s, B was 8.733s, and C was 16.168s.
- Function-profile overhead was B-A 2.439s, C-B 7.435s, and C-A 9.874s.
- The proof instrumented 53 source files and 544 functions.
- Text reports use `go-pretty/table`; numeric columns align right and text columns align left.
- The function table starts with complete-profile flat and cumulative totals; aggregate calls and averages stay blank.
- Observer function-profile proof is `workspace/perf/fprofiles/s6-b9-20260728T033218Z/report.txt` and `report.json`.
- Trade Bot 13 A/B/C behavior matched; A was 15.758s, B was 19.174s, and C was 29.662s.
- Trade total flat time was 28.917s, covering 97.49% of C elapsed time.
- Trade function-profile proof is `workspace/perf/fprofiles/s9-b13-20260728T033505Z/report.txt` and `report.json`.
- Grid Bot 15 A/B/C behavior matched; A was 64.066s, B was 77.719s, and C was 114.101s.
- Grid total flat time was 113.329s, covering 99.32% of C elapsed time.
- Grid hotspots were `decodeOne` 14.536s, `replaceTradeSummary` 10.433s, `Encode` 5.683s, `crosses` 4.767s, and `sortedActiveOrders` 4.516s.
- Grid function-profile proof is `workspace/perf/fprofiles/s11-b15-20260728T034420Z/report.txt` and `report.json`.
- Grid optimization is intentionally deferred until GridExecutor and Simulator hand-tuning.
- Replaced shared `[controller]` with validated `[live]` and `[backtest]` runtime policies.
- Both profiles configure Controller 1,000ms, Recon 10,000ms, Recon sweep 60,000ms, and telemetry 10,000ms.
- Live sets telemetry write-on-collect true. Backtest sets it false.
- Config rejects nonpositive Controller or telemetry cadence, Recon cadence not exceeding Controller, and sweep cadence not exceeding Recon.
- Account stores `lastReconMS` beside `failureCount` and owns the accepted timed skip decision.
- Initial and forced Recon execute immediately; dirty or pending state executes after 10 seconds; clean state executes after 60 seconds.
- Only a successfully published Recon advances `lastReconMS`; skips and failures do not.
- BotCycle reports whether any Account executed Recon; Controller delivers `OnRecon` only after executed Recon.
- BtBot selects `App.Backtest`; Runner selects `App.Live`; lower components read only selected `nuubot.Runtime`.
- BtBot checks telemetry independently from Controller cadence and retains 794,881 ten-second samples for terminal publication.
- `stest.sh` derives tick, Controller-run, and telemetry invariants from result data instead of hardcoded cadence.
- Validation failures now name the failed check, expected value, actual value, and result database.
- Standalone Observer, Trade, and Grid each passed after runtime and validator changes.
- Concurrent Observer, Trade, and Grid each passed without SQLite locking, path collision, or cross-Bot validation failure.
- Trade now produces 190 cycles, 190 Trades, 617 Orders, and 380 Fills under one-second Controller cadence.
- Grid remains 50 cycles, 1,982 Trades, 4,697 Orders, and 2,636 Fills.
- Added `internal/runharness.Profile` for shared whole-Run profiling infrastructure.
- Reduced both command mains to argument parsing, one package `Execute` call, and one terminal error path.
- Added `backtest.Execute` and `live.Execute`; Runner was compiled but never executed.
- Post-cleanup Observer passed in 9,112ms; replay was 5,390ms.
- Post-cleanup Observer log is `workspace/logs/nuubot5-stest-s6-b9-1-20260728T055624Z.log`.
- Post-cleanup Observer suite report is `workspace/logs/nuubot5-stest-s6-b9-1-20260728T055624Z.json`.
- Full Go tests, full vet, shell syntax, diff checks, stale-reference scans, and project diagnostics pass with canonical `-tags noasm`.
- Renamed `internal/btbot` to `internal/backtest` and its lifecycle owner to `backtest.Run`.
- Renamed `internal/runner` to `internal/live` and its lifecycle owner to `live.Run`.
- Command binaries now call `backtest.Execute` and `live.Execute`.
- Config loads both profiles into `nuubot.App`; each Run selects one into `nuubot.Runtime` before child initialization.
- Config requires Live telemetry write-on-collect and Backtest terminal-only telemetry publication.
- Added `telemetry.Store` for typed per-Bot Live telemetry rows with resumed sequence and idempotent Stop.
- Live checks telemetry cadence after successful Controller callbacks, writes due samples immediately, and writes one terminal sample during Stop.
- Focused Config and Telemetry persistence tests pass.
- Full Go tests, full vet, shell syntax, diff checks, stale Go-package scans, and project diagnostics pass after the package hardcut.
- Concurrent Observer, Trade, and Grid each passed once after the package hardcut.
- Observer proof: `workspace/logs/nuubot5-stest-s6-b9-1-20260728T061042Z.log` and `.json`.
- Trade proof: `workspace/logs/nuubot5-stest-s9-b13-1-20260728T061043Z.log` and `.json`.
- Grid proof: `workspace/logs/nuubot5-stest-s11-b15-1-20260728T061043Z.log` and `.json`.
- Runner was not executed.
- Reorganized Ledger, Trade, Order, and Fill as independent sibling packages under `internal/account`.
- Added central durable Bot and Sweep commands with processed, skipped, rejected, and acknowledged outcomes.
- Added central process generations, lifecycle status, heartbeat, health, and atomic Bot or Sweep status publication.
- Added one isolated execution database per globally unique Bot ID under `workspace/db/bots`.
- Backtest publishes through `.partial`; Live reopens durable evidence and never clears it.
- Observer, Trade, and Grid passed after the Account and control-plane changes.
- Trade and Grid passed concurrently without execution-database isolation failures.
- Command claim, acknowledgement, rejection, central lifecycle publication, full tests, and full vet passed.
- Final full Go tests and full Go vet pass with canonical `-tags noasm`.
- Final shell syntax and Git whitespace checks pass.
- Zed Go support is restored with native `gopls v0.23.0`; global Zed settings pin its binary and native Go environment.
- Zed log proves `gopls.exe` starts for `D:\rust\nuubot5`; project diagnostics are clean with no later language-server error.
- Global Zed settings were backed up at `C:\Users\PC\AppData\Roaming\Zed\settings.pre-gopls-20260728.json` before modification.
- Windows user `PATH` now includes the native Go toolchain and `gopls` directories and no longer includes `D:\rust\nuutrader6`.
- The original and updated Windows user `PATH` values are backed up at `workspace/backups/windows-user-path-20260728.txt`.
- Central Bot and process rows for Bots 9, 13, and 15 are `stopped`; each process generation is positive.
- Current Setup and `stest.sh` use only `workspace/db/bots/bot_<BotID>.db`; historical sweep database artifacts remain untouched.
- Runner was compiled but not executed.
- Committed and pushed Account package grouping, central control, and per-Bot execution databases as `5f16e26`.
- Created clean sibling worktree `D:\rust\nuubot5-server` on branch `nuubot5-server` for Server and GUI work.
- Created `.audits/07-28-account.md` with approved Account menu, lifecycle organization, caller cleanup, and CLOID ownership decisions.
- Account keeps CLOID creation; Executor supplies Order intent and `OrderLevel`; `cloid` performs mechanical encoding only.
- Account owns dirty state. Ledger reports pending evidence only. No public `IsDirty` exists without a confirmed caller.
- Account audit removes no-op `Result.Clone`, redundant `copySpec`, dead wrappers, and test-only `Order` and `Fill` wrappers.
- Same-package Account tests may inspect Account-owned Ledger records directly; test-only access does not justify production APIs.
- Read the complete `.audits/07-28-ledger-stack.md` target through line 1305.
- Ledger target uses one canonical Ledger and one flat record per Trade, Order, and Fill with ID indexes and no nested duplicate graph.
- Ledger audit records three earlier adversarial rounds ending PASS, but root Account/Ledger reconciliation remains unfinished.
- Created `.audits/07-28-simulator.md` with all 42 functions, callers, dispositions, target internals, impacts, sequence, and proof.
- DONE: Created concise entity-specific change files for Account, Ledger, Trade, Order, and Fill.
- Root review removed stale raw-payload history, Ledger-owned Recon redesign, `CountOrders`, Account `Start`, repair logic, and duplicate local/database IDs.
- ACTIVE REVIEW: Renamed the Simulator change document to `.audits/07-28-simulator-changes-pending-review.md`.
- Simulator Topic 1 remains pending review. `SetFillFeeAvailableForTest` was deleted from `simulator.go` and awaits review.
- Every other Simulator function has a production caller and is not dead.
- Earlier proposed removals confused used code with future behavior changes. Those decisions are withdrawn pending topic-by-topic review.
- Simulator persistence and all other committed behavior remain unchanged.
- Production-only proof: `go build -tags noasm ./internal/simulator` passes after the test-only API deletion.
- Completed the Simulator upstream/downstream trace across Account, Ledger, Hyperliquid, MarketData, Clock, Backtest, Live, BotSpec, and results.
- Simulator remains Account-owned. One Bot loop owns serialized calls.
- Temporary rule: `Simulator.Persist(mode)` writes dirty rows for both `none` and `max`.
- Future architecture TODO: restore true memory-only `none` after Simulator state survives Account and Executor cycle replacement.
- That future change must define account-state lifetime, cycle handoff, final publication, recovery, and shared-account ownership together.
- User requires agent status confirmation every 15 minutes during active delegated work.
- No production Account, Ledger-stack, or Simulator implementation has started in this audit task.

### TODO

- DONE: Reconciled Account and Ledger ownership decisions through the topic review.
- DONE: Completed the Simulator audit and holistic implementation blueprint.
- ACTIVE ROOT: Verify the Simulator audit and deliver the implementation-readiness summary.
- FUTURE ARCHITECTURE: Make `persist_mode=none` truly memory-only without losing Simulator state between Account and Executor cycles.
- QUEUED: Run one adversarial Account/Ledger review, fix accepted blockers, and repeat once only when blockers changed contracts.
- QUEUED: Run three independent adversarial reviews across Account, Ledger-stack, and Simulator audits; one must prove implementation readiness and complete caller impact.
- QUEUED: Fix accepted blockers and repeat final adversarial review up to two more rounds only when contracts changed.
- QUEUED: Deliver an implementation-readiness summary before production coding.
- Keep `.audits/**` uncommitted while the audits remain active.
- Re-read current Account-stack source and callers before implementation because another session may change them.
- Add Observer stop parameters for stop-loss, tick count, bar count, and elapsed-time limits.
- Use bounded Observer runs such as 200 ticks for fast smoke tests and small logs.
- Keep Controller, Recon, and Recon sweep timing user-configurable; tune fixed intervals from performance data before considering live self-tuning.
- Tune Backtest telemetry interval, retention, and memory because terminal-only publication accumulates samples in process memory.
- Revisit the recorded Grid function-profile hotspots during GridExecutor and Simulator hand-tuning.
- Complete Live Run and test it first with bounded Observer before Trade or Grid.
- Use the Live Observer smoke test to expose lifecycle and integration defects before trading execution.

### ACTIVE AUDIT EXECUTION CONTRACT

#### Overall Intent

- Rework Account, Ledger, Trade, Order, Fill, and Simulator as one coherent execution stack.
- Preserve accepted trading behavior, finance, recovery, persistence, system proof, and Exchange-shaped Simulator behavior.
- Improve runtime speed, memory shape, ownership clarity, and code consistency.
- Remove dead code, duplicate representations, clone chains, nested duplicate records, and test-only production APIs.
- A small local edit must be assessed by its complete cross-package and runtime impact.

#### Working Files

- `.audits/07-28-account.md` owns Account menu, lifecycle organization, Order preparation, CLOID ownership, callers, and removal decisions.
- `.audits/07-28-ledger-stack.md` owns flat Ledger, Trade, Order, Fill, persistence, recovery, reconciliation, schema, and domain proof.
- `.audits/07-28-simulator.md` owns Simulator lifecycle, API, state, matching, persistence, MarketData, parity, performance, callers, and proof.
- Production source, tests, and canonical wiki remain unchanged during audit work.
- Keep every active `.audits/**` file uncommitted.

#### Review Method

Every root and agent review must use this order:

1. State the complete application intent.
2. State the reviewed component's exact responsibility.
3. State each proposed change's intent.
4. Trace the current implementation and direct callers.
5. Assess complete ownership, behavior, data, persistence, recovery, and performance impact.
6. Identify every required caller, schema, result, test, and documentation change.
7. Reject local simplifications that break another execution path.
8. Mark findings as blocker, non-blocker, invalid, or already resolved.

Every proposed function decision must include:

- current one-line behavior;
- exact production and test callers;
- `KEEP`, `REMOVE`, `MODIFY`, `MOVE`, or `RENAME`;
- target signature and owner when changed;
- affected callers and replacement path;
- backtest and Live impact;
- required proof.

#### Phase 1 - Root Account and Ledger Alignment

- Read both audits completely before editing either.
- Re-read current Account-stack source and direct callers because another session may change them.
- Reconcile ownership language. Ledger allocates local Trade and Order IDs; Account orchestrates and creates CLOIDs.
- Define exact canonical Trade, Order, Fill, Ledger, Account input, output, and result types.
- Remove every unresolved phrase such as `if practical`, `if required`, or `implementation-time confirmation`.
- Define exact Account lifecycle. Do not add `Start` unless concrete owned background work requires it.
- Define exact Account observation APIs and immutable return shapes.
- Define exact terminal publication of all Backtest Trade, Order, and Fill records under `persist_mode=none`.
- Define exact Live recovery, archival reads, raw evidence, schema version, and failure atomicity.
- Define exact ResultPublisher, report, Executor, Grid, Trade, tests, schema, and wiki caller changes.
- Edit only the two audit files during alignment.

#### Phase 2 - Simulator Audit

- DONE: Read `simulator.go`, direct callers, tests, design, profiles, persistence, Account integration, and Nuutrader6 references.
- DONE: Listed all 42 functions with one-line behavior, exact callers, disposition, target, impact, sequence, and proof.
- DONE: Covered MarketData, matching, exchange parity, state, persistence, recovery, logging, lifecycle, Clock, and performance.
- DONE: Added the holistic upstream/downstream implementation blueprint.
- No production, test, wiki, Account-audit, or Ledger-audit change was made.

#### Phase 3 - Account and Ledger Adversarial Review

- After Phase 1, spawn one independent adversarial agent to review both complete audits.
- The agent starts from application intent and assesses actual proposed changes, not isolated wording.
- The agent checks implementation readiness, ownership, caller completeness, data correctness, recovery, persistence, performance, and proof.
- The agent may report blockers only. Root assesses each finding and edits the audits for accepted blockers.
- Reject invalid findings with exact source or audit evidence.
- Run at most two Account/Ledger adversarial rounds total.
- Run round two only when accepted round-one fixes changed behavior, ownership, contracts, or unresolved judgment.
- Phase 3 passes only with no unresolved blocker.

#### Phase 4 - Three Independent Full-Stack Reviews

- Start only after all three audits exist and Phase 3 passes.
- Spawn three independent agents concurrently. Every agent reads and reviews all three audits.
- Agent A emphasizes ownership, lifecycle, correctness, failure, reconciliation, and recovery.
- Agent B emphasizes performance, allocations, indexes, matching complexity, persistence volume, memory retention, and Simulator hot paths.
- Agent C proves implementation readiness, complete callers, schema, publication, tests, sequencing, and whether changes break other programs.
- Agents remain read-only and return findings to root. They do not concurrently edit shared audit files.
- Root assesses every finding, fixes accepted blockers, and records rejected findings with evidence.
- Repeat the three-agent round at most two more times.
- Repeat only when accepted fixes changed behavior, ownership, contracts, or unresolved judgment.
- Stop repeating after one clean round or after the maximum three full-stack rounds.

#### Blocker Standard

A blocker is any unresolved issue that can:

- prevent direct implementation;
- create ambiguous ownership or duplicate mutable state;
- change accepted trading behavior or finance;
- lose, merge, or misroute Trade, Order, Fill, or raw Exchange evidence;
- break Backtest, Live, recovery, persistence, ResultPublisher, report, Executor, Grid, or Simulator;
- violate Exchange parity or CLOID, OID, TID identity rules;
- publish memory before required durable commit;
- advance an unproven reconciliation cursor;
- reintroduce nested searches or material performance regressions;
- leave required proof impossible or undefined.

Style preference without behavioral, ownership, implementation, or proof impact is not a blocker.

#### Agent Continuity

- Confirm every delegated worker after 30 seconds when polling is available.
- Report every active worker at least every 15 minutes as `AGENT N - task - Alive and working`.
- Report cancellation, overload, silence, failure, or missing output immediately.
- Replace failed workers when servers permit. Root continues independent work during agent outages.
- Never claim an agent is active without confirmed evidence.

#### Completion Contract

The audit task completes only when:

- all three audit files exist;
- every function has an exact disposition and replacement path;
- every affected caller, schema, result, test, and wiki owner is listed;
- Backtest and Live behavior is explicit without scattered mode branches;
- persistence, recovery, archival, raw evidence, identity, and cursor rules are exact;
- implementation steps are ordered and dependency-safe;
- proof commands and accepted semantic invariants are exact;
- adversarial review limits are satisfied;
- no unresolved blocker or implementation choice remains;
- root gives one concise implementation-readiness summary;
- production implementation remains unstarted until the user requests it.

### PENDING USER APPROVAL

- None.

### CONFIRMED DESIGN


- Build one typed BotSpec object from exact persisted BotConfig TOML text; BotSpecID selects the exact decoder.
- `botspec` only decodes, validates, applies explicitly defined defaults, and shapes BotConfig into one clean immutable BotSpec object.
- Do not use `admission`, `admit`, or `admitted` for this flow; validation returns a BotSpec or an error.
- BotSpec contains BotSpecID plus Controller, Signaler, Risk, and Executor specifications.
- Complete Setup contains App Config, Meta, replay inputs, ResultPath, Bot instance identity, and provenance.
- Setup returns one shared `setup.Nuubot` application harness containing global infrastructure and the typed BotSpec.
- `Nuubot` contains shared infrastructure data, not procedural behavior, functionality, or features.
- BtBot and Controller retain the shared `nuubot` reference instead of duplicating Setup, BotSpec, identity, or config state.
- A component may retain `nuubot.Log` as its local logger reference for convenient logging.
- Controller owns runtime construction from `nuubot.BotSpec`.
- New and recovered Bots use the same Init code to load Bot, Ledger, Trade, Order, and Fill state.
- Every Bot uses the same Start code and invokes Account reconciliation.
- Reconciliation skips Exchange Order and Fill pulls when Ledger has no active Trades, then advances cursors to current time.
- Active Trades select normal Venue Order and Fill reconciliation.
- Persisted `starting`, `running`, and `stopping` intent resumes only after normal initialization and startup reconciliation.
- BtBot, Controller, BotSpec, BotCycle, and Executor runtime lifecycle do not use isolated unit tests.
- Deleted isolated BotCycle, ObserverExecutor, and TradeExecutor test files.
- Retained only pure deterministic Grid calculation tests in `internal/executor/grid_test.go`.
- Observer, Trade, and Grid system runs prove their real integrated paths.
- Ledger, Trade, Order, and Fill require strong direct domain tests.
- Simulator tests are Venue parity tests, including official JSON, canonical state, persistence, failure atomicity, matching, and exact comparison mechanics.
- Meta, OHLCV mechanics, Risk, Signaler calculations and packages, and CLOID are valid direct-test targets.
- Replay Reader has no isolated tests; real replay and system runs prove its concrete OHLCV-to-BBO path.
- Strengthened BalancedRisk proof for Allow, assessment counting, idempotent Stop, and unknown-kind rejection.
- Grid Bot 15 passed 1 of 1 through `./stest.sh -bot 15` after test-boundary cleanup.
- Grid suite total was 66,591 ms; BtBot was 61,689 ms; replay was 57,982 ms.
- Grid result log is `workspace/logs/nuubot5-stest-s11-b15-1-20260727T131553Z.log`.
- Grid suite report is `workspace/logs/nuubot5-stest-s11-b15-1-20260727T131553Z.json`.
- Retained Grid calculation tests and strengthened Risk tests pass.
- Full Go tests and full vet pass after final testing-boundary alignment.
- Manual code review of `internal/controller/controller.go` passed on 2026-07-27.
- Manual code review of `internal/botcycle/botcycle.go` passed on 2026-07-27.
- A later requested full-test and vet rerun was user-stopped before producing output.
- The successful full tests, full vet, and Observer proof recorded above remain the latest completed proof.
- High-frequency `Controller.Run()` has no entry or exit logging; terminal Controller stats prove its execution without log spam.
- Keep the `c.cycle != nil` guard in `Controller.Run`; nil is normal between cycles, and Controller must continue to Signal evaluation.
- Signaler produces one complete flat Signal package containing standard Actions and arbitrary BotSpec-defined custom fields.
- Controller passes the unchanged current package into `BotCycle.Run(signal)` before applying its standard lifecycle Action.
- BotCycle passes the unchanged package to supported running Executors; each Executor reads only fields required by its BotSpec.
- Never split one Signal package into separate standard and custom carriers.
- `recordBBOGap` measures the gap since the prior BBO; do not add min/max BBO-gap metrics until required.
- Clock is shared Nuubot infrastructure attached after initialization by BtBot or Runner.
- Controller, BotCycle, and children read current time through `nuubot.Clock.NowMS()`; `Run` and `OnRecon` do not receive timestamp arguments.
- BotCycle directly owns and manages Executors only.
- Each Executor directly owns its Account; each Account directly owns Ledger and Venue or Simulator.
- BotCycle affects Accounts only through Executor lifecycle and capability calls.

## Active Boundary Hardcut

### DONE

- Confirmed canonical Recon remains one bulk Venue consistency barrier.
- Confirmed current Account-Simulator types, persistence, results, and BBO flow violate ownership.
- Approved deleting Recon2 instead of adapting it.
- Deleted Recon2 source, selection, tests, template, and stale active documentation.
- Added official Hyperliquid actions and detached JSON responses.
- Made CLOID mandatory, opaque, and the first lookup identity.
- Kept OID as Venue-assigned fallback identity.
- Replaced Simulator state with one canonical Order record and one-time Fill records.
- Moved Simulator persistence and terminal state outside Account and Ledger results.
- Passed focused tests, full tests, full vet, and exact Bot 15 boundary parity.
- Added transient exact Simulator matching keys.
- Passed exact-key parity, fuzz, zero-allocation, normal Bot 15, and profiled Bot 15 proof.
- Updated owning design and performance pages.
- Audit round one found immediate IOC recovery could lose a CLOID-less Fill after submit persistence failure.
- Account now enriches that Fill from same-attempt OID-to-CLOID Order evidence before Ledger application.
- Immediate IOC recovery now proves Order, Fill, position, finance, and zero pending state.
- Audit round two passed with no findings, missing proof, bloat, or duplicate ownership.
- Full tests, full vet, gofmt, shell syntax, stale-route, and diff checks pass.

### TODO

- None.

### PENDING USER APPROVAL

- None.

## Active Recon Design Review Contract

- Every architect, documenter, adversarial reviewer, and fixer must assess each change holistically.
- Review must cover lifecycle, ownership, memory, indexes, Venue behavior, domain integrity, persistence, recovery, telemetry, STOP, tests, and cutover.
- A locally correct change fails review when it creates an unaddressed cross-package conflict or external impact.
- Reviewers must identify exact affected code and documentation outside Recon.
- Rewrite and reorganize complete sections when structure or ownership is wrong; never accumulate patch text.
- Keep the Recon concept minimal and every specification limited to implementation-driving facts.
- Reject duplicated, stale, bloated, locally patched, or structurally misleading design.
- No implementation begins before the user approves the reviewed design.

## Active Task

### DONE

- TMUX startup produced one full-height control pane and four viewer panes.
- All four viewer roles returned the required standby response.
- Direct targeted `send-keys` delivered viewer prompts successfully.
- PSMux `paste-buffer` ignored its viewer target and pasted into control.
- Dismissing the Codex composer suggestion with `Escape`, then sending `Enter`,
  submitted the prompt.
- Alternative-input test scope approved.
- Current Codex manual confirms Ctrl+G external-editor input.
- Current Codex manual confirms `codex exec -` full-prompt stdin input.
- Paste-setting A/B stopped when the user changed scope.
- Temporary `terminal.integrated.ignoreBracketedPasteMode` setting restored and verified absent.
- `View: Set Panel Alignment to Left` produced the requested VS Code layout.
- User confirmed the layout works perfectly.
- `View: Set Panel Alignment to Center` made both sidebars full height.
- User confirmed the final centered-panel layout.
- Paste-diagnosis scope confirmed as read-only.
- VS Code 1.129.1 built-in keybindings route both chords to terminal paste.
- PowerShell overrides Ctrl+V by sending control character 22 to Codex.
- Ctrl+Shift+V directly injects clipboard text through VS Code terminal paste.
- Both user tests produced the same lost tabs and added blank lines.
- No VS Code, Codex, source, or configuration changed; only handoff state changed.
- Complete report content and presentation requirements confirmed.
- Complete SuiteReport aggregation and fixed-width rendering implemented.
- Focused tests, full tests, vet, and script checks passed.
- Grid 5x completed with five successful processes.
- Delegated diagnosis traced the panic below current trading logic.
- Delegated diagnosis isolated the successful-suite divergence to cycle 40.
- The complete generated Grid 5x report is ready for user review.
- `stest.sh` is the only fresh-process system, stress, and profiling command.
- `stest.sh` rejects invalid selectors, repetitions, duplicates, profiling combinations, and extra arguments.
- `nuubot-bt-bot <sweep_id> <bot_id> -pp <prefix>` enables profiling; absence of `-pp` creates no profiles.
- Performance runs capture CPU, trace, heap, allocations, block, and mutex artifacts.
- Performance artifacts live under `workspace/perf/profiles/stest-s<sweep>-b<bot>-<timestamp>/`.
- Focused command tests, vet, shell syntax, diagnostics, and diff checks passed.
- Historical Observer one-run proof passed in 5,604 ms.
- Historical Observer profiled proof passed in 7,111 ms with 2,190 ms replay timing.
- All six performance artifacts are nonempty and readable by Go 1.26.5 tools.
- Proven profile: `workspace/perf/profiles/pptest-s6-b9-20260725T152130Z/`.
- Performance, filesystem, command, and BtBot wiki owners are aligned.
- Grid profiling binary completed successfully in 81,129 ms with 76,356 ms replay timing.
- The old profiling script falsely rejected Grid through inherited Observer counts. Generic successful-result validation replaced it.
- Grid profiling allocated 144,495.70 MB versus Observer 3,726.29 MB.
- Grid Account reconciliation owns 112,609.46 MB cumulative allocation.
- Grid Trade Orders owns 24,999.31 MB flat allocation.
- Grid math/big allocation owns 22,887.53 MB flat allocation.
- Grid GC consumed 53.19 CPU seconds; replay Loop consumed 73.32 CPU seconds.
- Grid ResultPublisher syscall delay was below one second; database waiting is not the main Grid bottleneck.
- Grid profile: `workspace/perf/profiles/pptest-s10-b14-20260725T152717Z/`.
- All six Grid artifacts are nonempty and readable by Go 1.26.5 tools.
- Adversarial Ledger clone audit completed at `.audits/07-25-ledger-clone-audit.md`.
- Corrected Ledger reassessment completed at `.audits/07-26-ledger-clone-reassessment.md`.
- Reassessment confirms failure-before-publication, dirty-only persistence, fatal first sweep error, and future live third-consecutive-failure stoppage.
- Renamed tracked engineering evidence from `audits/` to `.audits/`.
- Filesystem design now owns `.audits/` as development-only review evidence.
- Added concise `wiki/coding/sample.md` and linked it from canonical style and project pages.
- `wiki/reviews.md` contains exactly 94 current Go files with explicit review state, date, and hash columns.
- `workspace/perf/` is ignored and excluded from the commit.
- Full Go tests and vet pass with `CGO_ENABLED=0` and `-tags noasm`.
- Build, stress, trade, Grid, and performance shell syntax checks pass.
- Full diff whitespace check passes.
- Complete durable work committed at `HEAD` with message `Add Grid telemetry and performance reports`.
- Approved heartbeat reconciliation model documented in all five canonical design pages.
- Ledger Chunk 2 readable oracles passed focused count-10 and the full Ledger suite.
- Ledger Chunk 2 audit round three passed with no unresolved finding.
- Ledger Chunk 3 Account and publication oracles passed focused and full package proof.
- Ledger Chunk 3 audit round two passed with no unresolved finding.
- Ledger Chunk 4 failure characterization passed combined focused and full package proof.
- Ledger Chunk 4 adversarial audit passed with no material finding.
- Ledger Chunk 5 SQL fault and metrics proof passed the seven-package baseline.
- Ledger Chunk 5 adversarial audit passed with no material finding.
- User limited production behavior changes to Recon.
- Chunks 11–13 are stopped because they change non-Recon persistence routing.
- All worker agents stopped.
- Removed the interrupted, unfinished Chunk 6 `internal/ledger/ledger.go` diff.
- Restored checkpoint passes full tests, full vet, and diff whitespace checks.
- Traced current BotSpec admission, Bot templates, Nuubot3 Sweep expansion, replay fields, and canonical owners.
- Implemented `internal/btsweep` loading, validation, and deterministic expansion.
- Added one-Bot Observer, Trade, and Grid system Sweep templates.
- Added deterministic Cartesian expansion and failure-contract tests.
- Updated BotSpec template tests for robust project-root discovery and exact IDs.
- Aligned requested Sweep, Bot, architecture, design, and binary documentation.
- Focused command passed: `CGO_ENABLED=0 .../go.exe test -count=1 -tags noasm ./internal/btsweep ./internal/botspec`.
- Focused proof passed both packages in 1.295 seconds and 1.265 seconds.
- Diagnostics report no errors or warnings in all three changed Go files.
- Scoped `git diff --check` passed; only existing line-ending warnings were reported.
- `cmd/nuubot-bt-sweep/main.go` remains the pre-existing `Under Construction.` placeholder.
- No database write, process launch, replay, dependency change, commit, or push occurred.
- All canonical script owners now use `stest.sh`; deleted scripts remain absent.
- Parameter-free Sweeps emit one deterministic hashed Bot per date range.
- Sweep admission requires documentation and replay coverage for every generated Executor symbol.
- Relative tick paths resolve from the absolute Sweep source directory.
- `stest.sh` checks each proof log write, reporter status, and final log write.
- Controlled `tee` failure proof returned nonzero.
- Broken RunReport links, BtBot examples, and stale Sweep contracts are corrected.
- Strict config proof rejects the old `[btrunner]` key.
- Profiled Grid log, JSON, exact counts, and decimals are cited in `wiki/PERFORMANCE.md`.
- Reserved root path `nul` is absent.
- Focused tests passed: btsweep 1.861s, botspec 1.784s, config 0.151s.
- Full Go tests and vet passed with Go 1.26.5, `CGO_ENABLED=0`, and `-tags noasm`.
- Every changed Go file has zero `gofmt -d` output.
- Shell syntax, direct argument rejection, stale-name, inventory, filesystem, and diff checks passed.
- Observer `stest.sh -bot 9` passed 1/1 in 5,835 ms; replay was 2,155 ms.
- Observer proof: `workspace/logs/nuubot5-stest-s6-b9-1-20260726T110832Z.log` and matching JSON.
- Grid Baseline 1 passed through `./stest.sh -sweep 10`: 1,982 Trades, 4,697 Orders, 2,636 Fills, and exact accepted finance.
- Grid Baseline 1 measured BtBot 80,225 ms, loop 75,049 ms, and 142,471.675 MB allocation.
- Profiled Grid passed through `./stest.sh -bot 14 -pp` with all six nonempty profile artifacts and exact accepted finance.
- Profiled Grid measured BtBot 82,880 ms, loop 77,930 ms, and 142,514.548 MB allocation.
- Final direct full tests, full vet, shell syntax, diagnostics, stale-name, and diff checks passed.
- Commit and push are authorized and next.
- Hyperliquid filled acknowledgement is now distinct from fee-complete local reconciliation.
- `wiki/design/hyperliquid/exchange.md` owns the fee-completion and Grid/Hedge STOP contract.
- Generic Recon points to that fact and retains fee-incomplete filled Orders as reconciliation-pending.
- `Fill.HasFee` proves fee presence; zero fees and negative rebates remain valid.
- Hyperliquid creates no synthetic Fill from a filled acknowledgement lacking Venue TID.
- A TID-bearing Fill without fee remains immutable execution evidence with incomplete metadata.
- Reconciliation retains fee-incomplete timestamp anchors and queries bounded repair windows around them.
- Grid/Hedge STOP remains reconciling while any closure Fill is missing or fee-incomplete.
- Hyperliquid fee transitions require per-Fill identity logs and per-cycle aggregate telemetry.
- Fee telemetry identifies Venue, network, Account, symbol, both cursor ranges, rows, durations, cap splits, matches, errors, and pending ages.
- Fee-resolution lag measures first local missing-fee observation through successful enrichment.
- Testnet and mainnet require separate observed fee-completion baselines.
- Delayed-fee proof must advance the normal cursor beyond the Fill timestamp, then enrich that exact TID once.
- One advancing Fill cursor is insufficient and may permanently hide delayed Hyperliquid fee evidence.
- Recon requires independent new-Fill discovery and pending-fee repair queries, merged and deduplicated by Venue TID.
- Fee repair uses bounded windows around unresolved timestamps, never one growing oldest-pending-to-present query.
- Every physical Fill-history request produces one chartable telemetry entry with query kind, range, rows, duration, cap, matches, enrichment, and error.
- Account owns one `ReconTelemetry` outcome containing all `FillQueryTelemetry` entries for each successful, failed, or skipped invocation.
- Account publishes that outcome once; `Telemetry()` only copies it without Ledger traversal.
- Runner heartbeat is configurable and defaults to ten seconds.
- After every reconciliation, the process owner persists `ReconTelemetry` according to configuration.
- ReconTelemetry identifies `recon_kind` as `standard`, `sweep`, `recovery`, or `startup`.
- All reconciliation kinds share one telemetry schema; `recon_kind` is a filtering and charting tag only.
- Generic Recon links to the Telemetry owner and Hyperliquid per-pull measurement contract.
- One persisted ReconTelemetry value may contain multiple Fill-query entries.
- Filled acknowledgements without TID retain the Order acknowledgement timestamp as their repair anchor.
- Every admitted Fill observation must classify as added, enriched, or unchanged duplicate without silent omission.
- Dual-Recon focused Account, Ledger, Executor, and BotSpec tests pass with Go 1.26.5 and `-tags noasm`.
- Full Go tests and vet pass with `CGO_ENABLED=0`, Go 1.26.5, and `-tags noasm`.
- Before retirement, canonical Recon and Recon2 contained the same unoptimized algorithm.
- Holistic `wiki/design/concepts/recon.md` draft is complete and awaiting replacement by the implementable rewrite.
- The draft opens with only the corrected Recon concept, polling loop, and eight-step process.
- Read-only holistic impact assessment completed across lifecycle, domain, Venue, persistence, telemetry, STOP, tests, and cutover.
- Assessment found Grid startup, stopping reconciliation, delayed closed-Trade fees, Venue completeness, reload indexes, dirty persistence, telemetry, and migration blockers.
- Removed unproven 1,000/2,000 container preallocation; target maps and sets grow dynamically.
- Preallocation or buffer reuse requires focused proof after complete-Ledger cloning is removed.
- Recon design now has ten steps with separate Trade and Account Snapshot updates.
- Recon remains one synchronous Account process; it owns no loop, scheduler, lifecycle, or state machine.
- Controller or Bot decides to stop; BotCycle coordinates; Executor acts; Account and Ledger report facts.
- Adversarial holistic design review round one completed with verdict BLOCKED and ten accepted blockers.

### TODO

- Replace `wiki/design/concepts/recon.md` with the clean implementable Init and ten-step Recon specification. DONE.
- Fix accepted adversarial round-one blockers through complete section rewrites. DONE.
- Verify the 263-line minimal draft against current Recon source. DONE.
- Verify every current Recon2 action maps into the ten-step design. DONE.
- Add the exact ten-step source-comment layout contract. DONE.
- Commit and push complete checkpoint `b360335`. DONE.
- Define Section 1 as Init plus complete Recon flow, including start and skip decisions. DONE.
- Rewrite canonical Recon with dynamic indexes and touched-record staging. DONE.
- Preserve Recon2 unchanged as the behavior and performance control; hashes verified. DONE.
- Adversarial canonical Recon review round one found eight accepted current-scope blockers. DONE.
- Fixer agent launches were canceled before sessions started; no fixer is active. DONE.
- Recovered canceled fixer changes and removed its stray `NUL` artifact. DONE.
- Fixed one stale `ReconSnapshot` call; focused tests pass. DONE.
- Full tests, full vet, diff check, and Recon2 hash proof pass after fixes. DONE.
- Added mandatory worker ownership, 30-second health, 10-minute reporting, cancellation, replacement, and no-idle rules to `AGENTS.md`. DONE.
- No delegated worker is active; root owns current verification. DONE.
- Earlier adversarial implementation review round two was superseded by later completed reviews. STOPPED.
- Bot 14 exploratory runs used the current binary but are not accepted paired proof. DONE.
- Implemented create-only `nuubot-bt-sweep`; Server and CLI remain placeholders. DONE.
- Generated full immutable configs from templates: Sweep 11 Bot 15 Recon1; Sweep 12 Bot 16 Recon2. DONE.
- Paired 1x exact parity passed: zero Trade, Order, Fill, finance, equity, or drawdown differences. DONE.
- Recon1 measured 76,640 ms, 108,227.624 MB allocation, and 750 GCs. DONE.
- Recon2 measured 88,193 ms, 146,026.709 MB allocation, and 1,092 GCs. DONE.
- Recon1 improved BtBot time 13.1%, allocation 25.9%, and GC count 31.3%. DONE.
- Added cumulative Recon calls, clean skips, executions, successes, and failures. DONE.
- Added one indexed JSON `telemetry_event` table tagged by kind, frequency, owner, and parent. DONE.
- Persist Bot, BotCycle, Executor, Account, Ledger, and Simulator end telemetry. DONE.
- Recon1 and Recon2 each recorded 277,704 calls, 496 clean skips, 277,208 successes, and zero failures. DONE.
- Both recorded 32d 02:59:30 inside BotCycles and 59d 21:00:29 outside BotCycles. DONE.
- Latest Recon1 measured 77,118 ms, 108,227.373 MB allocation, and 752 GCs. DONE.
- Latest Recon2 measured 91,058 ms, 146,026.869 MB allocation, and 1,091 GCs. DONE.
- Profiled Recon1 at `workspace/perf/profiles/stest-s11-b15-20260726T172917Z/`. DONE.
- Confirmed Recon consumes 42.42 cumulative CPU seconds and remains allocation-bound. DONE.
- Confirmed open-position BBOs mark Account dirty and trigger full Recon despite unchanged execution evidence. DONE.
- Wrote ranked analysis at `.audits/07-27-recon1-performance-audit.md`. DONE.
- User reviewed and rejected nested domain snapshots for Recon, persistence, reporting, and terminal publication. DONE.
- Removed nested Trade, Order, and Fill snapshots from canonical Recon1 comparison and finance traversal. DONE.
- Canonical Recon1 now mutates Ledger-owned records directly without rollback cloning. DONE.
- Focused Ledger, Account, and Executor tests pass after direct mutation. DONE.
- Deleted Trade, Order, and Fill snapshot APIs and all consumers. DONE.
- Deleted terminal Ledger graph reconstruction and publication. DONE.
- Ledger terminal Result now contains only flat counts, cursors, and finance summary. DONE.
- Runtime child access now follows TradeID to Orders and OrderID to Fills through owned indexes. DONE.
- `PersistMode == none` skips Ledger, Trade, Order, Fill, and Simulator rows; terminal BtBot proof remains separate. DONE.
- Recon2 now stores marked finance directly without snapshots. DONE.
- Full Go test suite passes with `CGO_ENABLED=0` and `-tags noasm`. DONE.
- Bot 15 passed exact accepted Trades, Orders, Fills, finance, equity, and drawdown after direct Recon mutation. DONE.
- Bot 15 improved from 77,118 ms to 62,802 ms; loop improved from 71,698 ms to 57,022 ms. DONE.
- Allocation improved from 108,227.373 MB to 64,785.435 MB; GC runs improved from 752 to 456. DONE.
- Profiled snapshot-free Bot 15 at `workspace/perf/profiles/stest-s11-b15-20260726T184055Z/`. DONE.
- Profiled Bot 15 passed exact Trades, Orders, Fills, finance, equity, and drawdown. DONE.
- Snapshot-free profiled Bot 15 measured 50,906 ms BtBot, 46,733 ms loop, 46,715.107 MB allocation, and 320 GCs. DONE.
- Original Recon1 to snapshot-free change: BtBot -34.0%, loop -34.8%, allocation -56.8%, GC -57.4%. DONE.
- Confirmed `persist_mode = none` result contains zero Ledger, Trade, Order, Fill, or Simulator tables. DONE.
- Updated `stest.sh` to validate summary proof under `none` and domain rows under `max`. DONE.
- Ranked remaining bottlenecks in `.audits/07-27-recon1-remaining-performance-audit.md`. DONE.
- Updated Recon, Account, Ledger, Trade, Order, and Performance documentation. DONE.
- Full tests, full vet, shell syntax, diff check, and project diagnostics pass. DONE.
- Recon frequency, local dirty classification, and future WebSocket-driven exchange-dirty detection are deferred. DONE.
- Residual mutation-time Ledger cloning contract reviewed and direct touched-record replacement approved. DONE.
- First Agent 1 launch was canceled by the intervening user instruction; no coder session started and no Ledger source changed. DONE.
- Replacement Agent 1 removed whole-Ledger clones from `CreateTrade`, `AddOrders`, and `RecordSubmit`. DONE.
- `none` performs direct mutation without persistence preparation; `max` writes touched identity, Trade, and Order rows transactionally. DONE.
- Root reran focused ownership, touched-row persistence, and full Ledger package tests successfully. DONE.
- Agent 2 adversarially reviewed clone removal and found no blocker. DONE.
- Root aligned reason comments and Ledger Program Flow with direct touched-record mutation. DONE.
- Initial full test exposed one clone-era Account fault expectation; root updated it to prove untrusted memory and durable transaction rollback separately. DONE.
- Focused Account fault proof, full tests, full vet, formatting, shell syntax, and diff checks pass. DONE.
- Bot 15 1x passed exact accepted domain and finance parity after clone removal. DONE.
- Bot 15 1x measured 51,074 ms BtBot, 46,360 ms loop, 46,331.327 MB allocation, and 331 GCs. DONE.
- First background 5x launch exited before starting and produced no attempt, log, or sentinel. DONE.
- Root replaced it with direct canonical `./stest.sh -bot 15 -runs 5`; all five attempts passed identical results. DONE.
- Bot 15 5x averaged 50,298.4 ms BtBot, 46,238.6 ms loop, 46,324.660 MB allocation, and 331 GCs. DONE.
- Final Bot 15 `-pp` passed exact accepted domain and finance parity. DONE.
- Final profile measured 49,514 ms BtBot, 45,366 ms loop, 46,354.773 MB allocation, and 318 GCs. DONE.
- Final profile is `workspace/perf/profiles/stest-s11-b15-20260727T033944Z/`. DONE.
- `cloneTrades` has zero samples in the canonical Bot 15 profile. DONE.
- Compared with the immediate snapshot-free profile, allocation fell 375.563 MB and profiled runtime fell 1,343 ms. DONE.
- Updated Ledger design, performance history, and remaining-performance audit. DONE.
- Removed one stray untracked `NUL` artifact created during tooling and verified it absent. DONE.
- Active clone-removal scope has no remaining authorized TODO. DONE.
- Agent 1 implemented the approved three-consecutive-Recon-failure Controller barrier and focused tests. DONE.
- Agent 2 reviewed the barrier; root rejected its pre-existing Recon2 drift finding and accepted missing Controller recovery proof. DONE.
- Root added Controller recovery proof; focused Account, Executor, BotCycle, and Controller tests pass. DONE.
- Root completed current profile analysis. DONE.
- Full Go tests, full vet, formatting, and diff checks pass after the Recon barrier. DONE.
- Bot 15 passed exact accepted domain and finance parity after the Recon failure barrier. DONE.
- Barrier proof measured 50,260 ms BtBot, 45,514 ms loop, 46,351.126 MB allocation, and 330 GCs. DONE.
- Result proof is `workspace/logs/nuubot5-stest-s11-b15-1-20260727T041230Z.log` and matching JSON. DONE.
- Active three-consecutive-Recon-failure barrier scope has no remaining authorized TODO. DONE.
- User approved performance Targets 1–5: incremental Ledger totals, split Trade refresh, incremental Simulator position, lower Order evidence allocation, and normalized Simulator matching. BELAYED IN ZED; EXTERNAL TOOL OWNS CONTINUATION.
- Zed root, coder agents, reviewers, and orchestrators are stopped. No monitorable worker is active. DONE.
- Planner split Targets 1–5 into six ordered chunks: Ledger totals, Trade refresh, Simulator position, Order comparison, Order representations, and Simulator matching. DONE.
- Two Chunk 1 tool calls were canceled before returning session identities. Root initially reported no code changed without verifying the working tree; that report was wrong. DONE.
- Current source proves one canceled worker wrote Chunk 1 changes despite returning no session identity. DONE.
- Root inspected current Chunk 1 work and ran focused Ledger and Account tests successfully. DONE.
- User authorized Codex to finish Chunk 1 acceptance and implement Chunks 2–5 in order. DONE.
- Root completed delegated Chunk 1 acceptance. DONE.
- Chunk 1 focused, full-test, full-vet, formatting, shell, and diff proof passes. DONE.
- Bot 15 and Bot 16 preserve exact accepted domain counts and finance under partial Chunk 1. DONE.
- First Chunk 1 repair regressed Bot 15 from 46,354.773 MB to 65,750.272 MB allocation. DONE.
- Root rejected that repair after its profile exceeded the 180-second process limit. DONE.
- Agent 1 made failed-phase Summary repair error-only and removed 528 B plus 18 allocations from the focused seam. DONE.
- Corrected Chunk 1 Bot 15 passes exact parity at 53,101.770 MB, 55,346 ms, and 356 GCs. DONE.
- Corrected profile is `workspace/perf/profiles/stest-s11-b15-20260727T052216Z/`. DONE.
- Old `ReconSummary` traversal is absent; delta maintenance remains 6,746.997 MB above the pre-Chunk baseline. DONE.
- Independent Chunk 1 acceptance review passed with no material finding. DONE.
- Review accepted the intermediate regression because its remaining cost belongs to approved Chunk 2. DONE.
- Added project `coder`: `gpt-5.6`, medium reasoning. DONE.
- Added project `auditor`: `gpt-5.6`, medium reasoning, read-only. DONE.
- Auditor reviews changed code only, reports major issues only, and passes immediately otherwise. DONE.
- Do not audit Chunks 2–5 individually; run one full changed-code performance audit after Chunk 5. DONE.
- Added project `planner`: `gpt-5.6`, high reasoning. DONE.
- No delegated agent remains active. DONE.
- User reloaded Codex so exact project agent profiles are active. DONE.
- Updated `wiki/PERFORMANCE.md` with accepted Chunk 1 proof after reload. DONE.
- Implemented and accepted Chunk 2 Trade refresh split. DONE.
- Chunk 1 holistic review, full tests, Bot 15, Bot 16, profiling, and acceptance passed. DONE.
- Chunk 2 passed focused, full, Bot 15, and profiled proof. DONE.
- Chunk 3 maintains Simulator position and finance per accepted Fill. DONE.
- Chunk 3 focused tests, full tests, vet, formatting, shell, and diff checks pass. DONE.
- Chunk 3 Bot 15 and profiled Bot 15 preserve exact accepted domain and finance. DONE.
- Chunk 3 profile measured 47,379 ms BtBot, 43,055 ms loop, 44,336.333 MB allocation, and 294 GCs. DONE.
- Chunk 3 profile is `workspace/perf/profiles/stest-s11-b15-20260727T060536Z/`. DONE.
- All six Chunk 3 profile artifacts are nonempty and readable. DONE.
- Chunk 4A Order comparison and Fill ownership reads allocate nothing. DONE.
- Chunk 4A focused tests, full tests, vet, formatting, shell, and diff checks pass. DONE.
- Chunk 4A Bot 15 and profiled Bot 15 preserve exact accepted domain and finance. DONE.
- Chunk 4A profile measured 44,039 ms BtBot, 39,910 ms loop, 41,489.475 MB allocation, and 274 GCs. DONE.
- Chunk 4A profile is `workspace/perf/profiles/stest-s11-b15-20260727T062514Z/`. DONE.
- All six Chunk 4A profile artifacts are nonempty and readable. DONE.
- Chunk 4B caller-buffer and Simulator-cache work was rejected and fully reverted. DONE.
- Canonical Recon uses bulk Venue evidence; exact missing-active Order status is exception handling. DONE.
- User approved one `OrderStatusQueries` Recon telemetry count to measure that exception. ACTIVE.
- Performance Target 6, telemetry retention and terminal publication, is documented and deferred. DONE.
- Run accepted Recon1 5x after the clone-removal tuning round. DONE.
- Fresh snapshot-free Bot 15 `-pp` repeat passed exact domain and finance proof. DONE.
- Fresh repeat measured 50,857 ms BtBot, 46,689 ms loop, 46,730.336 MB allocation, and 320 GCs. DONE.
- Fresh profile is `workspace/perf/profiles/stest-s11-b15-20260726T184654Z/`. DONE.
- CPU, allocation, heap, block, mutex, scheduler, synchronization, and syscall evidence is readable. DONE.
- Recon remains dominant at 26.27 CPU seconds and 30.83 GB cumulative allocation. DONE.
- SQLite statement delay was about 0.27 seconds; database waiting is not the bottleneck. DONE.
- Residual `cloneTrades` costs 253.57 MB outside Recon1 through mutation candidate staging. DONE.
- Refreshed `.audits/07-27-recon1-remaining-performance-audit.md` and `wiki/PERFORMANCE.md`. DONE.

## Performance Implementation Contract

Codex owns authorized Chunks 2–5 continuation.

Recon2 and Bot 16 are retired. Do not run Bot 16.

Each remaining chunk runs one Bot 15 proof and one Bot 15 profiled proof.
Do not run 5x stability.

### Accepted Chunk 1

Confirmed current source contains:

```text
Ledger.summary cached state
Init and reload Summary rebuild
CreateTrade Summary insertion delta
AddOrders before-and-after Summary delta
RecordSubmit touched-Trade Summary deltas
ReconAttempt original Trade summaries
UpdateReconTrades Summary replacement deltas
ReconSummary and Summary cached reads
complete-traversal Summary test oracle
zero-allocation Summary read tests
```

Focused proof passed:

```text
CGO_ENABLED=0 go test -count=1 -tags noasm ./internal/ledger ./internal/account
```

Chunk 1 passed review, static proof, Bot 15, Bot 16, profiling, exact parity,
and canonical documentation.

### Approved Ordered Chunks

```text
Chunk 1  Maintain Ledger Summary totals incrementally.
Chunk 2  Split structural Trade refresh from current-mark refresh.
Chunk 3  Maintain Simulator position, entry, realized PnL, and fees incrementally.
Chunk 4A Make Order comparison state allocation-free.
Chunk 4B Dropped: caller buffers and Simulator public-output caching violate the Venue boundary.
Chunk 5  Normalize exact Simulator matching keys without rounding.
```

#### Chunk 2

- Preserve every scheduled Venue poll and current request order.
- Structural refresh only Trades touched by changed Order or Fill evidence.
- Mark refresh only open Trades using stored exposure.
- Closed Trades remain static.
- Preserve delayed-fee completion, cursor, generation, and Snapshot behavior.
- Prove Bot 15 parity, then profile against accepted Chunk 1.

#### Chunk 3

- Maintain Simulator signed size, entry price, realized PnL, and fees when Fills are accepted.
- Use maintained position for AccountState and reduce-only quantity.
- Rebuild and validate once from persisted Fills during reload.
- Preserve Fill sequence, StartPosition, ClosedPnL, Direction, JSON, schema, and persistence behavior.
- Prove Bot 15 parity, then profile against accepted Chunk 2.

#### Chunk 4A

- Add scalar allocation-free Order comparison state.
- Remove `Order.copyInput` and repeated decimal equality work from canonical Recon comparison.
- Preserve duplicate idempotency, fee enrichment, zero fees, negative rebates, and terminal transitions.
- Prove Bot 15 parity, then profile against accepted Chunk 3.

#### Chunk 4B

- Dropped without replacement.
- Current bulk Venue polling and detached outputs remain canonical.
- Exact missing-active status checks remain exceptional and gain one telemetry count.
- Account scratch, callbacks, maps, caches, storage, and domain objects never cross into Venue.
- Target 4 is partially complete through accepted Chunk 4A.

#### Chunk 5

- Build private exact comparison-only price keys once for Orders and once per BBO.
- Canonicalize coefficient and base-ten exponent without rounding.
- Comparison sign must equal `decimal.Decimal.Cmp` for all positive values.
- Preserve original Decimal values, Fill prices, fees, persistence text, outputs, IOC rules, and matching order.
- Add deterministic matrix, threshold-boundary, and fuzz parity tests.
- Prove Bot 15 parity, then run one final profile.

### Mandatory Boundaries

```text
ALWAYS use CGO_ENABLED=0 and -tags noasm.
Do not change scheduled Recon frequency or Venue request sequence.
Do not add WebSocket dirty shortcuts.
Do not change the three-failure Controller barrier.
Do not change persistence modes or database schemas.
Do not change public finance equations.
Do not optimize telemetry, RunReport, ResultPublisher, or terminal publication.
Do not add dependencies, CGO, unsafe, assembly, or runtime internals.
Do not commit or push without fresh explicit user authority.
Update HANDOFF.md after every accepted chunk.
```

### Target 6 Deferred

Do not implement during Targets 1–5:

```text
BtBot telemetry        1.61 GB cumulative allocation
Result publication     1.67 GB cumulative allocation
Result publication     3.60 s CPU
```

Canonical owner: `wiki/PERFORMANCE.md`.

### Exact Acceptance Baseline

```text
Trades                 1,982
Orders                 4,697
Fills                  2,636
Gross PnL              -15.13202
Fees                   42.28805409
Net PnL                -57.420074089999999993851
Ending equity          942.579925910000000006149
Maximum drawdown       75.791979199999999992245
Recon calls            277,704
Recon clean skips      496
Recon successes        277,208
Recon failures         0
```

Performance baseline before partial Chunk 1:

```text
BtBot                  49,514 ms
Historical loop       45,366 ms
Total allocation      46,354.773 MB
GC runs                318
Profile                workspace/perf/profiles/stest-s11-b15-20260727T033944Z/
```

Latest normal-path barrier proof:

```text
Result log             workspace/logs/nuubot5-stest-s11-b15-1-20260727T041230Z.log
Suite report           workspace/logs/nuubot5-stest-s11-b15-1-20260727T041230Z.json
```

### Final Acceptance Gate

After all chunks:

```text
CGO_ENABLED=0 go test -count=1 -tags noasm ./...
CGO_ENABLED=0 go vet -tags noasm ./...
bash -n stest.sh
git diff --check
./stest.sh -bot 15
./stest.sh -bot 15 -pp
```

Require one final adversarial review covering finance, lifecycle, reload, persistence, aliasing, exact matching, polling, deferred Target 6, and fake performance proof.

### PENDING USER APPROVAL

- None for the active clone-removal scope.
- Add the chief-of-staff delegation rule to `AGENTS.md`. DONE.
- Fail stability when deterministic results differ.
- Preserve every stability-attempt database.
- Add one rolling replay-input checksum.
- Compare replay proof under Go 1.26.5 and Go 1.25.12.
- Design the deferred equity/balance snapshot issue: configurable tiered retention and rollups.

## Superseded Ledger Orchestrator Restart Contract

Do not execute this historical contract. The `External Performance Tool Restart Contract` above replaces it.

- Historical planning, implementation, testing, profiling, audit, and fixing authority applied to the older Ledger chunk plan.
- Do not commit or push.
- The old orchestrator is not running.
- The historical plan remains below only as evidence.

Cleared plan:

```text
.audits/07-26-ledger-chunk-plan.md
```

Plan reviews:

```text
.audits/07-26-ledger-chunk-plan-review-1.md
.audits/07-26-ledger-chunk-plan-review-2.md
.audits/07-26-ledger-chunk-plan-review-3.md
```

The third review found one startup deadlock. The planner corrected it after the review cap:

- Chunk 1 is inventory-only.
- Chunk 2 repairs the known `ledger_test.go` placeholder and finishes green.
- Chunk 5 repairs known `store.go` hook formatting before SQL proof.
- Chunks 2–5 must establish the first fully green baseline.
- Root verified these corrections directly in the final plan.

Required orchestrator workflow for every chunk:

1. Spawn one coder with only that chunk.
2. Coder implements and runs focused proof.
3. Spawn one adversarial implementation auditor against exact chunk intent.
4. Spawn a fixer for accepted blockers.
5. Re-audit after fixes, maximum three audit rounds per chunk.
6. Advance only when all blockers are cleared or a genuine external blocker is proven.
7. Update `HANDOFF.md` after every completed chunk.
8. Update `.audits/07-26-ledger-orchestration.md` after every completed chunk.
9. Send the user a status update after every completed chunk.

After all chunks:

1. Run exact behavior, finance, Trade, Order, Fill, stability, and performance proof.
2. Use `CGO_ENABLED=0` and `-tags noasm`.
3. Spawn an overall adversarial implementation auditor.
4. Fix accepted blockers and re-audit, maximum three rounds.
5. Reject fake, incomplete, or incomparable proof.
6. Report exact before-and-after allocation, runtime, finance, and domain counts.

Implementation boundaries:

- Preserve exact trading logic and accepted financial results.
- This is a performance redesign.
- Do not redesign Simulator.
- Do not implement deferred future Runner heartbeat, live cleanup, grace, or telemetry persistence scope.
- Preserve unrelated uncommitted work.
- Keep existing ResultPublisher and BtBot rename-related files frozen where the plan requires it.

## Heartbeat Reconciliation Documentation Proof

- Canonical reconciliation, Account, Ledger, Runner, and telemetry pages now separate current implementation from approved live targets.
- Runner owns one drift-free ten-second heartbeat and reads time once per heartbeat.
- Every future live heartbeat appends one cheap JSON liveness and reconciliation row.
- Normal recon, unresolved cleanup, failure handling, atomic publication, and capacity reserves are documented.
- Hyperliquid Fill and Order history boundaries are documented without adding routine `historicalOrders` use.
- Decision-critical Account state remains synchronous; observability-only cadence changes require proof.
- Historical telemetry retention, downsampling, cleanup default, safety boundary, and escalation thresholds remain deferred.
- Documentation-only change touched the six authorized canonical files.
- No source, audit report, command, replay, or test changed or ran.
- Next step: reassess Ledger changes in execution order against these canonical contracts.

## Codex Paste Diagnosis

- Live VS Code is `1.129.1`; live terminal Codex CLI is `0.144.1`.
- Codex strict-config doctor passed with 17 OK and zero warnings or failures.
- Live Codex config sets `disable_paste_burst = false`.
- Current clipboard has 16 lines, 41 tabs, zero CR, and 15 LF characters.
- Temporary VS Code paste setting was restored without completing its A/B test.
- User keybindings define only terminal `Shift+Enter`; neither paste chord is overridden.
- Generic terminal paste binds Ctrl+V and Ctrl+Shift+V to the same command.
- PowerShell's more specific Ctrl+V route sends control character 22 into the terminal.
- Ctrl+Shift+V directly runs VS Code terminal paste.
- The route difference did not change the observed malformed paste.
- An earlier fresh TUI test ran `0.144.6`.
- Immediate pre-`0.140.0` version was `0.144.4`.
- Local failures exist on `0.140.0`, `0.144.4`, `0.144.6`, `0.145.0`, and `0.146.0-alpha.10`.
- `paste_burst.rs` is byte-identical from `0.144.1` through `0.146.0-alpha.10`.
- `handle_paste` is identical across those versions.
- Every compared version uses a 1,000-character large-paste placeholder threshold.
- Current submitted samples were 337 and 851 characters, below that threshold.
- The open Windows bug predates the `0.144` line.
- No Codex configuration or repository source was changed.

## Deferred Equity/Balance Snapshot Issue

- Keep decision-critical Account state current and synchronous.
- Keep reconciliation telemetry minimal: work counts, dirty identities, rows written, duration, and errors.
- Separate historical equity and balance snapshots from the reconciliation mutation path.
- Use configurable retention tiers, initially `1m -> 5m -> 15m -> 1h`.
- Retain recent high-resolution values and progressively lower-resolution older values.
- Rollups should preserve open, close, minimum, maximum, average, and sample count.
- Calculate current Account values once, then retain or roll them up when tier timers become due.
- Tune and implement this later. It is outside the current Ledger reconciliation redesign.

## Telemetry and RunReport

- Telemetry and RunReport are separate features.
- Every participating object owns custom private State.
- `Telemetry()` reads State without mutation and composes child telemetry once.
- BtBot collects telemetry in memory and publishes it terminally.
- Runner will persist the same telemetry periodically.
- RunReport consumes existing Results and telemetry.
- RunReport remains one top-level bolt-on with no domain-package dependency.
- User intent must precede proposed design in both owning wiki pages.
- `wiki/implement.md` records the reusable intent-design-audit-build workflow.
- Design audit round one failed with three accepted material blockers.
- Terminal ordering, memory sampling, and child-to-suite transport were corrected.
- Design audit round two passed with no material finding or bloat.
- Account, Executor, BotCycle, Controller, and BtBot own one pull-only telemetry path.
- BtBot retains telemetry in memory and publishes once after shutdown.
- `internal/report` builds RunReport and SuiteReport without domain Report methods.
- `nuubot-report` renders standardized tables and writes one SuiteReport JSON.
- ResultPublisher atomically stores telemetry, RunReport, Results, Ledger, and Simulator evidence.
- Implementation audit round one found false pre-recon equity and obsolete Grid `close` Orders.
- Account telemetry now reports explicit absence until its first observation.
- Grid and Trade shutdown now use only canonical `stop` Orders.
- Implementation audit round two passed with no material finding or bloat.

## Published Baseline

- Branch: `main`.
- `HEAD` and `origin/main`: `ff638268b0f08c8d9a7603d061cbbf179726fc26`.
- Message: `feat: implement bot controller trade baseline`.
- TradeBot tests, vet, replay, stability, and adversarial audit passed.

## Active Implementation

- Exact database BotSpecID, TOML bytes, and SHA-256 are authoritative.
- TOML remains the only persisted BotConfig representation.
- AppConfig no longer owns Bot behavior.
- BotSpec admits exact typed Config and builds immutable Bot definitions.
- Implemented BotSpecs are `macross_observer_bot` and `macross_trade_bot`.
- Controller replaced Runtime without compatibility code.
- Controller owns persistent Signaler, Risks, and zero or one active BotCycle.
- `max_cycles = 999` permits sequential cycles.
- Maximum concurrent BotCycles remains structurally fixed at one.
- Signal uses `NoAction`, `StartCycle`, and `StopCycle`.
- Risk uses immutable RiskInput and typed decisions.
- Executors own fixed side, capital, sizing, and distinct resources.
- Controller carries terminal Account equity into the next cycle.
- Every Risk assesses the same immutable input before Controller acts.
- Stronger Risk decisions dominate weaker decisions.
- Resource equity uses the exact Venue, network, Account, and symbol tuple.
- ResultPublisher writes Controller, cycle, Executor, Signal, Risk, and trading evidence.
- ResultPublisher also writes `telemetry_sample` and `run_report`.
- Results preserve exact admitted BotConfig TOML and hash.
- Final publication preserves both none-mode and maximum-mode Account children.
- Simulator replay loads no private credentials.
- Legacy Order role `close` is removed without fallback or compatibility.

## Local Database

- Database: `workspace/db/nuubot.db`.
- Recoverable pre-hardcut backup: `workspace/db/nuubot.pre-botspec-20260725.db`.
- Thirteen Bot rows contain exact TOML and valid hashes.
- Sweep 6 Bot 9 uses `macross_observer_bot`.
- Sweep 9 Bot 13 uses `macross_trade_bot`.
- Stored templates now use `max_cycles = 999`.
- Latest integrity check: `ok`.

## Verified Proof

- Full Go tests and vet pass with `CGO_ENABLED=0` and `-tags noasm`.
- TradeBot processed 7,948,800 ticks and 794,880 control passes.
- TradeBot completed 193 cycles, 193 Trades, 626 Orders, and 386 Fills.
- TradeBot capital was 1,000 USDC.
- TradeBot net PnL was -3.90459332761 USDC.
- TradeBot ending equity was 996.09540667239 USDC.
- TradeBot maximum drawdown was 4.21244716452 USDC.
- Cycle 2 starting equity equals Cycle 1 terminal equity.
- Result BotConfig equals the exact stored database TOML and hash.
- Result integrity and foreign-key checks passed.
- TradeBot passed 2 of 2 and 10 of 10 fresh processes.
- Observer passed 2 of 2 and 10 of 10 fresh processes.
- TradeBot 10x log: `workspace/logs/nuubot5-trtest-s9-b13-10-20260724T202751Z.log`.
- Observer 10x log: `workspace/logs/nuubot5-rtest-s6-b9-10-20260724T202915Z.log`.
- Post-audit TradeBot 1x passed in 9,393 ms process and 4,248 ms replay.
- Post-audit TradeBot log: `workspace/logs/nuubot5-trtest-s9-b13-1-20260724T204016Z.log`.
- Post-audit Observer 1x passed in 1,788 ms process and 1,694 ms replay.
- Post-audit Observer log: `workspace/logs/nuubot5-rtest-s6-b9-1-20260724T204033Z.log`.
- Implementation audit round one found maximum-mode final publication data loss.
- The accepted finding received one focused failing test and owning-path fix.
- Implementation audit round two passed with no material finding or bloat.
- Full Go tests and vet pass after Telemetry, RunReport, and stop-role hardcut.
- Observer 1x report passed with 794,881 telemetry samples.
- TradeBot 1x report passed with 193 cycles, 193 Trades, and 47 stop Orders.
- Grid 1x report passed with 50 cycles, 1,982 Trades, and 733 stop Orders.
- Active-cycle zero-equity telemetry defects: zero.
- Maximum-drawdown decreases: zero.
- Legacy `close` Orders: zero.
- Grid stability passed 2 of 2 and 10 of 10.
- Grid 10x BtBot average was 76,688.2 ms.
- Grid 10x historical-data-loop average was 72,237.3 ms.
- Grid 10x report: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T093515Z.json`.

## Completed

- Approved TradeBot architecture implementation.
- Full three-month TradeBot proof.
- TradeBot and Observer 2x and 10x stability.
- Final adversarial implementation audit.
- Post-audit static and replay proof.

## TODO

- None.

## Pending User Approval

- Commit and push the complete current worktree.

## Approved Grid Executor Contract

- Exact BotSpec is `macross_grid_bot`.
- Range uses cycle-start price minus 5 percent through plus 5 percent.
- Thirty total levels include two non-enterable boundaries and 28 active Trades.
- Deployed capital is 95 percent of cycle-start Executor equity.
- Every active level satisfies Venue minimum order value and expected-PnL gates.
- Initial long entries use current price or lower; short entries reverse this.
- Levels retain immutable calculations and mutable Trade/submission state.
- Boundary exit cancels TP, SL, then entry Orders before flattening Trades.
- One initial submission plus two proven-safe retries precedes fatal graceful exit.
- CLOID uses `order_level`; Trade and future Hedge Executors use Level zero.
- Every Grid Executor start logs its validated level table before order placement.
- Every test records runtime, domain counts, PnL, equity, drawdown, and evidence paths.
- Minimum price-gap-distance remains deferred and absent from current config/code.

## Grid Implementation

- Exact `macross_grid_bot` owns one arithmetic GridExecutor.
- Thirty total Levels create 28 active Level Trades.
- Grid deploys 95 percent of cycle-start equity.
- Every Level stores calculations, Trade state, submission state, and timestamps.
- Expected PnL gates include entry and exit commissions.
- CLOID identity now uses `order_level` from 0 through 1,023.
- TradeExecutor continues using Level zero.
- Persisted Order `order_pos` remains batch-local.
- ResultPublisher stores `grid_level_result`.
- `stest.sh -bot 14` verifies complete Grid replay evidence.

## Grid Initial Baseline

- Status: INVALID - ERROR found.
- Audit found re-entry sizing could exceed one capital slice.
- Audit found accepted uncertain submissions could be retried.
- Sweep 10 Bot 14.
- Ticks: 7,948,800.
- Controller runs: 794,880.
- Signals: 2,207.
- BotCycles: 50.
- Trades: 1,954.
- Orders: 4,641.
- Fills: 2,578.
- Cancellations: 2,063.
- Closure Orders: 733.
- Retries: 0.
- Round trips: 554.
- BtBot elapsed time: 25,848 ms.
- BtBot historical-data loop elapsed time: 23,908 ms.
- Net PnL: -70.864647459999999999278 USDC.
- Ending equity: 929.135352540000000000722 USDC.
- Maximum drawdown: 88.027421204999999999563 USDC.
- Integrity, foreign keys, Config identity, Levels, final Orders, and final positions passed.
- Log: `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T064809Z.log`.
- Baseline: `wiki/baselines/macross-grid-bot.md`.
- Grid stability passed 2 of 2 and 10 of 10.
- 2x log: `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T065218Z.log`.
- 10x log: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T065316Z.log`.

## Grid Corrected Baseline

- Status: INVALID - ERROR found.
- Boundary-tick TP fills were omitted from `round_trips`.
- Re-entry sizing uses the greater initial or re-entry price.
- Retries require proven non-submission.
- Ticks: 7,948,800.
- Controller runs: 794,880.
- Signals: 2,207.
- BotCycles: 50.
- Trades: 1,954.
- Orders: 4,641.
- Fills: 2,578.
- Cancellations: 2,063.
- Closure Orders: 733.
- Retries: 0.
- Round trips: 554.
- BtBot elapsed time: 26,007 ms.
- BtBot historical-data loop elapsed time: 24,066 ms.
- Net PnL: -69.766463889999999999562 USDC.
- Ending equity: 930.233536110000000000438 USDC.
- Maximum drawdown: 86.609100424999999999246 USDC.
- Integrity, foreign keys, Config identity, Levels, final Orders, and final positions passed.
- Stability passed 2 of 2 and 10 of 10 with identical results.
- 1x log: `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T070934Z.log`.
- 2x log: `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T071011Z.log`.
- 10x log: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T071111Z.log`.

## Grid Final Corrected Baseline

- Status: INVALID - ERROR found.
- Marketable Grid GTC Orders were not matched during submission.
- Account equity and drawdown snapshots stayed stale between Fill events.
- Completed round trips derive from terminal filled TP Orders.
- Ticks: 7,948,800.
- Controller runs: 794,880.
- Signals: 2,207.
- BotCycles: 50.
- Trades: 1,954.
- Orders: 4,641.
- Fills: 2,578.
- Cancellations: 2,063.
- Closure Orders: 733.
- Retries: 0.
- Round trips: 556.
- BtBot elapsed time: 25,397 ms.
- BtBot historical-data loop elapsed time: 23,533 ms.
- Net PnL: -69.766463889999999999562 USDC.
- Ending equity: 930.233536110000000000438 USDC.
- Maximum drawdown: 86.609100424999999999246 USDC.
- Integrity, foreign keys, Config identity, Levels, final Orders, and final positions passed.
- Stability passed 2 of 2 and 10 of 10 with identical results.
- 1x log: `workspace/logs/nuubot5-grtest-s10-b14-1-20260725T072506Z.log`.
- 2x log: `workspace/logs/nuubot5-grtest-s10-b14-2-20260725T072539Z.log`.
- 10x log: `workspace/logs/nuubot5-grtest-s10-b14-10-20260725T072637Z.log`.

## Grid Current Baseline

- Status: CURRENT.
- Signals: 2,208.
- BotCycles: 50.
- Trades: 1,982.
- Orders: 4,697.
- Fills: 2,636.
- Cancellations: 2,061.
- Stop Orders: 733.
- Retries: 0.
- Round trips: 585.
- Net PnL: -57.420074089999999993851 USDC.
- Ending equity: 942.579925910000000006149 USDC.
- Maximum drawdown: 75.791979199999999992245 USDC.
- Telemetry samples: 794,881.
- Result database size: 178,806,784 bytes.
- Stability passed 2 of 2 and 10 of 10 with identical results.
- Baseline: `wiki/baselines/macross-grid-bot.md`.

## Deferred

- Live cross-process Account claims.
- Multi-source replay merge.
- Physical Account and global risk.
- Runner periodic telemetry persistence.
- Server monitoring and recovery.

## Next Action

Commit and push the completed BtBot manual-review hardcut.

User explicitly authorized committing and pushing the complete current worktree.

After push, begin manual code review of `internal/controller/controller.go`.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
