# Setup Package

Status: Implemented for standalone BtBot.
Covers: `internal/setup/setup.go`
Purpose: Prepare global infrastructure and runtime inputs for one Bot.

## Flow

```text
resolve repository root
load strict App Config
load exact stored Bot by SweepID and BotID
verify BotConfig SHA-256
resolve replay path below shared-data root
load fresh mainnet Meta through caller context
prepare per-Bot result path
return immutable Infrastructure
```

Setup loads no credentials for Simulator backtests.

Setup never creates `context.Background()`.

Setup starts no goroutine or WebSocket.

## Infrastructure

The returned `setup.Infrastructure` contains:

- complete App Config;
- stored Bot identity and exact BotConfig TOML;
- ReplayInput;
- global Meta reference data; and
- the per-Bot result path.

BtBot transforms exact BotConfig TOML into one typed BotSpec.

BtBot passes complete Setup and BotSpec separately to Controller.

Setup values never enter BotSpec.

## Failure

Missing BotSpec identity, invalid Config hash, invalid replay path, missing Meta,
delisted Meta, or caller cancellation fails Setup.
