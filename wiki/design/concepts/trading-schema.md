# Trading Schema

Status: Implemented.
Covers: `internal/account/ledger/store.go`, `internal/simulator/store.go`, and `internal/resultpublisher`
Purpose: Define per-Bot Account evidence and durable Simulator child state.

## Scope

One `(sweep_id, bot_id)` worker owns:

```text
workspace/db/bots/bot_<bot_id>.db
```

The shared `workspace/db/nuubot.db` contains Sweeps, Bots, and mainnet Meta.

High-volume Trade, Order, Fill, and Simulator rows never enter the shared database.

## Decimal Storage

Prices, quantities, fees, and PnL use canonical decimal text.

SQLite `REAL` is prohibited for trading values.

Go uses `shopspring/decimal`.

## DDL

The source constant is canonical.

```sql
CREATE TABLE IF NOT EXISTS account_ledger (
    ledger_id           INTEGER PRIMARY KEY,
    cycle_no            INTEGER NOT NULL,
    executor_no         INTEGER NOT NULL,
    account_name        TEXT NOT NULL,
    network             TEXT NOT NULL,
    symbol              TEXT NOT NULL,
    next_trade_id       INTEGER NOT NULL,
    next_trade_no       INTEGER NOT NULL,
    next_order_id       INTEGER NOT NULL,
    fills_through_ms    INTEGER,
    last_recon_ms       INTEGER,
    account_state_json  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS account_trade (
    ledger_id         INTEGER NOT NULL REFERENCES account_ledger(ledger_id),
    trade_id          INTEGER NOT NULL,
    trade_no          INTEGER NOT NULL,
    symbol            TEXT NOT NULL,
    status            TEXT NOT NULL,
    side              TEXT NOT NULL,
    open_qty          TEXT NOT NULL,
    avg_entry_price   TEXT,
    realized_pnl      TEXT NOT NULL,
    fees              TEXT NOT NULL,
    net_pnl           TEXT NOT NULL,
    opened_ms         INTEGER,
    closed_ms         INTEGER,
    updated_ms        INTEGER NOT NULL,
    PRIMARY KEY (ledger_id, trade_id),
    UNIQUE (ledger_id, trade_no)
);

CREATE TABLE IF NOT EXISTS account_order (
    ledger_id          INTEGER NOT NULL,
    trade_id           INTEGER NOT NULL,
    order_id           INTEGER NOT NULL,
    account_name       TEXT NOT NULL,
    cycle_no           INTEGER NOT NULL,
    batch_no           INTEGER NOT NULL,
    order_pos          INTEGER NOT NULL,
    symbol             TEXT NOT NULL,
    cloid              TEXT NOT NULL UNIQUE,
    order_role         TEXT NOT NULL,
    side               TEXT NOT NULL,
    order_type         TEXT NOT NULL,
    time_in_force      TEXT NOT NULL,
    requested_qty      TEXT NOT NULL,
    requested_price    TEXT,
    trigger_price      TEXT,
    reduce_only        INTEGER NOT NULL,
    submitted_ms       INTEGER NOT NULL,
    venue_order_id     INTEGER,
    status             TEXT NOT NULL,
    active             INTEGER NOT NULL,
    reject_reason      TEXT,
    updated_ms         INTEGER,
    last_fill_ms       INTEGER,
    filled_qty         TEXT NOT NULL,
    remaining_qty      TEXT NOT NULL,
    avg_fill_price     TEXT,
    fees               TEXT NOT NULL,
    raw_json           TEXT NOT NULL,
    PRIMARY KEY (ledger_id, order_id),
    UNIQUE (ledger_id, trade_id, batch_no, order_pos),
    UNIQUE (ledger_id, trade_id, order_id, cloid),
    FOREIGN KEY (ledger_id, trade_id)
        REFERENCES account_trade (ledger_id, trade_id)
);

CREATE TABLE IF NOT EXISTS account_fill (
    ledger_id         INTEGER NOT NULL,
    trade_id          INTEGER NOT NULL,
    order_id          INTEGER NOT NULL,
    venue_tid         INTEGER NOT NULL,
    cloid             TEXT NOT NULL,
    venue_order_id    INTEGER NOT NULL,
    account_name      TEXT NOT NULL,
    cycle_no          INTEGER NOT NULL,
    symbol            TEXT NOT NULL,
    side              TEXT NOT NULL,
    qty               TEXT NOT NULL,
    price             TEXT NOT NULL,
    event_ms          INTEGER NOT NULL,
    fee               TEXT,
    liquidity         TEXT,
    raw_json          TEXT NOT NULL,
    PRIMARY KEY (ledger_id, venue_tid),
    FOREIGN KEY (ledger_id, trade_id, order_id, cloid)
        REFERENCES account_order (ledger_id, trade_id, order_id, cloid)
);

CREATE TABLE IF NOT EXISTS simulator_venue_state (
    account_name   TEXT NOT NULL,
    symbol         TEXT NOT NULL,
    schema_version INTEGER NOT NULL,
    payload_json   TEXT NOT NULL,
    updated_ms     INTEGER NOT NULL,
    PRIMARY KEY (account_name, symbol)
);
```

Every SQLite connection enables foreign keys and a 30-second busy timeout.

`simulator_venue_state` is Simulator-owned schema version 3.

Its payload stores official identity, policy, Venue counters, canonical Orders, and canonical Fills.

Each Simulator Order and Fill appears once.

It stores no Ledger ID, Trade ID, local Order ID, role, or purpose.

Legacy `simulator_state` payloads are not read or adapted.

## Grid Result Evidence

`grid_level_result` belongs to one terminal Executor result.

Its primary key is `(cycle_number, executor_id, level)`.

It stores:

- boundary, Grid, initial-entry, re-entry, and exit prices;
- quantity, notionals, commissions, and expected PnL;
- intended action;
- current Trade ID, number, and status;
- Level status and initial-submission completion;
- submission attempts; and
- last submitted and completed timestamps.

Grid Level identity is not duplicated into persisted Order `order_pos`.

CLOID `order_level` provides Order-to-Level identity.

## Persistence Modes

| Mode | Ledger | Simulator | Use |
|---|---|---|---|
| `none` | Memory, then terminal export | Memory, then terminal export | Sweeps |
| `max` | Every accepted mutation | Every changed state | Durable child-state reload |

`none` writes a complete `.partial` database only after successful Controller
shutdown.

Closing every writer precedes the final rename.

`max` reloads Ledger by Ledger identity.

Simulator reloads independently by official simulated account and symbol.

It does not resume replay, Controller, Signaler, or TradeExecutor policy
cursors.

Transient BBO state is never restored.

## Transaction Rules

- One Ledger Recon save writes changed rows in one transaction.
- Reconciliation validates one complete attempt before Snapshot publication.
- The Fill cursor advances only with the accepted batch.
- Venue calls never occur inside Ledger transactions.
- Simulator state uses one versioned JSON row per account and symbol.
- Simulator persistence succeeds before changed canonical memory becomes visible.

## Invariants

- One Order belongs to one Trade and Ledger.
- One Fill belongs to one Order, Trade, Ledger, and CLOID.
- One CLOID identifies one Order.
- One Venue TID enters one Ledger once.
- Missing bounded history deletes nothing.
- Terminal Orders never reopen.
- Terminal Trades never reopen.
- Simulator payload identity matches its official account, asset, symbol, and policy.
- Simulator Orders and Fills contain no Account domain identity.
- Simulator terminal Orders never re-enter its active index.
- Credentials enter no table.

## Proof

- Memory and `max` Ledger paths pass round-trip tests.
- Simulator and Account recover unreconciled Venue state.
- Cross-Trade Fill insertion fails by foreign key.
- Sweep `9`, Bot `13` publishes 50 Ledgers, 50 Trades, 151 Orders, and 100 Fills.
- SQLite integrity and foreign-key checks pass.
- Repeated successful publication replaces the completed result.
- No `.partial` file remains after success.
