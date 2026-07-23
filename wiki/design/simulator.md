# Simulator

**Status:** Approved — unimplemented in Nuubot5.

## Purpose

Provide authoritative simulated venue truth behind the same Account-facing Venue contract.

## Scope

Simulator stores venue-shaped Orders, Fills, positions, balances, and lifecycle state.

It accepts admitted market input for matching.

## Responsibilities

- Implement the Venue batch and query contract.
- Accept validated Order and cancellation requests.
- Assign simulated venue Order and Fill identities.
- Match eligible Orders against admitted BBO values.
- Maintain simulated exposure, balances, fees, and Order lifecycle.
- Return Hyperliquid-shaped public evidence where the shared contract requires it.
- Persist simulator-owned venue truth under approved persistence policy.
- Report whether BBO ingestion changed simulated truth.

## Does Not

- Store domain Trade, Order, or Fill objects.
- Mutate Ledger.
- Reconcile Account state.
- Decide strategy, Risk, or Executor behavior.
- Return domain Fills directly from BBO ingestion.
- Treat persisted BBO as current market truth after restart.

## Ownership

Account owns Simulator through the Venue contract.

Simulator owns only simulated venue state.

Ledger owns the separate domain cache.

## Matching Boundary

The first admitted BBO MUST warm transient state without filling Orders.

Later BBO values may match eligible Orders under approved execution rules.

One BBO MUST NOT fill two exit legs for the same exposure.

Reduce-only execution MUST NOT increase or reverse exposure.

Changed simulated truth marks the Account dirty. It MUST NOT trigger recon directly.

## Logical Relationships

| Relation | Cardinality | Identity | Lifetime | Writer |
|---|---:|---|---|---|
| Account owns Simulator | 1 to 0..1 | Simulated account identity | Account | Account selects |
| Simulator stores venue Orders | 1 to 0..many | CLOID and simulated venue Order id | Simulator | Simulator |
| Simulator stores venue Fills | 1 to 0..many | Simulated venue Fill id | Simulator | Simulator |
| Ledger mirrors validated truth | Separate trees | Matched venue identity | Recon interval | Ledger, never Simulator |

This table defines logical design. It does not define a physical schema.

## Invariants

- Simulator and Ledger MUST remain separate truth stores.
- Simulator MUST emit the same semantic response shape Account expects from Hyperliquid.
- Simulator MUST retain source BBO and fill timestamps.
- Matching MUST be deterministic for identical admitted input and state.
- Persistence failure MUST fail loudly.
- Simulator persistence alone MUST NOT claim full Runner restartability.

## Reference Evidence

Canonical:

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\ownership.md
```

Supplemental:

```text
D:\rust\nuubot3\wiki\account\simulator.md
D:\rust\nuubot3\wiki\coding\simulator.md
D:\rust\nuutrader6\src\nuubot\hcbots\simulator.py
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
```

## Conflict

Nuubot3 specifies detailed matching and persistence behavior absent from Nuubot4. Nuubot5 accepts the boundary, not an unreviewed full copy.

## Recommendation

Port matching rules with the first Simulator-backed Executor. Prove parity against fixed Hyperliquid-shaped fixtures.
