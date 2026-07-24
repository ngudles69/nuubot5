# Trading State Tranche

Status: Proposed assessment.
Covers: No implemented source.
Purpose: Define the smallest complete simulated trading path for the next implementation tranche.

## Verdict

Nuubot5 is structurally ready but not implementation-close.

Executor capability dispatch and Simulator BBO routing already exist.

Account, Ledger, Trade, Order, Fill, Simulator, reconciliation, and writable result storage remain reservations.

Nuubot3 provides the strongest reusable domain design.

Its ownership, reconciliation, persistence, tests, and DDL are substantially complete.

Nuutrader6 provides the proven Simulator behavior.

`async_hyperliquid` provides the client-visible response contract.

The Go work is a disciplined rewrite, not a mechanical translation.

## Evidence

| Area | Nuubot5 now | Best reference | Reusable | Main gap |
|---|---|---|---|---|
| TradeExecutor | Factory and optional BBO handlers | `nuubot3/nuubot/executor/tradebot.py` | Planning and bracket intent | Account binding and recon capabilities |
| Account | Reserved package | `nuubot3/nuubot/account/account.py` | Ownership, validation, submit, recon | Go types, store, and Venue construction |
| Ledger | Reserved package | `nuubot3/nuubot/account/ledger.py` | Atomic recon and cursor rules | Go domain tree and transactions |
| Trade | Reserved package | `nuubot3/nuubot/account/trade.py` | State and PnL rules | Go decimal arithmetic |
| Order | Reserved package | `nuubot3/nuubot/account/order.py` | Lifecycle and Fill aggregation | Typed transitions |
| Fill | Reserved package | `nuubot3/nuubot/account/fill.py` | Immutable execution identity | Go value and enrichment rules |
| Simulator | Reserved package | Nuutrader6 Simulator | Matching behavior | Go implementation and parity fixtures |
| Schema | No writable result schema | Nuubot3 Account DDL | Relationships and indexes | Nuubot5 per-Bot SQLite DDL |
| Credentials | TOML decoding only | Nuubot3 Account boundary | Selected-account validation | Runtime-to-Account delivery |
| Hyperliquid | Design only | Official API and `async_hyperliquid` | Wire shapes | Go signing and live calls |

## Reference Order

1. Nuubot5 source and approved ownership.
2. Official Hyperliquid API semantics.
3. Nuutrader6 Simulator and reconciliation behavior.
4. Installed `async_hyperliquid` 0.4.8 response outputs.
5. Nuubot3 Account, Ledger, domain, tests, and DDL.
6. Older Nuubot and Nuubot2 only for smaller implementation ideas.

Older Nuubot directly mixes Simulator results into Ledger.

That shortcut violates the approved reconciliation boundary.

Nuubot2 provides a small Trade plan but no complete Account domain.

## Target Ownership

```text
Runtime
`-- active BotCycle
    `-- TradeExecutor
        `-- Account
            |-- selected Venue
            |   `-- Simulator
            `-- Ledger
                `-- Trades
                    `-- Orders
                        `-- Fills
```

Runtime owns no Account reference.

BotCycle knows Executor capabilities, never concrete TradeExecutor fields.

TradeExecutor owns Account configuration and its Account.

Account owns one selected Venue and one Ledger.

Simulator owns exchange-shaped truth.

Ledger owns domain-shaped evidence.

## Target BBO Flow

```text
BtRunner reads one BBO
  Runtime stores the latest BBO
  Runtime.IngestBBO
    BotCycle.IngestBBO
      TradeExecutor.IngestBBO
        Account.IngestBBO
          Simulator.IngestBBO
    BotCycle.OnBBO
      TradeExecutor.OnBBO
  TickClock advances
  Runtime.Run may execute
```

Simulator changes only Simulator truth.

A Simulator change marks its Account dirty.

`OnBBO` records policy input only.

Neither BBO path reconciles Ledger state.

## Target Control Flow

```text
Runtime.Run
  check stop request
  reconcile active BotCycle Accounts when dirty or forced
  stop the pass on recon failure
  assess Risk from Account snapshots
  stop the pass on Risk stop
  dispatch successful recon to capable Executors
  check BotCycle completion
  close a completed BotCycle
  check maximum cycles
  read and consume the next Signal
  initialize a new BotCycle when admitted
```

The first Account state starts dirty.

Dirty reconciliation runs on the normal Runtime cadence.

A forced reconciliation periodically catches missed dirty hints.

Clean Accounts skip normal Venue queries.

WebSocket `userEvents`, submissions, and Simulator mutations mark Accounts dirty.

Exact live cadence remains Runner policy.

The BtRunner target may use its existing ten-second Runtime timer.

Venue history is bounded.

Missing history rows never delete local Orders or Fills.

Account passes one configured `persist_mode` to Ledger and Simulator.

`none` exports only after success. `max` persists every accepted state change.

Before each BotCycle teardown, terminal evidence flows upward by owned value:

```text
Ledger.Result + Simulator.Result
  -> Account.Result
  -> Executor.AccountResult
  -> BotCycle.Result
  -> Runtime.Result
  -> BtRunner ResultPublisher
```

Result values alias no mutable child state.

Simulator evidence is present only for Simulator-backed Accounts.

ResultPublisher writes `none` evidence only after shutdown and successful replay verification.

## TradeExecutor Slice

The first TradeExecutor uses existing standard entry Signals.

It needs no new Signaler fields.

Configuration supplies notional, take-profit percentage, stop-loss percentage, Account name, and Simulator capital.

The latest admitted BBO supplies entry reference price.

TradeExecutor submits one entry, one take-profit, and one stop-loss batch.

Account creates one Trade and three Orders before Venue submission.

TradeExecutor finishes only after its Trade is terminal and no owned Order remains active.

## Submit Boundary

```text
TradeExecutor builds semantic order requests
  Account validates the complete batch
  Ledger commits created Trade and Orders
  Account builds CLOIDs
  Account submits one Venue batch
  Account validates every returned item
  Ledger records confirmed submission outcomes
  later reconciliation records Venue Order and Fill truth
```

Every request receives exactly one explicit success or rejection.

One payload-wide Hyperliquid error expands across every ordered batch item.

Unknown submission outcomes leave durable `created` Orders for exact CLOID recovery.

Immediate Venue fills still enter domain truth through reconciliation.

## Simulator Parity Boundary

Simulator matches proven Hyperliquid behavior.

Its public responses match the admitted `async_hyperliquid` contract.

Account translates live and Simulator evidence through one path.

Parity covers output shapes, field meanings, status values, and item ordering.

Parity does not copy Python classes, async structure, dependencies, or defects.

See [Simulator Parity](simulator-parity.md).

## Credentials

Setup continues decoding the complete credentials file.

Runtime passes the admitted read-only credentials catalog through BotCycle.

TradeExecutor selects one configured Account name.

Account validates only the selected live or testnet credentials during initialization.

Simulator never receives private credentials.

BtRunner must reject live and testnet TradeExecutor Accounts.

No credential value may enter logs, errors, snapshots, DDL, or raw evidence.

## Implementation Order

1. Approve one pure-Go decimal representation.
2. Implement Fill, Order, and Trade with focused state tests.
3. Implement the per-Bot result store and Ledger transactions.
4. Implement Simulator and Hyperliquid-shape parity fixtures.
5. Implement Account submission, translation, dirty state, and reconciliation.
6. Implement TradeExecutor with Simulator-only admission.
7. Add BotCycle reconciliation and recon-event capabilities.
8. Reorder Runtime control around reconciliation, Risk, and decisions.
9. Prove one complete replay from Signal through terminal Trade evidence.

Each stage must leave one runnable proof before the next stage starts.

## Deferred

- Live Hyperliquid mutation calls.
- Shared live WebSocket ownership.
- Partial fills caused by order-book depth.
- Queue position and multi-level books.
- Funding, liquidation, leverage changes, and portfolio margin policy.
- Polymarket translation.
- Multiple Accounts per Executor.
- Shared Accounts across Executors.

## Required End-to-End Proof

- One long and one short Signal each create one Trade and three Orders.
- Simulator submission output matches admitted Hyperliquid fixtures.
- New resting Orders cannot match an already-consumed BBO.
- Explicit market-like Orders may fill immediately from their supplied reference.
- Simulator BBO matching marks Account dirty without mutating Ledger.
- Reconciliation creates each Fill once and advances its cursor once.
- Duplicate reconciliation is idempotent.
- Contradictory evidence fails before mutation.
- TP or SL fill cancels its sibling.
- Reduce-only execution never reverses exposure.
- Replay shutdown cancels exits, closes exposure, reconciles, and finishes flat.
- SQLite foreign keys, uniqueness, and terminal-state rules hold.
- No credential value appears in logs or persisted evidence.
