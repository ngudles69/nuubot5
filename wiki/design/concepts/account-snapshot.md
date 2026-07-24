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

It is not terminal Ledger or Simulator evidence.

## Responsibilities

- Preserve Account identity and observation time.
- Carry one successful recon result.
- Remain valid for one Runtime control pass.

## Does Not

- Own or mutate Account, Venue, Ledger, Trade, Order, or Fill.
- Query Venue.
- Calculate Risk policy.
- Replace the Ledger tree.
- Supply detailed `persist_mode = none` result publication.

## Lifecycle

Account creates one snapshot after successful recon. Risk reads it during one control pass. The value then expires.

## Inputs and Outputs

Input is coherent reconciled Ledger and Venue account state.

Output is one immutable-by-contract value for Runtime and Risk.

## State and Invariants

- A failed recon MUST produce no snapshot.
- One Risk evaluation MUST use snapshots from one completed recon barrier.
- Snapshot values MUST contain no Account pointers or mutable child collections.

## Proposed Initial Fields

| Field | Meaning |
|---|---|
| `cycle_no` | Owning BotCycle number |
| `executor_no` | Owning Executor number |
| `account_name` | Configured non-secret Account identity |
| `network` | Mainnet, testnet, or simnet |
| `symbol` | Reconciled instrument |
| `observed_ms` | Completed reconciliation observation time |
| `account_value` | Venue account value |
| `withdrawable` | Venue withdrawable value |
| `position_qty` | Signed current symbol exposure |
| `entry_price` | Current average Venue entry price |
| `unrealized_pnl` | Current Venue unrealized PnL evidence |
| `gross_pnl` | Ledger-calculated gross PnL |
| `fees` | Ledger-calculated fees |
| `net_pnl` | Ledger-calculated net PnL |
| `open_trades` | Nonterminal Trade count |
| `active_orders` | Active Order count |
| `fills` | Admitted Fill count |

Financial values use the approved decimal representation.

The snapshot contains no raw credential or private Venue payload.

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
Executor returns one snapshot through BotCycle
Runtime passes snapshots to Risk
```

Runtime MUST NOT retain or own Accounts.

## Required Proof

- Successful recon produces the expected snapshot.
- Failed recon produces no snapshot.
- Runtime and Risk receive values without Account references.

## Open Decisions

Approve the initial field table with the trading-state tranche.

Nuubot3's Runtime-owned Account list is rejected.
