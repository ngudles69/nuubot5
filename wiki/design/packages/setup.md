# Setup Package

Status: Implemented for standalone BtBot; live Setup remains replay-oriented.
Covers: `internal/setup/setup.go`
Purpose: Prepare one shared Nuubot application harness for one Bot.

## Flow

```text
validate caller context
resolve repository root
load strict App Config
load exact stored Bot by SweepID and BotID
verify BotConfig SHA-256
resolve replay path below shared-data root
build typed BotSpec
validate Executor replay symbols
load fresh mainnet Meta through caller context
prepare Nuubot
log setup completed
```

Setup loads no credentials for Simulator backtests.

Setup never creates `context.Background()`.

Setup starts no goroutine or WebSocket.

## Nuubot

`setup.Setup` returns one shared `*setup.Nuubot` containing:

- Logger;
- complete App Config;
- stored Bot identity and exact BotConfig TOML;
- ReplayInput;
- typed BotSpec;
- the initialized TickClock or WallClock attached by the program owner;
- the Runner-owned shared Info endpoint when live;
- the Runner-owned shared WebSocket endpoint when live;
- global Meta reference data; and
- the per-Bot result path.

Nuubot contains shared infrastructure data.

It contains shared infrastructure, not procedural application behavior or features.

BtBot, Controller, BotCycle, Executors, and Accounts receive the same Nuubot pointer.

Components do not copy App Config, BotSpec, Meta, Bot identity, or ResultPath.

A component may retain `nuubot.Log` as its local logger reference.

BtBot or Runner creates and initializes its selected Clock, then attaches that
Clock to Nuubot before Controller initialization. Runtime code reads current
time through `nuubot.Clock.NowMS()`.

Runner also creates shared Info and WebSocket objects and attaches them before Controller initialization.

Setup still resolves replay paths, validates replay symbols, and loads Meta through the replay symbol.

Therefore current Setup does not yet provide complete live admission.

## Failure

Missing BotSpec identity, invalid Config hash, invalid replay path, invalid BotSpec,
missing replay symbols, missing Meta, delisted Meta, or caller cancellation fails Setup.
