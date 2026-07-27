# Testing

## Purpose

This page owns Nuubot test boundaries.

Proof must match the real ownership and runtime boundary.

Not every component can be tested meaningfully as an isolated unit.

## Unit Tests

Use unit tests for deterministic code with complete local inputs.

Good unit-test targets include:

- Pure calculations and exact decimal behavior.
- Data validation and translation.
- Deterministic clock mechanics.
- Ledger mutations, ownership, reconciliation, and persistence transactions.
- Trade lifecycle and finance state transitions.
- Order lifecycle, Fill aggregation, and reconciliation state.
- Fill identity, enrichment, and immutable execution evidence.
- Meta freshness, persistence, refresh, and instrument rounding.
- OHLCV deterministic timestamp normalization and calculation mechanics.
- Risk policy decisions, counters, lifecycle, and invalid selection.
- Signaler closed-bar calculations, package history, flat JSON, and validation.
- CLOID encoding, decoding, round trips, and invalid identity rejection.
- Small state transitions with no missing owner.

A unit test must not recreate global infrastructure with fake composition.

A unit test must not add test-only interfaces or alternate production paths.

## No Unit Testing

The following components must not have isolated unit tests:

- BtBot.
- Controller.
- BotSpec.
- BotCycle.
- Executor runtime lifecycle.
- ObserverExecutor.
- TradeExecutor.
- GridExecutor runtime lifecycle.
- Replay Reader.

These are integrated components.

Replay is a thin concrete OHLCV-to-BBO streaming adapter. Isolated tests would
require fake Reader plumbing or duplicate Parquet fixtures and would prove the
harness instead of real replay behavior.

Isolated tests require excessive harnesses, fake data, fake infrastructure, and fake process state.

A passing isolated test would prove that artificial harness, not the real module path.

These components require the complete real construction and execution path.

Their proof runs through `Setup -> Nuubot -> Controller -> BotCycle -> Executor -> BtBot`.

Canonical system entrypoints are:

- Observer: `./stest.sh -bot 9`.
- TradeExecutor: `./stest.sh -bot 13`.
- GridExecutor: `./stest.sh -bot 14`.

The deleted historical `rtest.sh` command must not return.

Do not create fake Setup, Nuubot, Controller, BotSpec, BotCycle, Signaler, Risk,
Executor, replay, or Meta infrastructure for these components.

Ledger, Trade, Order, and Fill require strong direct tests because each package
owns exact domain state and deterministic mutation rules.

The only retained Executor unit tests prove pure deterministic Grid calculations.

Those tests call Grid calculation helpers directly. They do not initialize or
run GridExecutor, Account, Ledger, Simulator, BotCycle, Controller, or BtBot.

Risk is directly testable. The current BalancedRisk stub must prove `Allow`,
assessment counting, idempotent Stop, and rejection of unknown policy kinds.
Future Risk decisions require direct boundary and precedence tests.

Signaler calculations and immutable packages are directly testable. Tests must
use closed-bar series and package values, not complete Controller infrastructure.

## Simulator Parity Testing

Simulator testing is Venue parity testing.

It must prove official request and response shape, canonical Order and Fill
state, matching, cancellation, position and finance behavior, detached JSON,
persistence, failure atomicity, and exact comparison mechanics.

Simulator tests do not prove Nuubot Controller, BotCycle, Executor, Account, or
Ledger integration. System runs prove that complete path.

External Hyperliquid fixture and testnet parity remain separately required.

## Integration Tests

Use integration tests when behavior requires multiple real owners or initialized infrastructure.

Good integration-test targets include:

- Setup loading App Config, stored Bot data, replay inputs, Meta, and result paths.
- Controller receiving one complete Nuubot harness.
- Controller constructing Signaler, Risks, BotCycles, and Executors.
- Account, Ledger, and Simulator lifecycle behavior at their real owned boundary.
- Result publication from completed runtime results.
- Persistence, recovery, and reconciliation across package boundaries.


## System Tests

Use system tests for complete executable behavior and external evidence.

BtBot system proof must include:

- exact replay completion;
- expected ticks and Controller runs;
- domain and finance counts;
- graceful shutdown;
- result publication;
- database integrity; and
- deterministic fresh-process results when stability is authorized.

Canonical system execution uses `stest.sh`, Bot 9 or the explicitly selected Bot, and `-tags noasm`.

## Selection Rule

Test at the smallest boundary that still contains every real owner required by the behavior.

Move outward when isolation would require fake infrastructure, hidden decisions, or a second construction path.

Do not move outward merely to avoid testing deterministic local code.

## Proof Reporting

Report each command actually run.

Separate focused, integration, system, and omitted proof.

Never claim a passing level that did not run.
