# Error Contract

## Purpose

Preserve one useful failure chain and log each returned error at one owning boundary.

## Status

Approved — unimplemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/common/error.rs`
- Nuubot4: `D:/rust/nuubot4/src/common/program.rs`
- Nuubot5: `wiki/coding/RULES.md`

## Scope

This contract applies to every command, service, lifecycle owner, domain helper, and background task.

## Required Contract

- Use standard `error`, `errors`, and `fmt`.
- Return operational failures.
- Add context only at approved boundaries.
- Preserve internal errors with `%w`.
- Translate third-party errors unless `errors.Is` or `errors.As` is an approved contract.
- Keep the primary work error when shutdown also fails.
- Log returned errors exactly once.
- Use `errors.Join` when independent shutdown failures matter.

## Error Boundaries

Context may be added at:

- exported package operations;
- lifecycle operations;
- Domain Helpers;
- executable boundaries;
- request boundaries; and
- background-task boundaries.

Unexported same-owner functions return received errors unchanged.

## Does Not

- Define a custom error framework.
- Panic for operational failures.
- Log and return the same error below a boundary.
- Retry, skip, repair, or fall back without an approved recovery contract.
- Hide identity needed to locate the failure.

## Program Flow

```text
leaf
  return concise operation error

owner boundary
  add one useful context layer
  return error

process, request, or task boundary
  log structured error once
  return terminal status
```

## Required Proof

- Error chains retain the failing operation and identity.
- Shutdown failures do not erase work failures.
- One returned error creates one boundary error log.
- Invalid external input fails before mutation or external action.

