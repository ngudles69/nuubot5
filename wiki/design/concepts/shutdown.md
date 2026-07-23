# Shutdown Process

Status: Implemented.
Covers: `internal/btrunner/btrunner.go`, `internal/runtime/runtime.go`, `internal/botcycle/botcycle.go`
Purpose: Stop admission, close active work, release owned resources, and preserve terminal evidence.

## Canonical Sources

- Nuubot4 command: `D:/rust/nuubot4/src/bin/nuubot-btrunner.rs`
- Nuubot4: `D:/rust/nuubot4/src/btrunner.rs`
- Nuubot4: `D:/rust/nuubot4/src/runtime.rs`
- Nuubot4: `D:/rust/nuubot4/src/botcycle.rs`

## Participants

- Command guarantees BtRunner stop after successful start.
- BtRunner stops direct children.
- Runtime closes active BotCycle and stops Risks and Signaler.
- BotCycle stops Executors.
- Each owner reports its own terminal statistics.

## Preconditions

Shutdown may follow completion, end date, maximum cycles, Risk, user request, parent stop, start failure, or work failure.

Repeated valid stop calls MUST be safe.

## Ordered Flow

```text
Command
  retain Run result
  call BtRunner Stop
  combine without hiding Run failure

BtRunner Stop
  stop Runtime
  stop Reader
  stop TickClock

Runtime Stop
  latch first stop reason
  close active BotCycle
  stop Risks in reverse order
  stop Signaler

BotCycle Stop
  stop Executors in reverse order
  consolidate exit reason
```

## Decisions

Runtime owns the first graceful stop reason.

An Executor may supply the completed cycle reason.

The original work error remains primary when shutdown also fails.

Each parent controls only direct children.

## State Changes

Admission stops before child teardown.

An active BotCycle becomes closed exactly once.

Children become stopped in reverse ownership order.

Terminal statistics become final.

## Failure Handling

Every owner attempts all direct-child stops.

BtRunner returns Runtime error before Reader error.

BotCycle returns its first Executor stop error.

No shutdown error may erase the primary `Run` failure.

## Recovery

Idempotent stop permits repeated cleanup calls.

Unreleased process resources require process termination and defect investigation.

## Completion

Shutdown completes when every started direct child received stop, active cycles closed, resources released, and terminal evidence was reported.

## Does Not

- Retry failed trading actions.
- Reconcile future Accounts.
- Invent final prices.
- Treat process exit as sufficient proof.
- Allow parents to stop grandchildren directly.

## Required Proof

- End date closes an active BotCycle.
- ObserverExecutor preserves last timestamp and final price.
- Stop methods are idempotent.
- Start failures clean up started children.
- Run failure still triggers teardown.
- Started and closed cycle counts match.

## Open Decisions

Live shutdown must define feed cancellation and Account close ordering before implementation.
