# Execution

Status: Implemented for TradeExecutor and GridExecutor with Simulator.
Covers: `internal/executor`, `internal/account`, `internal/account/ledger`, and domain packages
Purpose: Turn one Executor decision into validated domain evidence and Venue actions without breaking ownership.

## Scope

Execution crosses Executor, Account, Ledger, Trade, Order, CLOID, and Venue.

Recon remains a separate preceding process.

## Canonical Flow

```text
Controller completes recon
Controller evaluates Risk
Controller delivers successful recon event
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
Later recon applies Venue lifecycle and Fill truth
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

TP, SL, exit, cleanup, and stop Orders MUST attach to the Trade they close.

Executor shutdown MUST use the `stop` Order role.

An entry, TP, and SL bracket creates separate Orders under one Trade.

Missing ownership identity MUST fail before submission.

## Batch Outcome

Each request MUST receive exactly one explicit success or rejection.

One payload-wide Venue error MUST expand across every ordered request.

Malformed, incomplete, duplicated, or unknown results MUST preserve recoverable created evidence.

Mixed success and rejection MUST preserve each validated result.

Explicit item errors are terminal rejection evidence.

Successful acknowledgements remain submitted until recon.

Known local Simulator submission failure makes every created Order terminal `error`.

Immediate Fills MUST still enter canonical domain truth through recon.

Explicit market-like IOC Orders may execute during submission.

Their returned execution still marks Account dirty and enters Ledger through recon.

Mutation HTTP responses and WebSocket events are non-authoritative hints.

Only successful reconciliation commits canonical Order, Fill, position, and balance truth.

## Shutdown Execution

User stop and parent stop MUST use the same graceful Controller stop path.

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

## Boundary

Implemented slices are [TradeExecutor](trade-executor.md) and [GridExecutor](grid-executor.md) with Simulator.

Live and testnet mutations remain blocked.

The physical result design is [Trading Schema](trading-schema.md).
