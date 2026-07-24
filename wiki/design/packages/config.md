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
- Reject unknown shared Config fields.
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

Unknown shared Config TOML fields MUST fail.

Credentials unknown-field rejection remains deferred with detailed validation.

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
- Unknown shared Config fields fail.
- Invalid lifecycle limits, kinds, periods, and percentages fail.

## Open Decisions

Detailed credential validation is deferred.

Detailed validation of new shared configuration fields is deferred.

Logging owns its directory and filenames. They are not configuration.

## Trading Fields

TradeExecutor configuration adds:

```text
account_name
network
order_notional_usdc
take_profit_pct
stop_loss_pct
simulator_equity_usdc
simulator_fee_pct
simulator_slippage_pct
persist_mode
```

Financial configuration uses canonical decimal text.

BtRunner accepts only `network = "simnet"` for TradeExecutor.

`persist_mode` accepts `none` or `max`.

Account passes the selected mode to Ledger and Simulator.

Live and testnet Account names resolve against the credentials catalog.

Account validates only the selected credential before live Venue initialization.

Config and Setup never log secret fields.

## Approved Target Split

Status: Approved target design. Not implemented.

The current mixed Config receives a hardcut replacement.

Target configuration ownership is:

| Value | Owns |
|---|---|
| AppConfig | Server address, storage paths, logging, and operational infrastructure |
| BotConfig | One exact BotSpec's supported trading parameters |
| ReplayInput | Historical data source, symbols, dates, and replay controls |
| Credentials | Secret values resolved only through non-secret Account references |

AppConfig cannot change Bot trading behavior.

`NUUBOT_CONFIG` may select AppConfig only.

Replay cadence affecting Results belongs to ReplayInput.

Network and Account selection belong to exact BotConfig fields.

## Approved BotConfig Source

TOML is a generic user form.

Each exact BotSpec owns one TOML template and decoder.

After validation, the database stores:

```text
bot_spec_id
config_toml
config_hash
```

The database value is authoritative after configuration.

Controller never reads the imported TOML file.

BotGeneration stores one immutable copy of the exact TOML and hash.

The running process decodes only that saved copy.

JSON is not a second persisted BotConfig representation.

See [BotSpec](../concepts/bot-spec.md).

## Approved Validation

The exact BotSpec decoder:

- Requires every required field.
- Validates every recognized field's type and value.
- Allows additional fields and sections.
- Preserves additional fields in the stored TOML.
- Ignores unrecognized fields.
- Rejects duplicate TOML keys.

An ignored field never gains meaning under the same BotSpecID.

Supporting it later requires another exact BotSpecID and Config.

There is no fallback, migration, alias, or compatibility decoder.

Config contains credential references only.

It never contains private credentials.

## Approved Component Selection

BotConfig does not select arbitrary Signaler, Risk, or Executor kinds.

The exact BotSpec selects the complete Controller structure.

Config supplies supported parameters such as:

- Symbols.
- Sides.
- Grid levels.
- Hedge triggers.
- Indicator periods.
- Approved filter choices.
- Risk thresholds.
- Account references.

Unsupported component combinations require another BotSpec.

## Approved Capital and Sizing

Each Executor's Config declares capital for its Account-symbol resource.

The Executor owns its Order-sizing rule.

Supported exact BotSpecs may use fixed quantity, fixed quote amount, percentage
of assigned capital, or percentage of physical Account.

A physical-Account percentage resolves once during admission.

The admitted fixed amount never changes because another Bot changes Account
equity.

Bot capital is the sum of admitted Executor capital in one reporting currency.

Order plans exceeding assigned capital fail admission.

Capital is not physical Account equity.

Cross-process capital reservation remains TBD.

## Approved Static Template

Each exact BotSpec owns one static canonical commented TOML template.

The template includes the exact BotSpecID, field units, valid options, safe
defaults, and credential references only.

No reflection generator, schema-driven custom form, or dynamic template system
is approved.
