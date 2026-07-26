# Bot Package

Status: Implemented.
Covers: `internal/bot/bot.go`
Purpose: Define one immutable admitted Bot identity and composition.

Current `bot.Identity` contains SweepID, BotID, exact BotSpecID, and Config hash.

The target BotID is globally unique. SweepID becomes optional grouping
provenance after datastore hardcut.

`bot.Definition` contains Signal symbol, cycle limit, Signaler, Risks, and
ordered Executor Specs.

Controller accepts only this definition.

The package owns no parsing, datastore, lifecycle, or trading behavior.

A Bot template is mutable import input for one complete scalar Config.

A generated Bot record is immutable. It owns exact Config TOML and hash.

Revising a Bot or Sweep template creates new Bot IDs. An unchanged Sweep rerun
reuses existing Bot IDs and replaces their current result targets.

Executor `id` values belong to Bot Config. Generated Bot record IDs do not
belong in templates.
