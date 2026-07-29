# Live Run

Status: Live lifecycle scaffold and telemetry persistence implemented; live execution remains unavailable.
Covers: `cmd/nuubot-live/main.go`, `internal/live/execute.go`, `internal/live/live.go`
Purpose: Own one standalone live Bot process.

## Current Boundary

Live Run mirrors Backtest Run's command and lifecycle structure.

The current scaffold must not be run as a live Bot.

`setup.Setup` still requires replay input and replay paths.

`controller.Init` still creates Signaler from replay OHLCV files.

The Hyperliquid WebSocket endpoint is a lifecycle stub. `Live Run.Start` fails explicitly until that endpoint is implemented.

## Ownership

The command parses arguments and calls `live.Execute`.

`live.Execute` owns one complete Live Run boundary. Live Run owns its lifecycle children.

Live Run owns:

- one WallClock;
- one shared MarketData object;
- one shared public Hyperliquid Info endpoint;
- one shared Hyperliquid WebSocket endpoint;
- one Controller;
- one telemetry Store; and
- runtime supervision.

Live Run creates and attaches Clock, MarketData, Info, and WebSocket to the shared `setup.Nuubot` harness before Controller initialization.

See [MarketData](marketdata.md) for permanent BBO ingestion, buffering, and subscription ownership.

The future credentialed Exchange endpoint belongs to Venue, not Live Run.

## Shared Endpoint Model

The shared Info endpoint uses the configured network and performs public, credential-free REST requests.

Any component receiving Nuubot may request supported public information through `nuubot.Info`.

The shared WebSocket will connect and manage subscriptions for requesting components.

A component will request a subscription and supply its callback. WebSocket subscription and callback mechanics remain unimplemented.

Meta refresh intentionally uses a separate Info object hardcoded to mainnet.

## Program Flow

```text
Init
  general app global setup
  select Live runtime policy
  reject terminal Bot
  retain runtime inputs
  create clock
  initialize clock
  attach clock to Nuubot
  create and attach MarketData to Nuubot
  initialize Info endpoint
  initialize WebSocket endpoint
  initialize Controller
  register Controller timer
  initialize telemetry persistence
  log init completed

Start
  start WebSocket endpoint
  start Info endpoint
  start Controller
  start clock
  log start completed

Loop
  wait for runtime event
  check clock failure

Stop
  log stop started
  ignore repeated stop request
  mark Live Run stopped
  stop clock
  stop WebSocket endpoint
  stop Info endpoint
  stop Controller
  stop MarketData
  collect terminal telemetry
  stop telemetry persistence
  log stop results and stats
  return stop errors
  log stop completed
```

`live.Run.Start` currently proves Info reachability through a public Meta request.

`live.Run.Loop` supervises caller cancellation, Controller stop requests, and WallClock errors.

Each successful Controller callback checks selected Live telemetry cadence.

Due samples append immediately to the per-Bot execution database. Sequence resumes from existing persisted telemetry.

## Target Live Differences

```text
Backtest
  ReplayReader drives ticks
  TickClock advances from replay timestamps

Live
  WallClock drives Controller timers
  WebSocket supplies requested live data
  shared Info supplies requested public REST data
```

Signaler must obtain initial live bars through the shared Info endpoint or another approved public information call.

Live Setup and Signaler construction remain unimplemented.

## Permanent Data Preservation Rule

This rule is mandatory and must not be weakened, relocated, or omitted.

Live Run never clears, resets, truncates, replaces, or silently repairs persisted runtime data.

This includes Bot, Ledger, Trade, Order, Fill, reconciliation, and Venue evidence.

Terminal `error` and `stopped` Bots can never restart.

Rerunning terminal behavior requires cloning the Bot into a new Bot ID.

A code defect leaves the Bot in `error` with all evidence intact.

Repair is explicit backend operator work. Recovery and normal startup never perform destructive repair.

Preserved data supports reference, review, troubleshooting, strategy analysis, and code analysis.

## Invariants

- One Live Run owns one Controller.
- Live Run owns shared Clock, MarketData, Info, and WebSocket lifecycle.
- Constructors perform no network work.
- Future endpoint Start performs reachability checks before Controller and Clock run.
- Stop remains idempotent after successful Start.
- Exchange credentials and trading transport do not belong to Live Run.

## Startup

Every Live Run Bot uses one Init path and one Start path.

Start always invokes Account reconciliation. Ledger state selects either the no-active-Trade fast path or full Venue reconciliation.

Live Run Startup also owns crash recovery and persisted lifecycle continuation.

Backtest Run does not recover. An interrupted backtest reruns from the beginning.

See [Startup](startup.md) for the approved unified contract.

## Required Proof

Current scaffold proof excludes execution.

Required current proof:

- formatting passes;
- editor diagnostics pass; and
- Git whitespace checks pass.

Future live proof must cover WebSocket connection, subscriptions, initial bars, Controller operation, endpoint failure, and orderly Stop.
