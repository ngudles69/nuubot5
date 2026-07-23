# Execution

**Status:** Approved — unimplemented in Nuubot5.

## Purpose

Turn one Executor decision into validated domain evidence and Venue actions without breaking ownership.

## Scope

Execution crosses Executor, Account, Ledger, Trade, Order, CLOID, and Venue.

Recon remains a separate preceding process.

## Canonical Flow

```text
Runtime completes recon
Runtime evaluates Risk
Runtime asks BotCycle to run Executors
Executor chooses one action
Executor calls its Account
Account validates the complete batch
Ledger creates or resolves Trade
Ledger creates Orders
Account creates CLOIDs
Account records created intent
Account submits one Venue batch
Account validates every response item
Ledger records confirmed submission outcomes
Later recon applies venue lifecycle and Fill truth
```

## Responsibilities

- Preserve direct ownership calls.
- Validate complete batches before mutation or external calls.
- Attach every Order to one Trade.
- Record recoverable intent before uncertain Venue I/O.
- Keep one response result per request.
- Defer normal lifecycle truth to recon.

## Does Not

- Let Executor call Venue or Ledger directly.
- Let Venue create domain objects.
- Guess Trade attachment by symbol.
- Treat batch submission as one Order.
- Treat timeout as success or rejection.
- Skip recon before the next Risk or Executor decision.

## Trade Attachment

New entry batches MUST create one Trade.

TP, SL, exit, close, cleanup, and stop Orders MUST attach to the Trade they close.

An entry, TP, and SL bracket creates separate Orders under one Trade.

Missing ownership identity MUST fail before submission.

## Batch Outcome

Each request MUST receive exactly one explicit success or rejection.

Malformed, incomplete, duplicated, or unknown results MUST preserve recoverable created evidence.

Mixed success and rejection MUST preserve each validated result.

Immediate Fills MUST still enter canonical domain truth through recon.

## Shutdown Execution

User stop and end-date stop MUST use the same graceful Runtime stop path.

An active BotCycle MUST close through its Executors and Accounts.

Account MUST cancel or close only through approved Venue calls and preserve resulting evidence.

## Invariants

- Recon MUST precede Risk and execution decisions.
- Every domain Order MUST exist before its Venue submission.
- Every Order MUST have one Trade.
- CLOID identity MUST match returned Venue evidence.
- Venue MUST NOT mutate Ledger.
- Unknown external outcomes MUST remain recoverable.

## Reference Evidence

Canonical:

```text
D:\rust\nuubot4\wiki\recon.md
D:\rust\nuubot4\wiki\project.md
D:\rust\nuubot4\wiki\logic\executor.md
```

Supplemental:

```text
D:\rust\nuubot3\wiki\account\account.md
D:\rust\nuubot3\wiki\account\order.md
D:\rust\nuutrader6\src\nuubot\hcbots\account.py
D:\rust\nuutrader6\src\nuubot\hcbots\exchange.py
```

## Conflict

Nuubot4 currently uses ObserverExecutor without real Account execution. This page defines approved future behavior, not implemented parity.

## Recommendation

Implement one real Executor's minimum execution slice before generalizing the Account or Venue surface.
