# Server

## Purpose

Compose shared services and expose operator control through thin application boundaries.

## Status

Approved — unimplemented.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/server/__main__.py`
- Nuubot3: `D:/rust/nuubot3/wiki/server-cli.md`
- Nuutrader6: `D:/rust/nuutrader6/wiki/architecture/server-gateway.md`

## Scope

Server owns application startup, shared resources, managers, DataEngine, HTTP assembly, and graceful service shutdown.

## Owner and Children

The server command owns Server.

Server directly owns:

- ProcessStore;
- RunnerControl;
- BotManager;
- SweepManager;
- DataEngine; and
- the HTTP application.

## Responsibilities

- Load server configuration and logging.
- Open shared stores.
- Construct and start direct services in dependency order.
- Recover approved active work before accepting commands.
- Expose thin API and web routes.
- Supervise service failures.
- Stop admission before unwinding services.

## Does Not

- Run trading policy.
- Own Runtime, BotCycle, Account, or Ledger.
- Interpret strategy configuration.
- Reach into Runner internals.
- Duplicate manager operations inside routes.

## Lifecycle

`NewServer` constructs one stopped composition root.

`Init` opens resources and creates direct children.

`Start` performs recovery, starts services, then opens HTTP admission.

`Loop` supervises until cancellation or service failure.

`Stop` closes HTTP admission and unwinds direct services.

## Required Proof

- Routes delegate through managers.
- Recovery completes before command admission.
- One service failure reaches Server.
- Partial startup cleans every started child.
- Shutdown order preserves terminal Runner evidence.

