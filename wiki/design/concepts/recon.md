# Reconciliation

Status: Approved — unimplemented. Refined by trading-state assessment.
Covers: No implemented source.
Purpose: Create one coherent post-venue Account state before Risk or Executor decisions.

## Scope

Recon is a major process crossing Runtime, BotCycle, Executor, Account, Venue, Ledger, and Risk.

Nuubot4 owns the canonical order and ownership.

## Canonical Flow

```text
Runtime requests BotCycle account recon
BotCycle asks every AccountReconciler
Each capable Executor reconciles its owned Account
Each Account queries Venue
Each Account validates venue evidence
Each Ledger applies one coherent recon
Each Account returns AccountSnapshot
BotCycle returns all snapshots
Runtime evaluates Risk
Runtime dispatches OnRecon to capable Executors
```

Recon MUST complete before Risk and Executor decisions.

Runtime owns work order. Runtime MUST NOT own Accounts.

## Correctness Source

Nuubot does not depend on complete WebSocket delivery for correctness.

Nuubot does not treat one mutation HTTP response as final lifecycle truth.

WebSocket events and immediate responses reduce latency and mark Accounts dirty.

They may arrive late, repeat, disappear, or describe only one transition.

Reconciliation repeatedly queries Venue Orders, Fills, positions, and balances.

Ledger changes only from one validated coherent reconciliation batch.

Hyperliquid history responses are bounded and may omit older activity.

Absence from one history response never proves an Order or Fill did not exist.

Recon upserts returned facts. It never deletes local evidence because a bounded response omitted it.

Forced reconciliation catches drift when no dirty hint arrived.

Failed reconciliation blocks Risk and Executor decisions.

Nuubot never guesses missing Venue truth.

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
2. Fills from its inclusive time cursor;
3. exact status for unmatched active local Orders;
4. transient account state.

Account validates each response. Ledger then performs matching and mutation.

Account narrows history by cursor and time range.

A cap-sized response is potentially incomplete.

Account MUST narrow or continue the range before advancing beyond an unproven gap.

If Venue cannot prove a complete range, recon fails without advancing its cursor.

## Dirty and Forced Recon

User events and Simulator changes may mark Account truth dirty.

Dirty state requests later recon. It MUST NOT perform recon immediately.

A normal recon pass skips clean Accounts.

This makes WebSocket `userEvents` a performance optimization.

Submission acknowledgements and Simulator mutations also mark their Account dirty.

A slower forced Run MUST reconcile every Account despite missing dirty hints.

Exact cadences remain Runner configuration, not Account or Ledger policy.

The first initialized Account starts dirty.

BtRunner may force reconciliation from its existing ten-second Runtime cadence.

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
- Bounded-history absence MUST NOT delete or terminalize local evidence.
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

Use one Account first.

Do not generalize shared or multi-Account ownership in this tranche.
