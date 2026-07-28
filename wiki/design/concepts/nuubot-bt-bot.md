# nuubot-bt-bot

Status: Implemented.
Covers: `cmd/nuubot-bt-bot/main.go`, `internal/backtest/execute.go`, `internal/runharness/profile.go`
Purpose: Parse one Backtest request and call one complete Backtest execution boundary.

## Responsibilities

`main.go` has three responsibilities:

1. Parse arguments into `btbot.Options`.
2. Call `backtest.Execute` once.
3. Report one terminal error and exit nonzero.

`backtest.Execute` owns Bot logging, optional whole-Run profiling, BtBot lifecycle, terminal report output, and successful elapsed-time logging.

BtBot owns historical runtime behavior. The command does not call lifecycle phases directly.

## Program Flow

```text
main
  parse arguments
  execute Backtest
  report one terminal error

backtest.Execute
  validate Run options
  open Bot log
  start whole-Run profiling
  initialize Backtest
  start Backtest
  loop Backtest
  stop Backtest
  write terminal report
  stop whole-Run profiling
  log Run completed
```

## Profiling

Normal invocation keeps two positional identities.

Performance invocation appends `-pp` and one output prefix.

`internal/runharness.Profile` owns CPU, trace, heap, allocations, block, and mutex profiling mechanics.

Profiling surrounds one complete Backtest without entering BtBot lifecycle code.

Profile setup and finalization failures terminate execution nonzero.

## Crash Model

The command provides no resume or recovery mode.

A crashed process produces a failed backtest attempt.

A later attempt starts BtBot from the beginning and replays the complete requested range.

Runner startup and crash recovery remain separate.

## Does Not

- Load App Config directly.
- Know BtBot-owned Clock, Reader, Controller, or replay proof.
- Own profiling mechanics.
- Reproduce BtBot lifecycle in `main`.

## Required Proof

- Argument parsing tests pass.
- Shared Profile lifecycle test writes all six nonempty artifacts.
- Normal Backtest execution exits zero and emits one report.
- Failure returns through one terminal command path.
