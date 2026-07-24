# Nuubot5 Design

## Purpose

This page indexes package and concept design.

Nuubot5 source proves implementation. Reservation files prove names only.

## Organization

- [`design/packages`](design/packages/) contains exactly one page per Go package.
- [`design/concepts`](design/concepts/) contains flows, programs, venues, types, and cross-package rules.
- `internal/toolkit` groups reusable packages. It is not a Go package.

## Packages

| Package | Status | Purpose |
|---|---|---|
| [account](design/packages/account.md) | Reserved | Coordinate venue requests and ledger evidence. |
| [botcycle](design/packages/botcycle.md) | Implemented | Coordinate Executors for one accepted Signal. |
| [btrunner](design/packages/btrunner.md) | Implemented | Execute one complete historical replay. |
| [config](design/packages/config.md) | Implemented | Decode and validate immutable settings. |
| [datastore](design/packages/datastore.md) | Implemented | Load one validated BotSpec. |
| [executor](design/packages/executor.md) | Implemented | Own execution policy boundaries. |
| [fill](design/packages/fill.md) | Reserved | Preserve immutable execution facts. |
| [ledger](design/packages/ledger.md) | Reserved | Own trade, order, and fill evidence. |
| [market](design/packages/market.md) | Implemented | Carry validated market events. |
| [meta](design/packages/meta.md) | Reserved | Own market instrument metadata. |
| [order](design/packages/order.md) | Reserved | Own submitted order state and fills. |
| [ohlcv](design/packages/ohlcv.md) | Implemented | Load validated OHLCV ranges. |
| [replay](design/packages/replay.md) | Implemented | Stream validated historical market data. |
| [risk](design/packages/risk.md) | Implemented | Assess configured risk policy. |
| [runtime](design/packages/runtime.md) | Implemented | Own signals, risks, cycles, and stop decisions. |
| [setup](design/packages/setup.md) | Implemented | Prepare one validated BtRunner context. |
| [signaler](design/packages/signaler.md) | Implemented | Calculate and release ordered Signals. |
| [simulator](design/packages/simulator.md) | Reserved | Provide venue-shaped simulated execution. |
| [trade](design/packages/trade.md) | Reserved | Own strategy-level orders and evidence. |
| [toolkit/clock](design/packages/clock.md) | Implemented | Provide deterministic clock mechanics. |
| [toolkit/logging](design/packages/logging.md) | Implemented | Write exact-format append-only file logs. |

Reserved packages contain only an approved package declaration.

## Concepts

| Concept | Purpose |
|---|---|
| [AccountSnapshot](design/concepts/account-snapshot.md) | Immutable account state. |
| [BalancedRisk](design/concepts/balanced-risk.md) | Current balanced risk implementation. |
| [BotManager](design/concepts/bot-manager.md) | Active Runner lifecycle ownership. |
| [CLOID](design/concepts/cloid.md) | Deterministic client-order identity. |
| [DataEngine](design/concepts/data-engine.md) | Shared market-data acquisition. |
| [Execution](design/concepts/execution.md) | Persist, submit, normalize, and reconcile flow. |
| [Hyperliquid](design/concepts/hyperliquid.md) | Venue-specific behavior. |
| [Live events](design/concepts/live-events.md) | Live event routing. |
| [Macross signaler](design/concepts/macross-signaler.md) | EMA crossover implementation. |
| [nuubot-btrunner](design/concepts/nuubot-btrunner.md) | Standalone historical replay command. |
| [Observer executor](design/concepts/observer-executor.md) | Observer execution implementation. |
| [PocketBase](design/concepts/pocketbase.md) | Server-owned web, API, authentication, realtime, and SQLite framework. |
| [Process store](design/concepts/process-store.md) | Process persistence boundary. |
| [Reconciliation](design/concepts/recon.md) | Venue-to-ledger repair flow. |
| [Recovery](design/concepts/recovery.md) | Startup state restoration. |
| [Replay](design/concepts/replay.md) | End-to-end historical replay flow. |
| [Result publisher](design/concepts/result-publisher.md) | Terminal replay publishing. |
| [RSI signaler](design/concepts/rsi-signaler.md) | RSI implementation. |
| [Runner](design/concepts/runner.md) | Live Bot supervision. |
| [RunnerControl](design/concepts/runner-control.md) | Runner lifecycle commands. |
| [Runtime store](design/concepts/runtime-store.md) | Runtime persistence boundary. |
| [Server](design/concepts/server.md) | Shared service composition. |
| [Shutdown](design/concepts/shutdown.md) | Ordered resource release. |
| [Signal](design/concepts/signal.md) | Immutable strategy decision. |
| [SweepManager](design/concepts/sweep-manager.md) | Sweep-level coordination. |
| [Toolkit](design/concepts/toolkit.md) | Portable package rules. |
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
