# Nuubot5 Server Handoff

Last updated: 2026-07-29

## Focus

Verify main's stale-test cleanup and complete local and remote synchronization.

## DONE

- User approved the split.
- Existing handoff history was preserved in `HANDOFF-MAIN.md`.
- `AGENTS.md` routes each worktree to its matching instruction file.
- `HANDOFF.md` routes each worktree to its matching state file.
- Common instructions remain in `AGENTS.md`.
- Main and server task state are isolated.
- Canonical ownership and startup references are aligned.
- Independent routing and ownership review passed.
- Legacy handoff comparison shows only two required owner-name changes.
- `git diff --check` passed.
- `nuubot-server` now runs the first embedded standard-library WebServer.
- Home, health, embedded asset, missing-route, and cancellation tests pass.
- Full `go test -tags noasm ./...` passed.
- Full `go vet -tags noasm ./...` passed.
- Live home, health, and embedded CSS requests returned HTTP 200.
- Server listens on `127.0.0.1:9898`.
- `build.sh` now builds `bin/nuubot-server.exe`.
- Persistent Server executable returned HTTP 200 for home and health.
- User approved two WebServer console lifecycle lines as a logging exception.
- WebServer prints and logs its successful start and graceful stop.
- Tests prove both lifecycle messages reach console output and the Server log.
- Rebuilt persistent executable printed the correct port and returned HTTP 200.
- Full post-change `go test -tags noasm ./...` passed.
- Final `git diff --check` passed.
- Main Account, Ledger, Trade, Order, Fill, Executor, Simulator, Venue, test,
  and wiki changes are merged without Server-side reinterpretation.
- Every test deletion from main is preserved.
- Main-owned `AGENTS-MAIN.md`, `HANDOFF-MAIN.md`, and
  `NUUBOT_RUNNER_PLAN.md` match main exactly.
- Server-owned `AGENTS-SERVER.md` remains unchanged.
- Shared `AGENTS.md` combines main's current rules with approved Server rules.
- Stale audit contents are purged; `.audits/.gitkeep` preserves the directory.
- Combined Server packages pass focused tests and the canonical build succeeds.
- Combined full tests remain red in Cloid, Simulator, and ResultPublisher.
- The same three failures reproduce on unmerged main and predate this merge.
- Server work was committed as `41fb0af`.
- Main was merged into Server as `5b256e9`.
- Local main fast-forwarded to the resolved merge.
- Both local worktrees contain main domain changes and Server WebServer changes.
- Main cleanup `d0ef071` refreshed CLOID proof and removed stale Simulator and
  ResultPublisher suites.
- Server fast-forwarded to `d0ef071`.
- Full synchronized `go test -tags noasm ./...` passes.
- Full synchronized `go vet -tags noasm ./...` passes.
- Canonical `build.sh` succeeds after synchronization.
- Synchronization proof was committed as `634d45a`.
- Local main fast-forwarded to `634d45a`.
- `origin/main` advanced through `634d45a`.

## TODO

- None.

## PENDING USER APPROVAL

- None.

## Next Action

Wait for the next Server task.
