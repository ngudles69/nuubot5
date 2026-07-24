# ResultPublisher

Status: Implemented for per-Bot SQLite result databases.
Covers: `internal/resultpublisher/*.go`
Purpose: Publish exactly one terminal backtest result from approved Runtime evidence.

## Canonical Sources

- Nuubot4 ownership: `D:/rust/nuubot4/wiki/ownership.md`
- Nuubot3 design: `D:/rust/nuubot3/wiki/runner-lifecycle.md`

## Scope

ResultPublisher translates one immutable Runtime result snapshot into durable BtRunner and Sweep result evidence.

Runtime captures descendant evidence before each BotCycle teardown.

## Owner and Children

BtRunner owns ResultPublisher.

ResultPublisher owns no Runtime descendant.

## Responsibilities

- Accept one terminal Runtime result snapshot.
- Select memory-only Account results using one result path.
- Persist supplied Ledger and Simulator evidence when `persist_mode = none`.
- Replace the completed per-Bot result database.
- Return publication failure to BtRunner.

## Does Not

- Traverse Runtime, BotCycle, Account, Ledger, Trade, Order, or Fill.
- Calculate trading results from mutable objects.
- Decide whether Runtime should stop.
- Publish partial success as terminal success.
- Define result schema in this page.

## Invariants

- One BtRunner publishes one terminal file.
- Publication follows Runtime shutdown and replay verification.
- Runtime result values alias no stopped child state.
- Failed Runtime execution publishes no partial Ledger or Simulator evidence.
- A failed publication does not replace the prior completed database.
- Publication failure makes BtRunner fail.

## Atomic Publication

For `none`, Account opens no result database.

ResultPublisher creates a temporary database beside the final path after success.

It writes all evidence into that hidden file, closes every database handle, then atomically renames it.

Only the final filename is completed evidence.

Failure removes or leaves only an ignored temporary file.

## Required Proof

- Repeated publication replaces the prior completed file.
- Incomplete replay cannot publish success.
- Snapshot identity mismatch fails.
- Failed publication leaves the prior completed database unchanged.

Shared Sweep catalog updates remain pending.
