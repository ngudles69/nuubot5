package setup

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"nuubot/internal/config"
	"nuubot/internal/datastore"
)

// Context contains validated setup values.
type Context struct {
	Config config.Config
	Bot    datastore.BotSpec
}

// Section 1 - Program Flow

// Init loads and validates one Bot setup.
func Init(logger *slog.Logger, sweepID, botID uint64) (Context, error) {
	var root, err = os.Getwd()
	if err != nil {
		return Context{}, fmt.Errorf("get working directory: %w", err)
	}
	var cfg, configErr = config.Load(filepath.Join(root, "config.toml"))
	if configErr != nil {
		return Context{}, configErr
	}
	var bot, botErr = datastore.LoadBot(
		config.Rooted(root, cfg.Paths.SweepDatabase),
		sweepID,
		botID,
	)
	if botErr != nil {
		return Context{}, fmt.Errorf("load bot: %w", botErr)
	}
	bot.TicksPath, err = within(config.Rooted(root, cfg.Paths.SharedData), bot.TicksPath)
	if err != nil {
		return Context{}, fmt.Errorf("validate ticks path: %w", err)
	}
	logger.With("component", "setup").Info(
		"setup initialized",
		"event", "init",
		"status", "success",
		"sweep_id", sweepID,
		"bot_id", botID,
		"symbol", bot.Symbol,
	)
	return Context{Config: cfg, Bot: bot}, nil
}

// Section 2 - Domain Helpers

func within(root, path string) (string, error) {
	var err error
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve shared_data %s: %w", root, err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve data path %s: %w", path, err)
	}
	var relative string
	relative, err = filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("data path is outside shared_data: %s", path)
	}
	return path, nil
}

// Section 3 - Generic Helpers
