# nuubot-bt-bot

Status: Implemented.
Covers: `cmd/nuubot-bt-bot/main.go`
Purpose: Parse identity and optional profiling, run BtBot lifecycle, and log the terminal result with elapsed time.

## Responsibilities

`main.go` has exactly four responsibilities:

1. Parse the input.
2. Own optional command profiling.
3. Run BtBot.
4. Log success or failure and elapsed time.

Every line in `main.go` MUST contribute directly to one responsibility.

Parsing stays in Section 2 of `main.go`.

BtBot owns lifecycle behavior. The command calls each lifecycle phase.

The command owns opt-in CPU, trace, heap, allocations, block, and mutex profiling.
Profiling surrounds initialization through terminal result emission without entering the BtBot package.

Operational log paths belong in `internal/toolkit/logging`.
Profiling files use the explicit internal prefix supplied by the invoking script.

## Program Flow

```text
main
  open server log
  parse input
  open bot log
  start performance profile
  create btbot
  initialize btbot
  start btbot
  loop btbot
  stop btbot
  get result
  write run report
  stop performance profile
  log result

parseInput
  parse sweep id
  parse bot id
  parse profile prefix flag and value
```

## Profiling

Normal invocation keeps exactly two positional identities.

Performance invocation appends `-pp` and one output prefix.
The invoking `stest.sh -pp` path owns its session directory and the `run-001` prefix.

Shutdown stops streaming trace and CPU collectors before snapshots.
It forces GC before heap, allocations, block, and mutex profiles are written.

Profile setup and finalization failures terminate the command nonzero.

## Logging

Failures before valid identity use `server.log`.

After valid identity and Bot-log opening, all output uses only
`bot_<sweep_id>_<bot_id>.log`.

Console output is allowed only when `server.log` cannot open.

The terminal message MUST name the failed boundary.

Every log call receives one complete message string.

The successful terminal message includes elapsed duration.

## Does Not

- Load configuration.
- Know BtBot-owned Clock, Reader, Controller, or replay proof.
- Open operational log files directly.
- Put profiling policy or mechanics in the BtBot package.
- Wrap `main` with `program`, command, or local Run functions.

## Required Proof

- Invalid input exits nonzero and writes to `server.log`.
- Identified failures exit nonzero and write only to the Bot log.
- Successful execution exits zero and logs one completion message with elapsed duration.
- Operational output does not use stdout or stderr after logger creation.
- Profile mode writes all six nonempty artifacts or exits nonzero.
- Normal mode creates no profiling artifacts.
