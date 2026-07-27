package ledger

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"

	"nuubot/internal/fill"
	"nuubot/internal/order"
	"nuubot/internal/trade"
)

type ledgerStore struct {
	db *sql.DB
}

// Section 1 - Program Flow

func openLedgerStore(path string) (*ledgerStore, error) {
	// open Ledger store
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	var db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open Ledger store: %v", err)
	}
	db.SetMaxOpenConns(1)

	// prepare Ledger tables
	var _, prepareErr = db.Exec(ledgerDDL)
	if prepareErr != nil {
		db.Close()
		return nil, fmt.Errorf("open Ledger store: prepare tables: %v", prepareErr)
	}
	return &ledgerStore{db: db}, nil
}

func (s *ledgerStore) close() error {
	// close Ledger store
	var err = s.db.Close()
	if err != nil {
		return fmt.Errorf("close Ledger store: %v", err)
	}
	return nil
}

// Section 2 - Domain Helpers

func (s *ledgerStore) save(cfg Config, state candidate) error {
	// stage persistence transaction
	var transaction, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("persist Ledger: begin transaction: %v", err)
	}
	defer transaction.Rollback()

	// replace complete Ledger evidence
	err = storeLedgerIdentity(transaction, cfg, state, "persist Ledger")
	if err != nil {
		return err
	}
	for _, table := range []string{"account_fill", "account_order", "account_trade"} {
		_, err = transaction.Exec("DELETE FROM "+table+" WHERE ledger_id = ?", cfg.ID)
		if err != nil {
			return fmt.Errorf("persist Ledger: clear %s: %v", table, err)
		}
	}
	var tradeIDs = sortedTradeIDs(state.trades)
	for _, tradeID := range tradeIDs {
		var ownedTrade = state.trades[tradeID]
		err = storeReconTrade(transaction, cfg.ID, ownedTrade.Record())
		if err != nil {
			return err
		}
		err = ownedTrade.EachOrder(func(ownedOrder *order.Order) error {
			var orderRecord = ownedOrder.Record()
			var writeErr = storeReconOrder(transaction, cfg.ID, orderRecord)
			if writeErr != nil {
				return writeErr
			}
			return ownedOrder.EachFill(func(execution *fill.Fill) error {
				return storeReconFill(transaction, cfg.ID, execution.State())
			})
		})
		if err != nil {
			return err
		}
	}

	// commit persistence transaction
	err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("persist Ledger: commit transaction: %v", err)
	}
	return nil
}

func (s *ledgerStore) saveMutation(
	cfg Config,
	state candidate,
	trades []*trade.Trade,
	orders []*order.Order,
) error {
	// persist only touched rows
	// SQLite owns database rollback; failed memory mutations recover from Venue truth.
	var transaction, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("persist Ledger mutation: begin transaction: %v", err)
	}
	defer transaction.Rollback()

	err = storeLedgerIdentity(transaction, cfg, state, "persist Ledger mutation")
	if err != nil {
		return err
	}
	for _, owned := range trades {
		err = storeReconTrade(transaction, cfg.ID, owned.Record())
		if err != nil {
			return err
		}
	}
	for _, owned := range orders {
		err = storeReconOrder(transaction, cfg.ID, owned.Record())
		if err != nil {
			return err
		}
	}
	err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("persist Ledger mutation: commit transaction: %v", err)
	}
	return nil
}

func (s *ledgerStore) saveRecon(cfg Config, attempt *ReconAttempt) error {
	var transaction, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("persist Ledger recon: begin transaction: %v", err)
	}
	defer transaction.Rollback()

	var accountState = attempt.input.AccountStateRaw
	if accountState == "" {
		accountState = "{}"
	}
	_, err = transaction.Exec(`
		UPDATE account_ledger
		SET fills_through_ms = ?, last_recon_ms = ?, account_state_json = ?
		WHERE ledger_id = ?`,
		nullableUint(attempt.input.FillsThroughMS),
		nullableUint(attempt.input.ObservedMS),
		accountState,
		cfg.ID,
	)
	if err != nil {
		return fmt.Errorf("persist Ledger recon: store cursor: %v", err)
	}
	var tradeIDs = sortedTradeIDs(attempt.trades)
	for _, tradeID := range tradeIDs {
		var ownedTrade = attempt.trades[tradeID]
		if _, changed := attempt.touchedTrades[tradeID]; changed {
			err = storeReconTrade(transaction, cfg.ID, ownedTrade.Record())
			if err != nil {
				return err
			}
		}
		err = ownedTrade.EachOrder(func(ownedOrder *order.Order) error {
			var orderRecord = ownedOrder.Record()
			if _, changed := attempt.touchedOrders[orderRecord.OrderID]; changed {
				var writeErr = storeReconOrder(transaction, cfg.ID, orderRecord)
				if writeErr != nil {
					return writeErr
				}
			}
			return ownedOrder.EachFill(func(execution *fill.Fill) error {
				var fillRecord = execution.State()
				if _, changed := attempt.touchedFills[fillRecord.VenueTID]; changed {
					return storeReconFill(transaction, cfg.ID, fillRecord)
				}
				return nil
			})
		})
		if err != nil {
			return err
		}
	}
	err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("persist Ledger recon: commit transaction: %v", err)
	}
	return nil
}

func storeLedgerIdentity(
	transaction *sql.Tx,
	cfg Config,
	state candidate,
	operation string,
) error {
	var accountState = state.accountStateRaw
	if accountState == "" {
		accountState = "{}"
	}
	var _, err = transaction.Exec(`
		INSERT INTO account_ledger (
			ledger_id, cycle_no, executor_no, account_name, network, symbol,
			next_trade_id, next_trade_no, next_order_id, fills_through_ms,
			last_recon_ms, account_state_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (ledger_id) DO UPDATE SET
			cycle_no = excluded.cycle_no,
			executor_no = excluded.executor_no,
			account_name = excluded.account_name,
			network = excluded.network,
			symbol = excluded.symbol,
			next_trade_id = excluded.next_trade_id,
			next_trade_no = excluded.next_trade_no,
			next_order_id = excluded.next_order_id,
			fills_through_ms = excluded.fills_through_ms,
			last_recon_ms = excluded.last_recon_ms,
			account_state_json = excluded.account_state_json`,
		cfg.ID,
		cfg.CycleNumber,
		cfg.ExecutorNumber,
		cfg.Account,
		cfg.Network,
		cfg.Symbol,
		state.nextTradeID,
		state.nextTradeNo,
		state.nextOrderID,
		nullableUint(state.fillsThroughMS),
		nullableUint(state.lastReconMS),
		accountState,
	)
	if err != nil {
		return fmt.Errorf("%s: store identity: %v", operation, err)
	}
	return nil
}

func storeReconTrade(transaction *sql.Tx, ledgerID uint64, owned trade.Record) error {
	var _, err = transaction.Exec(`
		INSERT INTO account_trade (
			ledger_id, trade_id, trade_no, symbol, status, side, open_qty,
			avg_entry_price, realized_pnl, fees, net_pnl, opened_ms,
			closed_ms, updated_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (ledger_id, trade_id) DO UPDATE SET
			status = excluded.status,
			side = excluded.side,
			open_qty = excluded.open_qty,
			avg_entry_price = excluded.avg_entry_price,
			realized_pnl = excluded.realized_pnl,
			fees = excluded.fees,
			net_pnl = excluded.net_pnl,
			opened_ms = excluded.opened_ms,
			closed_ms = excluded.closed_ms,
			updated_ms = excluded.updated_ms`,
		ledgerID,
		owned.TradeID,
		owned.TradeNo,
		owned.Symbol,
		owned.Status,
		owned.Side,
		owned.OpenQuantity.String(),
		nullableDecimal(owned.AverageEntryPrice, owned.HasAveragePrice),
		owned.RealizedPnL.String(),
		owned.Fees.String(),
		owned.NetPnL.String(),
		nullableUint(owned.OpenedMS),
		nullableUint(owned.ClosedMS),
		owned.UpdatedMS,
	)
	if err != nil {
		return fmt.Errorf("persist Ledger recon: store Trade %d: %v", owned.TradeID, err)
	}
	return nil
}

func storeReconOrder(transaction *sql.Tx, ledgerID uint64, owned order.Record) error {
	var raw = owned.Raw
	if raw == "" {
		raw = "{}"
	}
	var _, err = transaction.Exec(`
		INSERT INTO account_order (
			ledger_id, trade_id, order_id, account_name, cycle_no,
			batch_no, order_pos, symbol, cloid, order_role, side,
			order_type, time_in_force, requested_qty, requested_price,
			trigger_price, reduce_only, submitted_ms, venue_order_id,
			status, active, reject_reason, updated_ms, last_fill_ms,
			filled_qty, remaining_qty, avg_fill_price, fees, raw_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		          ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (ledger_id, order_id) DO UPDATE SET
			venue_order_id = excluded.venue_order_id,
			status = excluded.status,
			active = excluded.active,
			reject_reason = excluded.reject_reason,
			updated_ms = excluded.updated_ms,
			last_fill_ms = excluded.last_fill_ms,
			filled_qty = excluded.filled_qty,
			remaining_qty = excluded.remaining_qty,
			avg_fill_price = excluded.avg_fill_price,
			fees = excluded.fees,
			raw_json = excluded.raw_json`,
		ledgerID,
		owned.TradeID,
		owned.OrderID,
		owned.Account,
		owned.CycleNumber,
		owned.BatchNo,
		owned.OrderPos,
		owned.Symbol,
		owned.CLOID,
		owned.Role,
		owned.Side,
		owned.Type,
		owned.TimeInForce,
		owned.RequestedQuantity.String(),
		nullableDecimalPointer(owned.RequestedPrice),
		nullableDecimalPointer(owned.TriggerPrice),
		owned.ReduceOnly,
		owned.TimestampMS,
		nullableUint(owned.VenueOrderID),
		owned.Status,
		owned.Active,
		nullableText(owned.RejectReason),
		nullableUint(owned.UpdatedMS),
		nullableUint(owned.LastFillMS),
		owned.FilledQuantity.String(),
		owned.RemainingQuantity.String(),
		nullableDecimal(owned.AverageFillPrice, owned.HasAveragePrice),
		owned.Fees.String(),
		raw,
	)
	if err != nil {
		return fmt.Errorf("persist Ledger recon: store Order %d: %v", owned.OrderID, err)
	}
	return nil
}

func storeReconFill(transaction *sql.Tx, ledgerID uint64, execution fill.Record) error {
	var raw = execution.Raw
	if raw == "" {
		raw = "{}"
	}
	var _, err = transaction.Exec(`
		INSERT INTO account_fill (
			ledger_id, trade_id, order_id, venue_tid, cloid,
			venue_order_id, account_name, cycle_no, symbol, side,
			qty, price, event_ms, fee, liquidity, raw_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (ledger_id, venue_tid) DO UPDATE SET
			fee = excluded.fee,
			liquidity = excluded.liquidity,
			raw_json = excluded.raw_json`,
		ledgerID,
		execution.TradeID,
		execution.OrderID,
		execution.VenueTID,
		execution.CLOID,
		execution.VenueOrderID,
		execution.Account,
		execution.CycleNumber,
		execution.Symbol,
		execution.Side,
		execution.Quantity.String(),
		execution.Price.String(),
		execution.TimestampMS,
		nullableDecimal(execution.Fee, execution.HasFee),
		nullableText(execution.Liquidity),
		raw,
	)
	if err != nil {
		return fmt.Errorf("persist Ledger recon: store Fill %d: %v", execution.VenueTID, err)
	}
	return nil
}

func (s *ledgerStore) load(cfg Config) (candidate, bool, error) {
	// load Ledger identity
	var state candidate
	var cycleNumber int
	var executorNumber int
	var account string
	var network string
	var symbol string
	var fillsThrough sql.NullInt64
	var lastRecon sql.NullInt64
	var err = s.db.QueryRow(`
		SELECT cycle_no, executor_no, account_name, network, symbol,
		       next_trade_id, next_trade_no, next_order_id, fills_through_ms,
		       last_recon_ms, account_state_json
		FROM account_ledger
		WHERE ledger_id = ?`,
		cfg.ID,
	).Scan(
		&cycleNumber,
		&executorNumber,
		&account,
		&network,
		&symbol,
		&state.nextTradeID,
		&state.nextTradeNo,
		&state.nextOrderID,
		&fillsThrough,
		&lastRecon,
		&state.accountStateRaw,
	)
	if err == sql.ErrNoRows {
		return candidate{}, false, nil
	}
	if err != nil {
		return candidate{}, false, fmt.Errorf("load Ledger: read identity: %v", err)
	}
	if cycleNumber != cfg.CycleNumber || executorNumber != cfg.ExecutorNumber ||
		account != cfg.Account || network != cfg.Network || symbol != cfg.Symbol {
		return candidate{}, false, fmt.Errorf("load Ledger: identity mismatch")
	}
	state.fillsThroughMS = uint64(max(fillsThrough.Int64, 0))
	state.lastReconMS = uint64(max(lastRecon.Int64, 0))
	state.trades = make(map[uint64]*trade.Trade)

	// load Trade evidence
	var rows *sql.Rows
	rows, err = s.db.Query(`
		SELECT trade_id, trade_no, symbol, status
		FROM account_trade
		WHERE ledger_id = ?
		ORDER BY trade_id`,
		cfg.ID,
	)
	if err != nil {
		return candidate{}, false, fmt.Errorf("load Ledger: query Trades: %v", err)
	}
	var storedStatuses = make(map[uint64]trade.Status)
	for rows.Next() {
		var tradeID uint64
		var tradeNo uint32
		var tradeSymbol string
		var status trade.Status
		err = rows.Scan(&tradeID, &tradeNo, &tradeSymbol, &status)
		if err != nil {
			return candidate{}, false, fmt.Errorf("load Ledger: scan Trade: %v", err)
		}
		var created *trade.Trade
		created, err = trade.New(trade.Input{
			LedgerID:    cfg.ID,
			TradeID:     tradeID,
			TradeNo:     tradeNo,
			Account:     cfg.Account,
			CycleNumber: cfg.CycleNumber,
			Symbol:      tradeSymbol,
		})
		if err != nil {
			return candidate{}, false, fmt.Errorf("load Ledger: %w", err)
		}
		state.trades[tradeID] = created
		storedStatuses[tradeID] = status
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return candidate{}, false, fmt.Errorf("load Ledger: read Trades: %v", err)
	}
	if err = rows.Close(); err != nil {
		return candidate{}, false, fmt.Errorf("load Ledger: close Trades: %v", err)
	}

	// load Order and Fill evidence
	err = s.loadOrders(cfg, state.trades)
	if err != nil {
		return candidate{}, false, err
	}
	for tradeID, owned := range state.trades {
		if owned.ReconState().Status != storedStatuses[tradeID] {
			return candidate{}, false, fmt.Errorf("load Ledger: Trade %d status mismatch", tradeID)
		}
	}
	return state, true, nil
}

func (s *ledgerStore) loadOrders(
	cfg Config,
	trades map[uint64]*trade.Trade,
) error {
	var rows, err = s.db.Query(`
		SELECT trade_id, order_id, account_name, cycle_no, batch_no, order_pos,
		       symbol, cloid, order_role, side, order_type, time_in_force,
		       requested_qty, requested_price, trigger_price, reduce_only,
		       submitted_ms, venue_order_id, status, reject_reason, updated_ms,
		       raw_json
		FROM account_order
		WHERE ledger_id = ?
		ORDER BY order_id`,
		cfg.ID,
	)
	if err != nil {
		return fmt.Errorf("load Ledger: query Orders: %v", err)
	}
	type loadedOrder struct {
		tradeID uint64
		orderID uint64
		owned   *order.Order
	}
	var loaded []loadedOrder
	for rows.Next() {
		var tradeID uint64
		var orderID uint64
		var account string
		var cycleNumber int
		var batchNo uint16
		var orderPos uint16
		var symbol string
		var value string
		var role string
		var side string
		var orderType string
		var timeInForce string
		var requestedQuantityText string
		var requestedPriceText sql.NullString
		var triggerPriceText sql.NullString
		var reduceOnly bool
		var submittedMS uint64
		var venueOrderID sql.NullInt64
		var status order.Status
		var rejectReason sql.NullString
		var updatedMS sql.NullInt64
		var raw string
		err = rows.Scan(
			&tradeID,
			&orderID,
			&account,
			&cycleNumber,
			&batchNo,
			&orderPos,
			&symbol,
			&value,
			&role,
			&side,
			&orderType,
			&timeInForce,
			&requestedQuantityText,
			&requestedPriceText,
			&triggerPriceText,
			&reduceOnly,
			&submittedMS,
			&venueOrderID,
			&status,
			&rejectReason,
			&updatedMS,
			&raw,
		)
		if err != nil {
			return fmt.Errorf("load Ledger: scan Order: %v", err)
		}
		var requestedQuantity decimal.Decimal
		requestedQuantity, err = decimal.NewFromString(requestedQuantityText)
		if err != nil {
			return fmt.Errorf("load Ledger: invalid Order quantity: %v", err)
		}
		var requestedPrice *decimal.Decimal
		requestedPrice, err = parseOptionalDecimal(requestedPriceText)
		if err != nil {
			return err
		}
		var triggerPrice *decimal.Decimal
		triggerPrice, err = parseOptionalDecimal(triggerPriceText)
		if err != nil {
			return err
		}
		var created *order.Order
		created, err = order.New(order.Input{
			LedgerID:          cfg.ID,
			TradeID:           tradeID,
			OrderID:           orderID,
			Account:           account,
			CycleNumber:       cycleNumber,
			Symbol:            symbol,
			BatchNo:           batchNo,
			OrderPos:          orderPos,
			CLOID:             value,
			Role:              role,
			Side:              side,
			Type:              orderType,
			TimeInForce:       timeInForce,
			RequestedQuantity: requestedQuantity,
			RequestedPrice:    requestedPrice,
			TriggerPrice:      triggerPrice,
			ReduceOnly:        reduceOnly,
			TimestampMS:       submittedMS,
		})
		if err != nil {
			return fmt.Errorf("load Ledger: %w", err)
		}
		if status != order.Created {
			err = created.RecordSubmit(
				uint64(max(venueOrderID.Int64, 0)),
				rejectReason.String,
				raw,
			)
			if err != nil {
				return fmt.Errorf("load Ledger: %w", err)
			}
		}
		if status != order.Created && status != order.Submitted {
			var observationMS = uint64(max(updatedMS.Int64, int64(submittedMS)))
			err = created.ApplyVenueState(order.VenueState{
				VenueOrderID: uint64(max(venueOrderID.Int64, 0)),
				Status:       status,
				RejectReason: rejectReason.String,
				TimestampMS:  observationMS,
				Raw:          raw,
			})
			if err != nil {
				return fmt.Errorf("load Ledger: %w", err)
			}
		}
		loaded = append(loaded, loadedOrder{
			tradeID: tradeID,
			orderID: orderID,
			owned:   created,
		})
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("load Ledger: read Orders: %v", err)
	}
	if err = rows.Close(); err != nil {
		return fmt.Errorf("load Ledger: close Orders: %v", err)
	}
	for _, loadedOrder := range loaded {
		err = s.loadFills(cfg, loadedOrder.owned)
		if err != nil {
			return err
		}
		loadedOrder.owned.RefreshRecon()
		var ownedTrade = trades[loadedOrder.tradeID]
		if ownedTrade == nil {
			return fmt.Errorf(
				"load Ledger: Order %d has unknown Trade %d",
				loadedOrder.orderID,
				loadedOrder.tradeID,
			)
		}
		err = ownedTrade.AddOrder(loadedOrder.owned)
		if err != nil {
			return fmt.Errorf("load Ledger: %w", err)
		}
	}
	return nil
}

func (s *ledgerStore) loadFills(cfg Config, owned *order.Order) error {
	var record = owned.Record()
	var rows, err = s.db.Query(`
		SELECT venue_tid, venue_order_id, account_name, cycle_no, symbol, cloid,
		       side, qty, price, event_ms, fee, liquidity, raw_json
		FROM account_fill
		WHERE ledger_id = ? AND order_id = ?
		ORDER BY event_ms, venue_tid`,
		cfg.ID,
		record.OrderID,
	)
	if err != nil {
		return fmt.Errorf("load Ledger: query Fills: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var input fill.Input
		var quantityText string
		var priceText string
		var feeText sql.NullString
		var liquidity sql.NullString
		err = rows.Scan(
			&input.VenueTID,
			&input.VenueOrderID,
			&input.Account,
			&input.CycleNumber,
			&input.Symbol,
			&input.CLOID,
			&input.Side,
			&quantityText,
			&priceText,
			&input.TimestampMS,
			&feeText,
			&liquidity,
			&input.Raw,
		)
		if err != nil {
			return fmt.Errorf("load Ledger: scan Fill: %v", err)
		}
		input.LedgerID = cfg.ID
		input.TradeID = record.TradeID
		input.OrderID = record.OrderID
		input.Quantity, err = decimal.NewFromString(quantityText)
		if err != nil {
			return fmt.Errorf("load Ledger: invalid Fill quantity: %v", err)
		}
		input.Price, err = decimal.NewFromString(priceText)
		if err != nil {
			return fmt.Errorf("load Ledger: invalid Fill price: %v", err)
		}
		input.Fee, err = parseOptionalDecimal(feeText)
		if err != nil {
			return err
		}
		input.Liquidity = liquidity.String
		err = owned.ApplyFill(input)
		if err != nil {
			return fmt.Errorf("load Ledger: %w", err)
		}
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("load Ledger: read Fills: %v", err)
	}
	return nil
}

// Section 3 - Generic Helpers

func nullableUint(value uint64) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableDecimal(value decimal.Decimal, present bool) any {
	if !present {
		return nil
	}
	return value.String()
}

func nullableDecimalPointer(value *decimal.Decimal) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func parseOptionalDecimal(value sql.NullString) (*decimal.Decimal, error) {
	if !value.Valid {
		return nil, nil
	}
	var parsed, err = decimal.NewFromString(value.String)
	if err != nil {
		return nil, fmt.Errorf("load Ledger: invalid decimal: %v", err)
	}
	return &parsed, nil
}

func sortedTradeIDs(trades map[uint64]*trade.Trade) []uint64 {
	var ids = make([]uint64, 0, len(trades))
	for id := range trades {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	return ids
}

const ledgerDDL = `
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
`
