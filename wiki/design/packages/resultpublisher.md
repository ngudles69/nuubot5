# ResultPublisher Package

Status: Implemented.
Covers: `internal/resultpublisher/*.go`
Purpose: Atomically publish one complete successful backtest result.

## Ownership

BtBot calls ResultPublisher after replay verification and Controller shutdown.

BtBot owns fresh attempt-database preparation.

Maximum persistence writes Ledger and Simulator runtime state into that attempt database.

ResultPublisher appends terminal evidence and owns the final rename.

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
BtBot removes stale .partial data before Controller initialization
Ledger and Simulator optionally persist complete runtime evidence
ResultPublisher appends immutable reconciled results
ResultPublisher appends Controller, replay, telemetry, and RunReport
commit every transaction
rename .partial to final .db
```

Maximum-mode Accounts use `.partial` throughout execution, preventing prior-run evidence from loading or merging.

ResultPublisher does not copy Simulator durable state into terminal results.

Failure removes `.partial` and never publishes completed evidence.

Foreign-key and integrity checks belong to the backtest harness.
