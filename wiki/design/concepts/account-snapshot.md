# AccountSnapshot

Status: Approved — unimplemented.
Covers: No implemented source.
Purpose: Carry one Account's coherent post-recon state into Runtime and Risk without exposing mutable Account ownership.

## Canonical Sources

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\ownership.md
D:\rust\nuubot4\wiki\logic\risk.md
```

Supplemental behavior:

```text
D:\rust\nuutrader6\src\nuubot\hcbots\simulator.py
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
```

## Scope

AccountSnapshot is an owned value containing only reconciled facts required by Runtime and Risk.

## Owner and Children

Account creates the value. Runtime receives it through Executor and BotCycle.

AccountSnapshot owns no mutable domain children.

## Responsibilities

- Preserve Account identity and observation time.
- Carry one successful recon result.
- Remain valid for one Runtime control pass.

## Does Not

- Own or mutate Account, Venue, Ledger, Trade, Order, or Fill.
- Query Venue.
- Calculate Risk policy.
- Replace the Ledger tree.

## Lifecycle

Account creates one snapshot after successful recon. Risk reads it during one control pass. The value then expires.

## Inputs and Outputs

Input is coherent reconciled Ledger and Venue account state.

Output is one immutable-by-contract value for Runtime and Risk.

## State and Invariants

- A failed recon MUST produce no snapshot.
- One Risk evaluation MUST use snapshots from one completed recon barrier.
- Snapshot values MUST contain no Account pointers or mutable child collections.

## Concurrency

Snapshots cross ownership boundaries by value.

No lock or shared mutable state belongs inside AccountSnapshot.

## Persistence

AccountSnapshot is not persisted.

## Errors

Snapshot creation MUST fail when required reconciled facts are missing or inconsistent.

## Program Flow

```text
Account reconciles Venue into Ledger
Account creates AccountSnapshot
Executor returns snapshots through BotCycle
Runtime passes snapshots to Risk
```

Runtime MUST NOT retain or own Accounts.

## Required Proof

- Successful recon produces the expected snapshot.
- Failed recon produces no snapshot.
- Runtime and Risk receive values without Account references.

## Open Decisions

Define exact Go fields with the first real Risk port.

Nuubot3's Runtime-owned Account list is rejected.
