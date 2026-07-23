# Logging Package

Status: Implemented.
Covers: `internal/toolkit/logging/logging.go`
Purpose: Provide one structured process logger and owner-local operational evidence.

## Canonical Sources

- Nuubot4 behavior: `D:/rust/nuubot4/src/common/logging.rs`

## Scope

Logging owns process log paths and setup. Components receive loggers and emit only their owned events and statistics.

## Required Contract

- Use standard `log/slog`.
- Write under `workspace/logs`.
- Open every file with create, append, and write-only flags.
- Use `server.log` before identity.
- Use `bot_<sweep_id>_<bot_id>.log` after Bot identity.
- Write to one destination only.
- Write one console error only when `server.log` cannot open.
- `Open` owns directory creation, file opening, and handler construction.
- `OpenBot` opens one identity-bound Bot logger.
- Configure logging once at each executable or Server boundary.
- Pass explicit `*slog.Logger` values.
- Bind component identity with `logger.With`.
- Use stable snake_case fields.
- Log returned errors only at process, request, or background-task boundaries.
- Let each owner report its own terminal statistics.

## Required Fields

- Add a field only when removing it hides useful operational information.
- BtRunner Run results include `duration`.
- Failures include `error`.

## Does Not

- Define a custom Logger wrapper.
- Use global default logger configuration.
- Open files inside domain components.
- Write one event to both server and identity logs.
- Write console output after a logger exists.
- Format machine fields into message strings.
- Re-log child statistics in a parent.
- Record secrets or credentials.

## Program Flow

```text
command
  open server.log
  parse identity
  open identity log
  replace server logger
  run identified work
  log one result and duration

component
  log lifecycle and owned statistics
  return failures without logging error values

boundary
  log returned error once
```

## Required Proof

- Every record identifies its component and event.
- Error values appear once per returned failure.
- Owner terminal statistics are not duplicated.
- Logging setup failure reaches the executable boundary.
- Repeated starts append instead of truncating.
- Pre-identity failures appear in server.log.
- Identified Bot failures appear only in that Bot log.
- Logs contain no secrets.
