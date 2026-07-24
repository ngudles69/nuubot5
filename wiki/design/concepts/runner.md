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
- Trigger fast BBO checks and slower reconciliation requests.
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
  register Clock timers
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

## Invariants

- One Runner owns one Controller.
- Controller admission opens only after initial truth exists.
- Feed work cannot execute trading policy.
- Every background task has one owner, stop condition, and error path.

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
