package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"nuubot/internal/config"
	"nuubot/internal/datastore"
	"nuubot/internal/toolkit/logging"
)

// Context contains one admitted setup result.
type Context struct {
	Config      config.Config
	Credentials config.Credentials
	Bot         datastore.BotSpec
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
	cfg, err = config.Load(filepath.Join(root, "workspace", "config", "config.toml"))
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
	// setup datastore
	var bot datastore.BotSpec
	bot, err = datastore.LoadBot(
		config.Rooted(root, cfg.Paths.SweepDatabase),
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

	// Meta REFRESH is pending NuubotDB:
	// - read the dataset refresh time through Datastore
	// - continue when Meta is present and younger than 24 hours
	// - refresh when Meta is empty or stale
	// - continue only after Meta is admitted

	// Shared WebSocket ownership remains TBD. Setup starts no background work.

	// return setup
	log.Info(fmt.Sprintf("setup initialized symbol=%s", bot.Symbol))
	return Context{
		Config:      cfg,
		Credentials: credentials,
		Bot:         bot,
	}, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
