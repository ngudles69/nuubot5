# Runner Package

Status: Live lifecycle scaffold implemented; live execution remains unavailable.
Covers: `internal/runner/runner.go`
Purpose: Own one standalone live Bot runtime lifecycle.

## Ownership

Runner owns one WallClock, one MarketData service, one shared Info endpoint, one shared WebSocket endpoint, one Controller, and runtime supervision.

Runner attaches Clock, MarketData, Info, and WebSocket to the shared Nuubot harness before Controller initialization.

Account will own the future credentialed Exchange endpoint.

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

## Data Preservation

Runner never clears persisted runtime data.

Terminal `error` and `stopped` Bots cannot restart. Rerun requires cloning into a new Bot ID.

Failed evidence remains intact unless explicit backend repair is authorized.

## Current Limits

- `setup.Setup` remains replay-oriented.
- Controller Signaler construction remains replay-oriented.
- WebSocket Start returns the explicit unimplemented error.
- WebSocket-to-MarketData publication, initial bars, and trading transport remain unavailable.

See [Runner](../runner.md) for the process-level design.
