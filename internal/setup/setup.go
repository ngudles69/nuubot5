package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"nuubot5/internal/common"
	"nuubot5/internal/config"
	"nuubot5/internal/datastore"
)

type Context struct {
	Config config.Config
	Bot    datastore.BotSpec
}

func Init(log *common.Logger, root string, cfg config.Config, sweepID, botID uint64) (Context, error) {
	bot, err := datastore.LoadBot(config.Rooted(root, cfg.Paths.SweepDatabase), sweepID, botID)
	if err != nil {
		return Context{}, err
	}
	bot.TicksPath, err = within(config.Rooted(root, cfg.Paths.SharedData), bot.TicksPath)
	if err != nil {
		return Context{}, err
	}
	log.Info("setup", "sweep_id=%d bot_id=%d symbol=%s", sweepID, botID, bot.Symbol)
	return Context{Config: cfg, Bot: bot}, nil
}

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
