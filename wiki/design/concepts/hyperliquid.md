# Hyperliquid

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Implement live and testnet Hyperliquid truth behind the common Venue contract.

## Scope

This page owns Hyperliquid-specific signing, transport, request mapping, response mapping, precision, limits, and WebSocket evidence.

No Go SDK, transport, or external dependency is selected.

Any new dependency requires explicit prior user approval.

## Required Venue Behavior

Hyperliquid MUST support the Account-required subset:

- batch Order placement;
- batch cancellation by supported Order identity;
- open Order queries;
- bounded Fill queries;
- exact Order status queries;
- transient account-state queries;
- clean initialization and shutdown.

The implementation MUST preserve one result per submitted batch item.

## WebSocket Behavior

The BBO feed supplies responsive stop and trailing-stop input.

The user-event feed marks Account or Ledger truth dirty.

WebSocket readers MUST NOT reconcile, mutate Ledger, place Orders, or execute Runtime policy.

Runtime's synchronous timers consume feed state at approved cadences.

## Responsibilities

- Authenticate and sign Hyperliquid requests.
- Map normalized Account requests into Hyperliquid payloads.
- Map Hyperliquid responses into Venue response values.
- Preserve CLOID, venue Order identity, Fill identity, and venue timestamps.
- Expose explicit per-item success and rejection.
- Enforce authoritative price and quantity precision.
- Surface rate limits, transport failures, and venue rejections.
- Keep mainnet and testnet configuration explicit.

## Does Not

- Mutate Ledger, Trade, Order, or Fill.
- Simulate Fills.
- Fall back to Simulator.
- Own Account or Runtime.
- Select strategy or Risk policy.
- Invent missing Venue outcomes.
- Use CGO or non-standard Go without explicit user approval.

## Batch Contract

Every Order response MUST be structurally complete.

A success MUST contain the required venue identity or complete immediate-Fill evidence.

A rejection MUST contain a non-empty venue reason.

Malformed or incomplete responses MUST fail as unknown outcomes.

Returned CLOID MUST match the submitted CLOID when present.

## Recon Evidence

Hyperliquid provides external truth.

Account MUST validate and normalize that evidence before Ledger receives it.

Venue PnL fields remain evidence. Trade calculates canonical domain PnL from local Fill records.

## Logical Relationships

| Relation | Cardinality | Identity | Lifetime | Writer |
|---|---:|---|---|---|
| Account owns Hyperliquid Venue | 1 to 0..1 | Network and account identity | Account | Account selects |
| Hyperliquid owns venue Orders | 1 account to 0..many | CLOID and venue Order id | Venue retention | Hyperliquid |
| Hyperliquid owns venue Fills | 1 account to 0..many | Venue Fill identity | Venue retention | Hyperliquid |
| Ledger mirrors validated evidence | Separate local tree | Matched CLOID and Fill identity | Account | Ledger |

This table defines logical design. It does not define a physical schema.

## Safety

- Credentials MUST remain outside source, wiki, logs, tests, and prompts.
- External responses MUST remain untrusted until validated.
- Retry behavior MUST distinguish safe query retries from uncertain mutations.
- Rate-limit handling MUST preserve request identity.
- Mainnet and testnet MUST never be selected by implicit fallback.

## Reference Evidence

Canonical boundary:

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\cloid.md
```

Supplemental behavior:

```text
D:\rust\nuubot3\wiki\coding\async_hyperliquid.md
D:\rust\nuubot3\wiki\coding\exchange.md
D:\rust\nuubot3\wiki\account\account.md
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
D:\rust\nuutrader6\src\nuubot\hcbots\exchange.py
```

## Conflict

Reference implementations use Python `AsyncHyper`. That library does not select or approve any Go dependency.

## Recommendation

Evaluate pure-Go SDK batch support separately. Approve a dependency only after real testnet place, cancel, query, and WebSocket proof.
