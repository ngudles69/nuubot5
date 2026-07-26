package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nuubot/internal/config"
	"nuubot/internal/datastore"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/meta"
	"nuubot/internal/toolkit/logging"
)

// Admission contains trusted external inputs for one Bot build.
type Admission struct {
	App        config.App
	Bot        datastore.Bot
	Meta       meta.Instrument
	ResultPath string
}

// Section 1 - Program Flow

// Setup returns one admitted standalone BtBot input.
func Setup(
	caller context.Context,
	log *logging.Logger,
	sweepID,
	botID uint64,
) (Admission, error) {
	if caller == nil {
		return Admission{}, fmt.Errorf("setup requires caller context")
	}
	// resolve root
	var root, err = os.Getwd()
	if err != nil {
		return Admission{}, fmt.Errorf("get working directory: %w", err)
	}
	// load config
	var cfg config.App
	var configPath = os.Getenv("NUUBOT_CONFIG")
	if configPath == "" {
		configPath = filepath.Join(root, "workspace", "config", "config.toml")
	} else if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, configPath)
	}
	cfg, err = config.LoadApp(configPath)
	if err != nil {
		return Admission{}, fmt.Errorf("load setup AppConfig: %w", err)
	}
	// prepare datastore
	var storedBot datastore.Bot
	storedBot, err = datastore.LoadBot(
		config.Rooted(root, cfg.Paths.Database),
		sweepID,
		botID,
	)
	if err != nil {
		return Admission{}, fmt.Errorf("prepare datastore: %w", err)
	}
	// validate ticks path
	storedBot.Replay.TicksPath, err = config.ResolveDataPath(
		config.Rooted(root, cfg.Paths.SharedData),
		storedBot.Replay.TicksPath,
	)
	if err != nil {
		return Admission{}, fmt.Errorf("validate ticks path: %w", err)
	}

	// admit mainnet Meta
	var timeout = time.Duration(cfg.Process.RequestTimeoutSeconds) * time.Second
	var client *hyperliquid.Client
	client, err = hyperliquid.New("mainnet", timeout)
	if err != nil {
		return Admission{}, fmt.Errorf("admit Meta: %w", err)
	}
	var requestContext, cancel = context.WithTimeout(caller, timeout)
	defer cancel()
	var instrument meta.Instrument
	instrument, err = meta.EnsureFresh(
		requestContext,
		config.Rooted(root, cfg.Paths.Database),
		storedBot.Replay.Symbol,
		time.Now().UTC(),
		client,
	)
	if err != nil {
		return Admission{}, fmt.Errorf("admit Meta: %w", err)
	}

	// Shared WebSocket ownership remains TBD. Setup starts no background work.

	// return setup
	log.Info(fmt.Sprintf(
		"setup initialized bot_spec=%s symbol=%s",
		storedBot.BotSpecID,
		storedBot.Replay.Symbol,
	))
	return Admission{
		App:  cfg,
		Bot:  storedBot,
		Meta: instrument,
		ResultPath: filepath.Join(
			root,
			"workspace",
			"db",
			"sweeps",
			fmt.Sprintf("sweep_%d", sweepID),
			fmt.Sprintf("bot_%d.db", botID),
		),
	}, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
