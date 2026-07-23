# Toolkit

Status: Implemented.
Covers: `internal/toolkit/*/*.go`
Purpose: Group small, reusable, domain-independent packages.

`internal/toolkit` is a directory. It is not a Go package.

## Packages

- `clock`: deterministic clock mechanics and duration helpers.
- `logging`: exact-format append-only file logging.

## Rules

- Each child is an independent Go package.
- Each child remains standard-library first.
- Toolkit packages contain no trading policy.
- Toolkit packages import no Nuubot domain package.
- Copying one child must not require copying every child.
- New toolkit packages require proven reuse.

## Does Not

- Replace clear domain ownership.
- Become a `common`, `shared`, `utils`, or `misc` bucket.
- Own program orchestration.
