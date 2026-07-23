# Nuubot5 Design

## Purpose

This page gives the high-level object catalog, status, purpose, and owner.

Detailed contracts live in [`design/**`](design/).

Nuubot4 process remains canonical. Approved design is not implemented proof.

## Status

- **Implemented:** Nuubot5 source exists and participates in a proven path.
- **Stub:** Source exists but intentionally omits real domain behavior.
- **Approved — unimplemented:** The boundary is accepted. Nuubot5 has no working implementation.

No other status is valid.

## Ownership

```text
BtRunner
|-- ReplayReader
|-- TickClock
|-- ResultPublisher                    approved
`-- Runtime
    |-- RuntimeStore                    approved
    |-- Signaler
    |-- Risks
    `-- BotCycle
        `-- Executors
            `-- Accounts               approved
                |-- Venue              approved
                `-- Ledger             approved
                    `-- Trades
                        `-- Orders
                            `-- Fills

Server                                  approved
|-- DataEngine
|-- ProcessStore
|-- BotManager
|   `-- Runners
|-- SweepManager
`-- HTTP application
```

Executor owns Accounts. Runtime receives AccountSnapshots without owning Accounts.

DataEngine owns shared acquisition. Runner owns subscriptions and local state, then routes events through Runtime.

## Implemented

| Object | Purpose | Owner | Detail |
|---|---|---|---|
| Setup | Load configuration and one validated BotSpec. | BtRunner construction | [setup](design/setup.md) |
| Configuration | Decode and validate immutable settings. | Command and consumers | [config](design/config.md) |
| BtRunner | Build, run, stop, and prove one replay. | Command | [btrunner](design/btrunner.md) |
| ReplayReader | Stream validated Parquet BBO values. | BtRunner | [replay-reader](design/replay-reader.md) |
| TickClock | Convert replay time into pass decisions. | BtRunner | [tick-clock](design/tick-clock.md) |
| Runtime | Own signals, risks, cycles, and stop decisions. | BtRunner | [runtime](design/runtime.md) |
| BotCycle | Coordinate Executors for one accepted Signal. | Runtime | [botcycle](design/botcycle.md) |
| Signaler | Calculate and release ordered Signals. | Runtime | [signaler](design/signaler.md) |
| MacrossSignaler | Calculate regime-filtered EMA crossovers. | Signaler | [macross-signaler](design/macross-signaler.md) |
| RsiSignaler | Calculate volume-confirmed RSI signals. | Signaler | [rsi-signaler](design/rsi-signaler.md) |
| Signal | Carry one immutable strategy decision. | Signaler creates; Runtime consumes | [signal](design/signal.md) |
| Executor | Define the execution policy boundary and factory. | BotCycle | [executor](design/executor.md) |
| ObserverExecutor | Observe exits without placing orders. | BotCycle | [observer-executor](design/observer-executor.md) |
| Risk | Define risk assessment and factory behavior. | Runtime | [risk](design/risk.md) |
| BBO | Carry one validated timestamp and price. | Passed by value | [bbo](design/bbo.md) |
| Bar | Carry validated OHLCV data and requirements. | Signaler retains loaded values | [bar](design/bar.md) |
| Datastore | Load one BotSpec from read-only SQLite. | Setup | [datastore](design/datastore.md) |
| Replay | Define the complete historical replay process. | BtRunner | [replay](design/replay.md) |
| Shutdown | Close active work and children in ownership order. | Each lifecycle owner | [shutdown](design/shutdown.md) |

## Stub

| Object | Purpose | Owner | Detail |
|---|---|---|---|
| BalancedRisk | Prove the Risk call path without real policy. | Runtime | [balanced-risk](design/balanced-risk.md) |

## Approved — Unimplemented

| Object | Purpose | Owner | Detail |
|---|---|---|---|
| Runner | Supervise one live Bot and route events through Runtime. | BotManager | [runner](design/runner.md) |
| WallClock | Publish bounded live cadence events. | Runner | [wall-clock](design/wall-clock.md) |
| Server | Compose and supervise shared services. | Server command | [server](design/server.md) |
| BotManager | Own active Runner lifecycles. | Server | [bot-manager](design/bot-manager.md) |
| SweepManager | Coordinate Sweep-level work. | Server | [sweep-manager](design/sweep-manager.md) |
| DataEngine | Share external acquisition, validation, and multiplexing. | Server | [data-engine](design/data-engine.md) |
| RuntimeStore | Persist Runtime and BotCycle state. | Runtime | [runtime-store](design/runtime-store.md) |
| ProcessStore | Persist process and manager state. | Server | [process-store](design/process-store.md) |
| RunnerControl | Carry validated Runner lifecycle commands. | BotManager | [runner-control](design/runner-control.md) |
| ResultPublisher | Publish terminal replay results. | BtRunner | [result-publisher](design/result-publisher.md) |
| AccountSnapshot | Return immutable coherent Account state. | Account creates; Runtime consumes | [account-snapshot](design/account-snapshot.md) |
| Account | Coordinate Venue requests and Ledger evidence. | Executor | [account](design/account.md) |
| Ledger | Own Trades, Orders, and Fills. | Account | [ledger](design/ledger.md) |
| Trade | Own strategy-level Orders and evidence. | Ledger | [trade](design/trade.md) |
| Order | Own one submitted order and its Fills. | Trade | [order](design/order.md) |
| Fill | Preserve one immutable execution fact. | Order | [fill](design/fill.md) |
| Venue | Normalize live, testnet, or simulated execution truth. | Account | [venue](design/venue.md) |
| Simulator | Provide venue-shaped simulated truth. | Account through Venue | [simulator](design/simulator.md) |
| Execution | Define persist, submit, normalize, and reconcile flow. | Executor | [execution](design/execution.md) |
| Reconciliation | Rebuild Ledger evidence from Venue truth. | Account | [recon](design/recon.md) |
| LiveEvents | Route subscribed BBO and user events into Runtime. | Runner | [live-events](design/live-events.md) |
| Recovery | Restore persisted owners before accepting work. | Each persisted owner | [recovery](design/recovery.md) |
| CLOID | Provide deterministic client-order identity. | Execution boundary | [cloid](design/cloid.md) |
| Hyperliquid | Record venue-specific required behavior. | Hyperliquid Venue implementation | [hyperliquid](design/hyperliquid.md) |
| Errors | Define standard error ownership and wrapping. | Package and process boundaries | [errors](design/errors.md) |
| Logging | Define process logger configuration and structured events. | Executable or server boundary | [logging](design/logging.md) |

## Boundaries

- Runtime owns decisions and RuntimeStore. It does not own Accounts.
- Executor owns Accounts. It does not bypass Account to reach Venue or Ledger.
- Account owns Venue and Ledger. Venue never mutates Ledger.
- Ledger owns Trade, Order, and Fill evidence. It does not call Venue.
- DataEngine owns shared acquisition. It does not call Runtime.
- Runner owns subscriptions and local state. It routes events through Runtime.
- Simulator stores venue-shaped truth. It does not store domain objects.
- Server composes services. It does not contain trading policy.

SDK, transport, physical schema, and dependency choices require later approval.
