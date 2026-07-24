# Reconciliation

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Create one coherent post-venue Account state before Risk or Executor decisions.

## Scope

Recon is a major process crossing Runtime, BotCycle, Executor, Account, Venue, Ledger, and Risk.

Nuubot4 owns the canonical order and ownership.

## Canonical Flow

```text
Runtime requests BotCycle account recon
BotCycle asks every Executor
Each Executor reconciles its own Accounts
Each Account queries Venue
Each Account validates venue evidence
Each Ledger applies one coherent recon
Each Account returns AccountSnapshot
BotCycle returns all snapshots
Runtime evaluates Risk
Runtime lets BotCycle run Executor decisions
```

Recon MUST complete before Risk and Executor decisions.

Runtime owns work order. Runtime MUST NOT own Accounts.

## Responsibilities

- Reconcile zero, one, or many Accounts through the same Executor contract.
- Gather owned AccountSnapshot values.
- Establish one post-recon barrier.
- Block later decisions when any selected Account fails.
- Preserve Account ownership inside each Executor.

## Does Not

- Share mutable Accounts.
- Let Runtime reach through BotCycle.
- Run from a WebSocket callback.
- Let BBO or user events mutate Ledger directly.
- Continue Risk or Executor decisions after failure.
- Create separate single-account and multi-account recon paths.

## Venue Query Order

Each Account MUST query:

1. open Orders;
2. bounded Fills;
3. exact status for unmatched active local Orders;
4. transient account state.

Account validates each response. Ledger then performs matching and mutation.

## Dirty and Forced Recon

User events and Simulator changes may mark Account truth dirty.

Dirty state requests later recon. It MUST NOT perform recon immediately.

A slower forced Run MUST reconcile every Account despite missing dirty hints.

Exact cadences remain Runner configuration, not Account or Ledger policy.

## Failure Contract

Failed recon MUST:

- produce no snapshot for the failed Account;
- restore or retain dirty state;
- avoid success timestamps and cursors;
- prevent Risk evaluation;
- prevent Executor decisions;
- propagate through the normal Bot failure path.

## Ownership

```text
Runtime -> BotCycle -> Executor -> Account -> Ledger
```

Each call controls only its direct child.

Snapshots travel upward as owned values.

## Invariants

- All snapshots in one Risk evaluation MUST follow the same completed recon barrier.
- Runtime MUST retain no Account reference.
- Ledger MUST apply no partial invalid batch.
- Clean Accounts may reuse their latest completed snapshot only under an approved cadence contract.
- Account truth MUST remain Venue-authoritative.

## Reference Evidence

Canonical:

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\ownership.md
D:\rust\nuubot4\wiki\logic\risk.md
```

Supplemental:

```text
D:\rust\nuubot3\wiki\account\account.md
D:\rust\nuubot3\wiki\account\ledger.md
D:\rust\nuutrader6\src\nuubot\hcbots\recon.py
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
```

## Conflict

Nuubot3 lets Runtime iterate shared Accounts. Nuubot5 preserves the Nuubot4 Executor-owned traversal.

## Recommendation

Retain dirty and forced cadence behavior, but approve exact timer values with the live Runner design.
