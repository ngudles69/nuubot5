# nuubot-btrunner

Status: Implemented.
Covers: `cmd/nuubot-btrunner/main.go`
Purpose: Parse identity, run BtRunner lifecycle, and log the terminal result with elapsed time.

## Responsibilities

`main.go` has exactly three responsibilities:

1. Parse the input.
2. Run BtRunner.
3. Log success or failure and elapsed time.

Every line in `main.go` MUST contribute directly to one responsibility.

Parsing stays in Section 2 of `main.go`.

BtRunner owns lifecycle behavior. The command calls each lifecycle phase.

Log paths and file opening belong in `internal/toolkit/logging`.

## Program Flow

```text
main
  open server log
  parse input
  open bot log
  create btrunner
  initialize btrunner
  start btrunner
  loop btrunner
  stop btrunner
  log result

parseInput
  parse sweep id
  parse bot id
```

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
- Know BtRunner-owned Clock, Reader, Runtime, or replay proof.
- Open log files directly.
- Wrap `main` with `program`, command, or local Run functions.

## Required Proof

- Invalid input exits nonzero and writes to `server.log`.
- Identified failures exit nonzero and write only to the Bot log.
- Successful execution exits zero and logs one completion message with elapsed duration.
- Operational output does not use stdout or stderr after logger creation.
