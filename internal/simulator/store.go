package simulator

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const simulatorSchemaVersion = 1

type simulatorStore struct{ db *sql.DB }

type storedOrder struct {
	VenueOrderID      uint64
	CLOID             string
	Asset             int
	Symbol            string
	BatchID           uint64
	Kind              string
	IsBuy             bool
	Price             string
	Quantity          string
	ReduceOnly        bool
	TimeInForce       string
	TriggerPrice      string
	HasTriggerPrice   bool
	Status            string
	Armed             bool
	RemainingQuantity string
	FilledQuantity    string
	AverageFillPrice  string
	Fees              string
	TimestampMS       uint64
}

type storedFill struct {
	VenueOrderID  uint64
	VenueTID      uint64
	Symbol        string
	IsBuy         bool
	Quantity      string
	Price         string
	TimestampMS   uint64
	StartPosition string
	ClosedPnL     string
	Direction     string
	Fee           string
	HasFee        bool
	Liquidity     string
}

type storedState struct {
	SchemaVersion    int
	Account          string
	Asset            int
	Symbol           string
	Equity           string
	FeePct           string
	SlippagePct      string
	NextVenueOrderID uint64
	NextVenueTID     uint64
	NextBatchID      uint64
	ObservedMS       uint64
	Orders           []storedOrder
	Fills            []storedFill
}

func openSimulatorStore(path string) (*simulatorStore, error) {
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	var db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open Simulator Store: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(simulatorStoreDDL); err != nil {
		db.Close()
		return nil, fmt.Errorf("open Simulator Store: create schema: %v", err)
	}
	return &simulatorStore{db: db}, nil
}

func (s *simulatorStore) close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close Simulator Store: %v", err)
	}
	return nil
}

func (s *simulatorStore) save(cfg Config, state storedState) error {
	var tx, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("persist Simulator: begin: %v", err)
	}
	defer tx.Rollback()
	var nowMS = time.Now().UnixMilli()
	_, err = tx.Exec(`
		INSERT INTO simulator VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (account, symbol) DO UPDATE SET
			next_venue_order_id=excluded.next_venue_order_id,
			next_venue_tid=excluded.next_venue_tid,
			next_batch_id=excluded.next_batch_id,
			observed_ms=excluded.observed_ms,
			updated_ms=excluded.updated_ms`,
		simulatorSchemaVersion, cfg.Account, cfg.Asset, cfg.Symbol,
		cfg.Equity.String(), cfg.FeePct.String(), cfg.SlippagePct.String(),
		state.NextVenueOrderID, state.NextVenueTID, state.NextBatchID,
		state.ObservedMS, nowMS, nowMS,
	)
	for _, o := range state.Orders {
		if err != nil {
			break
		}
		_, err = tx.Exec(`
			INSERT INTO simulator_order VALUES (
				?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
			)
			ON CONFLICT (account, symbol, venue_order_id) DO UPDATE SET
				status=excluded.status, armed=excluded.armed,
				remaining_quantity=excluded.remaining_quantity,
				filled_quantity=excluded.filled_quantity,
				average_fill_price=excluded.average_fill_price,
				fees=excluded.fees, updated_ms=excluded.updated_ms`,
			cfg.Account, cfg.Symbol, o.VenueOrderID, o.CLOID, o.Asset,
			o.BatchID, o.Kind, o.IsBuy, o.Price, o.Quantity, o.ReduceOnly,
			o.TimeInForce, o.TriggerPrice, o.HasTriggerPrice, o.Status, o.Armed,
			o.RemainingQuantity, o.FilledQuantity, o.AverageFillPrice, o.Fees,
			o.TimestampMS, nowMS, nowMS, simulatorSchemaVersion,
		)
	}
	for _, f := range state.Fills {
		if err != nil {
			break
		}
		_, err = tx.Exec(`
			INSERT INTO simulator_fill VALUES (
				?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
			)
			ON CONFLICT (account, symbol, venue_tid) DO UPDATE SET
				fee=excluded.fee, has_fee=excluded.has_fee,
				liquidity=excluded.liquidity, updated_ms=excluded.updated_ms`,
			cfg.Account, cfg.Symbol, f.VenueOrderID, f.VenueTID, f.IsBuy,
			f.Quantity, f.Price, f.TimestampMS, f.StartPosition, f.ClosedPnL,
			f.Direction, f.Fee, f.HasFee, f.Liquidity,
			nowMS, nowMS, simulatorSchemaVersion,
		)
	}
	if err != nil {
		return fmt.Errorf("persist Simulator rows: %v", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("persist Simulator: commit: %v", err)
	}
	return nil
}

func (s *simulatorStore) load(cfg Config) (storedState, bool, error) {
	var state = storedState{SchemaVersion: simulatorSchemaVersion}
	var version int
	var err = s.db.QueryRow(`
		SELECT schema_version, account, asset, symbol, equity, fee_pct,
			slippage_pct, next_venue_order_id, next_venue_tid, next_batch_id,
			observed_ms
		FROM simulator WHERE account=? AND symbol=?`,
		cfg.Account, cfg.Symbol,
	).Scan(&version, &state.Account, &state.Asset, &state.Symbol,
		&state.Equity, &state.FeePct, &state.SlippagePct,
		&state.NextVenueOrderID, &state.NextVenueTID, &state.NextBatchID,
		&state.ObservedMS)
	if err == sql.ErrNoRows {
		return storedState{}, false, nil
	}
	if err != nil {
		return storedState{}, false, fmt.Errorf("load Simulator: %v", err)
	}
	if version != simulatorSchemaVersion || state.Asset != cfg.Asset ||
		state.Equity != cfg.Equity.String() || state.FeePct != cfg.FeePct.String() ||
		state.SlippagePct != cfg.SlippagePct.String() {
		return storedState{}, false, fmt.Errorf("load Simulator: identity mismatch")
	}
	var rows *sql.Rows
	rows, err = s.db.Query(`
		SELECT venue_order_id,cloid,asset,symbol,batch_id,kind,is_buy,price,
			quantity,reduce_only,time_in_force,trigger_price,has_trigger_price,
			status,armed,remaining_quantity,filled_quantity,average_fill_price,
			fees,timestamp_ms
		FROM simulator_order WHERE account=? AND symbol=?`, cfg.Account, cfg.Symbol)
	if err != nil {
		return storedState{}, false, err
	}
	for rows.Next() {
		var o storedOrder
		if err = rows.Scan(&o.VenueOrderID, &o.CLOID, &o.Asset, &o.Symbol, &o.BatchID,
			&o.Kind, &o.IsBuy, &o.Price, &o.Quantity, &o.ReduceOnly, &o.TimeInForce,
			&o.TriggerPrice, &o.HasTriggerPrice, &o.Status, &o.Armed,
			&o.RemainingQuantity, &o.FilledQuantity, &o.AverageFillPrice, &o.Fees,
			&o.TimestampMS); err != nil {
			rows.Close()
			return storedState{}, false, err
		}
		state.Orders = append(state.Orders, o)
	}
	if err = rows.Close(); err != nil {
		return storedState{}, false, err
	}
	rows, err = s.db.Query(`
		SELECT venue_order_id,venue_tid,symbol,is_buy,quantity,price,timestamp_ms,
			start_position,closed_pnl,direction,fee,has_fee,liquidity
		FROM simulator_fill WHERE account=? AND symbol=?`, cfg.Account, cfg.Symbol)
	if err != nil {
		return storedState{}, false, err
	}
	for rows.Next() {
		var f storedFill
		if err = rows.Scan(&f.VenueOrderID, &f.VenueTID, &f.Symbol, &f.IsBuy,
			&f.Quantity, &f.Price, &f.TimestampMS, &f.StartPosition, &f.ClosedPnL,
			&f.Direction, &f.Fee, &f.HasFee, &f.Liquidity); err != nil {
			rows.Close()
			return storedState{}, false, err
		}
		state.Fills = append(state.Fills, f)
	}
	if err = rows.Close(); err != nil {
		return storedState{}, false, err
	}
	return state, true, nil
}

const simulatorStoreDDL = `
CREATE TABLE IF NOT EXISTS simulator (
	schema_version INTEGER NOT NULL, account TEXT NOT NULL, asset INTEGER NOT NULL,
	symbol TEXT NOT NULL, equity TEXT NOT NULL, fee_pct TEXT NOT NULL,
	slippage_pct TEXT NOT NULL, next_venue_order_id INTEGER NOT NULL,
	next_venue_tid INTEGER NOT NULL, next_batch_id INTEGER NOT NULL,
	observed_ms INTEGER NOT NULL, created_ms INTEGER NOT NULL,
	updated_ms INTEGER NOT NULL, PRIMARY KEY (account, symbol)
);
CREATE TABLE IF NOT EXISTS simulator_order (
	account TEXT NOT NULL, symbol TEXT NOT NULL, venue_order_id INTEGER NOT NULL,
	cloid TEXT NOT NULL, asset INTEGER NOT NULL, batch_id INTEGER NOT NULL,
	kind TEXT NOT NULL, is_buy INTEGER NOT NULL, price TEXT NOT NULL,
	quantity TEXT NOT NULL, reduce_only INTEGER NOT NULL, time_in_force TEXT NOT NULL,
	trigger_price TEXT NOT NULL, has_trigger_price INTEGER NOT NULL,
	status TEXT NOT NULL, armed INTEGER NOT NULL, remaining_quantity TEXT NOT NULL,
	filled_quantity TEXT NOT NULL, average_fill_price TEXT NOT NULL, fees TEXT NOT NULL,
	timestamp_ms INTEGER NOT NULL, created_ms INTEGER NOT NULL, updated_ms INTEGER NOT NULL,
	schema_version INTEGER NOT NULL,
	PRIMARY KEY (account, symbol, venue_order_id), UNIQUE (account, symbol, cloid),
	FOREIGN KEY (account, symbol) REFERENCES simulator(account, symbol)
);
CREATE TABLE IF NOT EXISTS simulator_fill (
	account TEXT NOT NULL, symbol TEXT NOT NULL, venue_order_id INTEGER NOT NULL,
	venue_tid INTEGER NOT NULL, is_buy INTEGER NOT NULL, quantity TEXT NOT NULL,
	price TEXT NOT NULL, timestamp_ms INTEGER NOT NULL, start_position TEXT NOT NULL,
	closed_pnl TEXT NOT NULL, direction TEXT NOT NULL, fee TEXT NOT NULL,
	has_fee INTEGER NOT NULL, liquidity TEXT NOT NULL, created_ms INTEGER NOT NULL,
	updated_ms INTEGER NOT NULL, schema_version INTEGER NOT NULL,
	PRIMARY KEY (account, symbol, venue_tid),
	FOREIGN KEY (account, symbol, venue_order_id)
		REFERENCES simulator_order(account, symbol, venue_order_id)
);`
