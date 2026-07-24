# Server

Status: Approved — unimplemented.
Covers: `cmd/nuubot-server/main.go`
Purpose: Run the optional master application process and expose operator control
through thin application boundaries.

## Canonical Sources

- Nuubot3: `D:/rust/nuubot3/nuubot/server/__main__.py`
- Nuubot3: `D:/rust/nuubot3/wiki/server-cli.md`
- Nuutrader6: `D:/rust/nuutrader6/wiki/architecture/server-gateway.md`

## Scope

Server is Nuubot's PocketBase-style application host.

It owns application startup, shared infrastructure, managers, API assembly,
web serving, readiness, and graceful service shutdown.

Runner, BtRunner, and SweepRunner remain standalone programs.

Their execution never requires Server to be running.

## Owner and Children

The server command owns Server.

Server directly owns:

- ProcessStore;
- RunnerControl;
- BotManager;
- SweepManager;
- PocketBase and Datastore;
- the thin API;
- the WebServer; and
- the HTTP application.

Shared exchange WebSocket ownership remains TBD.

Server may later own shared exchange connectivity without making standalone
Runner depend on it.

## Responsibilities

- Load server configuration and logging.
- Open shared stores.
- Check database connectivity and required tables.
- Trigger Meta availability and freshness checks.
- Construct and start direct services in dependency order.
- Expose thin API and web routes.
- Supervise service failures.
- Stop admission before unwinding services.
- Launch and supervise standalone Runner and SweepRunner processes through
  Managers.

## Does Not

- Run trading policy.
- Own Controller, BotCycle, Account, or Ledger.
- Interpret strategy configuration.
- Reach into Runner internals.
- Duplicate manager operations inside routes.
- Provide required runtime services to standalone execution.

## Lifecycle

`NewServer` constructs one stopped composition root.

`Init` opens resources and creates direct children.

`Start` performs bootstrap checks, starts services, then opens HTTP admission.

`Loop` supervises until cancellation or service failure.

`Stop` closes HTTP admission and unwinds direct services.

Healthy standalone execution does not automatically stop because Server stops
or crashes.

Reconnection to an already-running standalone process remains TBD.

## API Boundary

API decodes transport envelopes and forwards Bot requests to BotManager and
Sweep requests to SweepManager.

API does not query domain tables, choose data sources, validate trading policy,
or control execution directly.

Managers decide whether data comes from Datastore, process state, or stored
results.

## Deferred Server Capabilities

The following remain TBD and outside the current backtest implementation:

- Shared exchange WebSockets.
- Operational monitoring.
- Physical Account drawdown switch.
- Black Swan switch.
- External volatility, spike, sentiment, news, and economic inputs.
- Shared Server safety action path.

Server safety controls apply only to Server-supervised processes.

They are not Bot Risk and are not evaluated by isolated backtests.

## Required Proof

- Routes delegate through managers.
- Bootstrap checks complete before command admission.
- One service failure reaches Server.
- Partial startup cleans every started child.
- Shutdown order preserves Runner evidence.
- Direct Runner, BtRunner, and SweepRunner execution succeeds without Server.
