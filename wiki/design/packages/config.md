# Config Package

Status: Implemented.
Covers: `internal/config/*.go`
Purpose: Load process-wide AppConfig and local credentials.

## AppConfig

`config.LoadApp` strictly decodes:

- Server listener values;
- network policy;
- Hyperliquid shared policy;
- process timeouts;
- workspace paths; and
- separate Live and Backtest runtime policies.

AppConfig contains no BotSpec, Signaler, Risk, Executor, capital, or order
settings.

Unknown AppConfig fields fail.

Each runtime policy defines Controller, Recon, Recon sweep, and telemetry cadence plus telemetry write-on-collect policy.

Controller and telemetry cadence must be positive. Recon cadence must exceed Controller cadence. Recon sweep cadence must exceed Recon cadence.

The loader validates both profiles. Backtest `Run` selects `App.Backtest` once. Live `Run` selects `App.Live` once.

Live requires telemetry write-on-collect. Backtest requires terminal-only telemetry publication.

Lower components read only selected `nuubot.Runtime`. They never branch on Live versus Backtest.

## BotConfig Boundary

BotConfig belongs to one exact BotSpec and is stored as TOML in Datastore.

`internal/botspec` owns BotConfig decoding and validation.

Extra BotConfig fields are stored and ignored.

Required recognized fields must exist and validate.

Duplicate TOML keys fail.

## Credentials

Credentials remain a separate ignored local file.

Simulator backtests never load credentials.

Live Account admission may load only its referenced credential later.

## Paths

Relative AppConfig paths resolve below the repository root.

Replay paths must resolve below the configured shared-data root.

## Invariants

- Bot behavior never comes from `NUUBOT_CONFIG`.
- Controller receives no Config parser or mutable Config.
- Secrets never enter tracked AppConfig or BotConfig.
- No fallback or compatibility Config path exists.
