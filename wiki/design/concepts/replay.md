# Replay

Status: Implemented for one symbol-qualified historical stream.
Covers: `internal/backtest`, `internal/replay`, `internal/controller`
Purpose: Drive one exact historical market sequence through shared MarketData and TickClock.

## Flow

```text
read validated one-second row
create BBO
attach ReplayInput symbol
MarketData.IngestBBO
complete matching subscriber callbacks
TickClock.Advance
registered Controller timer runs
```

Reader exhaustion is normal completion.

BtBot then verifies exact tick count, control-pass count, first timestamp,
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
- Sweep 6 Bot 9: 63 cycles and 16 stop-loss exits after Start-time latest-BBO entry.
- Sweep 9 Bot 13: full TradeBot Account, Ledger, Simulator, and Result proof.
