# Control Package

Status: Implemented for central command and process coordination.
Covers: `internal/control/*.go`
Purpose: Persist durable Bot and Sweep commands, acknowledgements, and exact process generations.

## Ownership

The central operational database owns `control_command` and `process_state`.

Server managers enqueue commands. A matching process generation claims and acknowledges commands.

Control owns no Controller, Account, execution evidence, or operating-system process.

## Command Lifecycle

```text
requested
  → claimed
  → acknowledged
      outcome: processed | skipped | rejected
```

One poll claims the newest requested command for one target.

Older pending commands for that target and generation become acknowledged with `skipped` outcome.

Backtest and Live currently process `stop`. Unsupported actions are acknowledged as `rejected`.

Commands queued without an active process use generation zero. The next process generation may claim them.

## Process Lifecycle

```text
starting → running → stopping → stopped
                            └──→ error
```

Process identity contains target kind, target ID, generation, PID, and process token.

Registration is atomic. One active target generation wins. A stale generation cannot claim or acknowledge current commands.

Process updates atomically update `process_state` and the matching `bot.status` or `sweep.status`.

## Timing

Backtest and Live call command polling from their Controller callback.

Polling cadence uses wall time from `process.poll_seconds`, not replay time.

Transactions end before Controller, Recon, Stop, or process waits continue.

## Proof

- Concurrent process registration has one winner.
- Newest command wins and older pending commands become skipped.
- Stale generations cannot acknowledge current commands.
- A queued unsupported command is claimed and rejected during a successful Observer Run.
