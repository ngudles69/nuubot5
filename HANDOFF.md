# Handoff

Last updated: 2026-07-26 00:22:22 +08:00

## Focus

Commit the complete validated Grid, telemetry, reporting, profiling, audit, and review work.

## Active Task

### DONE

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
- `rtest.sh` remains the repetitive fresh-process stress command and rejects extra arguments.
- `pptest.sh` owns one explicit performance-profiling run.
- `nuubot-btrunner <sweep_id> <bot_id> -pp <prefix>` enables profiling; absence of `-pp` creates no profiles.
- Performance runs capture CPU, trace, heap, allocations, block, and mutex artifacts.
- Performance artifacts live under `workspace/perf/profiles/pptest-s<sweep>-b<bot>-<timestamp>/`.
- Focused command tests, vet, shell syntax, diagnostics, and diff checks passed.
- Normal `rtest.sh 1 6 9` passed in 5,604 ms.
- `pptest.sh 6 9` passed in 7,111 ms with 2,190 ms replay timing.
- All six performance artifacts are nonempty and readable by Go 1.26.5 tools.
- Proven profile: `workspace/perf/profiles/pptest-s6-b9-20260725T152130Z/`.
- Performance, filesystem, command, and BtRunner wiki owners are aligned.
- Grid profiling binary completed successfully in 81,129 ms with 76,356 ms replay timing.
- `pptest.sh` falsely rejected Grid through inherited Observer-specific counts; generic successful-result validation replaced it.
- Grid profiling allocated 144,495.70 MB versus Observer 3,726.29 MB.
- Grid Account reconciliation owns 112,609.46 MB cumulative allocation.
- Grid Trade Orders owns 24,999.31 MB flat allocation.
- Grid math/big allocation owns 22,887.53 MB flat allocation.
- Grid GC consumed 53.19 CPU seconds; replay Loop consumed 73.32 CPU seconds.
- Grid ResultPublisher syscall delay was below one second; database waiting is not the main Grid bottleneck.
- Grid profile: `workspace/perf/profiles/pptest-s10-b14-20260725T152717Z/`.
- All six Grid artifacts are nonempty and readable by Go 1.26.5 tools.
- Adversarial Ledger clone audit completed at `.audits/07-25-ledger-clone-audit.md`.
- Renamed tracked engineering evidence from `audits/` to `.audits/`.
- Filesystem design now owns `.audits/` as development-only review evidence.
- Added concise `wiki/coding/sample.md` and linked it from canonical style and project pages.
- `wiki/reviews.md` contains exactly 91 Go files with explicit review state, date, and hash columns.
- `workspace/perf/` is ignored and excluded from the commit.
- Full Go tests and vet pass with `CGO_ENABLED=0` and `-tags noasm`.
- Build, stress, trade, Grid, and performance shell syntax checks pass.
- Full diff whitespace check passes.
- Complete durable work committed at `HEAD` with message `Add Grid telemetry and performance reports`.

### TODO

- None.

### PENDING USER APPROVAL

- Add the chief-of-staff delegation rule to `AGENTS.md`.
- Fail stability when deterministic results differ.
- Preserve every stability-attempt database.
- Add one rolling replay-input checksum.
- Compare replay proof under Go 1.26.5 and Go 1.25.12.
- Push the committed current worktree.

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

## Telemetry and RunReport

- Telemetry and RunReport are separate features.
- Every participating object owns custom private State.
- `Telemetry()` reads State without mutation and composes child telemetry once.
- BtRunner collects telemetry in memory and publishes it terminally.
- Runner will persist the same telemetry periodically.
- RunReport consumes existing Results and telemetry.
- RunReport remains one top-level bolt-on with no domain-package dependency.
- User intent must precede proposed design in both owning wiki pages.
- `wiki/implement.md` records the reusable intent-design-audit-build workflow.
- Design audit round one failed with three accepted material blockers.
- Terminal ordering, memory sampling, and child-to-suite transport were corrected.
- Design audit round two passed with no material finding or bloat.
- Account, Executor, BotCycle, Controller, and BtRunner own one pull-only telemetry path.
- BtRunner retains telemetry in memory and publishes once after shutdown.
- `internal/runreport` builds RunReport and SuiteReport without domain Report methods.
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
- Grid 10x BtRunner average was 76,688.2 ms.
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
- `grtest.sh` verifies complete Grid replay evidence.

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
- BtRunner elapsed time: 25,848 ms.
- BtRunner historical-data loop elapsed time: 23,908 ms.
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
- BtRunner elapsed time: 26,007 ms.
- BtRunner historical-data loop elapsed time: 24,066 ms.
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
- BtRunner elapsed time: 25,397 ms.
- BtRunner historical-data loop elapsed time: 23,533 ms.
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

User chooses whether to test another workaround or leave Codex paste handling unchanged.

Commit and push require explicit user approval.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
