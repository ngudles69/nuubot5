# ObserverExecutor

Status: Implemented.
Covers: `internal/executor/observer.go`
Purpose: Provide the complete starting template without placing trades.

## Ownership

BotCycle owns ObserverExecutor through the Executor interface.

Observer owns no Account or trading state.

Executor factory constructs it and calls `OnInit`.

## Lifecycle

Observer uses the canonical Executor status.

```text
configured
  starting
    running
      stopping
        stopped
```

Invalid initialization enters error.

Stopped and error states never transition.

No separate terminal flag exists.

`OnStop` is idempotent.

## Admission

Observer requires exactly one standard entry trigger.

Missing or conflicting triggers reject BotCycle admission.

Valid long or short entry starts Observer immediately.

## Capabilities

Observer implements:

- `BBOHandler.OnBBO`.
- `BBOIngestHandler.IngestBBO`.

It implements no unused event method.

## Program Flow

```text
OnInit
  validate config
  admit signal
  initialize observer

IngestBBO
  count ingested bbo

OnBBO
  count received bbo
  record last bbo
  record entry
  assess stop loss

OnStop
  preserve stop reason
  preserve end time
  stop observer
  calculate duration
  report proof
```

## Stop Loss

The first delivered `OnBBO` price becomes the observed entry.

Long stops at or below entry multiplied by one minus stop percentage.

Short stops at or above entry multiplied by one plus stop percentage.

Stop loss moves Observer to stopping.

Controller closes the owning BotCycle during its next timed pass.

## Logging

Observer never logs each BBO.

Its final summary reports:

- Triggering Signal facts.
- Entry and final prices.
- Stop-loss price.
- Duration and reason.
- `ingest_bbo_count`.
- `on_bbo_count`.

## Does Not

- Place or cancel Orders.
- Create Account, Ledger, Trade, Fill, Simulator, or Venue state.
- Match Orders or create simulated Fills.
- Model fees, liquidity, or slippage.
- Directly stop Controller.
