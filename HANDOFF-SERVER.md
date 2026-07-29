# Nuubot5 Server Handoff

Last updated: 2026-07-29

## Focus

Implement the first minimal embedded WebServer vertical slice.

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
- Nuubot3 report, Server ownership, lazy chart loading, BotCycle focus, and
  Signaler storage requirements are consolidated in
  `.audits/05-28-gui-prep.md`.
- Server shell, managers, WebServer, embedded assets, chart baseline, warmup
  evidence, lazy loading, and control boundaries are recorded in
  `.audits/07-28-server.md`.
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

## TODO

- None.

## PENDING USER APPROVAL

- None.

## Next Action

Wait for approval to add the next Server vertical slice.
