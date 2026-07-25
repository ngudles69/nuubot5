# Shutdown

Status: Implemented for BtRunner and Simulator TradeBot.
Covers: `internal/btrunner`, `internal/controller`, `internal/botcycle`,
`internal/executor`, and `internal/account`
Purpose: Stop admission, flatten trading state, and preserve terminal evidence.

## Order

```text
BtRunner Stop
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

BtRunner publishes only when Reader stop, Controller stop, and replay proof all
succeed.

Fatal cleanup errors propagate to the command.
