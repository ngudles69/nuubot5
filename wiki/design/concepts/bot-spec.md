# BotSpec

Status: Implemented for Macross Observer, TradeBot, and GridBot replay.
Covers: `internal/bot`, `internal/botspec`, stored BotConfig validation, and
typed BotSpec construction.
Purpose: Define one complete configurable Bot architecture without runtime mix-and-match behavior.

## Meaning

A BotSpec is the complete Controller design.

It includes:

- Signaler implementation and supported options.
- Risk implementation and supported rules.
- Executor roles and coordination.
- Account and market-data requirements.
- BotCycle start and exit behavior.
- Exact BotConfig schema and TOML template.

A BotSpec is not one Executor.

It is not a family with separately resolved versions.

## Identity

`BotSpecID` is one opaque exact identity.

Examples:

```text
macross_grid_hedge_bot
macross_grid_hedge_bot_v2
```

The complete strings identify separate BotSpecs.

Nuubot does not parse a family or version from either name.

Rules:

- One BotSpecID resolves one compiled BotSpec.
- Changed BotSpec behavior requires another BotSpecID.
- An existing BotSpecID never resolves changed behavior.
- Missing BotSpecs fail.
- No alias or substitute is allowed.
- No fallback or compatibility path is allowed.
- No forward or backward Config migration is allowed.
- No plugin, runtime Go loading, reflection registration, or expression DSL is allowed.

Shared platform fixes do not create a new BotSpec.

Executable artifact identity belongs to Result provenance.

## Configured Bots

Many Bots may use the same BotSpec with different BotConfig values.

Example:

```text
Bot 1
  BotSpec: macross_grid_hedge_bot
  Grid: long BTC, 30 levels
  Hedge: short BTC, trigger 1 percent below start
  Signaler: EMA 9 and 21, KAMA 200 regime filter

Bot 2
  BotSpec: macross_grid_hedge_bot
  Grid: long ETH, 20 levels
  Hedge: short SOL, trigger 1 percent below start
  Signaler: EMA 20 and 50, HMA 200 regime filter
```

The BotSpec explicitly supports those fields and choices.

BotConfig does not select arbitrary component implementations.

Unsupported component combinations require another BotSpec.

## BotConfig TOML

TOML is the generic user form.

It avoids one custom input form for every BotSpec.

Every template contains its exact `bot_spec`.

The selected BotSpec and submitted `bot_spec` MUST match.

The exact BotSpec owns:

- Required fields.
- Recognized optional fields.
- Supported values.
- Validation.
- Field comments, ordering, units, and safe defaults.

Validation rules:

- Every required field MUST exist.
- Required and recognized fields MUST have valid types and values.
- Additional fields and sections are allowed.
- Additional fields remain stored in the original TOML.
- The selected BotSpec ignores unrecognized fields.
- Ignored fields never gain meaning under the same BotSpecID.
- Duplicate TOML keys fail.
- Extra fields cannot override recognized fields.

Supporting a previously ignored field requires another BotSpecID and separate Config.

## Persistence

TOML files are import inputs only.

After configuration, the database owns:

```text
bot_spec_id
config_toml
config_hash
```

The database stores the exact submitted TOML text.

Start copies that text and hash into one immutable BotGeneration.

Setup strictly transforms saved TOML through the exact BotSpecID.

Controller receives the resulting typed BotSpec through the shared Nuubot harness.

Execution code never rereads the original TOML file.

Implemented exact IDs are:

- `macross_observer_bot`;
- `macross_trade_bot`;
- `macross_grid_bot`.

Changing or deleting the import file changes nothing after saving.

JSON is for HTTP envelopes, status, and Results.

JSON is not a second persisted BotConfig representation.

## Typed BotSpec

Typed BotSpec is one immutable Bot definition produced from exact BotConfig TOML.

It contains:

- Exact BotSpecID.
- Typed Controller specification.
- Typed Signaler specification.
- Typed Risk specifications.
- Typed Executor specifications.

It contains no App Config, Meta, replay runtime input, ResultPath, Bot instance
identity, provenance, runtime object, or runtime state.

## Construction

Setup passes BotSpecID and exact BotConfig TOML to `botspec.Build`.

`botspec.Build` decodes, validates, applies explicitly defined defaults, and
shapes typed specification values.

It returns one immutable BotSpec or an error.

Setup stores BotSpec in one shared Nuubot harness and returns it to BtBot.

Nuubot contains Logger, App Config, Bot identity and provenance, ReplayInput,
BotSpec, Meta, and ResultPath.

Controller, BotCycle, Executors, and Accounts receive the same Nuubot pointer.

Controller constructs Signaler and Risk runtime objects.

BotCycle constructs Executor runtime objects when a cycle starts.

Simulator Setup loads no private credentials.

Caller context owns cancellation and timeouts.

Setup creates no background context.

## Live Start Gates

This is the approved main list. The implementation checklist may add required
Venue and domain checks after direct code review.

Every traded resource uses:

```text
venue
network
physical_account_id
symbol
```

The active owner is one Bot ID and BotGeneration ID.

Start requires:

1. No active Bot owns the same resource.
2. The Venue reports zero active Orders for the symbol.
3. The Venue reports zero position for the symbol.
4. The Account has sufficient available funds and margin for the complete Bot.
5. Required metadata exists and is authoritative.
6. Every required symbol is active, supported, and tradable.
7. Required market data and a current price are available.

Direction does not change resource exclusivity.

One BotGeneration cannot declare the same resource more than once.

Each Executor receives one distinct Account-symbol resource.

Hyperliquid normal Accounts expose one net position per symbol and do not
support hedge mode.

Examples:

- Account A and BTC plus Account A and ETH is allowed.
- Account A and BTC plus Account B and BTC is allowed.
- Two Executors using Account A and BTC is rejected.
- Long Account A and BTC plus short Account A and BTC is rejected.

Configured, stopped, and error Bots do not hold active claims.

Starting, running, and stopping Bots hold active claims.

Fresh Start never adopts existing Orders or positions.

Users clear existing Venue state manually before retrying Start.

No automatic cancel, flatten, transfer, leverage change, or state substitution occurs during Start validation.

## Symbol and Meta Gates

Live Start fails when:

- Meta cannot be obtained.
- Required Meta is missing or incomplete.
- A symbol is delisted, retired, discontinued, deactivated, or inactive.
- The Venue restricts the symbol to close-only or reduce-only operation.
- Tick size, quantity precision, minimum notional, leverage, or margin mode is unsupported.

No stale cached or previous-generation Meta substitutes for unavailable current live Meta.

Historical replay requires its pinned Meta snapshot and matching hash.

Current live Meta never substitutes for missing historical Meta.

Data-only symbols require Meta and market-data validation.

Traded symbols additionally require Account, clean-slate, and capital validation.

## Capital and Order Sizing

BotConfig declares capital for each Executor's Account-symbol resource.

Bot capital is the sum of configured Executor capital in one reporting currency.

Bot equity is:

```text
Bot capital + cumulative net Bot PnL
```

Net Bot PnL includes realized PnL, unrealized PnL, fees, and funding.

Each Executor controls its Order sizing from its validated specification.

Supported sizing may include:

- Fixed base quantity.
- Fixed quote amount.
- Percentage of assigned capital.
- Percentage of physical Account.

A physical-Account percentage resolves once during Start validation from one
recorded Account snapshot.

It becomes an immutable amount for that BotGeneration.

It is never recalculated for each Order.

Order plans that exceed assigned Executor capital fail validation.

Start validation aggregates required capital by physical Account for funds and
margin checks.

Capital is initially a measurement and sizing basis.

Cross-process capital reservation remains TBD.

Price and margin changes may still cause later Venue rejection.

## Static Template

Each exact BotSpec owns one static canonical commented TOML template.

BotManager may return that template through the thin API.

CLI and GUI consumers edit TOML directly.

Nuubot adds no reflection generator, schema-driven custom form, or dynamic
template language.

## Hard Boundaries

- Controller never reads BotConfig from a file.
- Controller never decodes TOML.
- Controller never loads credentials.
- BotConfig never stores secrets.
- Config does not select arbitrary Signaler, Risk, or Executor kinds.
- Existing BotSpec identities never acquire new Config meaning.
- Validation failures create no Controller or Venue mutation.

## Open Decisions

- Live cross-process Account ownership and claims.
- Standalone Runner datastore writes.
- Complete Venue-specific Start validation checklist.
- BotConfig database schema and Server route shape.
