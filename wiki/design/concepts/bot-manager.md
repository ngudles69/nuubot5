# BotManager

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Own operator-facing Bot configuration and lifecycle commands.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/bots/botmgr.py`
- Nuubot3: `D:/rust/nuubot3/nuubot/server/process.py`

## Scope

BotManager validates Bot requests, persists admitted Bot configuration, and coordinates Runner lifecycle through approved process boundaries.

## Owner and Children

Server owns BotManager.

BotManager owns active Runner handles when Runners share the Server process.

RunnerControl owns commands when Runners use standalone processes.

## Responsibilities

- Create, clone, read, list, update, archive, and delete stored Bots.
- Validate complete Bot configuration before persistence.
- Start, pause, resume, and stop one Bot through its lifecycle owner.
- Return stable status and telemetry views.
- Reject invalid state transitions.

## Does Not

- Call Runtime directly.
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

## Required Proof

- Invalid configuration never persists.
- Duplicate start cannot create two active Runners.
- Commands reach only the selected Bot.
- Status reflects durable and live truth without exposing internals.
