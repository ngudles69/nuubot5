# Config Package

Status: Implemented.
Covers: `internal/config/config.go`, `internal/config/credentials.go`
Purpose: Load shared configuration and local credentials for Setup.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/src/config.rs`

## Scope

Configuration covers server, network, Hyperliquid policy, process, data paths,
BtRunner cadence, Runtime limits, Signaler selection, Executors, and Risks.

Credentials cover datastore access and Hyperliquid accounts.

## Owner and Children

Setup loads configuration and credentials. Components receive the admitted Context.

## Responsibilities

- Decode TOML.
- Reject unknown fields.
- Validate required paths and positive limits.
- Validate admitted Signaler, Executor, and Risk kinds.
- Resolve repository-relative paths.
- Decode credentials TOML without authenticating accounts.

## Does Not

- Load Bot-specific Sweep data.
- Open files or databases.
- Select behavior outside declared configuration.
- Reload files while a process runs.
- Authenticate or semantically validate credentials.

## Lifecycle

`Load` decodes and applies the existing configuration validation.

`Rooted` resolves one configured path without filesystem access.

`ResolveDataPath` admits one path inside the configured shared-data root.

`LoadCredentials` decodes credentials without semantic validation.

## Inputs and Outputs

Inputs are the shared config path and local credentials path.

Outputs are `config.Config` and `config.Credentials`.

## State and Invariants

Unknown TOML fields MUST fail.

BtRunner interval and Runtime maximum cycles MUST be positive.

At least one Executor MUST exist.

Executor stop-loss percentages MUST be finite and between zero and one.

`hyperliquid.min_order_notional_usdc` is currently `11`.

The configured floor buffers Hyperliquid's USDC 10 minimum against price and
size rounding before exchange acceptance.

## Concurrency

Configuration is immutable after loading.

Loading is read-only and idempotent when the source file is unchanged.

Running processes do not watch or reload configuration.

## Persistence

Configuration reads `workspace/config/config.toml`.

Credentials read `workspace/config/credentials.toml`.

Both loaders write nothing.

## Errors

Config decode, unknown-field, missing-value, and invalid-range failures return errors.

Malformed credentials TOML returns an error without exposing secret values.

## Program Flow

```text
Load
  decode toml
  reject unknown fields
  validate paths
  validate cadence
  validate runtime

LoadCredentials
  decode toml
```

## Required Proof

- Current `workspace/config/config.toml` loads twice with equal results.
- Credentials load twice with equal results.
- Malformed credentials TOML fails.
- Unknown fields fail.
- Invalid lifecycle limits, kinds, periods, and percentages fail.

## Open Decisions

Detailed credential validation is deferred.

Detailed validation of new shared configuration fields is deferred.

Logging owns its directory and filenames. They are not configuration.
