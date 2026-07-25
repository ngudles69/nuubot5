# ResultPublisher Package

Status: Implemented.
Covers: `internal/resultpublisher/*.go`
Purpose: Atomically publish one complete successful backtest result.

## Ownership

BtRunner calls ResultPublisher only after replay verification and Controller
shutdown succeed.

ResultPublisher owns the temporary database and final rename.

Ledger and Simulator own detailed Account child serialization.

ResultPublisher owns Controller, BotCycle, Executor, Signal, Risk, and replay
summary serialization.

## Output

The per-Bot database contains:

- `backtest_result`;
- `botcycle_result`;
- `executor_result`;
- `signal_decision`;
- `risk_decision`;
- Account Ledger, Trade, Order, and Fill tables; and
- Simulator state.

`backtest_result` preserves the exact admitted BotConfig TOML and hash.

## Atomic Flow

```text
remove stale .partial
publish every terminal Account child
publish Controller and replay hierarchy
commit every summary transaction
rename .partial to final .db
```

Maximum-mode children are re-materialized from terminal immutable results.

The final rename never replaces durable children with a summary-only database.

Failure removes `.partial` and never publishes completed evidence.

Foreign-key and integrity checks belong to the TradeBot harness.
