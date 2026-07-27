# Runner

Status: Live lifecycle scaffold implemented; live execution remains unavailable.
Covers: `cmd/nuubot-runner/main.go`, `internal/runner/runner.go`
Purpose: Own one standalone live Bot process.

## Current Boundary

Runner mirrors BtBot's command and lifecycle structure.

The current scaffold must not be run as a live Bot.

`setup.Setup` still requires replay input and replay paths.

`controller.Init` still creates Signaler from replay OHLCV files.

The Hyperliquid WebSocket endpoint is a lifecycle stub. `Runner.Start` fails explicitly until that endpoint is implemented.

## Ownership

The command owns Runner.

Runner owns:

- one WallClock;
- one shared MarketData object;
- one shared public Hyperliquid Info endpoint;
- one shared Hyperliquid WebSocket endpoint;
- one Controller; and
- runtime supervision.

Runner creates and attaches Clock, MarketData, Info, and WebSocket to the shared `setup.Nuubot` harness before Controller initialization.

See [MarketData](marketdata.md) for permanent BBO ingestion, buffering, and subscription ownership.

The future credentialed Exchange endpoint belongs to Account, not Runner.

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
  mark Runner stopped
  stop clock
  stop WebSocket endpoint
  stop Info endpoint
  stop Controller
  stop MarketData
  log stop results and stats
  return stop errors
  log stop completed
```

`Runner.Start` currently proves Info reachability through a public Meta request.

`Runner.Loop` supervises caller cancellation, Controller stop requests, and WallClock errors.

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

Runner never clears, resets, truncates, replaces, or silently repairs persisted runtime data.

This includes Bot, Ledger, Trade, Order, Fill, reconciliation, and Venue evidence.

Terminal `error` and `stopped` Bots can never restart.

Rerunning terminal behavior requires cloning the Bot into a new Bot ID.

A code defect leaves the Bot in `error` with all evidence intact.

Repair is explicit backend operator work. Recovery and normal startup never perform destructive repair.

Preserved data supports reference, review, troubleshooting, strategy analysis, and code analysis.

## Invariants

- One Runner owns one Controller.
- Runner owns shared Clock, MarketData, Info, and WebSocket lifecycle.
- Constructors perform no network work.
- Future endpoint Start performs reachability checks before Controller and Clock run.
- Stop remains idempotent after successful Start.
- Exchange credentials and trading transport do not belong to Runner.

## Startup

Every Runner Bot uses one Init path and one Start path.

Start always invokes Account reconciliation. Ledger state selects either the no-active-Trade fast path or full Venue reconciliation.

Runner Startup also owns crash recovery and persisted lifecycle continuation.

BtBot does not recover. An interrupted backtest reruns from the beginning.

See [Startup](startup.md) for the approved unified contract.

## Required Proof

Current scaffold proof excludes execution.

Required current proof:

- formatting passes;
- editor diagnostics pass; and
- Git whitespace checks pass.

Future live proof must cover WebSocket connection, subscriptions, initial bars, Controller operation, endpoint failure, and orderly Stop.
