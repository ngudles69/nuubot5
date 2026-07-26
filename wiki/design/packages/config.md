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
- BtBot cadence.

AppConfig contains no BotSpec, Signaler, Risk, Executor, capital, or order
settings.

Unknown AppConfig fields fail.

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
