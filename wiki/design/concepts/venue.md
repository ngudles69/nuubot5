# Venue

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Define the common Account-facing batch, cancellation, and query contract for live and simulated execution truth.

## Scope

Venue is the Account-owned behavioral boundary.

Hyperliquid and Simulator implement Venue.

The consumer package MUST own the smallest interface Account requires.

## Required Contract

Venue MUST support:

- initialize and close;
- ingest BBO for Simulator-only matching;
- place one batch of validated Orders;
- cancel one batch by supported venue identity;
- query open Orders;
- query bounded Fills;
- query exact Order status;
- query transient account state.

The exact Go method signatures belong to the Account implementation plan.

## IngestBBO

[`IngestBBO`](ingestbbo.md) is one common Account-facing Venue operation.

```text
Venue.IngestBBO
  Live
    return without changing state
  Simulator
    match eligible existing Orders
    record simulated Venue outcomes
    report whether Account state became dirty
```

Venue ingestion does not mutate Ledger or run Executor policy. Reconciliation
later copies validated Venue truth into Ledger evidence.

## Responsibilities

- Translate Account requests into venue-specific operations.
- Return venue-shaped responses and source timestamps.
- Preserve batch item identity and order.
- Surface partial batch success and rejection explicitly.
- Keep authoritative venue or simulated execution state.

## Does Not

- Mutate Ledger, Trade, Order, or Fill.
- Own Account.
- Reconcile domain state.
- Decide Risk or Executor behavior.
- Hide adapter failure behind fallback.
- Make live and Simulator implementations import each other.

## Ownership

```text
Account
`-- Venue
    |-- Hyperliquid
    `-- Simulator
```

One Account owns exactly one selected Venue implementation.

Venue lifetime matches Account lifetime.

## Logical Relationships

| Relation | Cardinality | Identity | Lifetime | Writer |
|---|---:|---|---|---|
| Account owns Venue | 1 to 1 | Network and account identity | Account | Account selects |
| Hyperliquid implements Venue | One selected implementation | Mainnet or testnet account | Account | Hyperliquid writes venue truth |
| Simulator implements Venue | One selected implementation | Simulated account identity | Account | Simulator writes simulated truth |

This table defines logical design. It does not define a physical schema.

## Batch Contract

Every submitted request MUST receive one explicit success or rejection result.

Malformed or incomplete batch responses MUST fail before domain state advances.

Returned CLOID MUST match its request when present.

Successful venue Order identities MUST be unique inside the response.

## Invariants

- Venue responses remain untrusted until Account validates them.
- Venue timestamps MUST retain source meaning.
- No implementation may silently fall back to another Venue.
- Venue MUST NOT expose mutable internal state.
- Venue implementation selection MUST occur during Account initialization.

## Reference Evidence

Canonical:

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\ownership.md
```

Supplemental:

```text
D:\rust\nuubot3\wiki\coding\exchange.md
D:\rust\nuubot3\wiki\account\simulator.md
D:\rust\nuutrader6\src\nuubot\hcbots\exchange.py
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
```

## Conflict

Nuubot4 left `Venue` versus `Exchange` unresolved. Nuubot5 uses `Venue` for this approved design.

## Recommendation

Keep the interface consumer-owned. Add methods only when Account has a current caller.
