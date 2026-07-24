# Account Package

Status: Implemented for Simulator-backed Accounts.
Covers: `internal/account/*.go`
Purpose: Give one Executor a trading boundary backed by one Venue and one local Ledger.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/account/account.py`
- `D:/rust/nuubot3/wiki/account/account.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/account.py`

## Ownership

TradeExecutor owns one Account.

Account owns one selected Venue and one Ledger.

The current BtRunner implementation selects Simulator only.

Account hides Venue selection and response translation from Executor.

## Construction

Account is one concrete component.

It does not need a factory.

TradeExecutor owns one Account value and calls `Init`.

`Init` validates identity before opening Ledger or Venue state.

Account publishes neither child after partial initialization failure.

## Inputs

| Input | Purpose |
|---|---|
| Logger | Report Account terminal statistics |
| Cycle and Executor numbers | Create stable result identity |
| Account config | Select network, name, capital, fees, and persistence |
| Selected credentials | Initialize a future live or testnet Venue |
| Current Clock | Preserve deterministic operation timestamps |
| Store operations | Required only when `persist_mode = max` |

Simulator initialization receives no private credential.

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
  validate complete order batch
  resolve Trade ownership
  create CLOIDs
  commit created Trade and Orders
  submit Venue batch
  terminalize known Simulator submission failure
  validate submit response
  commit submit outcomes
  mark Account dirty

CancelOrders
  validate owned active Orders
  cancel Venue batch
  validate cancel response
  mark Account dirty

IngestBBO
  ingest Venue BBO
  mark Account dirty when Venue changes

Recon
  claim dirty state
  read open Venue Orders
  read bounded Venue Fills
  read missing active Order statuses
  read Venue account state
  validate complete Venue evidence
  reconcile Ledger
  publish Account snapshot

Result
  get immutable Ledger result
  get immutable Simulator result
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

TP, SL, exit, close, cleanup, and stop Orders attach to an existing Trade.

One entry, TP, and SL bracket creates three Orders under one Trade.

Account persists `created` intent before Venue I/O.

Every request receives one explicit success or rejection.

One known Simulator submission failure maps every local Order to `error`.

Malformed or incomplete responses leave recoverable `created` evidence.

An explicit item error maps its Order to `rejected`.

Successful acknowledgements remain submitted until reconciliation.

Account never retries uncertain mutation outcomes automatically.

Immediate Fills still enter Ledger through reconciliation.

HTTP mutation responses are acknowledgement evidence, not final lifecycle truth.

## Reconciliation

Account queries Venue in this order:

1. Open Orders.
2. Fills from the inclusive Ledger cursor.
3. Exact status for missing active local Orders.
4. Transient account state.

Account validates untrusted Venue shapes once.

Ledger receives normalized concrete values.

Account filters history by the smallest useful time range.

A cap-sized response remains incomplete until Account narrows or continues the range.

Missing history rows never authorize deleting local Ledger evidence.

Reconciliation repairs drift from missing or delayed HTTP and WebSocket evidence.

Normal reconciliation returns the latest snapshot without querying a clean Account.

Forced reconciliation queries Venue even when Account is clean.

Failed reconciliation restores dirty state.

It advances no cursor or success timestamp.

## Dirty State

Account solely owns its reconciliation-dirty flag.

Initialization, user events, submissions, and changed Simulator truth mark it dirty.

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

It contains no Account, Ledger, Trade, Order, Fill, or Venue pointer.

See [AccountSnapshot](../concepts/account-snapshot.md).

## Persistence

Account receives store operations only for `max`.

`none` opens no database during Account execution.

Account receives `persist_mode` and passes it to Ledger and Simulator.

`none` keeps both children in memory until one successful final export.

`max` persists every accepted Ledger mutation and every Simulator state change.

`max` currently proves durable Ledger and Simulator child-state reload.

Full Bot resume requires Runner, replay, Runtime, Signaler, and TradeExecutor cursor ownership.

TradeExecutor rejects persisted Trades until that recovery path exists.

Neither child detects Runner, Sweep, paper, or live mode.

For `none`, ResultPublisher owns the final per-Bot SQLite path.

Live Runner later uses Server-owned store operations.

Account never opens the Server-owned PocketBase database directly.

See [Trading Schema](../concepts/trading-schema.md).

## Terminal Result

Before child teardown, Account creates one immutable terminal result.

`account.Result` contains identity, Venue kind, `persist_mode`, and `ledger.Result`.

Simulator-backed Accounts also contain one explicit optional `simulator.Result`.

Live and testnet Hyperliquid Accounts leave Simulator evidence absent.

They never fabricate a zero Simulator result.

Slices and maps are copied. The result aliases no mutable child state.

AccountSnapshot remains the small one-control-pass Risk value.

The terminal result travels upward without Account, Ledger, or Simulator pointers.

## Does Not

- Calculate Trade PnL.
- Match Simulator Orders.
- Mutate Order or Fill fields directly.
- Expose raw Venue payloads to Executor.
- Share mutable Accounts between Executors.
- Log returned errors.

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
