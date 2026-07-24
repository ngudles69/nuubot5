# IngestBBO

Status: Partially implemented.
Covers: `internal/runtime/runtime.go`, `internal/botcycle/botcycle.go`, and `internal/executor/*.go`.
Purpose: Drive Simulator matching from BBO values without mixing simulated Venue work with Executor `OnBBO` policy.

## Contract

`IngestBBO` exists only for Simulator use.

Runtime calls the common ownership path without checking the selected Venue.
Only Simulator performs matching and fill work. Observer records delivery count
without changing simulated or domain state. A live Venue returns unchanged.

```text
Runtime.IngestBBO
  -> BotCycle.IngestBBO
  -> Executor.IngestBBO
  -> Account.IngestBBO
  -> Venue.IngestBBO
       Live: no-op
       Simulator: match Orders
```

Simulator changes mark the Account dirty. Simulated Order and Fill truth reaches
the Ledger and Executor only through the separate reconciliation process.

Each owner controls only its direct child. Runtime never reaches through
BotCycle or Executor to access Account, Venue, or Simulator.

## IngestBBO Versus OnBBO

| Operation | Owner | Purpose | Must not do |
|---|---|---|---|
| `IngestBBO` | Account-selected Venue path | Advance Simulator matching and simulated fills. | Run Executor policy or reconcile Account state. |
| `OnBBO` | Executor | Consume one BBO for Executor policy. | Advance Simulator, match Orders, create Fills, or reconcile Accounts. |

`IngestBBO` runs before `OnBBO` for the same admitted BBO.

`OnBBO` receives no simulated Fill result. Executor observes confirmed simulated
Order and Fill state only after reconciliation updates its Account and Ledger.

## Ordering

```text
receive validated BBO
  ingest BBO through existing Executor-owned Accounts
  let Simulator match existing Orders
  mark changed Accounts dirty
  deliver BBO through Executor.OnBBO

next due control pass
  reconcile dirty Accounts
  evaluate Risk
  run Executor decisions
```

New Orders created after this matching phase cannot match against the BBO that
already passed. They wait for a later BBO.

## Failure Handling

- Simulator ingestion failure stops delivery before `OnBBO`.
- Live Venue ingestion must return without mutation.
- Failed reconciliation preserves dirty state.
- Executor decisions must not consume unreconciled simulated outcomes.

## Required Proof

- Observer counts every delivered `IngestBBO` without matching Orders.
- Simulator matches an existing eligible Order from `IngestBBO`.
- Live Venue `IngestBBO` changes nothing.
- `IngestBBO` completes before `OnBBO` receives the same BBO.
- `OnBBO` cannot drive Simulator matching or create simulated Fills.
- Simulated Fill truth reaches Executor state only after reconciliation.
