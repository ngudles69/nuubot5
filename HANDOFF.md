# Handoff

Last updated: 2026-07-23

## Focus

Review compact core designs and cascading Go lifecycle.

## Current status

- Nuubot4 core ownership is ported into compact Go designs and one cascading implementation.
- Signaler, Executor, and Risk are permanent factories for many concrete implementations.
- Approved baseline remains Macross, Observer stop-loss, and BalancedRisk.
- Setup owns configuration and Bot admission.
- Runtime initialization owns Signaler Bars loading and preparation.
- Runtime returns stop intent; BtRunner owns final cascading teardown.
- `internal` contains 22 Go packages.
- `internal/toolkit` groups `clock`, `errors`, and `logging`.
- `internal/toolkit` is not a Go package.
- `internal/common`, `internal/clock`, and `internal/logging` are removed.
- `account`, `fill`, `ledger`, `meta`, `order`, `simulator`, and `trade` reserve approved package names only.
- `wiki/design/packages` contains exactly one page per Go package.
- `wiki/design/concepts` contains 27 process, program, venue, type, and cross-package pages.
- `wiki/DESIGN.md` indexes both categories.
- `main.go` now only parses input, runs BtRunner, and logs the final result with duration.
- `internal/btrunner.Run` owns construction, lifecycle, and shutdown.
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
- No commit or push has been authorized.

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

- `go list -tags noasm ./...` passed.
- `go test -tags noasm ./...` passed.
- Focused BtRunner command tests passed.
- Two invalid starts appended exactly two `server.log` entries.
- An identified failure appended zero Server entries and one Bot entry.
- Invalid and identified failures emitted zero console lines after logger creation.
- All 22 Go package names match 22 package page names.
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
- All 22 package page names match all 22 Go package names.
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

- Package pages live in `wiki/design/packages`.
- Concept pages live in `wiki/design/concepts`.
- Toolkit is a grouping directory, not a package.
- Toolkit children remain independent and domain-neutral.
- Explicitly approved reservation files contain only package documentation and declaration.
- `StateError` belongs to `internal/toolkit/errors`.
- Millisecond `Duration` belongs to `internal/toolkit/clock`.
- `main.go` has exactly three responsibilities: parse, run, and log the timed result.
- Log fields exist only when removing them would hide useful information.
- `duration` remains on BtRunner Run results.
- `error` remains on failures.
- Boundary messages name `parseInput()`, `logging.OpenBot()`, or `btrunner.Run()`.
- New and changed Go prefers `var` declarations and later `=` assignments.
- `:=` remains allowed only when `var` or `=` is impossible or materially less clear.
- Setup, BtRunner, and Runtime form one cascading lifecycle.
- Signaler, Executor, and Risk remain configuration-selected factories.
- Runtime owns Signaler preparation; BtRunner does not know Bars or Signaler.
- Observer retains canonical one-percent adverse-move exit behavior.

## Not run

- No commit or push was performed.

## Next action

User reviews the compact designs and cascading Go implementation.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
