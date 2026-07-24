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

One configured shared database owns Sweeps, Bots, and mainnet Meta.

## Program Flow

```text
Setup
  resolve root
  load config
  load credentials
  prepare datastore
  validate ticks path
  admit mainnet Meta
  return setup
```

## Notes

- Setup performs admission only. It owns no running child.
- Setup has one function and returns one Context.
- Config and credentials are read-only and idempotent when files are unchanged.
- Setup performs no hot reload. Running processes retain their admitted Context.
- Credentials receive TOML decoding only. Account validation is deferred.
- Account validates only its selected live credential during initialization.
- Simulator receives no private credential.
- Meta reads freshness from the configured shared database.
- Meta younger than 24 hours will continue without an exchange request.
- Empty or stale Meta will refresh before Setup continues.
- Meta always refreshes from mainnet.
- Tests needing different Meta manually update their local SQLite database.
- Shared WebSocket ownership remains TBD. Setup starts no background work.
- Setup uses the existing short-lived `LoadBot` path.
