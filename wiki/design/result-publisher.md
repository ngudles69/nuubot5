# ResultPublisher

## Purpose

Publish exactly one terminal backtest result from approved Runtime evidence.

## Status

Approved — unimplemented.

## Canonical Sources

- Nuubot4 ownership: `D:/rust/nuubot4/wiki/ownership.md`
- Nuubot3 design: `D:/rust/nuubot3/wiki/runner-lifecycle.md`

## Scope

ResultPublisher translates one immutable Runtime result snapshot into durable BtRunner and Sweep result evidence.

## Owner and Children

BtRunner owns ResultPublisher.

ResultPublisher owns no Runtime descendant.

## Responsibilities

- Accept one terminal Runtime result snapshot.
- Validate identity, completion, and required evidence.
- Persist the Bot result once.
- Update the owning Sweep result boundary through an approved store.
- Return publication failure to BtRunner.

## Does Not

- Traverse Runtime, BotCycle, Account, Ledger, Trade, Order, or Fill.
- Calculate trading results from mutable objects.
- Decide whether Runtime should stop.
- Publish partial success as terminal success.
- Define result schema in this page.

## Invariants

- One BtRunner publishes at most one terminal result.
- Result identity matches Sweep and Bot identity.
- Publication follows Runtime shutdown and replay verification.
- Publication failure makes BtRunner fail.

## Required Proof

- Duplicate publication is rejected or idempotent.
- Incomplete replay cannot publish success.
- Snapshot identity mismatch fails.
- Sweep aggregate sees the same terminal Bot result.

