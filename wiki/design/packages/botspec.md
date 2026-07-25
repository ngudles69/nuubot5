# BotSpec Package

Status: Implemented for Macross Observer and TradeBot.
Covers: `internal/botspec/**`
Purpose: Admit exact BotSpec Config and build one BotDefinition.

The explicit catalogue recognizes:

- `macross_observer_bot`;
- `macross_trade_bot`.

Each ID has one typed Config decoder and one static commented TOML template.

Required fields validate.

Extra fields are stored and ignored.

Duplicate TOML keys, duplicate Executor resources, invalid decimals, invalid
roles, and unknown IDs fail.

Build constructs Signaler and Risks, attaches admitted Meta and result paths,
and returns one immutable BotDefinition.

No reflection, plugin, DSL, runtime compilation, fallback, alias, or
compatibility path exists.
