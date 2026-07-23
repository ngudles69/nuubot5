# Ledger

## Purpose

Hold one Account's coherent local Trade, Order, Fill, accounting, and reconciliation state.

## Status

Approved — unimplemented.

## Canonical Sources

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\ownership.md
```

Supplemental behavior:

```text
D:\rust\nuubot3\wiki\account\ledger.md
D:\rust\nuubot3\wiki\coding\modules\ledger.md
D:\rust\nuutrader6\src\nuubot\hcbots\ledger\ledger.py
```

## Scope

Ledger is the canonical local domain cache for one Account context.

Venue or Simulator remains external truth.

## Owner and Children

```text
Account
`-- Ledger
    `-- Trades
        `-- Orders
            `-- Fills
```

Each child owns its immediate children.

## Responsibilities

- Atomically persist a new Trade and its Orders before Venue submission.
- Match validated Venue evidence to existing Orders.
- Apply Fills before Order lifecycle rows.
- Preserve domain identity and relationships.
- Refresh affected Orders and Trades once.
- Calculate accounting from owned Fill evidence.
- Track dirty and successful recon state.

## Does Not

- Query Venue.
- Decide when recon runs.
- Own Account.
- Adopt unrelated Venue activity.
- Revalidate normalized Account-boundary rows.
- Let arrival order select final truth.

## Lifecycle

```text
NewLedger
Init or load
create and reconcile domain state
Stop
```

`Stop` persists required evidence under the approved persistence mode.

## Inputs and Outputs

Inputs are validated Order intents and normalized Venue evidence.

Outputs are owned Trades, coherent accounting state, and snapshot source values.

## State and Invariants

- One Ledger belongs to one Account context.
- `trade_id` and BotCycle-local `trade_no` are different identities.
- Invalid matched batches cause no partial mutation.
- Fill deduplication uses stable Venue identity.
- Finalized Trades do not reopen through normal recon.

## Concurrency

Only the owning Account control path mutates Ledger.

Dirty marking MUST be atomic or serialized without exposing the domain tree.

## Persistence

Ledger owns domain persistence.

RuntimeStore reserves `trade_no` before creation.

Ledger atomically persists Trade and Orders before Account submits them.

SQLite is approved for backtesting. PostgreSQL is approved for live, simulator, and paper operation.

## Errors

Identity, matching, accounting, and persistence failures return errors before success metadata advances.

Ledger MUST NOT fall back to delayed or memory-only persistence after a required write fails.

## Program Flow

```text
receive reserved trade_no and Order intents
persist Trade and Orders atomically
return created domain objects
receive normalized Venue evidence
validate complete matched batch
apply Fills
apply Order lifecycle
refresh affected Trades
persist required changes
advance recon evidence
```

## Required Proof

- Trade and Orders either persist together or not at all.
- Invalid recon batches cause no mutation.
- Duplicate and stale evidence converges deterministically.
- Persistence failure advances no recon cursor or snapshot.

## Open Decisions

The Nuubot5 physical schema remains unapproved.

Do not copy unused Nuubot3 or Nuutrader6 columns.
