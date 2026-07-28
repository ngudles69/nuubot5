# ProcessStore

Status: Implemented as the central `internal/control` Store.
Covers: `internal/control/*.go`, `internal/backtest/*.go`, `internal/live/*.go`
Purpose: Persist command reservation, acknowledgement, process identity, lifecycle, and health evidence.

## Ownership

The central operational database owns:

```text
control_command
process_state
bot.status
sweep.status
```

Per-Bot execution databases contain no command or process-supervision rows.

## Commands

Commands target one Bot or Sweep ID.

```text
requested → claimed → acknowledged
```

Acknowledged outcomes are:

```text
processed
skipped
rejected
```

The newest pending command wins. Older pending commands for the same target generation are acknowledged as skipped.

A command queued while no process is active uses generation zero and may be claimed by the next generation.

## Process Identity

One exact process identity contains:

```text
target kind
target ID
generation
PID
process token
```

Atomic registration rejects duplicate active generations.

A stale generation cannot claim or acknowledge current commands.

Process lifecycle updates atomically update both `process_state` and the canonical Bot or Sweep status.

## Polling

Backtest and Live poll from the Controller callback using wall-time `process.poll_seconds`.

Backtest replay time never controls command polling.

No transaction remains open during Controller, Recon, Stop, or process waits.

## Current Actions

`stop` is processed by Backtest and Live.

`start`, `pause`, and `resume` remain manager or future lifecycle work. A Run acknowledges unsupported actions as rejected.

## Proof

- Concurrent registrations produce one winner.
- Superseded requests become skipped.
- Stale acknowledgement fails.
- Observer claims and rejects an unsupported queued command, then completes normally.
