# ResultPublisher

Status: Implemented for complete per-Bot SQLite results.
Covers: `internal/resultpublisher`
Purpose: Persist one successful immutable backtest hierarchy atomically.

BtRunner calls ResultPublisher only after replay verification and Controller
shutdown succeed.

Publication receives:

- exact Bot identity;
- exact admitted BotConfig TOML and hash;
- Controller capital, PnL, equity, and drawdown;
- Signal and changed Risk decisions;
- BotCycle and Executor results;
- Grid Level calculations and terminal state;
- Account, Ledger, Trade, Order, Fill, and Simulator evidence; and
- ordered telemetry samples;
- one calculated terminal RunReport; and
- replay counts, range, duration, and completion.

ResultPublisher creates `.partial`, writes all evidence, commits, and renames
to `.db`.

It writes every terminal Account result, including maximum persistence mode.

Failure removes `.partial`.

Failed replay or Controller execution never publishes completed evidence.

The publisher owns no trading descendant and makes no lifecycle decision.
