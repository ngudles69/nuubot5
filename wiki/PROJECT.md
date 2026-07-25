# Nuubot5 Project

## Purpose

Nuubot5 tests whether simple, idiomatic Go can support a fast and stable trading system.

Current proof covers historical replay, signals, BotCycles, ObserverExecutor,
Simulator-backed TradeExecutor, Account reconciliation, Ledger evidence,
result publication, and a Risk stub.

Live trading and Server behavior remain unimplemented.

## Language

Nuubot5 uses the standard Go toolchain and approved pure-Go dependencies.

Canonical builds MUST use `-tags noasm`.

Non-standard Go requires explicit prior user approval.

[STYLE.md](coding/STYLE.md) owns Nuubot-specific style.

[RULES.md](coding/RULES.md) owns dependencies, errors, logging, concurrency, safety, and proof.

Idiomatic Go governs everything those contracts do not specify.

## Reference Order

Nuubot5 source and runtime evidence own implemented truth.

Nuubot4 owns canonical process, ordering, lifecycle, and decisions.

Nuubot3 fills behavior missing from Nuubot4.

Nuutrader6 fills remaining proven gaps, especially shared data, reconciliation, persistence, and Hyperliquid behavior.

Reference code does not prove Nuubot5 implementation.

Conflicts MUST be reported. Recommendations MUST remain separate from canonical behavior.

Do not modify a reference repository without explicit user authority.

## Implemented Scope

- Go BtRunner command and configuration.
- Read-only SQLite Bot loading.
- Parquet tick replay and OHLCV loading.
- Shared TickClock and WallClock timer mechanics.
- TickClock-driven Controller passes.
- Exact compiled BotSpec selection and stored TOML BotConfig admission.
- Immutable BotDefinition construction.
- Macross and RSI signalers.
- ObserverExecutor stop-loss behavior.
- Simulator-backed TradeExecutor.
- Account, Ledger, Trade, Order, and Fill reconciliation.
- Simulator venue-shaped execution.
- Per-Bot result publication.
- BalancedRisk stub.
- Reader-exhaustion shutdown through BtRunner.
- Exact replay and semantic completion checks.

## Approved Unimplemented Scope

- `nuubot-server`, `nuubot-cli`, and `nuubot-runner` command shells reserve
  canonical executable names and print `Under Construction.`.
- Live Runner and live event handling.
- Server, API, web server, BotManager, and SweepManager.
- Standalone SweepRunner.
- Live Venue execution, recovery, and CLOID handling.
- ProcessStore and RunnerControl.
- PocketBase-backed HTTP, API, authentication, administration, realtime, and
  SQLite persistence.
- Multi-source replay and live BotSpec admission.

Only the explicitly approved implementation sequence authorizes target work.

The three command shells do not prove their named systems are implemented.

DataEngine and ControllerStore remain candidates. Their ownership and final scope
are unresolved.

## Success Contract

BtRunner succeeds only when:

- the process exits zero;
- every input timestamp and value passes validation;
- served ticks, passes, and replay range match expectations;
- Controller statistics remain internally consistent;
- any active BotCycle closes during graceful shutdown after replay completion; and
- every direct child stops successfully.

Go passes the current speed gate when replay remains below twice the accepted Rust reference.

Correctness and fresh-process stability take priority over speed.

## Accepted Proof

Sweep 6 Bot 9 replays 7,948,800 one-second ticks through 794,880 Controller
passes.

Each accepted run reports 2,207 Signal packages and 724 skipped StartCycle
actions.

Observer produces 64 sequential cycles and 17 stop-loss exits.

Sweep 9 Bot 13 runs the same three-month input through Macross, TradeExecutor,
Account, Ledger, and Simulator.

TradeBot produces:

- 193 completed BotCycles;
- 193 Trades;
- 626 Orders;
- 386 Fills;
- 1,000 USDC capital;
- -3.90459332761 USDC net PnL;
- 996.09540667239 USDC ending equity; and
- 4.200462813402 USDC maximum drawdown.

The result stores exact BotConfig TOML and hash.

Cycle 2 starts with Cycle 1 terminal equity.

Database integrity and foreign-key checks pass.

Fresh-process stability passed:

- TradeBot 2 of 2 and 10 of 10;
- Observer 2 of 2 and 10 of 10.

Proof logs:

```text
workspace/logs/nuubot5-trtest-s9-b13-2-20260724T202731Z.log
workspace/logs/nuubot5-trtest-s9-b13-10-20260724T202751Z.log
workspace/logs/nuubot5-rtest-s6-b9-2-20260724T202905Z.log
workspace/logs/nuubot5-rtest-s6-b9-10-20260724T202915Z.log
workspace/logs/nuubot5-trtest-s9-b13-1-20260724T204016Z.log
workspace/logs/nuubot5-rtest-s6-b9-1-20260724T204033Z.log
```

Historical commit benchmarks live in [PERFORMANCE.md](PERFORMANCE.md).

The earlier canonical `noasm` decoder gate passed 1,000 of 1,000 fresh processes.

The optimized decoder returned one corrupt timestamp at run 183.

Validation rejected it. The source Parquet row was valid.

This evidence selects `-tags noasm`. It does not identify the dependency fault.

## Data and Deployment

SQLite is approved for backtesting.

PocketBase-owned SQLite is approved for future Server persistence.

One embedded PocketBase application in `nuubot-server` owns the writable
database, web server, API, authentication, administration, and realtime.

Nuubot owns the trading interface, operational dashboards, analytics, and
reports.

Runner, BtRunner, and SweepRunner must remain independently executable while
Server is stopped.

Standalone saved-Config reads, status writes, physical schemas, and migrations
remain unresolved.

Writable output MUST remain inside this repository or an explicitly approved datastore.

Windows BtRunner execution is proven.

Ubuntu 24 is the intended VPS target. Linux runtime behavior remains unproven.

## Documentation Ownership

- `AGENTS.md` owns startup, authority, prose, and project-wide decisions.
- `PROJECT.md` owns purpose, scope, status, proof, and reference order.
- [ARCHITECTURE.md](ARCHITECTURE.md) owns layers, ownership, flows, concurrency, and persistence boundaries.
- [DESIGN.md](DESIGN.md) owns the high-level object catalog.
- [`design/**`](design/) owns detailed object and process contracts.
- [`logic/**`](logic/) remains legacy detail until separately migrated.
- `HANDOFF.md` owns restart state, active work, proof, and next action.

When source and wiki conflict, stop and report the conflict.
