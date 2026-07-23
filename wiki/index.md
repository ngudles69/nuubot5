# Nuubot5 Wiki

This wiki is the durable project truth. `HANDOFF.md` contains only current
restart state.

## Required at Startup

1. [Restart Handoff](../HANDOFF.md)
2. [Project Contract](PROJECT.md)
3. [User Working Profile](USER.md) — prescriptive
4. [Chief-of-Staff Contract](SOUL.md) — prescriptive
5. [Architecture](ARCHITECTURE.md)
6. [Design](DESIGN.md)

## Required Before Coding

1. [Go Coding Style](coding/STYLE.md)
2. [Coding Rules](coding/RULES.md)

[Abbreviations](abbreviations.md) owns canonical project spellings.

## Runtime Logic

- [BtRunner](logic/btrunner.md)
- [Signaler](logic/signaler.md)
- [Executor](logic/executor.md)
- [Risk](logic/risk.md)
- [Live Runner](logic/runner.md)

Account, Ledger, Trade, Order, Fill, Simulator, server, CLI, and web components
are not implemented. Their pages MUST NOT be created until their real contracts
exist.

## Current State

- Go source: `cmd/**` and `internal/**`
- Canonical proof harness: `rtest.sh`
