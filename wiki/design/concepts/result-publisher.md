# ResultPublisher

Status: Implemented for complete per-Bot SQLite results.
Covers: `internal/resultpublisher`
Purpose: Publish one successful reconciled backtest hierarchy atomically.

BtBot calls ResultPublisher after replay verification and Controller shutdown.

Publication receives:

- exact Bot identity and admitted configuration;
- Controller capital, PnL, equity, and drawdown;
- Signal and Risk decisions;
- BotCycle and Executor results;
- Grid Level state;
- Account and reconciled Ledger evidence;
- ordered telemetry;
- one terminal RunReport; and
- replay counts, range, duration, and completion.

Account results contain no Simulator private result.

ResultPublisher never reads Simulator memory, counters, indexes, or persistence.

Simulator remains exchange truth during execution.

Recon copies validated official evidence into Ledger.

Ledger evidence is the terminal publishable trading record.

BtBot clears `.partial` before runtime initialization.

Maximum persistence writes complete runtime evidence into `.partial` during execution.

ResultPublisher appends terminal evidence, commits, and renames `.partial` to `.db`.

It writes every terminal Account result, including maximum persistence mode.

Failure removes `.partial`.

Failed replay or Controller execution never publishes completed evidence.

The publisher owns no trading descendant and makes no lifecycle decision.
