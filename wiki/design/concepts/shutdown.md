# Shutdown

Status: Implemented for BtBot and Simulator TradeBot.
Covers: `internal/btbot`, `internal/controller`, `internal/botcycle`,
`internal/executor`, and `internal/account`
Purpose: Stop admission, flatten trading state, and preserve terminal evidence.

## Order

```text
BtBot Stop
  stop TickClock
  stop ReplayReader
  stop Controller

Controller Stop
  stop active BotCycle
  stop Risks
  stop Signaler

BotCycle Stop
  stop Executors in reverse order

TradeExecutor Stop
  reconcile
  cancel active Orders
  close exposure
  reconcile
  prove flat and inactive
  capture Account result
  stop Account
```

Every Stop is idempotent.

The first Controller stop reason wins.

BtBot publishes only when Reader stop, Controller stop, and replay proof all
succeed.

Fatal cleanup errors propagate to the command.
