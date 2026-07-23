# Nuubot5 Project

## Purpose

Nuubot5 tests whether simple, idiomatic Go can support a fast and stable trading system.

Current proof covers historical replay, signals, BotCycles, observer execution, and a risk stub.

Live trading, accounting, persistence, simulation, and server behavior remain approved but unimplemented.

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
- TickClock-driven Runtime passes.
- Macross and RSI signalers.
- ObserverExecutor stop-loss behavior.
- BalancedRisk stub.
- End-date shutdown through Runtime.
- Exact replay and semantic completion checks.

## Approved Unimplemented Scope

- `nuubot-server`, `nuubot-cli`, and `nuubot-runner` command shells reserve
  canonical executable names and print `Under Construction.`.
- Live Runner, WallClock, DataEngine, and live events.
- Server, API, web server, BotManager, and SweepManager.
- Account, Ledger, Trade, Order, Fill, Venue, and Simulator.
- Execution, reconciliation, recovery, and CLOID handling.
- RuntimeStore, ProcessStore, RunnerControl, and ResultPublisher.
- PostgreSQL live, simulator, and paper persistence.

Approved design does not authorize implementation, dependencies, transport, or schema choices.

The three command shells do not prove their named systems are implemented.

## Success Contract

BtRunner succeeds only when:

- the process exits zero;
- every input timestamp and value passes validation;
- served ticks, passes, and replay range match expectations;
- Runtime statistics remain internally consistent;
- any active BotCycle closes at the effective end date; and
- every direct child stops successfully.

Go passes the current speed gate when replay remains below twice the accepted Rust reference.

Correctness and fresh-process stability take priority over speed.

## Accepted Proof

Sweep 6 Bot 9 replays 7,948,800 one-second ticks through 794,880 Runtime passes.

Each accepted run reports 55 signals, 18 cycles, 17 stop-loss exits, and one end-date exit.

The canonical `noasm` build passed 1,000 of 1,000 fresh-process runs without delay.

Process time averaged 445 ms. Replay time averaged 371 ms.

The run reported zero failure markers, zero incorrect statistics, and zero stderr.

Proof log:

```text
workspace/logs/nuubot5-rtest-s6-b9-1000-20260723T041701Z.log
```

Historical commit benchmarks live in [PERFORMANCE.md](PERFORMANCE.md).

The optimized decoder returned one corrupt timestamp at run 183.

Validation rejected it. The source Parquet row was valid.

This evidence selects `-tags noasm`. It does not identify the dependency fault.

## Data and Deployment

SQLite is approved for backtesting.

PostgreSQL is approved for future live, simulator, and paper operation.

Physical schemas, migrations, and result publication remain unresolved.

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
