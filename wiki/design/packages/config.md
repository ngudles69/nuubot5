# Config Package

Status: Implemented.
Covers: `internal/config/config.go`
Purpose: Provide one validated, non-secret configuration for BtRunner and Runtime composition.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/config.rs`

## Scope

Configuration covers data paths, BtRunner cadence, Runtime limits, Signaler selection, Executors, and Risks.

## Owner and Children

The command loads configuration. Components receive only their relevant values.

## Responsibilities

- Decode TOML.
- Reject unknown fields.
- Validate required paths and positive limits.
- Validate admitted Signaler, Executor, and Risk kinds.
- Resolve repository-relative paths.

## Does Not

- Store secrets.
- Load Bot-specific Sweep data.
- Open files or databases.
- Select behavior outside declared configuration.

## Lifecycle

`Load` decodes and validates once before Setup.

`Rooted` resolves one configured path without filesystem access.

## Inputs and Outputs

Input is one TOML file path.

Output is one validated `config.Config`.

## State and Invariants

Unknown TOML fields MUST fail.

BtRunner interval and Runtime maximum cycles MUST be positive.

At least one Executor MUST exist.

Executor stop-loss percentages MUST be finite and between zero and one.

## Concurrency

Configuration is immutable after loading.

## Persistence

Configuration reads one local TOML file. It writes nothing.

## Errors

Decode, unknown-field, missing-value, and invalid-range failures return errors.

## Program Flow

```text
Load
  decode TOML
  reject unknown fields
  validate paths
  validate BtRunner
  validate Runtime
  return Config
```

## Required Proof

- Current `config.toml` loads.
- Unknown fields fail.
- Invalid lifecycle limits, kinds, periods, and percentages fail.

## Open Decisions

Nuubot4 contains additional environment, loader, and simulator fields. They are not current Nuubot5 configuration.

Logging owns its directory and filenames. They are not configuration.
