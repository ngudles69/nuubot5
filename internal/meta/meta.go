// Package meta owns admitted Hyperliquid instrument metadata.
package meta

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"

	"nuubot/internal/hyperliquid"
)

const (
	kindPerpetual  = "perp"
	networkMainnet = "mainnet"
	freshFor       = 24 * time.Hour
)

// Source provides one complete exchange-owned perpetual Meta dataset.
type Source interface {
	PerpetualMeta(context.Context) (hyperliquid.PerpetualMeta, error)
}

// Instrument contains one normalized perpetual trading contract.
type Instrument struct {
	Network       string
	Kind          string
	Symbol        string
	AssetID       uint32
	ExchangeIndex uint32
	MaxLeverage   uint32
	MarginTableID uint32
	SizeDecimals  int32
	PriceDecimals int32
	OnlyIsolated  bool
	IsDelisted    bool
	Retired       bool
	Raw           string
}

// Section 1 - Program Flow

// EnsureFresh admits one symbol from a complete fresh perpetual Meta dataset.
func EnsureFresh(
	ctx context.Context,
	path string,
	symbol string,
	now time.Time,
	source Source,
) (Instrument, error) {
	// validate meta request
	if path == "" || symbol == "" || source == nil {
		return Instrument{}, fmt.Errorf("ensure meta: path, symbol, and source are required")
	}
	symbol = strings.ToUpper(symbol)

	// open shared database
	var err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return Instrument{}, fmt.Errorf("ensure meta: create shared database directory: %w", err)
	}
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	var db *sql.DB
	db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return Instrument{}, fmt.Errorf("ensure meta: open shared database: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	// prepare Meta tables
	err = prepareTables(db)
	if err != nil {
		return Instrument{}, err
	}

	// claim network refresh
	// ponytail: one SQLite writer lock prevents duplicate refreshes; use a lease if refresh blocks future main-store writes.
	var transaction *sql.Tx
	transaction, err = db.BeginTx(ctx, nil)
	if err != nil {
		return Instrument{}, fmt.Errorf("ensure meta: begin refresh: %v", err)
	}
	defer transaction.Rollback()

	// read dataset freshness
	var refreshedMS uint64
	var count uint64
	refreshedMS, count, err = datasetState(transaction, networkMainnet)
	if err != nil {
		return Instrument{}, err
	}
	var fresh = count > 0 && now.Sub(time.UnixMilli(int64(refreshedMS))) < freshFor

	// refresh stale dataset
	if !fresh {
		var response hyperliquid.PerpetualMeta
		response, err = source.PerpetualMeta(ctx)
		if err != nil {
			return Instrument{}, fmt.Errorf("ensure meta: fetch perpetual meta: %w", err)
		}
		var instruments []Instrument
		instruments, err = normalize(networkMainnet, response)
		if err != nil {
			return Instrument{}, err
		}
		err = replaceDataset(transaction, networkMainnet, now, response, instruments)
		if err != nil {
			return Instrument{}, err
		}
	}

	// load admitted symbol
	var instrument Instrument
	instrument, err = loadSymbol(transaction, networkMainnet, symbol)
	if err != nil {
		return Instrument{}, err
	}

	// commit Meta admission
	err = transaction.Commit()
	if err != nil {
		return Instrument{}, fmt.Errorf("ensure meta: commit refresh: %v", err)
	}
	return instrument, nil
}

// Section 2 - Domain Helpers

// RoundSize truncates one size to the exchange lot precision.
func (i Instrument) RoundSize(value decimal.Decimal) decimal.Decimal {
	return value.Truncate(i.SizeDecimals)
}

// RoundPrice rounds one price to Hyperliquid's decimal and significant-figure limits.
func (i Instrument) RoundPrice(value decimal.Decimal) decimal.Decimal {
	if value.Equal(value.Truncate(0)) {
		return value
	}
	var places = significantPlaces(value)
	if places > i.PriceDecimals {
		places = i.PriceDecimals
	}
	if places < 0 {
		places = 0
	}
	return value.Round(places)
}

func normalize(
	network string,
	response hyperliquid.PerpetualMeta,
) ([]Instrument, error) {
	var instruments = make([]Instrument, 0, len(response.Universe))
	for index, asset := range response.Universe {
		var priceDecimals = int32(6) - int32(asset.SizeDecimals)
		if priceDecimals < 0 {
			return nil, fmt.Errorf("normalize meta: invalid size decimals for %s", asset.Name)
		}
		instruments = append(instruments, Instrument{
			Network:       network,
			Kind:          kindPerpetual,
			Symbol:        strings.ToUpper(asset.Name),
			AssetID:       uint32(index),
			ExchangeIndex: uint32(index),
			MaxLeverage:   asset.MaxLeverage,
			MarginTableID: asset.MarginTableID,
			SizeDecimals:  int32(asset.SizeDecimals),
			PriceDecimals: priceDecimals,
			OnlyIsolated:  asset.OnlyIsolated,
			IsDelisted:    asset.IsDelisted,
			Raw:           asset.Raw,
		})
	}
	return instruments, nil
}

func prepareTables(db *sql.DB) error {
	var _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS meta_dataset (
			network       TEXT NOT NULL,
			kind          TEXT NOT NULL,
			refreshed_ms  INTEGER NOT NULL,
			symbol_count  INTEGER NOT NULL,
			PRIMARY KEY (network, kind)
		);
		CREATE TABLE IF NOT EXISTS meta (
			network          TEXT NOT NULL,
			kind             TEXT NOT NULL,
			symbol           TEXT NOT NULL,
			asset_id         INTEGER NOT NULL,
			exchange_index   INTEGER NOT NULL,
			max_leverage     INTEGER NOT NULL,
			margin_table_id  INTEGER NOT NULL,
			size_decimals    INTEGER NOT NULL,
			price_decimals   INTEGER NOT NULL,
			only_isolated    INTEGER NOT NULL,
			is_delisted      INTEGER NOT NULL,
			retired          INTEGER NOT NULL,
			raw_json         TEXT NOT NULL,
			PRIMARY KEY (network, kind, symbol)
		);
		CREATE TABLE IF NOT EXISTS meta_margin_table (
			network     TEXT NOT NULL,
			kind        TEXT NOT NULL,
			table_id    INTEGER NOT NULL,
			payload_json TEXT NOT NULL,
			PRIMARY KEY (network, kind, table_id)
		);
	`)
	if err != nil {
		return fmt.Errorf("ensure meta: prepare tables: %v", err)
	}
	return nil
}

func datasetState(transaction *sql.Tx, network string) (uint64, uint64, error) {
	var refreshedMS uint64
	var count uint64
	var err = transaction.QueryRow(
		`SELECT refreshed_ms,
		        CASE WHEN symbol_count = (
		            SELECT COUNT(*)
		            FROM meta
		            WHERE network = ? AND kind = ? AND retired = 0
		        ) THEN symbol_count ELSE 0 END
		 FROM meta_dataset
		 WHERE network = ? AND kind = ?`,
		network,
		kindPerpetual,
		network,
		kindPerpetual,
	).Scan(&refreshedMS, &count)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("ensure meta: read dataset freshness: %v", err)
	}
	return refreshedMS, count, nil
}

func replaceDataset(
	transaction *sql.Tx,
	network string,
	now time.Time,
	response hyperliquid.PerpetualMeta,
	instruments []Instrument,
) error {
	var _, err = transaction.Exec(
		`UPDATE meta SET retired = 1 WHERE network = ? AND kind = ?`,
		network,
		kindPerpetual,
	)
	if err != nil {
		return fmt.Errorf("ensure meta: retire previous symbols: %v", err)
	}
	for _, instrument := range instruments {
		_, err = transaction.Exec(`
			INSERT INTO meta (
				network, kind, symbol, asset_id, exchange_index,
				max_leverage, margin_table_id, size_decimals, price_decimals,
				only_isolated, is_delisted, retired, raw_json
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)
			ON CONFLICT (network, kind, symbol) DO UPDATE SET
				asset_id = excluded.asset_id,
				exchange_index = excluded.exchange_index,
				max_leverage = excluded.max_leverage,
				margin_table_id = excluded.margin_table_id,
				size_decimals = excluded.size_decimals,
				price_decimals = excluded.price_decimals,
				only_isolated = excluded.only_isolated,
				is_delisted = excluded.is_delisted,
				retired = 0,
				raw_json = excluded.raw_json`,
			instrument.Network,
			instrument.Kind,
			instrument.Symbol,
			instrument.AssetID,
			instrument.ExchangeIndex,
			instrument.MaxLeverage,
			instrument.MarginTableID,
			instrument.SizeDecimals,
			instrument.PriceDecimals,
			instrument.OnlyIsolated,
			instrument.IsDelisted,
			instrument.Raw,
		)
		if err != nil {
			return fmt.Errorf("ensure meta: store symbol %s: %v", instrument.Symbol, err)
		}
	}
	_, err = transaction.Exec(
		`DELETE FROM meta_margin_table WHERE network = ? AND kind = ?`,
		network,
		kindPerpetual,
	)
	if err != nil {
		return fmt.Errorf("ensure meta: clear margin tables: %v", err)
	}
	for _, table := range response.MarginTables {
		var payload []byte
		payload, err = json.Marshal(table)
		if err != nil {
			return fmt.Errorf("ensure meta: encode margin table %d: %v", table.ID, err)
		}
		_, err = transaction.Exec(
			`INSERT INTO meta_margin_table (network, kind, table_id, payload_json)
			 VALUES (?, ?, ?, ?)`,
			network,
			kindPerpetual,
			table.ID,
			string(payload),
		)
		if err != nil {
			return fmt.Errorf("ensure meta: store margin table %d: %v", table.ID, err)
		}
	}
	_, err = transaction.Exec(`
		INSERT INTO meta_dataset (network, kind, refreshed_ms, symbol_count)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (network, kind) DO UPDATE SET
			refreshed_ms = excluded.refreshed_ms,
			symbol_count = excluded.symbol_count`,
		network,
		kindPerpetual,
		now.UnixMilli(),
		len(instruments),
	)
	if err != nil {
		return fmt.Errorf("ensure meta: store dataset freshness: %v", err)
	}
	return nil
}

func loadSymbol(
	transaction *sql.Tx,
	network string,
	symbol string,
) (Instrument, error) {
	var instrument Instrument
	var err = transaction.QueryRow(`
		SELECT network, kind, symbol, asset_id, exchange_index, max_leverage,
		       margin_table_id, size_decimals, price_decimals, only_isolated,
		       is_delisted, retired, raw_json
		FROM meta
		WHERE network = ? AND kind = ? AND symbol = ?`,
		network,
		kindPerpetual,
		symbol,
	).Scan(
		&instrument.Network,
		&instrument.Kind,
		&instrument.Symbol,
		&instrument.AssetID,
		&instrument.ExchangeIndex,
		&instrument.MaxLeverage,
		&instrument.MarginTableID,
		&instrument.SizeDecimals,
		&instrument.PriceDecimals,
		&instrument.OnlyIsolated,
		&instrument.IsDelisted,
		&instrument.Retired,
		&instrument.Raw,
	)
	if err == sql.ErrNoRows {
		return Instrument{}, fmt.Errorf(
			"ensure meta: symbol %s is absent from %s perpetual meta",
			symbol,
			network,
		)
	}
	if err != nil {
		return Instrument{}, fmt.Errorf("ensure meta: load symbol %s: %v", symbol, err)
	}
	if instrument.IsDelisted || instrument.Retired {
		return Instrument{}, fmt.Errorf(
			"ensure meta: symbol %s is unavailable on %s",
			symbol,
			network,
		)
	}
	return instrument, nil
}

// Section 3 - Generic Helpers

func significantPlaces(value decimal.Decimal) int32 {
	var parts = strings.SplitN(value.Abs().String(), ".", 2)
	var integer = strings.TrimLeft(parts[0], "0")
	if integer != "" {
		return int32(5 - len(integer))
	}
	var fraction = parts[1]
	var leading = len(fraction) - len(strings.TrimLeft(fraction, "0"))
	return int32(leading + 5)
}
