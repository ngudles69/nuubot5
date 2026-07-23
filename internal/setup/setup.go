package setup

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"nuubot5/internal/config"
	"nuubot5/internal/datastore"
)

// Context contains validated setup values.
type Context struct {
	Config config.Config
	Bot    datastore.BotSpec
}

// Program Flow

// Init loads and validates one bot setup.
func Init(logger *slog.Logger, root string, cfg config.Config, sweepID, botID uint64) (Context, error) {
	bot, err := datastore.LoadBot(config.Rooted(root, cfg.Paths.SweepDatabase), sweepID, botID)
	if err != nil {
		return Context{}, fmt.Errorf("load bot: %w", err)
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

// Domain Helpers

func within(root, path string) (string, error) {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve shared_data %s: %w", root, err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve data path %s: %w", path, err)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("data path is outside shared_data: %s", path)
	}
	return path, nil
}
