# BotManager

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Own operator-facing Bot configuration and lifecycle commands.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/bots/botmgr.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/server/process.py`

## Scope

BotManager validates Server-side Bot requests, persists admitted Bot
configuration, and coordinates standalone Runner processes.

## Owner and Children

Server owns BotManager.

RunnerControl owns process commands.

Runner owns Controller and every execution-local child.

## Responsibilities

- Create, clone, read, list, update, archive, and delete stored Bots.
- Validate complete Bot configuration before persistence.
- Start and stop one Server-supervised Runner through RunnerControl.
- Return stable status and available result views.
- Reject invalid state transitions.
- Retrieve static BotConfig templates from the exact BotSpec catalogue.

## Does Not

- Construct or call Controller directly.
- Import Runner internals.
- Reconcile Accounts.
- Manage Sweep Bots.
- Decode live events.
- Spawn unmanaged work.
- Own exchange credentials beyond validated references.

## Invariants

- Bot identity is stable.
- One lifecycle owner controls one active Bot.
- Stored configuration remains the restart source.
- Commands are reserved and completed once.
- Direct Runner execution remains valid without BotManager or Server.

## Required Proof

- Invalid configuration never persists.
- Duplicate start cannot create two active Runners.
- Commands reach only the selected Bot.
- Status reflects durable and live truth without exposing internals.
- Direct Runner can complete without Server availability.

## Open Decisions

- Cross-process duplicate-start and Account-symbol claims.
- Reconnection to independently started Runner processes.
- Standalone Runner status publication while Server is stopped.
