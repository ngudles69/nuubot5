# Runtime Entity Ownership

## Backtest

```text
1 nuubot-backtest binary
└── 1:1 BtBot
    └── 1:1 Controller
        ├── 1:1 Signaler
        ├── 1:N Risks
        ├── 0:N closed BotCycles
        └── 0:1 active BotCycle
            └── 1:N Executors
                └── 0:1 Account
                    ├── 1:1 Exchange Connection or Simulator
                    └── 1:1 Ledger
                        └── 1:N Trades
                            └── 1:N Orders
                                └── 1:N Fills
```

## Live

```text
1 nuubot-live binary
└── 1:1 Runner
    └── 1:1 Controller
        ├── 1:1 Signaler
        ├── 1:N Risks
        ├── 0:N closed BotCycles
        └── 0:1 active BotCycle
            └── 1:N Executors
                └── 0:1 Account
                    ├── 1:1 Exchange Connection or Simulator
                    └── 1:1 Ledger
                        └── 1:N Trades
                            └── 1:N Orders
                                └── 1:N Fills
```

## Runtime Drivers

```text
Backtest
  ReplayReader drives ticks
  TickClock advances from replay timestamps

Live
  WallClock drives Controller timers
  WebSocket supplies requested live data
  shared Info supplies requested public REST data
```

BtBot owns ReplayReader, TickClock, and MarketData lifecycle.

Runner owns WallClock, MarketData, Info, and WebSocket lifecycle.

Meta refresh uses a separate mainnet Info object.

The credentialed Exchange endpoint belongs to each Account, not Runner.
