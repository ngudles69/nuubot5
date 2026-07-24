# Runner

Status: Approved — unimplemented.
Covers: `cmd/nuubot-runner/main.go`
Purpose: Run one stored live, testnet, paper, or simulator Bot.

## Canonical Sources

- Nuubot4: `D:/rust/nuubot4/wiki/logic/runner.md`
- Nuubot3: `D:/rust/nuubot3/nuubot/runner/runner.py`

## Scope

Runner owns one Bot's live inputs, local feed state, clock, Runtime, supervision, and terminal evidence.

## Owner and Children

BotManager owns Runner.

Runner directly owns:

- one WallClock;
- one Runtime;
- DataEngine subscriptions;
- local BBO, bar, and user-event state; and
- Runner-owned background work.

## Responsibilities

- Load and validate one stored Bot.
- Build Runtime from admitted configuration.
- Request required DataEngine subscriptions.
- Bootstrap required bars before opening Runtime admission.
- Deliver validated bars and BBO values to Runtime.
- Mark Account and Ledger dirty from user events.
- Trigger fast BBO checks and slower reconciliation requests.
- Supervise its clock, subscriptions, Runtime, and completion.
- Stop new input before Runtime teardown.
- Publish lifecycle and terminal status through RuntimeStore.

## Does Not

- Share mutable Runtime state with DataEngine.
- Decode venue WebSocket messages.
- Implement signal, risk, execution, or reconciliation policy.
- Own another Runner.
- Manage Sweeps.
- Expose Runtime descendants to Server or BotManager.

## Lifecycle

`Create` constructs one stopped Runner.

`Init` loads its Bot and prepares direct children.

`Start` establishes initial truth, starts Runtime, subscribes inputs, then starts WallClock.

`Loop` supervises until stop, completion, or child failure.

`Stop` closes time and event admission, releases subscriptions, stops Runtime, and records terminal evidence.

## Program Flow

```text
Init
  load stored Bot
  create WallClock
  create Runtime
  obtain Runtime data requirements
  prepare local feed state

Start
  bootstrap Bars
  start Runtime
  subscribe DataEngine
  register Clock timers
  start WallClock
  mark running

Loop
  wait for feed events
  supervise WallClock, subscriptions, Runtime, and stop request

Stop
  stop Clock admission
  cancel Runner work
  unsubscribe DataEngine
  stop Runtime
  persist terminal status
```

## Invariants

- One Runner owns one Runtime.
- Runtime admission opens only after initial truth exists.
- Feed work cannot execute trading policy.
- Every background task has one owner, stop condition, and error path.

## Required Proof

- Initial truth completes before Runtime admission.
- BBO and user events reach the correct Bot only.
- User events mark state dirty without reconciling immediately.
- Child failure reaches the Runner boundary.
- Stop remains idempotent after successful Start.
