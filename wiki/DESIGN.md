# Nuubot5 Design

## Purpose

This page indexes package and concept design.

Nuubot5 source proves implementation. Reservation files prove names only.

## Organization

- [`design/packages`](design/packages/) contains exactly one page per Go package.
- [`design/concepts`](design/concepts/) contains flows, programs, venues, types, and cross-package rules.
- [`design/runner.md`](design/runner.md) owns the standalone live Runner design.
- [`design/server.md`](design/server.md) owns the Server, unified binary, and child-process design.
- [`design/startup.md`](design/startup.md) owns Runner startup and crash recovery across all live networks.
- [`design/entities.md`](design/entities.md) shows backtest and live runtime ownership cardinality.
- [`design/hyperliquid`](design/hyperliquid/) contains the internal Hyperliquid boundary details.
- `internal/toolkit` groups reusable packages. It is not a Go package.

## User Review Checklist

Last updated: 2026-07-24 13:33:22 +08:00

```text
Component                          Review         Last reviewed
nuubot-bt-bot                    DONE           2026-07-24 12:55:22 +08:00
`-- BtBot                       DONE           2026-07-24 12:55:22 +08:00
    |-- Setup                      PARTIAL        2026-07-24 13:33:22 +08:00
    |-- ReplayReader               PARTIAL        2026-07-24 12:55:22 +08:00
    |-- TickClock                  NOT REVIEWED   —
    `-- Controller                 DONE           2026-07-24 12:55:22 +08:00
        |-- Signaler               PARTIAL        2026-07-24 12:55:22 +08:00
        |   |-- Macross            NOT REVIEWED   —
        |   `-- RSI                NOT REVIEWED   —
        |-- Risk                   PARTIAL        2026-07-24 12:55:22 +08:00
        |   `-- BalancedRisk       NOT REVIEWED   —
        `-- BotCycle               PARTIAL        2026-07-24 12:55:22 +08:00
            `-- Executor           NOT REVIEWED   —
                `-- Observer       NOT REVIEWED   —
```

The user owns these checklist states: `DONE`, `PARTIAL`, `NOT REVIEWED`, and
`TO CODE`. `PARTIAL` remains open. `NOT REVIEWED` has no review timestamp.

## Implementation and Review

| Component | State | Last reviewed | Note |
|---|---|---|---|
| WallClock | DONE | 2026-07-24 17:42:58 +08:00 | Implemented and proven. |
| Runner | TO CODE | — | Command and lifecycle scaffold exist; live runtime remains unavailable. |
| Select the SDK | DONE | 2026-07-24 17:42:58 +08:00 | Rewrite the required official API inside Nuubot. |
| Simulator parity | PARTIAL | — | Internal behavior implemented; external parity pending. |
| Account | NOT REVIEWED | — | Implemented for Simulator. |
| Ledger | NOT REVIEWED | — | Implemented with `none` and `max`. |
| Trade | NOT REVIEWED | — | Implemented. |
| Order | NOT REVIEWED | — | Implemented. |
| Fill | NOT REVIEWED | — | Implemented. |
| Simulator | NOT REVIEWED | — | Implemented for BtBot. |
| TradeExecutor | NOT REVIEWED | — | Simulator-first vertical slice implemented. |
| GridExecutor | NOT REVIEWED | — | Arithmetic, equal-capital Simulator slice implemented. |
| PocketBase | TO CODE | — | Approved design; implementation deferred. |
| Meta | NOT REVIEWED | — | Mainnet perpetual Meta implemented. |
| Setup | PARTIAL | 2026-07-24 | Shared database and mainnet Meta added after user review. |
| Hyperliquid SDK selection | DONE | 2026-07-24 17:42:58 +08:00 | No external SDK adoption. |

This is the user's coding checklist. It does not replace package implementation
status.

## Hyperliquid Source

| Question | Decision |
|---|---|
| What | Build Nuubot's required Hyperliquid protocol boundary. |
| Where | Source belongs in `internal/hyperliquid`. |
| Design | [Hyperliquid design](design/hyperliquid.md) owns the complete boundary. |
| Why internal | Nuubot5 is the only confirmed consumer. Account and Venue changes remain atomic. |
| Why rewrite | The audited SDK contains unrelated dependencies, generated code, hidden network construction, and unsafe WebSocket lifecycle behavior. |
| How | Rewrite from the official API. Consult audited reference code only when useful. |
| In | Transport, signing, protocol types, validation, and Venue mapping. |
| Out | Trading policy, domain ownership, Meta persistence, Simulator, and reconciliation decisions. |
| Status | Public Info and parity probe implemented. Exchange, signing, and WebSocket transport remain pending. |

Nuubot independently rewrites its Hyperliquid boundary from the
[official Hyperliquid API documentation](https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api).

[Sonirico's Go client](https://github.com/sonirico/go-hyperliquid) is the
secondary implementation reference.

Python `async_hyperliquid` is the third known-working reference.

Nuubot targets official Hyperliquid semantics.

Nuutrader6 supplies proven matching behavior.

Selected client-visible responses target `async_hyperliquid` 0.4.8 output parity.

Nuubot does not target either reference library's implementation structure.

Nuubot admits only required, audited code. It does not import or preserve either reference library.

## Packages

| Package | Status | Purpose |
|---|---|---|
| [account](design/packages/account.md) | Implemented | Coordinate venue requests and ledger evidence. |
| [bot](design/packages/bot.md) | Implemented | Define immutable Bot identity. |
| [botcycle](design/packages/botcycle.md) | Implemented | Coordinate Executors for one configured entry Signal. |
| [botspec](design/packages/botspec.md) | Implemented | Validate and shape exact BotConfig into typed BotSpec. |
| [btbot](design/packages/btbot.md) | Implemented | Execute one complete historical replay. |
| [btsweep](design/packages/btsweep.md) | Partial | Validate and expand Sweep templates; persistence and execution remain deferred. |
| [config](design/packages/config.md) | Implemented | Decode and validate immutable settings. |
| [controller](design/packages/controller.md) | Implemented | Own signals, risks, cycles, capital, and stop decisions. |
| [datastore](design/packages/datastore.md) | Implemented | Load one stored BotConfig and replay input. |
| [executor](design/packages/executor.md) | Implemented | Own execution policy boundaries. |
| [fill](design/packages/fill.md) | Implemented | Preserve immutable execution facts. |
| [hyperliquid](design/packages/hyperliquid.md) | Implemented | Own Hyperliquid protocol transport and translations. |
| [ledger](design/packages/ledger.md) | Implemented | Own trade, order, and fill evidence. |
| [market](design/packages/market.md) | Implemented | Carry validated market events. |
| [meta](design/packages/meta.md) | Implemented | Own mainnet perpetual instrument metadata. |
| [order](design/packages/order.md) | Implemented | Own submitted order state and fills. |
| [ohlcv](design/packages/ohlcv.md) | Implemented | Load validated OHLCV ranges. |
| [parity](design/packages/parity.md) | Implemented | Admit and run permanent parity probes. |
| [parity/info](design/packages/info.md) | Implemented | Capture `/info` payloads and translations. |
| [replay](design/packages/replay.md) | Implemented | Stream validated historical market data. |
| [resultpublisher](design/packages/resultpublisher.md) | Implemented | Publish terminal per-Bot SQLite evidence. |
| [risk](design/packages/risk.md) | Implemented | Assess configured risk policy. |
| [runner](design/packages/runner.md) | Scaffold | Own one standalone live Bot lifecycle. |
| [setup](design/packages/setup.md) | Implemented | Prepare one validated BtBot context. |
| [signaler](design/packages/signaler.md) | Implemented | Calculate and serve ordered Signal packages. |
| [simulator](design/packages/simulator.md) | Implemented | Provide venue-shaped simulated execution. |
| [trade](design/packages/trade.md) | Implemented | Own strategy-level orders and evidence. |
| [toolkit/clock](design/packages/clock.md) | Implemented | Provide deterministic clock mechanics. |
| [toolkit/logging](design/packages/logging.md) | Implemented | Write exact-format append-only file logs. |

Package pages state their implemented and pending boundaries.

## Approved Design

Backtest hardcuts are implemented. Live and process-control contracts remain
deferred.

| Target | Owner |
|---|---|
| Exact complete BotSpec and stored TOML BotConfig | [BotSpec](design/concepts/bot-spec.md) |
| Controller decision ownership | [Controller](design/packages/controller.md) |
| Persistent traffic-light strategy source | [Signaler](design/packages/signaler.md) |
| Persistent Risk gates and exit source | [Risk](design/packages/risk.md) |
| Coordinated exchange-style campaign and flat-stop proof | [BotCycle](design/packages/botcycle.md) |
| Ongoing detachable state observation | [Telemetry](design/telemetry.md) |
| Terminal detachable analysis and rendering | [RunReport](design/runreport.md) |
| AppConfig, BotConfig, ReplayInput, and Credentials split | [Config](design/packages/config.md) |
| Typed saved-Config and fail-closed Meta admission | [Setup](design/packages/setup.md) |
| Stored TOML and active Account-symbol claims | [Datastore](design/packages/datastore.md) |
| Standalone Runner, BtBot, and BtSweep execution | [Runner](design/runner.md) |
| Thin Server API and Manager-to-process boundaries | [Server](design/server.md) |
| Implemented Sweep template validation and expansion | [BtSweep package](design/packages/btsweep.md) |
| Reusable Sweep records and standalone execution | [SweepManager](design/concepts/sweep-manager.md) |
| ControllerResult, BotCycleResult, and ExecutorResult hierarchy | [BotCycle](design/packages/botcycle.md) |
| Arithmetic Grid levels, lifecycle, flattening, and proof | [GridExecutor](design/concepts/grid-executor.md) |

Current implementation sections remain authoritative until each approved
hardcut is implemented and proven.

## Concepts

| Concept | Purpose |
|---|---|
| [Telemetry](design/telemetry.md) | Ongoing detachable state snapshots for backtest, monitoring, and run playback. |
| [RunReport](design/runreport.md) | Terminal detachable calculations, aggregation, and standardized rendering. |
| [AccountSnapshot](design/concepts/account-snapshot.md) | Immutable account state. |
| [BalancedRisk](design/concepts/balanced-risk.md) | Current balanced risk implementation. |
| [BotSpec](design/concepts/bot-spec.md) | Exact complete Bot design, TOML Config, and admission contract. |
| [BotManager](design/concepts/bot-manager.md) | Server-side Bot requests and standalone Runner process control. |
| [CLOID](design/concepts/cloid.md) | Deterministic client-order identity. |
| [DataEngine](design/concepts/data-engine.md) | Earlier shared-feed candidate with ownership TBD. |
| [Execution](design/concepts/execution.md) | Persist, submit, normalize, and reconcile flow. |
| [Filesystem](design/concepts/filesystem.md) | Mutable workspace layout and deployment mount. |
| [Hyperliquid](design/hyperliquid.md) | Internal Hyperliquid protocol boundary. |
| [Hyperliquid parity probe](design/hyperliquid/parity.md) | Permanent testnet and Simulator API-drift harness. |
| [IngestBBO](design/concepts/ingestbbo.md) | Simulator-only BBO matching input. |
| [Live events](design/concepts/live-events.md) | Live event routing. |
| [Macross signaler](design/concepts/macross-signaler.md) | EMA crossover implementation. |
| [nuubot-bt-bot](design/concepts/nuubot-bt-bot.md) | Standalone historical replay command. |
| [Observer executor](design/concepts/observer-executor.md) | Observer execution implementation. |
| [PocketBase](design/concepts/pocketbase.md) | Server-owned web, API, authentication, realtime, and SQLite framework. |
| [Process store](design/concepts/process-store.md) | Process persistence boundary. |
| [Reconciliation](design/concepts/recon.md) | Venue-to-ledger repair flow. |

| [Replay](design/concepts/replay.md) | End-to-end historical replay flow. |
| [Result publisher](design/concepts/result-publisher.md) | Terminal replay publishing. |
| [RSI signaler](design/concepts/rsi-signaler.md) | RSI implementation. |

| [RunnerControl](design/concepts/runner-control.md) | Runner lifecycle commands. |
| [Controller store](design/concepts/controller-store.md) | Candidate Controller persistence boundary. |

| [Shutdown](design/concepts/shutdown.md) | Ordered resource release. |
| [Signal](design/concepts/signal.md) | Immutable strategy decision. |
| [Simulator parity](design/concepts/simulator-parity.md) | Exchange behavior and response parity boundary. |
| [SweepManager](design/concepts/sweep-manager.md) | Server-side Sweep requests and standalone BtSweep control. |
| [Toolkit](design/concepts/toolkit.md) | Portable package rules. |
| [TradeExecutor](design/concepts/trade-executor.md) | First Account-owning Executor design. |
| [GridExecutor](design/concepts/grid-executor.md) | Arithmetic Grid calculations, Orders, re-entry, boundaries, and flattening. |
| [Trading schema](design/concepts/trading-schema.md) | Per-Bot SQLite trading evidence DDL. |
| [Trading state tranche](design/concepts/trading-state.md) | Next-tranche assessment and implementation order. |
| [Venue](design/concepts/venue.md) | Normalized execution truth. |
| [WallClock](design/concepts/wall-clock.md) | Live cadence behavior. |

## Boundaries

- `Status`, `Covers`, and `Purpose` form the standard design header.
- `Covers` names current Nuubot source.
- `Canonical Sources` names Nuubot4 source.
- Confirmed implementation facts update the owning design page in the same change.
- Package pages own package contracts.
- Concept pages may span packages.
- A concept page does not create a Go package.
- Each package has one canonical design page.
- Source and package pages must remain aligned.
