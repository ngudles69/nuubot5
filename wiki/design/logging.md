# Logging Contract

## Purpose

Provide one structured process logger and owner-local operational evidence.

## Status

Approved — unimplemented.

## Canonical Sources

- Nuubot4 behavior: `D:/rust/nuubot4/src/common/logging.rs`
- Nuubot5 target: `wiki/coding/STYLE.md`
- Nuubot5 rules: `wiki/coding/RULES.md`

## Scope

Logging setup is shared infrastructure. Components receive loggers and emit only their owned events and statistics.

## Required Contract

- Use standard `log/slog`.
- `internal/logging.New(io.Writer)` owns handler construction.
- Configure logging once at each executable or Server boundary.
- Pass explicit `*slog.Logger` values.
- Bind component identity with `logger.With`.
- Use stable snake_case fields.
- Log returned errors only at process, request, or background-task boundaries.
- Let each owner report its own terminal statistics.

## Required Fields

Lifecycle records contain:

- `component`;
- `event`;
- `status`; and
- owning identity.

Boundary failures also contain `error`.

## Does Not

- Define a custom Logger wrapper.
- Use global default logger configuration.
- Open files inside domain components.
- Format machine fields into message strings.
- Re-log child statistics in a parent.
- Record secrets or credentials.

## Program Flow

```text
command or Server
  open approved outputs
  create one slog Logger
  pass child loggers with bound component

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
- Logs contain no secrets.

