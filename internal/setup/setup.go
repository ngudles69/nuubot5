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

// Context contains one admitted setup result.
type Context struct {
	Config      config.Config
	Credentials config.Credentials
	Bot         datastore.BotSpec
	Meta        meta.Instrument
	ResultPath  string
}

// Section 1 - Program Flow

// Setup returns one admitted process context.
func Setup(log *logging.Logger, sweepID, botID uint64) (Context, error) {
	// resolve root
	var root, err = os.Getwd()
	if err != nil {
		return Context{}, fmt.Errorf("get working directory: %w", err)
	}
	// load config
	var cfg config.Config
	var configPath = os.Getenv("NUUBOT_CONFIG")
	if configPath == "" {
		configPath = filepath.Join(root, "workspace", "config", "config.toml")
	} else if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, configPath)
	}
	cfg, err = config.Load(configPath)
	if err != nil {
		return Context{}, fmt.Errorf("load setup config: %w", err)
	}
	// load credentials
	var credentials config.Credentials
	credentials, err = config.LoadCredentials(
		filepath.Join(root, "workspace", "config", "credentials.toml"),
	)
	if err != nil {
		return Context{}, fmt.Errorf("load setup credentials: %w", err)
	}
	// prepare datastore
	var bot datastore.BotSpec
	bot, err = datastore.LoadBot(
		config.Rooted(root, cfg.Paths.Database),
		sweepID,
		botID,
	)
	if err != nil {
		return Context{}, fmt.Errorf("prepare datastore: %w", err)
	}
	// validate ticks path
	bot.TicksPath, err = config.ResolveDataPath(
		config.Rooted(root, cfg.Paths.SharedData),
		bot.TicksPath,
	)
	if err != nil {
		return Context{}, fmt.Errorf("validate ticks path: %w", err)
	}

	// admit mainnet Meta
	var timeout = time.Duration(cfg.Process.RequestTimeoutSeconds) * time.Second
	var client *hyperliquid.Client
	client, err = hyperliquid.New("mainnet", timeout)
	if err != nil {
		return Context{}, fmt.Errorf("admit meta: %w", err)
	}
	var requestContext, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var instrument meta.Instrument
	instrument, err = meta.EnsureFresh(
		requestContext,
		config.Rooted(root, cfg.Paths.Database),
		bot.Symbol,
		time.Now().UTC(),
		client,
	)
	if err != nil {
		return Context{}, fmt.Errorf("admit meta: %w", err)
	}

	// Shared WebSocket ownership remains TBD. Setup starts no background work.

	// return setup
	log.Info(fmt.Sprintf("setup initialized symbol=%s", bot.Symbol))
	return Context{
		Config:      cfg,
		Credentials: credentials,
		Bot:         bot,
		Meta:        instrument,
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
