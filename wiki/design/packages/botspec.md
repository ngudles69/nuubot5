# BotSpec Package

Status: Implemented for Macross Observer, TradeBot, and GridBot.
Covers: `internal/botspec/**`
Purpose: Validate and shape exact BotConfig TOML into one typed BotSpec.

The explicit catalogue recognizes:

- `macross_observer_bot`;
- `macross_trade_bot`; and
- `macross_grid_bot`.

Each BotSpec contains:

- exact BotSpecID;
- Controller specification;
- Signaler specification;
- Risk specifications; and
- Executor specifications.

`Build` receives BotSpecID and exact BotConfig TOML.

`Build` decodes, validates, applies explicitly defined defaults, and shapes
typed values.

`Build` returns one immutable BotSpec or an error.

BotSpec never contains:

- App Config;
- Meta;
- replay runtime inputs;
- ResultPath;
- SweepID or BotID;
- Config provenance;
- runtime objects; or
- runtime state.

BotSpec creates no Signaler, Risk, BotCycle, Executor, Account, or Venue object.

Setup stores BotSpec in the shared Nuubot harness.

Controller reads BotSpec from Nuubot and constructs runtime objects.

Extra TOML fields remain stored and ignored.

Duplicate TOML keys, duplicate Executor resources, invalid decimals, invalid
roles, and unknown IDs fail.

No reflection, plugin, DSL, runtime compilation, fallback, alias, or
compatibility path exists.

Observer market identity contains Venue, network, and symbol.

When both Venue and network are absent, BotSpec applies the explicit
`simulator/simnet` default. Supplying only one value is invalid.

Supported Observer market identities are Simulator simnet and Hyperliquid
testnet or mainnet.

`macross_grid_bot` requires exactly one GridExecutor.

Its specification owns capital, side, 3-to-1,024 Levels, range, minimum
expected PnL, fees, slippage, and persistence.

Its compiled calculation uses arithmetic spacing with equal capital slices.

BtBot, Controller, BotSpec, BotCycle, and Executor runtime lifecycle use system
proof instead of isolated unit tests.

Only pure deterministic Grid calculations retain Executor unit tests.

See `wiki/testing.md`.
