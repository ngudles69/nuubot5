# Setup

## Purpose

Prepare one admitted BtRunner context before component construction.

## Status

Implemented.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/setup.rs`
- Nuubot5: `internal/setup/setup.go`

## Scope

Setup loads one Bot specification and proves its market-data path remains inside shared data.

## Owner and Children

BtRunner calls Setup. Setup owns no long-lived child.

## Responsibilities

- Load the requested Sweep Bot.
- Resolve configured paths from the repository root.
- Resolve symlinks before containment checks.
- Return validated configuration and Bot values.

## Does Not

- Construct Runtime components.
- Read market rows.
- Configure strategy behavior.
- Retain datastore connections.

## Lifecycle

One synchronous `Init` call returns a ready context or an error.

## Inputs and Outputs

Inputs are logger, repository root, configuration, Sweep ID, and Bot ID.

Output is `setup.Context` containing `config.Config` and admitted `datastore.BotSpec`.

## State and Invariants

The returned tick path MUST resolve beneath configured shared data.

The selected Bot MUST exist and pass datastore validation.

## Concurrency

Setup is synchronous and starts no goroutine.

## Persistence

Setup reads the SQLite Sweep copy through Datastore. It writes nothing.

## Errors

Missing Bots, unresolved paths, and path escapes fail setup.

## Program Flow

```text
Init
  load Bot by Sweep ID and Bot ID
  resolve shared-data root
  resolve Bot tick path
  reject path outside shared data
  return ready Context
```

## Required Proof

- A valid Bot returns its normalized specification.
- A path outside shared data fails.
- Missing paths and Bot identities fail.

## Open Decisions

None.
