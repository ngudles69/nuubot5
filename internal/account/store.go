package account

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"

	"nuubot/internal/account/fill"
	"nuubot/internal/account/ledger"
	"nuubot/internal/account/order"
	"nuubot/internal/account/trade"
)

const accountStoreSchemaVersion = 1

type store struct {
	db *sql.DB
}

func (a *Account) persist(all bool) error {
	if a.config.PersistMode == "none" && !all {
		return nil
	}
	if a.store == nil {
		var err error
		a.store, err = openStore(a.config.Nuubot.RuntimePath)
		if err != nil {
			return err
		}
	}
	return a.store.save(a, a.ledger.StoreChanges(all))
}

func openStore(path string) (*store, error) {
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	var db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open Account Store: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(accountStoreDDL); err != nil {
		db.Close()
		return nil, fmt.Errorf("open Account Store: create schema: %v", err)
	}
	return &store{db: db}, nil
}

func (s *store) close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close Account Store: %v", err)
	}
	return nil
}

func (s *store) save(a *Account, state ledger.StoreState) error {
	var tx, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("persist Account: begin: %v", err)
	}
	defer tx.Rollback()
	var nowMS = time.Now().UnixMilli()
	err = saveAccount(tx, a, nowMS)
	if err == nil && state.LedgerDirty {
		err = saveLedger(tx, state, nowMS)
	}
	for _, current := range state.Trades {
		if err == nil {
			err = saveTrade(tx, current, nowMS)
		}
	}
	for _, current := range state.Orders {
		if err == nil {
			err = saveOrder(tx, current, nowMS)
		}
	}
	for _, current := range state.Fills {
		if err == nil {
			err = saveFill(tx, current, nowMS)
		}
	}
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("persist Account: commit: %v", err)
	}
	a.ledger.StoreCommitted(state)
	return nil
}

func saveAccount(tx *sql.Tx, a *Account, nowMS int64) error {
	var c = a.config
	var s = a.lastSnapshot
	_, err := tx.Exec(`
		INSERT INTO account VALUES (
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
		)
		ON CONFLICT (ledger_id) DO UPDATE SET
			snapshot_generation=excluded.snapshot_generation,
			snapshot_observed_ms=excluded.snapshot_observed_ms,
			account_value=excluded.account_value,
			withdrawable=excluded.withdrawable,
			position_quantity=excluded.position_quantity,
			entry_price=excluded.entry_price,
			realized_pnl=excluded.realized_pnl,
			unrealized_pnl=excluded.unrealized_pnl,
			gross_pnl=excluded.gross_pnl,
			fees=excluded.fees,
			net_pnl=excluded.net_pnl,
			open_trades=excluded.open_trades,
			active_orders=excluded.active_orders,
			fills=excluded.fills,
			pending_orders=excluded.pending_orders,
			pending_fills=excluded.pending_fills,
			recon_calls=excluded.recon_calls,
			recon_skipped_clean=excluded.recon_skipped_clean,
			recon_executed=excluded.recon_executed,
			recon_succeeded=excluded.recon_succeeded,
			recon_failed=excluded.recon_failed,
			last_recon_ms=excluded.last_recon_ms,
			failure_count=excluded.failure_count,
			dirty=excluded.dirty,
			updated_ms=excluded.updated_ms`,
		accountStoreSchemaVersion,
		c.SweepID, c.BotID, c.LedgerID, c.CycleNumber, c.ExecutorNumber,
		c.Name, c.Venue, c.Network, c.Symbol,
		c.EquityUSDC.String(), c.FeePct.String(), c.SlippagePct.String(), c.Recon,
		s.Generation, s.ObservedMS,
		s.AccountValue.String(), s.Withdrawable.String(),
		s.PositionQuantity.String(), s.EntryPrice.String(),
		s.RealizedPnL.String(), s.UnrealizedPnL.String(), s.GrossPnL.String(),
		s.Fees.String(), s.NetPnL.String(),
		s.OpenTrades, s.ActiveOrders, s.Fills, s.PendingOrders, s.PendingFills,
		a.reconStats.Calls, a.reconStats.SkippedClean, a.reconStats.Executed,
		a.reconStats.Succeeded, a.reconStats.Failed,
		a.lastReconMS, a.failureCount, a.dirty, nowMS, nowMS,
	)
	if err != nil {
		return fmt.Errorf("persist Account row: %v", err)
	}
	return nil
}

func saveLedger(tx *sql.Tx, s ledger.StoreState, nowMS int64) error {
	_, err := tx.Exec(`
		INSERT INTO ledger VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (ledger_id) DO UPDATE SET
			next_trade_id=excluded.next_trade_id,
			next_order_id=excluded.next_order_id,
			next_fill_id=excluded.next_fill_id,
			account_raw_json=excluded.account_raw_json,
			updated_ms=excluded.updated_ms`,
		accountStoreSchemaVersion,
		s.SweepID, s.BotID, s.ID, s.CycleNumber, s.ExecutorNumber,
		s.Venue, s.Network, s.Account, s.Symbol,
		s.NextTradeID, s.NextOrderID, s.NextFillID, s.AccountRawJSON,
		nowMS,
	)
	if err != nil {
		return fmt.Errorf("persist Ledger row: %v", err)
	}
	return nil
}

func saveTrade(tx *sql.Tx, t *trade.Trade, nowMS int64) error {
	_, err := tx.Exec(`
		INSERT INTO trade VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (ledger_id, trade_id) DO UPDATE SET
			status=excluded.status, side=excluded.side,
			open_quantity=excluded.open_quantity,
			average_entry_price=excluded.average_entry_price,
			realized_pnl=excluded.realized_pnl,
			unrealized_pnl=excluded.unrealized_pnl,
			gross_pnl=excluded.gross_pnl, fees=excluded.fees,
			net_pnl=excluded.net_pnl, opened_ms=excluded.opened_ms,
			closed_ms=excluded.closed_ms, updated_ms=excluded.updated_ms,
			stored_ms=excluded.stored_ms`,
		t.SweepID, t.BotID, t.Venue, t.Network, t.Account, t.LedgerID,
		t.TradeID, t.CycleNumber, t.Symbol, t.Status, t.Side,
		t.OpenQuantity.String(), t.AverageEntryPrice.String(),
		t.RealizedPnL.String(), t.UnrealizedPnL.String(), t.GrossPnL.String(),
		t.Fees.String(), t.NetPnL.String(), t.OpenedMS, t.ClosedMS, t.UpdatedMS,
		nowMS, nowMS,
	)
	if err != nil {
		return fmt.Errorf("persist Trade %d: %v", t.TradeID, err)
	}
	return nil
}

func saveOrder(tx *sql.Tx, o *order.Order, nowMS int64) error {
	_, err := tx.Exec(`
		INSERT INTO "order" VALUES (
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
		)
		ON CONFLICT (ledger_id, order_id) DO UPDATE SET
			venue_order_id=excluded.venue_order_id,
			venue_status=excluded.venue_status,
			status=excluded.status, reject_reason=excluded.reject_reason,
			updated_ms=excluded.updated_ms,
			filled_quantity=excluded.filled_quantity,
			filled_notional=excluded.filled_notional,
			average_fill_price=excluded.average_fill_price,
			remaining_quantity=excluded.remaining_quantity,
			fees=excluded.fees, fill_count=excluded.fill_count,
			pending_fee_count=excluded.pending_fee_count,
			last_fill_ms=excluded.last_fill_ms, raw_json=excluded.raw_json,
			stored_ms=excluded.stored_ms`,
		o.SweepID, o.BotID, o.Venue, o.Network, o.Account,
		o.LedgerID, o.TradeID, o.OrderID, o.CLOID, o.VenueOrderID,
		o.VenueStatus, o.CycleNumber, o.Symbol, o.Level, o.Role, o.Side,
		o.Type, o.TimeInForce, o.SubmittedQuantity.String(),
		nullableDecimal(o.SubmittedPrice), nullableDecimal(o.TriggerPrice),
		o.ReduceOnly, o.SubmittedMS, o.Status, o.RejectReason, o.UpdatedMS,
		o.FilledQuantity.String(), o.FilledNotional.String(),
		o.AverageFillPrice.String(), o.RemainingQuantity.String(),
		o.Fees.String(), o.FillCount, o.PendingFeeCount, o.LastFillMS,
		o.RawJSON, nowMS, nowMS,
	)
	if err != nil {
		return fmt.Errorf("persist Order %d: %v", o.OrderID, err)
	}
	return nil
}

func saveFill(tx *sql.Tx, f *fill.Fill, nowMS int64) error {
	_, err := tx.Exec(`
		INSERT INTO fill VALUES (
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
		)
		ON CONFLICT (ledger_id, fill_id) DO UPDATE SET
			fee=excluded.fee, liquidity=excluded.liquidity,
			raw_json=excluded.raw_json, updated_ms=excluded.updated_ms`,
		f.FillID, f.SweepID, f.BotID, f.Venue, f.Network, f.Account,
		f.LedgerID, f.TradeID, f.OrderID, f.CLOID, f.VenueOrderID, f.VenueTID,
		f.CycleNumber, f.Symbol, f.Side, f.Quantity.String(), f.Price.String(),
		f.TimestampMS, nullableDecimal(f.Fee), f.Liquidity, f.RawJSON,
		nowMS, nowMS,
	)
	if err != nil {
		return fmt.Errorf("persist Fill %d: %v", f.FillID, err)
	}
	return nil
}

func nullableDecimal(value *decimal.Decimal) any {
	if value == nil {
		return nil
	}
	return value.String()
}

const accountStoreDDL = `
CREATE TABLE IF NOT EXISTS account (
	schema_version INTEGER NOT NULL,
	sweep_id INTEGER NOT NULL,
	bot_id INTEGER NOT NULL,
	ledger_id INTEGER PRIMARY KEY,
	cycle_number INTEGER NOT NULL,
	executor_number INTEGER NOT NULL,
	name TEXT NOT NULL,
	venue TEXT NOT NULL,
	network TEXT NOT NULL,
	symbol TEXT NOT NULL,
	equity_usdc TEXT NOT NULL,
	fee_pct TEXT NOT NULL,
	slippage_pct TEXT NOT NULL,
	recon TEXT NOT NULL,
	snapshot_generation INTEGER NOT NULL,
	snapshot_observed_ms INTEGER NOT NULL,
	account_value TEXT NOT NULL,
	withdrawable TEXT NOT NULL,
	position_quantity TEXT NOT NULL,
	entry_price TEXT NOT NULL,
	realized_pnl TEXT NOT NULL,
	unrealized_pnl TEXT NOT NULL,
	gross_pnl TEXT NOT NULL,
	fees TEXT NOT NULL,
	net_pnl TEXT NOT NULL,
	open_trades INTEGER NOT NULL,
	active_orders INTEGER NOT NULL,
	fills INTEGER NOT NULL,
	pending_orders INTEGER NOT NULL,
	pending_fills INTEGER NOT NULL,
	recon_calls INTEGER NOT NULL,
	recon_skipped_clean INTEGER NOT NULL,
	recon_executed INTEGER NOT NULL,
	recon_succeeded INTEGER NOT NULL,
	recon_failed INTEGER NOT NULL,
	last_recon_ms INTEGER NOT NULL,
	failure_count INTEGER NOT NULL,
	dirty INTEGER NOT NULL,
	created_ms INTEGER NOT NULL,
	updated_ms INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS ledger (
	schema_version INTEGER NOT NULL,
	sweep_id INTEGER NOT NULL,
	bot_id INTEGER NOT NULL,
	ledger_id INTEGER PRIMARY KEY REFERENCES account(ledger_id),
	cycle_number INTEGER NOT NULL,
	executor_number INTEGER NOT NULL,
	venue TEXT NOT NULL,
	network TEXT NOT NULL,
	account TEXT NOT NULL,
	symbol TEXT NOT NULL,
	next_trade_id INTEGER NOT NULL,
	next_order_id INTEGER NOT NULL,
	next_fill_id INTEGER NOT NULL,
	account_raw_json TEXT NOT NULL,
	updated_ms INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS trade (
	sweep_id INTEGER NOT NULL,
	bot_id INTEGER NOT NULL,
	venue TEXT NOT NULL,
	network TEXT NOT NULL,
	account TEXT NOT NULL,
	ledger_id INTEGER NOT NULL REFERENCES ledger(ledger_id),
	trade_id INTEGER NOT NULL,
	cycle_number INTEGER NOT NULL,
	symbol TEXT NOT NULL,
	status TEXT NOT NULL,
	side TEXT NOT NULL,
	open_quantity TEXT NOT NULL,
	average_entry_price TEXT NOT NULL,
	realized_pnl TEXT NOT NULL,
	unrealized_pnl TEXT NOT NULL,
	gross_pnl TEXT NOT NULL,
	fees TEXT NOT NULL,
	net_pnl TEXT NOT NULL,
	opened_ms INTEGER NOT NULL,
	closed_ms INTEGER NOT NULL,
	updated_ms INTEGER NOT NULL,
	created_ms INTEGER NOT NULL,
	stored_ms INTEGER NOT NULL,
	PRIMARY KEY (ledger_id, trade_id)
);
CREATE TABLE IF NOT EXISTS "order" (
	sweep_id INTEGER NOT NULL,
	bot_id INTEGER NOT NULL,
	venue TEXT NOT NULL,
	network TEXT NOT NULL,
	account TEXT NOT NULL,
	ledger_id INTEGER NOT NULL,
	trade_id INTEGER NOT NULL,
	order_id INTEGER NOT NULL,
	cloid TEXT NOT NULL,
	venue_order_id INTEGER NOT NULL,
	venue_status TEXT NOT NULL,
	cycle_number INTEGER NOT NULL,
	symbol TEXT NOT NULL,
	level INTEGER NOT NULL,
	role TEXT NOT NULL,
	side TEXT NOT NULL,
	type TEXT NOT NULL,
	time_in_force TEXT NOT NULL,
	submitted_quantity TEXT NOT NULL,
	submitted_price TEXT,
	trigger_price TEXT,
	reduce_only INTEGER NOT NULL,
	submitted_ms INTEGER NOT NULL,
	status TEXT NOT NULL,
	reject_reason TEXT NOT NULL,
	updated_ms INTEGER NOT NULL,
	filled_quantity TEXT NOT NULL,
	filled_notional TEXT NOT NULL,
	average_fill_price TEXT NOT NULL,
	remaining_quantity TEXT NOT NULL,
	fees TEXT NOT NULL,
	fill_count INTEGER NOT NULL,
	pending_fee_count INTEGER NOT NULL,
	last_fill_ms INTEGER NOT NULL,
	raw_json TEXT NOT NULL,
	created_ms INTEGER NOT NULL,
	stored_ms INTEGER NOT NULL,
	PRIMARY KEY (ledger_id, order_id),
	UNIQUE (ledger_id, cloid),
	FOREIGN KEY (ledger_id, trade_id) REFERENCES trade(ledger_id, trade_id)
);
CREATE TABLE IF NOT EXISTS fill (
	fill_id INTEGER NOT NULL,
	sweep_id INTEGER NOT NULL,
	bot_id INTEGER NOT NULL,
	venue TEXT NOT NULL,
	network TEXT NOT NULL,
	account TEXT NOT NULL,
	ledger_id INTEGER NOT NULL,
	trade_id INTEGER NOT NULL,
	order_id INTEGER NOT NULL,
	cloid TEXT NOT NULL,
	venue_order_id INTEGER NOT NULL,
	venue_tid INTEGER NOT NULL,
	cycle_number INTEGER NOT NULL,
	symbol TEXT NOT NULL,
	side TEXT NOT NULL,
	quantity TEXT NOT NULL,
	price TEXT NOT NULL,
	timestamp_ms INTEGER NOT NULL,
	fee TEXT,
	liquidity TEXT NOT NULL,
	raw_json TEXT NOT NULL,
	created_ms INTEGER NOT NULL,
	updated_ms INTEGER NOT NULL,
	PRIMARY KEY (ledger_id, fill_id),
	UNIQUE (ledger_id, venue_tid),
	FOREIGN KEY (ledger_id, order_id) REFERENCES "order"(ledger_id, order_id)
);`
