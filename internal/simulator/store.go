package simulator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const simulatorSchemaVersion = 3

type simulatorStore struct {
	db *sql.DB
}

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

// Section 1 - Program Flow

func openSimulatorStore(path string) (*simulatorStore, error) {
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	var db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open Simulator store: %v", err)
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS simulator_venue_state (
			account_name   TEXT NOT NULL,
			symbol         TEXT NOT NULL,
			schema_version INTEGER NOT NULL,
			payload_json   TEXT NOT NULL,
			updated_ms     INTEGER NOT NULL,
			PRIMARY KEY (account_name, symbol)
		)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("open Simulator store: create schema: %v", err)
	}
	return &simulatorStore{db: db}, nil
}

func (s *simulatorStore) close() error {
	var err = s.db.Close()
	if err != nil {
		return fmt.Errorf("close Simulator store: %v", err)
	}
	return nil
}

// Section 2 - Domain Helpers

func (s *simulatorStore) save(cfg Config, state storedState) error {
	var payload, err = json.Marshal(state)
	if err != nil {
		return fmt.Errorf("persist Simulator: encode state: %v", err)
	}
	_, err = s.db.Exec(`
		INSERT INTO simulator_venue_state (
			account_name, symbol, schema_version, payload_json, updated_ms
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (account_name, symbol) DO UPDATE SET
			schema_version = excluded.schema_version,
			payload_json = excluded.payload_json,
			updated_ms = excluded.updated_ms`,
		cfg.Account,
		cfg.Symbol,
		simulatorSchemaVersion,
		string(payload),
		time.Now().UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("persist Simulator: store state: %v", err)
	}
	return nil
}

func (s *simulatorStore) load(cfg Config) (storedState, bool, error) {
	var schemaVersion int
	var payload string
	var err = s.db.QueryRow(`
		SELECT schema_version, payload_json
		FROM simulator_venue_state
		WHERE account_name = ? AND symbol = ?`,
		cfg.Account,
		cfg.Symbol,
	).Scan(&schemaVersion, &payload)
	if err == sql.ErrNoRows {
		return storedState{}, false, nil
	}
	if err != nil {
		return storedState{}, false, fmt.Errorf("load Simulator: read state: %v", err)
	}
	if schemaVersion != simulatorSchemaVersion {
		return storedState{}, false, fmt.Errorf(
			"load Simulator: unsupported schema version %d",
			schemaVersion,
		)
	}
	var state storedState
	err = json.Unmarshal([]byte(payload), &state)
	if err != nil {
		return storedState{}, false, fmt.Errorf("load Simulator: decode state: %v", err)
	}
	if state.SchemaVersion != simulatorSchemaVersion ||
		state.Account != cfg.Account ||
		state.Asset != cfg.Asset ||
		state.Symbol != cfg.Symbol ||
		state.Equity != cfg.Equity.String() ||
		state.FeePct != cfg.FeePct.String() ||
		state.SlippagePct != cfg.SlippagePct.String() {
		return storedState{}, false, fmt.Errorf("load Simulator: identity mismatch")
	}
	return state, true, nil
}

// Section 3 - Generic Helpers
