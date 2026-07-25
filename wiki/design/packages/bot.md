# Bot Package

Status: Implemented.
Covers: `internal/bot/bot.go`
Purpose: Define one immutable admitted Bot identity and composition.

`bot.Identity` contains SweepID, BotID, exact BotSpecID, and Config hash.

`bot.Definition` contains Signal symbol, cycle limit, Signaler, Risks, and
ordered Executor Specs.

Controller accepts only this definition.

The package owns no parsing, datastore, lifecycle, or trading behavior.
