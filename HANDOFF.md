# Handoff

Last updated: 2026-07-23

## Focus

Rewrite the BtRunner command and logging ownership after aligning mandatory coding contracts.

## Current status

- Go BtRunner remains implemented and proven with canonical `-tags noasm`.
- TMUX/PSMux startup builds one control pane and four role-aware viewer panes.
- `nuubot-server`, `nuubot-cli`, and `nuubot-runner` reserve canonical names.
- Each placeholder command prints exactly `Under Construction.` and exits zero.
- `nuubot-btrunner` remains the only implemented command.
- The new BtRunner command and log-routing shape is agreed but not implemented.

## Active agents

- Root only.

## Blockers

- BtRunner rewrite requires approval to align `wiki/coding/STYLE.md` and `RULES.md`.

## Files changed

- `AGENTS.md` and `TMUX.md`: guarded TMUX/PSMux startup and placeholder exception.
- `cmd/nuubot-server/main.go`: Server naming placeholder.
- `cmd/nuubot-cli/main.go`: CLI naming placeholder.
- `cmd/nuubot-runner/main.go`: Runner naming placeholder.
- `wiki/PROJECT.md`: placeholder status without implementation claims.
- `HANDOFF.md`: current restart state.

## Proof

- `go test -tags noasm ./...` passed.
- All four command packages built together with `-tags noasm`.
- All three placeholders printed exactly `Under Construction.` and exited zero.
- TMUX proof established the required five-pane layout and role responses.
- Prior 1000x BtRunner proof remains recorded in `wiki/PROJECT.md`.

## Decisions

- Keep `cmd/nuubot-btrunner/main.go` and `internal/btrunner/btrunner.go` separate.
- Keep shared implementation in concrete `internal/<package>` directories.
- Placeholder commands reserve names only and do not prove implementation.
- General errors use `server.log` until identity-specific logging is established.
- Identity-specific work then logs only to its identity log.
- Every log file opens append-only.

## Not run

- No fresh BtRunner replay was required for command placeholders.
- A separate fresh root session has not rerun automatic TMUX startup.

## Next action

Approve coding-contract alignment, then rewrite BtRunner `main.go` and `internal/logging`.

Go toolchain:

```text
C:\Users\PC\.local\go1.26.5\go\bin
```
