# Account Package

Status: Implemented for Simulator-backed Accounts.
Covers: `internal/account/*.go`
Purpose: Give one Executor a trading boundary backed by one Venue and one local Ledger.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/account/account.py`
- `D:/rust/nuubot3/wiki/account/account.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/account.py`

## Ownership

TradeExecutor or GridExecutor owns one Account.

Account owns one selected Venue and one Ledger.

The current BtBot implementation selects Simulator only.

Account hides Venue selection and response translation from Executor.

Executor supplies `order_level`.

Account supplies Trade, batch, purpose, and remaining CLOID identity.

These values stay inside Account and Ledger.

Account sends Venue only official operation fields.

Persisted `order_pos` remains request position inside one batch.

## Construction

Account is one concrete component.

It does not need a factory.

TradeExecutor or GridExecutor owns one Account value and calls `Init`.

`Init` validates identity before opening Ledger or Venue state.

Account publishes neither child after partial initialization failure.

## Inputs

| Input | Purpose |
|---|---|
| Nuubot | Supply Logger, App Config, MarketData, Meta, ResultPath, and RuntimePath |
| Cycle and Executor numbers | Create stable result identity |
| Account config | Select network, name, capital, fees, and persistence |
| Selected credentials | Initialize a future live or testnet Venue |
| Current Clock | Preserve deterministic operation timestamps |
| Store operations | Required only when `persist_mode = max` |

Simulator initialization receives no private credential.

It also receives no Ledger, Trade, local Order, role, or purpose identity.

## Program Flow

```text
Init
  bind Account inputs
  validate Account identity
  validate persistence mode
  initialize Ledger with persistence mode
  initialize Venue with persistence mode
  initialize Account

PlaceOrders
  validate complete Order batch
  resolve Trade ownership
  create CLOIDs
  commit created Trade and Orders
  submit Venue batch
  validate submit response
  commit submit outcomes
  mark Account dirty

CancelOrders
  validate owned active Orders
  cancel Venue batch
  validate cancel response
  mark Account dirty



Reconcile
  record reconciliation call
  execute reconciliation
  publish reconciliation outcome

reconcile
  prepare attempt
  download current Order evidence
  download Fill history
  download current Account state
  update Fill records
  update Order records
  update Trade records
  update Account Snapshot
  persist and publish
  finalize Recon outcome and return

Result
  get immutable Ledger result
  return immutable Account result

Stop
  stop Venue
  stop Ledger
  stop Account
```

Each indented action becomes one exact source comment during implementation.

## Submit Contract

Account validates the complete batch before mutation.

One new entry batch creates one Trade.

TP, SL, exit, cleanup, and stop Orders attach to an existing Trade.

Executor shutdown uses the canonical `stop` Order role.

One entry, TP, and SL bracket creates three Orders under one Trade.

Account persists `created` intent before Venue I/O.

Every request receives one explicit success or rejection.

Every Nuubot request carries one mandatory opaque official CLOID.

Venue assigns OID once after accepting the request.

Account binds ordered acknowledgement to the local Order.

One known Venue submission failure maps every local Order to `error`.

Malformed or incomplete responses leave recoverable `created` evidence.

An explicit item error maps its Order to `rejected`.

Successful acknowledgements remain submitted until reconciliation.

Account never retries uncertain mutation outcomes automatically.

Immediate Fills still enter Ledger through reconciliation.

Fill history may omit CLOID while retaining Venue OID.

Account enriches it from same-attempt Order evidence before Ledger Fill application.

Duplicate or contradictory identity evidence fails the reconciliation attempt.

HTTP mutation responses are acknowledgement evidence, not final lifecycle truth.

## Reconciliation

### Current Implementation

Account queries Venue in this order:

1. Open Orders.
2. Exact status for selected active local Orders missing from the bulk response.
3. Fills from the inclusive Ledger cursor.
4. Transient account state.

Every query returns fresh detached official JSON.

Account validates each untrusted response through `internal/hyperliquid`.

Open Order evidence resolves CLOID-first and OID-fallback.

Fill evidence normally resolves OID because official Fill rows may omit CLOID.

If both CLOID and OID exist, they must identify the same local Order.

Exact status lookup is exception handling. Recon telemetry counts every attempted lookup.

Ledger receives normalized concrete values.

Normal reconciliation returns the latest snapshot without querying a clean Account.

Forced reconciliation queries Venue even when Account is clean.

Failed reconciliation restores dirty state. It advances no cursor or success timestamp.

### Approved Live Target

Normal live reconciliation queries `openOrders`, paginated `userFillsByTime`, exact
`orderStatus` for missing active Orders, then Account state.

Exact `orderStatus` is not the normal Order download. Telemetry proves its observed frequency.

Hyperliquid Fill history has no symbol filter. Responses cap at 2,000 rows, and
`userFillsByTime` retains only the latest 10,000 Fills.

Account continues capped responses and deduplicates inclusive cursor boundaries by Venue TID.

Routine reconciliation does not query `historicalOrders`. `openOrders` has no documented 2,000-row cap.

Account and Ledger compare through stable identity indexes.

Work touches only active Orders, new Fills, touched Trades, and the Account candidate.

The cursor advances only after every Venue read, validation, Ledger change, Account calculation, and persistence step succeeds.

A missing active Order remains unresolved after inconclusive exact lookup.

Future live behavior quarantines its owning Grid level without replacement, reuse, or assumed outcome.

Other levels continue only within a separately approved safety boundary. Sweep fails immediately.

Configurable cleanup may run when unresolved Orders exist and cleanup is due.

Cleanup reads the latest 2,000 historical Orders and Fills, matches CLOID, OID,
and TID, and repairs only exact evidence.

No cleanup default or escalation threshold is approved.

## Dirty State

Account solely owns its reconciliation-dirty flag.

Initialization, user events, submissions, changed Simulator truth, and open-position marks make Account dirty.

Simulator invokes one narrow Account-owned change callback after matching or marked-position changes.

Account reads current mark price from Nuubot MarketData. It owns no BBO ingestion method or latest-BBO copy.

Normal recon skips a clean Account.

Forced recon ignores the flag.

Successful recon clears it. Failed recon restores it.

Venue and Ledger own no dirty flag.

## Credentials

Account selects one configured credential by Account name and network.

Selection rejects duplicates, missing names, network mismatch, empty address, and empty API key.

Semantic credential validation occurs only before live or testnet Venue initialization.

Simulator ignores the credentials catalog.

Credential values never enter formatted errors or logs.

## Snapshot

Successful reconciliation returns one immutable-by-contract Account snapshot.

The snapshot contains identity, observation time, exposure, equity, margins, PnL, fees, and domain counts.

Decision-critical Account state remains synchronous.

Balance and equity calculation cadence may become configurable only after proof that those values are observability-only.

Telemetry includes separate freshness timestamps for balance and equity.

It contains no Account, Ledger, Trade, Order, Fill, or Venue pointer.

See [AccountSnapshot](../concepts/account-snapshot.md).

`Telemetry()` returns explicit absence until one successful Account observation.

Observed telemetry returns the latest snapshot without mutation or Venue access.

## Recon Failure and Publication

Account increments its consecutive failure count exactly once for each failed Recon.

One successful executed Recon resets the count. A valid clean skip leaves it unchanged.

Account returns the count explicitly with the Snapshot, refresh fact, and error.

Account makes no Controller, BotCycle, Sweep, retry, skip, or stoppage decision.

BotCycle publishes no Snapshot barrier when any capable running Executor Recon fails.

Controller skips the remaining pass after failures one and two and returns an error at three.

Persistence or execution failures outside Account Recon remain immediately fatal.

Account validates normalized identities before Ledger mutation.

Maximum persistence writes dirty rows in one SQL transaction.

No complete graph clone or memory rollback exists.

A persistence failure leaves Ledger memory untrusted and publishes no successful Account Snapshot.

## Capacity

Account and Ledger use dynamic maps and per-attempt slices.

They reserve no fixed Trade, Order, Fill, or evidence-buffer capacity.

## Persistence

Account receives store operations only for `max`.

`none` opens no database during Account execution.

Account receives `persist_mode` and passes policy to Ledger and Simulator.

`none` performs no Ledger, Trade, Order, Fill, or Simulator database writes.

`max` persists every accepted Ledger mutation and every Simulator state change.

`max` proves independent Ledger and Simulator state reload.

Full Bot resume requires Runner, replay, Controller, Signaler, and
TradeExecutor cursor ownership.

TradeExecutor rejects persisted Trades until that recovery path exists.

Neither child detects Runner, Sweep, paper, or live mode.

For `none`, ResultPublisher owns the final per-Bot SQLite path.

Standalone live Runner persistence remains TBD.

Runner cannot require Server availability.

Account never opens the Server-owned PocketBase database directly.

See [Trading Schema](../concepts/trading-schema.md).

## Terminal Result

Before child teardown, Account creates one immutable terminal result.

`account.Result` contains identity, Venue kind, `persist_mode`, and `ledger.Result`.

It contains no Simulator result, pointer, private record, counter, or persistence payload.

Slices and maps are copied. The result aliases no mutable child state.

AccountSnapshot remains the small one-control-pass Risk value.

The terminal result travels upward without Account, Ledger, or Simulator pointers.

## Does Not

- Calculate Trade PnL.
- Match Simulator Orders.
- Mutate Order or Fill fields directly.
- Expose raw Venue payloads to Executor.
- Expose Simulator private state through Result.
- Share mutable Accounts between Executors.
- Log returned errors.
- Persist telemetry.

## Required Proof

- Partial initialization publishes no child.
- Invalid batches create no rows.
- Unknown submit outcomes retain `created` Orders.
- Mixed submit results preserve each item.
- Known Simulator failure terminalizes every created Order.
- Missing created or submitted Simulator Orders repair to `error` during recon.
- Failed recon changes no domain state or cursor.
- Simulator BBO changes only mark dirty.
- Stop releases Venue before Ledger.
- Credential values never appear in output.
