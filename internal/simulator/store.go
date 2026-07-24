package simulator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const simulatorSchemaVersion = 2

type simulatorStore struct {
	db *sql.DB
}

type storedOrder struct {
	Request           OrderRequest
	VenueOrderID      uint64
	Status            string
	Armed             bool
	RemainingQuantity string
	FilledQuantity    string
	AverageFillPrice  string
	Fees              string
	TimestampMS       uint64
}

type storedState struct {
	SchemaVersion    int
	LedgerID         uint64
	Name             string
	Account          string
	CycleNumber      int
	Symbol           string
	Equity           string
	FeePct           string
	SlippagePct      string
	NextVenueOrderID uint64
	NextVenueTID     uint64
	OpenOrders       []storedOrder
	OrderHistory     []OrderState
	Fills            []FillState
}

// Section 1 - Program Flow

func openSimulatorStore(path string) (*simulatorStore, error) {
	// open Simulator store
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	var db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open Simulator store: %v", err)
	}
	db.SetMaxOpenConns(1)

	// verify Simulator table
	var name string
	err = db.QueryRow(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table' AND name = 'simulator_state'`,
	).Scan(&name)
	if err != nil {
		db.Close()
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("open Simulator store: Ledger schema is missing")
		}
		return nil, fmt.Errorf("open Simulator store: verify table: %v", err)
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
		INSERT INTO simulator_state (
			ledger_id, schema_version, payload_json, updated_ms
		) VALUES (?, ?, ?, ?)
		ON CONFLICT (ledger_id) DO UPDATE SET
			schema_version = excluded.schema_version,
			payload_json = excluded.payload_json,
			updated_ms = excluded.updated_ms`,
		cfg.LedgerID,
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
		FROM simulator_state
		WHERE ledger_id = ?`,
		cfg.LedgerID,
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
		state.LedgerID != cfg.LedgerID ||
		state.Name != cfg.Name ||
		state.Account != cfg.Account ||
		state.CycleNumber != cfg.CycleNumber ||
		state.Symbol != cfg.Symbol ||
		state.Equity != cfg.Equity.String() ||
		state.FeePct != cfg.FeePct.String() ||
		state.SlippagePct != cfg.SlippagePct.String() {
		return storedState{}, false, fmt.Errorf("load Simulator: identity mismatch")
	}
	return state, true, nil
}

// Section 3 - Generic Helpers
