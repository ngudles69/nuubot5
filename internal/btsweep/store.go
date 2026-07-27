package btsweep

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// CreatedBot identifies one generated immutable Bot record.
type CreatedBot struct {
	ID     uint64
	Number uint64
}

// Creation identifies one generated immutable Sweep and its Bots.
type Creation struct {
	SweepID uint64
	Bots    []CreatedBot
}

type storedBotConfig struct {
	General struct {
		Symbol string `json:"symbol"`
		Start  string `json:"start"`
		End    string `json:"end"`
	} `json:"general"`
	Data struct {
		Ticks string `json:"ticks"`
	} `json:"data"`
	DateRange DateRange `json:"date_range"`
}

// Section 1 - Program Flow

// Create validates one Sweep template and persists its immutable records.
func Create(templatePath string, databasePath string) (Creation, error) {
	// load sweep template
	var expansion, err = Load(templatePath)
	if err != nil {
		return Creation{}, err
	}

	// open database
	var db *sql.DB
	db, err = sql.Open("sqlite", databasePath)
	if err != nil {
		return Creation{}, fmt.Errorf("create Sweep: open database: %w", err)
	}
	defer db.Close()

	// create transaction
	var transaction *sql.Tx
	transaction, err = db.Begin()
	if err != nil {
		return Creation{}, fmt.Errorf("create Sweep: begin transaction: %w", err)
	}
	defer transaction.Rollback()

	// create Sweep record
	var created Creation
	created.SweepID, err = nextID(transaction, "sweep", "sweep_id")
	if err != nil {
		return Creation{}, err
	}
	err = insertSweep(transaction, created.SweepID, expansion)
	if err != nil {
		return Creation{}, err
	}

	// create Bot records
	var nextBotID uint64
	nextBotID, err = nextID(transaction, "bot", "bot_id")
	if err != nil {
		return Creation{}, err
	}
	for _, generated := range expansion.Bots {
		var botID = nextBotID
		nextBotID++
		err = insertBot(transaction, created.SweepID, botID, expansion, generated)
		if err != nil {
			return Creation{}, err
		}
		created.Bots = append(created.Bots, CreatedBot{ID: botID, Number: generated.Number})
	}

	// commit immutable records
	err = transaction.Commit()
	if err != nil {
		return Creation{}, fmt.Errorf("create Sweep: commit transaction: %w", err)
	}
	return created, nil
}

// Section 2 - Domain Helpers

func insertSweep(transaction *sql.Tx, sweepID uint64, expansion Expansion) error {
	var definition, err = json.Marshal(expansion)
	if err != nil {
		return fmt.Errorf("create Sweep: encode definition: %v", err)
	}
	var now = time.Now().UTC().Format(time.RFC3339Nano)
	_, err = transaction.Exec(
		`INSERT INTO sweep (
			sweep_id, source_path, doc, definition_json, status,
			process_generation, process_status, process_health,
			process_failed_count, generated_count, result_json,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sweepID,
		expansion.SourcePath,
		expansion.Doc,
		string(definition),
		"configured",
		0,
		"stopped",
		"unknown",
		0,
		len(expansion.Bots),
		"{}",
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("create Sweep %d: insert record: %v", sweepID, err)
	}
	return nil
}

func insertBot(
	transaction *sql.Tx,
	sweepID uint64,
	botID uint64,
	expansion Expansion,
	generated Bot,
) error {
	var stored storedBotConfig
	stored.General.Symbol = expansion.Symbol
	stored.Data.Ticks = expansion.TicksPath
	stored.DateRange = generated.DateRange
	var configJSON, err = json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("create Bot %d: encode replay Config: %v", botID, err)
	}
	var now = time.Now().UTC().Format(time.RFC3339Nano)
	_, err = transaction.Exec(
		`INSERT INTO bot (
			bot_id, sweep_id, bot_no, status, sweep_generation,
			config_json, execution_count, created_at, updated_at,
			bot_spec_id, config_toml, config_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		botID,
		sweepID,
		generated.Number,
		"configured",
		0,
		string(configJSON),
		0,
		now,
		now,
		expansion.BotSpecID,
		generated.ConfigTOML,
		generated.ConfigHash,
	)
	if err != nil {
		return fmt.Errorf("create Bot %d: insert record: %v", botID, err)
	}
	return nil
}

// Section 3 - Generic Helpers

func nextID(transaction *sql.Tx, table string, column string) (uint64, error) {
	var query = fmt.Sprintf("SELECT COALESCE(MAX(%s), 0) + 1 FROM %s", column, table)
	var id uint64
	var err = transaction.QueryRow(query).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create record: read next %s: %v", column, err)
	}
	return id, nil
}
