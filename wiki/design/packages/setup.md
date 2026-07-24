# Setup Package

Status: Partially reviewed.
Covers: `internal/setup/setup.go`
Purpose: Return one fully admitted context before BtRunner composition.

## Canonical Source

- `D:/rust/nuubot4/src/setup.rs`

## Scope & Responsibilities

`Setup` coordinates configuration, credentials, and existing Bot admission.

Config and credentials own their decoding. Datastore retains its current
short-lived read-only Bot-loading behavior.

## Program Flow

```text
Setup
  resolve root
  load config
  load credentials
  prepare datastore
  validate ticks path
  return setup
```

## Notes

- Setup performs admission only. It owns no running child.
- Setup has one function and returns one Context.
- Config and credentials are read-only and idempotent when files are unchanged.
- Setup performs no hot reload. Running processes retain their admitted Context.
- Credentials receive TOML decoding only. Account validation is deferred.
- Source marks the future Meta-admission location after current datastore admission.
- Meta will read dataset freshness through Datastore.
- Meta younger than 24 hours will continue without an exchange request.
- Empty or stale Meta will refresh before Setup continues.
- Meta implementation waits for NuubotDB and Datastore ownership.
- Shared WebSocket ownership remains TBD. Setup starts no background work.
- Datastore redesign is deferred. Setup uses the existing `LoadBot` path.
