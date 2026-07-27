# ResultPublisher Package

Status: Implemented.
Covers: `internal/resultpublisher/*.go`
Purpose: Atomically publish one complete successful backtest result.

## Ownership

BtBot calls ResultPublisher after replay verification and Controller shutdown.

ResultPublisher owns the temporary database and final rename.

It serializes immutable Controller, BotCycle, Executor, Account, Ledger, telemetry, RunReport, and replay results.

It never owns or queries Simulator.

## Output

The per-Bot database contains:

- `backtest_result`;
- `botcycle_result`;
- `executor_result`;
- `grid_level_result`;
- `signal_decision`;
- `risk_decision`;
- `telemetry_sample`;
- `run_report`; and
- reconciled Account Ledger, Trade, Order, and Fill tables.

`backtest_result` preserves admitted BotConfig TOML and hash.

`grid_level_result` preserves calculated economics and final Level state.

Simulator private Orders, Fills, indexes, counters, and persistence payload are excluded.

## Atomic Flow

```text
remove stale .partial
publish immutable reconciled results
publish Controller, replay, telemetry, and RunReport
commit every transaction
rename .partial to final .db
```

Maximum-mode Accounts publish the same reconciled Ledger result shape.

ResultPublisher does not copy Simulator durable state into terminal results.

Failure removes `.partial` and never publishes completed evidence.

Foreign-key and integrity checks belong to the backtest harness.
