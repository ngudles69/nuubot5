# Nuubot5 Project

## Purpose

Nuubot5 tests whether simple, idiomatic Go can support a fast and stable trading system.

The current objective is a complete BtRunner path over the same market data used by Nuubot4.

Nuubot5 is incomplete. Current proof covers historical replay, signals, BotCycles, observer execution, and a risk stub.

It does not prove live trading, accounting, persistence, simulator behavior, server behavior, or full Nuubot parity.

## Language

Nuubot5 uses the standard Go toolchain and approved pure-Go dependencies.

Canonical builds MUST use `-tags noasm`.

CGO, native bindings, `unsafe`, handwritten assembly, and other non-standard Go require explicit prior user approval.

[Coding Style](coding/STYLE.md) owns Nuubot-specific style. Idiomatic Go governs everything not specified there.

[Coding Rules](coding/RULES.md) owns dependencies, errors, logging, concurrency, safety, and proof.

## References

Nuubot5 owns its implementation and proof.

Nuubot4 is the current BtRunner behavior and performance reference.

Nuubot3 is the future server, live Runner, Account, Ledger, Exchange, and Simulator behavior reference.

References define behavior to assess. They do not establish Nuubot5 parity or authorize copied implementation.

Read current reference code and its owning wiki before porting a component.

Do not modify a reference repository from a Nuubot5 session without explicit user authority.

## Current Scope

Implemented:

- Go command boundary and configuration loading.
- Read-only SQLite Sweep Bot loading.
- Parquet tick replay and OHLCV bar loading.
- TickClock-driven Runtime passes.
- Macross and RSI signal calculations.
- ObserverExecutor stop-loss behavior.
- BalancedRisk stub.
- End-date shutdown through the normal Runtime stop path.
- Exact replay and semantic completion checks.

Not implemented:

- live Runner and WebSocket feeds;
- Account, Ledger, Trade, Order, and Fill;
- Exchange adapters and Simulator;
- server, API, web server, CLI, and web frontend;
- BotManager and SweepManager;
- PostgreSQL live persistence; and
- result persistence and publication.

## Success Contract

BtRunner succeeds only when:

- the process exits zero;
- every input timestamp and value passes validation;
- served ticks, passes, and replay range match expectations;
- Runtime statistics remain internally consistent;
- an active BotCycle closes at the effective end date; and
- every direct child stops successfully.

Go passes the current speed gate when replay time remains below twice the accepted Rust reference.

Speed never overrides correctness or repeated fresh-process stability.

## Accepted Proof

Sweep 6 Bot 9 replays 7,948,800 one-second ticks and triggers 794,880 Runtime passes.

Every accepted run reports 55 signals, 18 cycles, 17 stop-loss exits, and one end-date exit.

The canonical `noasm` build passed 1,000 of 1,000 fresh-process runs without an inter-run delay.

Process time averaged 445 ms. Replay time averaged 371 ms.

The run reported zero failure markers, zero incorrect statistics, and zero stderr.

The proof log is:

```text
workspace/logs/nuubot5-rtest-s6-b9-1000-20260723T041701Z.log
```

An optimized decoder returned one corrupt timestamp at run 183.

Timestamp validation rejected it. The source Parquet row was valid.

This evidence selects `-tags noasm`. It does not identify the exact optimized dependency fault.

## Deployment

Windows execution is proven by the accepted BtRunner runs.

Ubuntu 24 deployment is expected because Nuubot5 uses Go and approved pure-Go dependencies.

Linux runtime behavior is not yet proven. No recorded cross-build evidence is claimed here.

## Data and Persistence

Shared market data is read-only.

Nuubot5 writable files MUST remain inside this repository or an explicitly approved datastore.

SQLite is approved for backtesting.

PostgreSQL is the approved future datastore for live, simulator, and paper operation.

Result persistence has no implemented owner or schema.

## Documentation Ownership

- `AGENTS.md` owns startup, authority, prose, and project-wide key decisions.
- `PROJECT.md` owns purpose, scope, status, success, proof, and reference hierarchy.
- [ARCHITECTURE.md](ARCHITECTURE.md) owns layers, ownership, flows, concurrency, persistence, and deployment boundaries.
- [DESIGN.md](DESIGN.md) owns object responsibilities, exclusions, lifecycle, and implementation status.
- `wiki/logic/**` owns detailed component behavior.
- `HANDOFF.md` owns current restart state, recent proof, blockers, and next action.

Source and runtime evidence determine implemented truth.

When source and wiki conflict, stop and report the conflict. Do not silently choose either version.
