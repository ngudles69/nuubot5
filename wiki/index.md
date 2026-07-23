# Nuubot5 Wiki

This wiki is durable project truth. `HANDOFF.md` owns current restart state.

## Required at Startup

1. [Restart Handoff](../HANDOFF.md)
2. [Project](PROJECT.md)
3. [User Contract](USER.md) — prescriptive
4. [Chief-of-Staff Contract](SOUL.md) — prescriptive
5. [Architecture](ARCHITECTURE.md)
6. [Design](DESIGN.md)

## Required Before Coding

1. [Go Coding Style](coding/STYLE.md)
2. [Coding Rules](coding/RULES.md)

[Abbreviations](abbreviations.md) owns canonical project spellings.

## Detailed Design

[DESIGN.md](DESIGN.md) indexes all implemented, stub, and approved-unimplemented design pages.

Detailed object and process contracts live in [`design/**`](design/).

Legacy process pages remain in [`logic/**`](logic/) until separately migrated.

## Current Implementation

- Go source: `cmd/**` and `internal/**`
- Canonical proof harness: `rtest.sh`
- Canonical build tag: `-tags noasm`
