# Replay

Status: Implemented for one symbol-qualified historical stream.
Covers: `internal/btrunner`, `internal/replay`, `internal/controller`
Purpose: Drive one exact historical market sequence through Controller.

## Flow

```text
read validated one-second row
create BBO
attach ReplayInput symbol
Controller.IngestBBO
TickClock.Advance
registered Controller timer runs
```

Reader exhaustion is normal completion.

BtRunner then verifies exact tick count, control-pass count, first timestamp,
and last timestamp.

Controller stop never bypasses replay proof.

## Ordering

The current path has one stream, so file order is canonical.

Each BBO is admitted before timers at the same timestamp.

Signaler calculations use closed OHLCV bars only.

Deterministic multi-source merging and equal-timestamp ordering remain
deferred.

## Current Proof

- 7,948,800 one-second BBO values.
- 794,880 Controller passes.
- 2,207 Macross packages.
- Sweep 6 Bot 9: 64 cycles and 17 stop-loss exits.
- Sweep 9 Bot 13: full TradeBot Account, Ledger, Simulator, and Result proof.
