# Runner

Status: Approved — unimplemented.
Covers: `cmd/nuubot-runner/main.go`
Purpose: Run one stored live, testnet, paper, or simulator Bot.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/wiki/logic/runner.md`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`

## Scope

Runner is one standalone process.

It owns one Bot's live inputs, local feed state, clock, Controller, supervision,
and result evidence.

## Owner and Children

The command owns Runner.

BotManager may launch and supervise the same standalone program.

Runner directly owns:

- one WallClock;
- one Controller;
- required process-local live inputs;
- local BBO, bar, and user-event state; and
- Runner-owned background work.

## Responsibilities

- Load and validate one stored Bot.
- Build Controller from admitted configuration.
- Obtain required live subscriptions without requiring Server.
- Bootstrap required bars before opening Controller admission.
- Deliver validated bars and BBO values to Controller.
- Mark the matching Account recon-dirty from user events.
- Trigger fast BBO checks and own one drift-free heartbeat for scheduled work.
- Supervise its clock, subscriptions, Controller, and completion.
- Stop new input before Controller teardown.
- Publish its own lifecycle and result evidence.

## Does Not

- Share mutable Controller state with feed transport.
- Decode venue WebSocket messages.
- Implement signal, risk, execution, or reconciliation policy.
- Own another Runner.
- Manage Sweeps.
- Expose Controller descendants to Server or BotManager.

## Lifecycle

`Create` constructs one stopped Runner.

`Init` loads its Bot and prepares direct children.

`Start` establishes initial truth, starts Controller, subscribes inputs, then starts WallClock.

`Loop` supervises until stop, completion, or child failure.

`Stop` closes time and event admission, releases subscriptions, stops
Controller, and records result evidence.

## Program Flow

```text
Init
  load stored Bot
  create WallClock
  create Controller
  obtain Controller data requirements
  prepare local feed state

Start
  bootstrap Bars
  start Controller
  subscribe live inputs
  register one heartbeat timer
  start WallClock
  mark running

Loop
  wait for feed events
  supervise WallClock, subscriptions, Controller, and stop request

Stop
  stop Clock admission
  cancel Runner work
  release live subscriptions
  stop Controller
  persist result status
```

## Heartbeat

Runner owns one scheduler timer. Its configurable heartbeat defaults to ten seconds.

Each heartbeat reads WallClock once and uses that value for every due decision.

The heartbeat may run normal reconciliation, unresolved-Order cleanup, balance and
equity calculation, telemetry append, or stopping action.

Every interval is configurable. Scheduled boundaries advance independently of work duration to avoid drift.

Every heartbeat appends one cheap JSON telemetry row, whether scheduled work succeeds or fails.

A failed reconciliation writes operational telemetry but no domain state or cursor.

The current live display reads the latest indexed telemetry row.

Hot-path telemetry uses primitive counters and freshness timestamps. It never traverses the domain graph.

Decision-critical Account state remains synchronous.

Balance and equity cadence may become configurable only after those calculations are proven observability-only.

## Reconciliation Failure

Future live execution retains the last published generation after the first and second consecutive whole-reconciliation failures.

A successful reconciliation resets the count. The third consecutive failure begins stoppage.

Runner stages exact deltas, validates Ledger and Account candidates, commits dirty
rows, then performs non-failing memory publication.

It does not deep-clone the complete graph for rollback.

## Initialization Capacity

Runner initialization reserves capacity for 1,000 Trades, 2,000 Orders, and
2,000 Fills per Runner, plus reusable evidence buffers.

Capacity reserves containers, not objects. It is not automatically a hard limit.

BtBot uses the same initialization reserve.

## Invariants

- One Runner owns one Controller.
- Controller admission opens only after initial truth exists.
- Feed work cannot execute trading policy.
- Every background task has one owner, stop condition, and error path.
- One Runner heartbeat is the only scheduler timer.
- Every heartbeat appends one telemetry row.
- Failed reconciliation publishes no domain state or cursor.

## Required Proof

- Initial truth completes before Controller admission.
- BBO and user events reach the correct Bot only.
- User events mark state dirty without reconciling immediately.
- Child failure reaches the Runner boundary.
- Stop remains idempotent after successful Start.
- Direct execution works while Server is stopped.

## Open Decisions

- Live cross-process Account-symbol claims.
- Standalone status writes.
- Shared versus process-local exchange WebSockets.
- Server reconnection to independently started Runner processes.
