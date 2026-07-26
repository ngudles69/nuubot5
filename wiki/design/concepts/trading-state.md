# Trading State

Status: Implemented for standalone Simulator backtests.
Covers: `internal/controller`, `internal/botcycle`, `internal/executor`,
`internal/account`, `internal/ledger`, and `internal/simulator`
Purpose: Define the complete current trading ownership path.

```text
BtBot
`-- Controller
    `-- BotCycle
        `-- Executor
            `-- Account
                |-- Simulator
                `-- Ledger
                    `-- Trade
                        `-- Order
                            `-- Fill
```

Each mutable object has one direct owner.

Values move upward.

Controller never reaches through BotCycle.

## Per-Cycle State

TradeExecutor owns one Account for one BotCycle.

Every Executor resource is distinct inside the Bot.

Controller carries each resource's terminal Account equity into its next
cycle.

Bot capital remains the sum of declared Executor capital.

Bot net PnL equals current Bot equity minus declared capital.

## Persistence

`none` keeps trading state in memory and publishes once after success.

`max` persists accepted Account child state during execution.

Full process recovery remains deferred.

## Result

The terminal hierarchy preserves Bot, cycle, Executor, resource, Account,
Trade, Order, Fill, Simulator, Signal, Risk, capital, PnL, equity, drawdown,
and replay proof.
