# Risk

## Covers

- `internal/risk/risk.go`
- `internal/risk/balanced.go`
- `internal/runtime/runtime.go`

## Intent

Risk MUST request a graceful Runtime stop when a current risk rule is breached.
Risk MUST NOT directly stop Runtime, BotCycle, or Executor.

## Ownership

Runtime MUST own `[]Risk`. Each Risk implementation MUST own only its rule state
and statistics.

The approved configuration-selected Risk factory MUST return Runtime's
consumer-owned Risk interface. Runtime MUST NOT depend on a concrete Risk type.

## Program Flow

```text
Runtime.MainLoop(now)
  for each Risk
    if Risk.Assess()
      latch stop reason "risk"

  if a stop reason is latched
    Runtime.Stop(reason)
    return stop requested

  continue active BotCycle
```

## Current Implementation

`balanced` is a deliberate stub:

- `Assess` MUST increment its assessment count.
- `Assess` MUST return false.
- `Stop` MUST be idempotent.
- `Stop` MUST report assessments and zero exits requested.

Account snapshots, drawdown, equity, and live risk rules are not implemented.
They MUST NOT be simulated with invented data.

## Future Boundary

When Account exists, Runtime MUST pass one immutable coherent snapshot to Risk.

Risk MUST NOT hold an Account reference, query a datastore, or reconcile state.
