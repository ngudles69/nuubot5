# BtBot, stest, and Template Audit

Date: 2026-07-26

Status: **FAIL**

Reason: one BLOCKER and multiple MATERIAL findings remain.

## Scope

Audited current tracked and untracked worktree changes for:

- BtRunner-to-BtBot hard rename.
- Config, logging, JSON, SQL, reports, builds, tests, wiki, and review inventory.
- `stest.sh` selection, validation, repetition, profiling, failure handling, paths, and shell safety.
- Deletion parity for `rtest.sh`, `trtest.sh`, `grtest.sh`, and `pptest.sh`.
- Bot and Sweep templates.
- `internal/btsweep` validation and deterministic expansion.
- `cmd/nuubot-bt-sweep` placeholder scope.
- Grid Baseline 1 and profiled evidence.
- Unrelated worktree damage.

Historical `.audits/**` content was excluded from rename-residue findings.

## BLOCKER

### B1. Mandatory test contract names a deleted command

Evidence:

- `AGENTS.md` requires `./rtest.sh N 6 9` for every requested `Nx` test.
- `rtest.sh` is deleted.
- `wiki/bin.md` says `stest.sh` is the only system-test entrypoint.
- `wiki/index.md` still names `rtest.sh` as the canonical proof harness.
- `wiki/design/concepts/filesystem.md` still owns `rtest.sh` and `pptest.sh`.
- `wiki/design/concepts/nuubot-bt-bot.md` still assigns profiling ownership to `pptest.sh`.

Impact:

- The project’s mandatory operator command cannot run after this change.
- Canonical authority conflicts with the new executable contract.
- Deletion parity is incomplete.

Required fix:

- Replace the `AGENTS.md` command contract with the exact approved `stest.sh` equivalent.
- Update every canonical script owner and index reference.
- Keep old scripts deleted; do not add compatibility wrappers unless explicitly approved.

## MATERIAL

### M1. Parameter-free Sweeps are rejected

Evidence:

- `internal/btsweep.collectParameters` rejects zero parameter lists.
- Nuubot3 `_parameter_values` permits zero dimensions.
- Nuubot3 Cartesian expansion emits one Bot per date range with zero dimensions.
- Current one-Bot templates add artificial one-value dimensions to bypass this rejection.

Impact:

- A valid one-Bot Sweep cannot express “use this Bot template unchanged.”
- System templates encode fake variation instead of actual intent.

Required fix:

- Permit an empty `[sweep.parameters]` table.
- Emit one generated Bot per date range when no dimensions exist.
- Add focused deterministic and hash proof.

### M2. Sweep admission misses required cross-input validation

Evidence:

- `sweep.doc` is decoded but never validated.
- Empty documentation passes `validateSweep`.
- Sweep `symbol` is never compared with generated Executor symbols.
- `botspec.Validate` cannot check replay identity because replay sits outside BotConfig.
- `botspec.Build` later rejects an Executor symbol lacking matching replay input.

Impact:

- `btsweep.Load` accepts definitions that fail only during later BtBot construction.
- Template admission does not prove one runnable generated Bot input.

Required fix:

- Validate nonempty `sweep.doc`.
- Validate Sweep replay symbol against every generated Bot’s required symbols.
- Leave a focused mismatch test.

### M3. Relative tick paths depend on process working directory

Evidence:

- Bot template paths resolve relative to the Sweep source file.
- Tick paths receive only `filepath.Clean`.
- Relative tick paths therefore change meaning with caller working directory.

Impact:

- The same Sweep file can return different replay paths from different working directories.
- This violates deterministic template path behavior.

Required fix:

- Resolve relative tick paths against the Sweep source directory.
- Alternatively reject non-absolute tick paths and document that contract.
- Add a test loading outside the template directory.

### M4. `stest.sh` can ignore log-write failure

Evidence:

- Most `tee -a "$result_log"` statuses are unchecked.
- The final report pipeline records only `PIPESTATUS[1]`, the reporter status.
- It ignores `PIPESTATUS[2]`, the final `tee` status.
- `set -o pipefail` has no effect because the pipeline result is not tested.

Impact:

- A Bot and reporter can pass while the required result log fails to persist.
- `stest.sh` can exit zero with incomplete proof.

Required fix:

- Check every proof-critical `tee` failure.
- Check reporter and final log-writer statuses separately.
- Add a controlled failing-output test.

### M5. Canonical documentation contains broken and stale links

Evidence:

- `wiki/DESIGN.md` links twice to missing `wiki/design/report.md`.
- The actual owner remains `wiki/design/runreport.md`.
- Script ownership references remain stale in three canonical wiki pages.

Impact:

- Canonical navigation is broken.
- New `stest.sh` ownership is not consistently discoverable.

Required fix:

- Restore links to `design/runreport.md`, or rename the owner coherently.
- Replace all stale script references with the approved `stest.sh` contract.

### M6. Prescriptive coding examples do not match the real BtBot API

Evidence:

- `wiki/coding/RULES.md` calls nonexistent `btbot.Init` and expects `*btbot.BtBot`.
- `wiki/coding/STYLE.md` calls `runner.Init` without the required `context.Context` argument.
- Real API is `(*BtBot).Init(context.Context, *logging.Logger, uint64, uint64)`.

Impact:

- Mandatory examples cannot compile.
- Future code following canonical examples will violate the implemented API.

Required fix:

- Make both examples match `cmd/nuubot-bt-bot/main.go` exactly.
- Include context and value construction consistently.

### M7. Go review inventory omits all new BtSweep files

Evidence:

- Actual Go files: 94.
- Listed Go files: 91.
- Missing `cmd/nuubot-bt-sweep/main.go`.
- Missing `internal/btsweep/btsweep.go`.
- Missing `internal/btsweep/btsweep_test.go`.

Impact:

- `wiki/reviews.md` no longer inventories every Go file.
- The review contract cannot detect later changes for omitted files.

Required fix:

- Add all three paths with `NO`, no review date, and no hash.
- Preserve the fixed-width layout.

### M8. New Go source does not pass `gofmt`

Evidence:

- `gofmt -d` reports field alignment changes in `internal/report/report.go`.
- It reports CRLF normalization for `cmd/nuubot-bt-sweep/main.go`.
- Project style requires every hand-written Go file to pass `gofmt`.

Impact:

- Green tests do not satisfy mandatory source style.

Required fix:

- Run canonical `gofmt` on changed Go files.
- Re-run zero-diff formatting proof.

### M9. Untracked reserved-name file can contaminate or block commit

Evidence:

- Worktree contains untracked zero-byte file `nul`.
- `nul` is unrelated to the approved task.
- `nul` is a reserved Windows device name.

Impact:

- Broad staging can fail or create a non-portable path.
- The worktree is not safe for an indiscriminate commit.

Required fix:

- Remove `nul` before staging.
- Verify status afterward.

### M10. Profiled semantic proof is not linked from its canonical claim

Evidence:

- `wiki/PERFORMANCE.md` links only the profile directory for the profiled Grid run.
- Profile files contain runtime data, not exact finance and domain counts.
- Exact profiled evidence exists in matching `.log` and `.json` files but is not cited.

Missing citations:

```text
workspace/logs/nuubot5-stest-s10-b14-1-20260726T104650Z.log
workspace/logs/nuubot5-stest-s10-b14-1-20260726T104650Z.json
```

Impact:

- The durable “same semantic result” claim is not directly traceable to exact result evidence.

Required fix:

- Cite the profiled result log and suite JSON beside the profile directory.
- Retain exact counts and decimal strings.

## NOTE

### N1. Tracked hard rename is otherwise complete

Evidence:

- Current source and wiki contain no old command, package, type, config, metric, or SQL names.
- The sole tracked match is the intentional prohibition in `wiki/bin.md`.
- `bin/` contains `nuubot-bt-bot.exe` and no old executable.

### N2. Ignored historical workspace logs retain old names

Evidence:

- `workspace/logs/**` contains historical BtRunner messages.
- These files are ignored runtime evidence, not compatibility code.

Assessment:

- This violates a literal repository-wide residue search.
- It does not affect the commit unless ignored artifacts are force-added.
- Decide whether the hardcut excludes ignored historical runtime evidence.

### N3. Config, JSON, SQL, and report names align

Confirmed:

- AppConfig uses `[btbot]` and `App.BtBot`.
- Logs use `btbot` and `btbot_historical_data_loop_elapsed_ms`.
- Attempt JSON uses `btbot_elapsed_ms`.
- Result SQLite uses `btbot_historical_data_loop_elapsed_ms`.
- Report aggregation and rendering use `BtBot`.

Missing focused proof:

- No config test explicitly rejects the old `[btrunner]` key.

### N4. `stest.sh` core selection is sound

Confirmed:

- Numeric selectors and repetitions reject zero, signs, text, duplicates, and extra arguments.
- Bot mode resolves global `bot_id` and rejects duplicate rows.
- Current schema proves `bot_id` is the primary key.
- Sweep mode orders by `bot_no`, then `bot_id`.
- Supported BotSpecs select distinct Observer, Trade, and Grid validators.
- Profiling rejects repetitions above one.
- Paths and SQL identity values are quoted or numeric-constrained.
- Reporter failure makes the suite fail.

### N5. Script replacement mostly preserves old semantic checks

Confirmed:

- Observer log checks match deleted `rtest.sh`.
- Trade SQL checks preserve deleted `trtest.sh` behavior.
- Grid SQL checks preserve deleted `grtest.sh` behavior.
- Profiling arguments preserve deleted `pptest.sh` behavior.
- New argument handling is stricter than old scripts.

### N6. Placeholder scope is honest

Confirmed:

- `cmd/nuubot-bt-sweep` prints exactly `Under Construction.`.
- `internal/btsweep` writes no database and launches no process.
- Canonical architecture marks persistence, execution, cancellation, and aggregation deferred.

### N7. Existing Grid proof is authentic

Normal evidence:

```text
workspace/logs/nuubot5-stest-s10-b14-1-20260726T104521Z.log
workspace/logs/nuubot5-stest-s10-b14-1-20260726T104521Z.json
```

Confirmed normal result:

```text
BtBot ms             80,225
Historical loop ms   75,049
Total allocation MB  142,471.67504882812
Trades               1,982
Orders               4,697
Fills                 2,636
Net PnL               -57.420074089999999993851
Ending equity         942.579925910000000006149
Maximum drawdown      75.791979199999999992245
```

Profiled evidence:

```text
workspace/logs/nuubot5-stest-s10-b14-1-20260726T104650Z.log
workspace/logs/nuubot5-stest-s10-b14-1-20260726T104650Z.json
workspace/perf/profiles/stest-s10-b14-20260726T104650Z/
```

Confirmed profiled result:

```text
BtBot ms             82,880
Historical loop ms   77,930
Total allocation MB  142,514.548286438
Trades               1,982
Orders               4,697
Fills                 2,636
Net PnL               -57.420074089999999993851
Ending equity         942.579925910000000006149
Maximum drawdown      75.791979199999999992245
```

All six profile files are nonempty.

All five pprof files are readable by Go 1.26.5.

The trace parses successfully with Go 1.26.5.

### N8. Unrelated Ledger work remains present

Evidence:

- Worktree contains separate Ledger documentation and `.audits/07-26-ledger-*` artifacts.
- No unrelated Ledger source content difference appeared in `git diff`.

Assessment:

- Preserve this work.
- Stage deliberately after accepted fixes.
- Do not delete or rewrite it as command-task cleanup.

## Validation Run

Passed:

```text
bash -n build.sh stest.sh
```

Passed invalid-argument probes:

```text
stest.sh with no selector
stest.sh -bot 0
stest.sh with both selectors
stest.sh -runs 2 -pp
stest.sh with duplicate -runs
```

Passed focused Go tests:

```text
CGO_ENABLED=0 go test -count=1 -tags noasm \
  ./cmd/nuubot-bt-bot ./cmd/nuubot-report ./cmd/nuubot-bt-sweep \
  ./internal/btbot ./internal/btsweep ./internal/botspec ./internal/config \
  ./internal/report ./internal/resultpublisher
```

Passed full Go tests:

```text
CGO_ENABLED=0 go test -count=1 -tags noasm ./...
```

Passed full vet:

```text
CGO_ENABLED=0 go vet -tags noasm ./...
```

Passed isolated builds:

```text
go build -buildvcs=false -tags noasm ./cmd/nuubot-bt-bot
go build -buildvcs=false -tags noasm ./cmd/nuubot-report
go build -buildvcs=false -tags noasm ./cmd/nuubot-bt-sweep
```

Passed:

```text
git diff --check
```

Failed formatting proof:

```text
gofmt -d <changed Go files>
```

Failed documentation-link proof:

```text
wiki/DESIGN.md -> design/report.md
wiki/DESIGN.md -> design/report.md
```

No new full `stest.sh` replay was run during this read-only audit.

## Verdict

**FAIL**

Do not commit or push yet.

Clear B1 and M1-M10, then rerun focused proof, full proof, formatting, links, inventory, exact Grid, and profiled Grid verification.
