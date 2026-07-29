# Nuubot5 Server Handoff

Last updated: 2026-07-29

## Focus

Synchronize the complete main and Server worktrees without mixing owned state.

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

## TODO

- Commit the resolved local merge.
- Fast-forward local main to the merge.
- Verify both local branches and worktrees.

## PENDING USER APPROVAL

- Repair or remove main's remaining stale Cloid, Simulator, and
  ResultPublisher tests before remote push.

## Next Action

Commit the resolved merge, then fast-forward local main.
