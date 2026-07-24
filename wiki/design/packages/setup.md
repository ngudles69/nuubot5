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

## Approved Admission Target

Status: Approved target design. Not implemented.

Setup will stop returning the current broad mutable-behavior Context.

Target admission receives:

- Saved BotGeneration TOML and hash.
- Exact BotSpecID.
- ReplayInput or live Runner inputs.
- Caller context.
- AppConfig.
- Required metadata.

It returns one immutable typed BotDefinition or an error.

Controller receives admitted values only.

Controller never receives:

- A database handle.
- A TOML parser.
- A Config file path.
- Unselected credentials.
- A background context created by Setup.

Caller context owns cancellation and timeouts.

Simulator admission loads no private credential.

Live Account admission resolves only referenced credentials.

Runner, BtRunner, and SweepRunner each own their process-local admission.

Admission cannot require a running Server, API, BotManager, or SweepManager.

See [BotSpec](../concepts/bot-spec.md).

## Approved Meta Admission

Live Start fails closed when required Meta:

- Cannot be obtained.
- Does not exist.
- Is incomplete.
- Names an inactive or unsupported symbol.

No stale cache, previous generation, default precision, default margin, or
substitute symbol is accepted.

Historical replay requires its pinned Meta snapshot and matching hash.

Current live Meta never substitutes for missing historical Meta.

The complete Venue-specific admission checklist remains pending direct
implementation review.
