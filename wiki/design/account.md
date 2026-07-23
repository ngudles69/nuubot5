# Account

## Purpose

Provide one Executor-owned trading boundary with one Venue and one coherent local Ledger.

## Status

Approved — unimplemented.

## Canonical Sources

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\ownership.md
```

Supplemental behavior:

```text
D:\rust\nuubot3\wiki\account\account.md
D:\rust\nuubot3\wiki\coding\accounts.md
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
D:\rust\nuutrader6\src\nuubot\hcbots\recon.py
```

## Scope

Account coordinates submission, cancellation, reconciliation, and shutdown for one configured account context.

Venue truth remains authoritative. Ledger stores the reconciled local domain view.

## Owner and Children

```text
Executor
`-- zero or more Accounts
    |-- one Venue
    `-- one Ledger
```

Account lifetime remains inside its owning Executor and BotCycle.

## Responsibilities

- Validate Account-level requests before mutation or Venue calls.
- Build CLOID-bearing Order intents.
- Coordinate atomic Trade and Order persistence through Ledger.
- Submit and cancel batches through Venue.
- Query and validate Venue truth.
- Return AccountSnapshot after successful recon.
- Stop Ledger and Venue through defined ownership.

## Does Not

- Belong to Runtime.
- Expose Venue or Ledger for direct Executor mutation.
- Calculate Trade PnL.
- Let strategy code create CLOIDs.
- Guess Trade ownership from symbol.
- Let Venue mutate domain state.

## Lifecycle

```text
NewAccount
Init
operations
Recon
Stop
```

Initialization MUST publish neither child until Ledger and Venue are ready.

Nuubot lifecycle uses `Stop`. A transport's private `Close` call remains inside its Venue implementation.

## Inputs and Outputs

Inputs include validated account config, execution requests, timestamps, BBO values, credentials, and recon triggers.

Outputs include batch outcomes, cancellation outcomes, AccountSnapshot values, and terminal errors.

## State and Invariants

- Runtime MUST NOT own or retain Account references.
- Venue and Ledger MUST represent the same account context.
- No Venue call may follow failed request validation.
- Every Order MUST attach to one Trade.
- Recon failure MUST block later decisions.

## Concurrency

Account mutations occur inside the owning Executor's synchronous control path.

Feed goroutines may mark dirty state. They MUST NOT reconcile or mutate Ledger.

## Persistence

RuntimeStore reserves `trade_no` before Trade creation.

Account builds CLOID-bearing Order intents from reserved identity.

Ledger atomically persists the Trade and Orders before Venue submission.

## Errors

Validation, persistence, Venue, and recon failures return errors.

Timeout, malformed, or incomplete Venue responses remain unknown. Account retains `created` evidence for exact CLOID recon.

Account MUST NOT invent rejection.

## Program Flow

```text
Executor calls Account
RuntimeStore reserves trade_no
Account validates the complete batch
Account builds CLOID-bearing Order intents
Ledger atomically persists Trade and Orders as created
Account submits the Venue batch
Account validates the complete decoded response
Ledger records explicit per-item outcomes
Later recon applies Venue lifecycle and Fill truth
```

## Required Proof

- Invalid batches cause no persistence or Venue call.
- Trade and Orders commit atomically before submission.
- Complete decoded responses map one result per request.
- Unknown outcomes preserve `created` evidence.
- Recon returns snapshots without exposing Accounts.

## Open Decisions

Exact Go fields and physical schema remain implementation decisions.

If two Executors need one mutable Account, stop and redesign with explicit user approval.
