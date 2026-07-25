# Setup Package

Status: Implemented for standalone BtRunner admission.
Covers: `internal/setup/setup.go`
Purpose: Admit external App, Bot, replay, Meta, and result-path inputs.

## Flow

```text
resolve repository root
load strict AppConfig
load exact stored Bot by SweepID and BotID
verify BotConfig SHA-256
resolve replay path below shared-data root
load fresh mainnet Meta through caller context
return immutable Admission
```

Setup loads no credentials for Simulator backtests.

Setup never creates `context.Background()`.

Setup starts no goroutine or WebSocket.

## Admission

The returned value contains AppConfig, stored BotConfig, ReplayInput, Meta, and
the per-Bot result path.

BtRunner passes those values to the exact BotSpec builder.

Controller never imports Setup.

## Failure

Missing BotSpec identity, invalid Config hash, invalid replay path, missing
Meta, delisted Meta, or caller cancellation fails before Controller admission.
