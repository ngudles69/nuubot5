# Logging Package

Status: Implemented.
Covers: `internal/toolkit/logging/logging.go`
Purpose: Write complete messages to append-only files using one exact format.

## Canonical Sources

- Nuubot4 behavior: `D:/rust/nuubot4/src/common/logging.rs`

## Scope

Logging owns destinations, append-only opening, timestamps, levels, record format, and line writing.

## Required Contract

- Use the standard library `log.Logger` internally.
- Write under `workspace/logs`.
- Open every file with create, append, and write-only flags.
- Use `server.log` before identity.
- Use `bot_<sweep_id>_<bot_id>.log` after Bot identity.
- Write to one destination only.
- Write one console error only when `server.log` cannot open.
- `Open` owns directory creation, file opening, and Logger construction.
- `OpenBot` opens one identity-bound Bot logger.
- Configure logging once at each executable or Server boundary.
- Pass explicit `*logging.Logger` values.
- Name every Logger parameter `log`.
- Construct one complete message string before calling `log`.
- Write `YYYY-MMM-DD HH:MM:SS [LEVEL] message`.
- Right-align levels to a minimum width of five characters.
- Support `DEBUG`, `INFO`, `WARNING`, `ERROR`, and `CRITICAL`.
- Log returned errors only at process, request, or background-task boundaries.
- Let each owner report its own terminal statistics.

## Message Ownership

- Logging formats the timestamp, level, and complete record.
- The sender owns everything inside the message.
- The sender formats values, errors, statistics, and elapsed duration.
- Every `Info`, `Error`, and other level call receives exactly one string.

## Does Not

- Interpret or modify message content.
- Accept key-value message arguments.
- Bind component, event, status, or identity fields.
- Open files inside domain components.
- Write one event to both server and identity logs.
- Write console output after a logger exists.
- Re-log child statistics in a parent.
- Record secrets or credentials.

## Program Flow

```text
create
  create logger

Open
  create log directory
  open log file
  return logger

OpenBot
  open bot log

write
  write record
```

`Open`, `OpenBot`, and level methods remain domain helpers because their names
state the exact logging operation.

## Required Proof

- Every line matches the exact timestamp, level, and message format.
- Short levels are padded; long levels remain complete.
- Every message identifies its owned operation.
- Error values appear once per returned failure.
- Owner terminal statistics are not duplicated.
- Logging setup failure reaches the executable boundary.
- Repeated starts append instead of truncating.
- Pre-identity failures appear in server.log.
- Identified Bot failures appear only in that Bot log.
- Logs contain no secrets.
