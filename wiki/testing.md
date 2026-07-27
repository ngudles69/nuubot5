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
- Ledger mutations and persistence transactions.
- Small state transitions with no missing owner.

A unit test must not recreate global infrastructure with fake composition.

A unit test must not add test-only interfaces or alternate production paths.

## No Unit Testing

The following components must not have isolated unit tests:

- Controller.
- BotSpec.
- BtBot.

These are integrated components.

Isolated tests require excessive harnesses, fake data, fake infrastructure, and fake process state.

A passing isolated test would prove that artificial harness, not the real module path.

These components require the complete real construction and execution path.

Their proof is RTest through `Setup -> BotSpec -> Controller -> BtBot`.

The current RTest entrypoint is `./stest.sh -bot 9`.

The deleted historical `rtest.sh` command must not return.

Do not create fake Setup, Controller, BotSpec, Signaler, Risk, Executor, replay, or Meta infrastructure for these components.

## Integration Tests

Use integration tests when behavior requires multiple real owners or initialized infrastructure.

Good integration-test targets include:

- Setup loading App Config, stored Bot data, replay inputs, Meta, and result paths.
- Controller receiving complete Setup and one validated typed BotSpec.
- Controller constructing Signaler, Risks, BotCycles, and Executors.
- BotCycle, Executor, Account, Ledger, and Simulator lifecycle behavior.
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
