# BotSpec

Status: Implemented for the Macross Observer and TradeBot replay slice.
Covers: `internal/bot`, `internal/botspec`, stored BotConfig admission, and
BotDefinition.
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

Start strictly decodes the saved TOML through the exact BotSpec.

Controller receives the resulting typed BotDefinition.

Execution code never rereads the original TOML file.

Implemented exact IDs are:

- `macross_observer_bot`;
- `macross_trade_bot`.

Changing or deleting the import file changes nothing after saving.

JSON is for HTTP envelopes, status, and Results.

JSON is not a second persisted BotConfig representation.

## BotDefinition

BotDefinition is one immutable admitted Controller unit.

It contains:

- Exact BotSpecID.
- Exact saved BotConfig TOML and hash.
- Typed Signaler definition.
- Typed Risk definition.
- Typed Executor definitions.
- Account references.
- Required market and metadata inputs.
- Exactly one active BotCycle rule.
- Configured total sequential cycle limit, such as `max_cycles = 999`.
- BotCycle coordination rules.

It contains no database handle.

It contains no TOML parser.

It contains credential references, never secrets.

## Admission

Admission receives:

- Saved BotGeneration TOML.
- Exact BotSpecID.
- ReplayInput or live Runner inputs.
- Caller context.
- Operational AppConfig.
- Required metadata.

Admission returns one immutable BotDefinition or an error.

Simulator admission loads no private credentials.

Live Account admission resolves only referenced credentials.

Caller context owns cancellation and timeouts.

Setup creates no background context for admitted work.

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

No automatic cancel, flatten, transfer, leverage change, or state substitution occurs during admission.

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

Data-only symbols require Meta and market-data admission.

Traded symbols additionally require Account, clean-slate, and capital admission.

## Capital and Order Sizing

BotConfig declares capital for each Executor's Account-symbol resource.

Bot capital is the sum of admitted Executor capital in one reporting currency.

Bot equity is:

```text
Bot capital + cumulative net Bot PnL
```

Net Bot PnL includes realized PnL, unrealized PnL, fees, and funding.

Each Executor controls its Order sizing from its admitted Config.

Supported sizing may include:

- Fixed base quantity.
- Fixed quote amount.
- Percentage of assigned capital.
- Percentage of physical Account.

A physical-Account percentage resolves once during admission from one recorded
Account snapshot.

It becomes an immutable amount for that BotGeneration.

It is never recalculated for each Order.

Order plans that exceed assigned Executor capital fail admission.

Admission aggregates required capital by physical Account for funds and margin
checks.

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
- Admission failures create no Controller or Venue mutation.

## Open Decisions

- Live cross-process Account ownership and claims.
- Standalone Runner datastore writes.
- Shared versus process-local exchange WebSockets.
- Complete Venue-specific admission checklist.
- BotConfig database schema and Server route shape.
