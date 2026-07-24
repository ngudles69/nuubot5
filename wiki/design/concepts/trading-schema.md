# Trading Schema

Status: Proposed assessment.
Covers: No implemented source.
Purpose: Define the first per-Bot SQLite result schema for Account trading evidence and Simulator recovery.

## Scope

This DDL targets one BtRunner result database.

One `(sweep_id, bot_id)` worker owns the database writer.

The file belongs under:

```text
workspace/db/sweeps/sweep_<sweep_id>/bot_<bot_id>.db
```

Live, paper, and server persistence will use Server-owned PocketBase operations.

PocketBase adoption does not permit Runners to open its SQLite database directly.

The live physical migration remains separate.

## Tables

| Table | Owner | Purpose |
|---|---|---|
| `account_ledger` | Ledger | One Account reconciliation cursor and snapshot root |
| `account_trade` | Trade | One strategy trading intent |
| `account_order` | Order | One submitted Venue leg |
| `account_fill` | Fill | One immutable Venue execution |
| `simulator_state` | Simulator | One versioned simulated Venue snapshot |

Simulator state never replaces Ledger evidence.

Ledger rows never become Simulator exchange truth.

## Decimal Storage

Prices, quantities, fees, and PnL use canonical decimal text.

SQLite `REAL` is prohibited for trading values.

Go parses each admitted decimal once.

The runtime decimal implementation needs explicit dependency approval before coding.

## DDL

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS account_ledger (
    ledger_id           INTEGER PRIMARY KEY,
    cycle_no            INTEGER NOT NULL CHECK (cycle_no > 0),
    executor_no         INTEGER NOT NULL CHECK (executor_no > 0),
    account_name        TEXT NOT NULL CHECK (length(account_name) > 0),
    network             TEXT NOT NULL CHECK (network IN ('mainnet', 'testnet', 'simnet')),
    symbol              TEXT NOT NULL CHECK (length(symbol) > 0),
    fills_through_ms    INTEGER,
    last_recon_ms       INTEGER,
    account_state_json  TEXT NOT NULL DEFAULT '{}',
    created_ms          INTEGER NOT NULL CHECK (created_ms >= 0),
    updated_ms          INTEGER NOT NULL CHECK (updated_ms >= created_ms),
    UNIQUE (cycle_no, executor_no, account_name)
);

CREATE TABLE IF NOT EXISTS account_trade (
    trade_id            INTEGER PRIMARY KEY,
    ledger_id           INTEGER NOT NULL REFERENCES account_ledger(ledger_id),
    trade_no            INTEGER NOT NULL CHECK (trade_no BETWEEN 1 AND 2097151),
    symbol              TEXT NOT NULL CHECK (length(symbol) > 0),
    status              TEXT NOT NULL CHECK (
        status IN ('pending', 'open', 'closing', 'closed', 'canceled', 'error')
    ),
    side                TEXT NOT NULL CHECK (side IN ('long', 'short', 'flat')),
    open_qty            TEXT NOT NULL DEFAULT '0',
    avg_entry_price     TEXT,
    realized_pnl        TEXT NOT NULL DEFAULT '0',
    fees                TEXT NOT NULL DEFAULT '0',
    net_pnl             TEXT NOT NULL DEFAULT '0',
    opened_ms           INTEGER,
    closed_ms           INTEGER,
    updated_ms          INTEGER NOT NULL CHECK (updated_ms >= 0),
    UNIQUE (ledger_id, trade_no),
    UNIQUE (ledger_id, trade_id)
);

CREATE TABLE IF NOT EXISTS account_order (
    order_id            INTEGER PRIMARY KEY,
    ledger_id           INTEGER NOT NULL,
    trade_id            INTEGER NOT NULL,
    batch_no            INTEGER NOT NULL CHECK (batch_no BETWEEN 1 AND 1000),
    order_pos           INTEGER NOT NULL CHECK (order_pos BETWEEN 1 AND 1000),
    symbol              TEXT NOT NULL CHECK (length(symbol) > 0),
    cloid               TEXT NOT NULL UNIQUE CHECK (length(cloid) = 34),
    order_role          TEXT NOT NULL CHECK (
        order_role IN ('entry', 'tp', 'sl', 'exit', 'close', 'cleanup', 'stop')
    ),
    side                TEXT NOT NULL CHECK (side IN ('B', 'A')),
    order_type          TEXT NOT NULL CHECK (order_type IN ('limit', 'trigger', 'market')),
    time_in_force       TEXT CHECK (time_in_force IN ('Gtc', 'Ioc', 'Alo')),
    requested_qty       TEXT NOT NULL,
    requested_price     TEXT,
    trigger_price       TEXT,
    reduce_only         INTEGER NOT NULL CHECK (reduce_only IN (0, 1)),
    submitted_ms        INTEGER NOT NULL CHECK (submitted_ms >= 0),
    venue_order_id      TEXT,
    status              TEXT NOT NULL CHECK (
        status IN (
            'created',
            'submitted',
            'open',
            'partially_filled',
            'filled',
            'canceled',
            'rejected',
            'expired',
            'error'
        )
    ),
    active              INTEGER NOT NULL CHECK (active IN (0, 1)),
    reject_reason       TEXT,
    updated_ms          INTEGER,
    last_fill_ms        INTEGER,
    filled_qty          TEXT NOT NULL DEFAULT '0',
    remaining_qty       TEXT NOT NULL,
    avg_fill_price      TEXT,
    fees                TEXT NOT NULL DEFAULT '0',
    raw_json            TEXT NOT NULL DEFAULT '{}',
    UNIQUE (trade_id, batch_no, order_pos),
    UNIQUE (ledger_id, trade_id, order_id, cloid),
    FOREIGN KEY (ledger_id, trade_id)
        REFERENCES account_trade (ledger_id, trade_id)
);

CREATE TABLE IF NOT EXISTS account_fill (
    fill_id             INTEGER PRIMARY KEY,
    ledger_id           INTEGER NOT NULL,
    trade_id            INTEGER NOT NULL,
    order_id            INTEGER NOT NULL,
    venue_tid           TEXT NOT NULL CHECK (length(venue_tid) > 0),
    cloid               TEXT NOT NULL,
    venue_order_id      TEXT NOT NULL,
    symbol              TEXT NOT NULL CHECK (length(symbol) > 0),
    side                TEXT NOT NULL CHECK (side IN ('B', 'A')),
    qty                 TEXT NOT NULL,
    price               TEXT NOT NULL,
    event_ms            INTEGER NOT NULL CHECK (event_ms >= 0),
    fee                 TEXT,
    liquidity           TEXT,
    raw_json            TEXT NOT NULL DEFAULT '{}',
    UNIQUE (ledger_id, venue_tid),
    FOREIGN KEY (ledger_id, trade_id, order_id, cloid)
        REFERENCES account_order (ledger_id, trade_id, order_id, cloid)
);

CREATE TABLE IF NOT EXISTS simulator_state (
    ledger_id           INTEGER PRIMARY KEY REFERENCES account_ledger(ledger_id),
    schema_version      INTEGER NOT NULL CHECK (schema_version > 0),
    payload_json        TEXT NOT NULL,
    updated_ms          INTEGER NOT NULL CHECK (updated_ms >= 0)
);

CREATE INDEX IF NOT EXISTS account_trade_status_idx
    ON account_trade (ledger_id, status);

CREATE INDEX IF NOT EXISTS account_order_active_idx
    ON account_order (ledger_id, active);

CREATE INDEX IF NOT EXISTS account_order_trade_idx
    ON account_order (trade_id, order_id);

CREATE INDEX IF NOT EXISTS account_fill_order_idx
    ON account_fill (order_id, event_ms);

CREATE INDEX IF NOT EXISTS account_fill_cursor_idx
    ON account_fill (ledger_id, event_ms);
```

## Transaction Boundaries

One transaction creates a Trade and its initial Orders.

One transaction adds later Orders to an existing Trade.

One transaction applies one validated reconciliation batch.

One transaction updates confirmed submission outcomes.

External Venue calls occur outside transactions.

Unknown outcomes preserve committed `created` Orders.

The Fill cursor advances only inside the successful reconciliation transaction.

## Persistence Modes

| Mode | Ledger | Simulator | Intended use |
|---|---|---|---|
| `none` | One successful final export | Memory only | Sweeps |
| `max` | Every accepted mutation | Every state change | Restartable runs |

Account passes one configured mode to Ledger and Simulator.

Sweep runs select `none`.

They create result evidence only after successful completion.

Account opens no database for `none`.

ResultPublisher builds a temporary database and atomically renames it after one committed export.

Failed Sweep runs retain no recovery checkpoint and are rerun.

Restartable runs select `max` and force recon after loading state.

`max` requires measured performance proof.

## Invariants

- Foreign keys remain enabled on every connection.
- One Order belongs to one Trade and one Ledger.
- One Fill belongs to one Order, Trade, and Ledger.
- One CLOID identifies one Order.
- One Venue TID enters one Ledger once.
- Terminal Orders never return active.
- Terminal Trades never reopen.
- Simulator payload identity must match its Ledger.
- Corrupt Simulator state fails without replacement.
- Credentials never enter any table.

## Required Proof

- DDL applies twice without destructive migration.
- Foreign-key violations fail.
- Cross-Ledger Trade, Order, Fill, and CLOID ancestry fails.
- Duplicate CLOIDs and Venue TIDs fail.
- Failed reconciliation rolls back every row and cursor.
- Unknown submission outcomes retain `created` Orders.
- Reopening the result database reconstructs the same Ledger.
- Failed final publication leaves no completed result database.

## Executed Assessment Proof

The proposed DDL applied twice to one in-memory SQLite database.

SQLite rejected a cross-Ledger Order.

SQLite rejected a cross-Ledger Fill.

SQLite rejected a Fill with the wrong CLOID.
