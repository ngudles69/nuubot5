package datastore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type storedBot struct {
	General struct {
		Symbol string `json:"symbol"`
		Start  string `json:"start"`
		End    string `json:"end"`
	} `json:"general"`
	Data struct {
		Ticks string `json:"ticks"`
	} `json:"data"`
	DateRange struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"date_range"`
}

// Section 1 - Program Flow

// LoadBot loads one validated Bot specification.
func LoadBot(path string, sweepID, botID uint64) (BotSpec, error) {
	// open database
	var bot BotSpec
	dsn := "file:" + filepath.ToSlash(path) + "?mode=ro&immutable=1"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return bot, fmt.Errorf("open sweep database %s: %w", path, err)
	}
	defer db.Close()

	// query bot
	var text string
	err = db.QueryRow(
		"SELECT config_json FROM bot WHERE sweep_id = ? AND bot_id = ?",
		sweepID,
		botID,
	).Scan(&text)
	if err != nil {
		return bot, fmt.Errorf("load bot sweep_id=%d bot_id=%d: %w", sweepID, botID, err)
	}

	// decode bot
	var stored storedBot
	if err := json.Unmarshal([]byte(text), &stored); err != nil {
		return bot, fmt.Errorf("parse bot config: %w", err)
	}
	// parse dates
	replayStart, err := time.Parse(time.DateOnly, stored.DateRange.Start)
	if err != nil {
		return bot, fmt.Errorf("invalid bot replay start date: %w", err)
	}
	replayEnd, err := time.Parse(time.DateOnly, stored.DateRange.End)
	if err != nil {
		return bot, fmt.Errorf("invalid bot replay end date: %w", err)
	}
	startAt, err := parseOptionalTime(stored.General.Start)
	if err != nil {
		return bot, fmt.Errorf("invalid bot start: %w", err)
	}
	endAt, err := parseOptionalTime(stored.General.End)
	if err != nil {
		return bot, fmt.Errorf("invalid bot end: %w", err)
	}
	// validate bot
	if stored.General.Symbol == "" || stored.Data.Ticks == "" || !replayStart.Before(replayEnd) {
		return bot, fmt.Errorf("invalid bot symbol, tick path, or date range")
	}
	if startAt != nil && endAt != nil && !startAt.Before(*endAt) {
		return bot, fmt.Errorf("bot start must precede end")
	}
	// return bot
	return BotSpec{
		Symbol:      stored.General.Symbol,
		TicksPath:   filepath.Clean(stored.Data.Ticks),
		ReplayStart: replayStart,
		ReplayEnd:   replayEnd,
		StartAt:     startAt,
		EndAt:       endAt,
	}, nil
}

// Section 2 - Domain Helpers

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			parsed = parsed.UTC()
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("expected rfc3339 timestamp or yyyy-mm-dd")
}

// Section 3 - Generic Helpers
