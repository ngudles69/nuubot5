# Live Package

Status: Live lifecycle scaffold and telemetry persistence implemented; live execution remains unavailable.
Covers: `internal/live/*.go`
Purpose: Own one standalone Live Bot Run lifecycle.

## Ownership

Live `Run` owns one WallClock, one MarketData service, one shared Info endpoint, one shared WebSocket endpoint, one Controller, one telemetry Store, and runtime supervision.

Live `Run` attaches Clock, MarketData, Info, and WebSocket to the shared Nuubot harness before Controller initialization.

Live `Run` selects `App.Live` once and registers Controller from selected `nuubot.Runtime.ControllerIntervalMS`.

Account reads selected Recon cadence without branching on execution mode.

Account will own the future credentialed Exchange endpoint.

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
  mark Run stopped
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

Each successful Controller callback checks selected telemetry cadence.

A due collection writes immediately when `nuubot.Runtime.TelemetryWriteOnCollect` is true.

Live Config requires write-on-collect. The Store resumes sequence from existing per-Bot telemetry rows.

## Data Preservation

Live `Run` never clears persisted runtime data.

Terminal `error` and `stopped` Bots cannot restart. Rerun requires cloning into a new Bot ID.

Failed evidence remains intact unless explicit backend repair is authorized.

## Current Limits

- `setup.Setup` remains replay-oriented.
- Controller Signaler construction remains replay-oriented.
- WebSocket Start returns the explicit unimplemented error.
- WebSocket-to-MarketData publication, initial bars, and trading transport remain unavailable.

See [Live Run](../live.md) for process-level design.
