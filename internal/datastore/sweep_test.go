package datastore

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// Section 1 - Program Flow

func TestOptionalBotTime(t *testing.T) {
	for _, value := range []string{"2026-07-23", "2026-07-23T12:30:00Z"} {
		parsed, err := parseOptionalTime(value)
		if err != nil || parsed == nil {
			t.Fatalf("parse %q: %v", value, err)
		}
	}
	parsed, err := parseOptionalTime("")
	if err != nil || parsed != nil {
		t.Fatalf("empty time: parsed=%v err=%v", parsed, err)
	}
}

func TestLoadBotReturnsExactStoredConfig(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "nuubot.db")
	var db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE bot (
			bot_id INTEGER PRIMARY KEY,
			sweep_id INTEGER NOT NULL,
			bot_spec_id TEXT NOT NULL,
			config_toml TEXT NOT NULL,
			config_hash TEXT NOT NULL,
			config_json TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create bot table: %v", err)
	}
	var configTOML = "bot_spec = \"macross_trade_bot\"\n"
	var configHash = fmt.Sprintf("%x", sha256.Sum256([]byte(configTOML)))
	_, err = db.Exec(
		`INSERT INTO bot VALUES (?, ?, ?, ?, ?, ?)`,
		9,
		6,
		"macross_trade_bot",
		configTOML,
		configHash,
		`{
			"general":{"symbol":"BTC","start":"","end":""},
			"data":{"ticks":"ticks"},
			"date_range":{"start":"2026-03-01","end":"2026-06-01"}
		}`,
	)
	if err != nil {
		t.Fatalf("insert bot: %v", err)
	}

	var bot Bot
	bot, err = LoadBot(path, 6, 9)
	if err != nil {
		t.Fatalf("load bot: %v", err)
	}
	if bot.BotSpecID != "macross_trade_bot" ||
		bot.ConfigTOML != configTOML ||
		bot.ConfigHash != configHash ||
		bot.Replay.Symbol != "BTC" {
		t.Fatalf("unexpected bot: %#v", bot)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
