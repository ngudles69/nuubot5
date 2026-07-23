package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"nuubot5/internal/btrunner"
	"nuubot5/internal/config"
	"nuubot5/internal/logging"
)

// Program Flow

func main() {
	os.Exit(program(os.Args[1:]))
}

func program(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: nuubot-btrunner <sweep_id> <bot_id>")
		return 1
	}
	sweepID, err := positiveID(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	botID, err := positiveID(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("get working directory: %w", err))
		return 1
	}
	cfg, err := config.Load(filepath.Join(root, "config.toml"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	logDir := config.Rooted(root, cfg.Paths.Logs)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("create log directory %s: %w", logDir, err))
		return 1
	}
	logPath := filepath.Join(logDir, fmt.Sprintf("nuubot5-bot-%d-%d.log", sweepID, botID))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("open log %s: %w", logPath, err))
		return 1
	}
	defer file.Close()
	logger := logging.New(io.MultiWriter(os.Stdout, file)).With(
		"sweep_id", sweepID,
		"bot_id", botID,
	)

	if err := run(logger, root, cfg, sweepID, botID); err != nil {
		logger.Error(
			"program failed",
			"component", "nuubot-btrunner",
			"event", "run",
			"status", "failed",
			"error", err,
		)
		return 1
	}
	return 0
}

func run(logger *slog.Logger, root string, cfg config.Config, sweepID, botID uint64) error {
	runner, err := btrunner.New(logger, root, cfg, sweepID, botID)
	if err != nil {
		return fmt.Errorf("create btrunner: %w", err)
	}
	if err := runner.Start(); err != nil {
		return fmt.Errorf("start btrunner: %w", err)
	}
	runErr := runner.Run()
	stopErr := runner.Stop()
	if runErr != nil {
		return fmt.Errorf("run btrunner: %w", runErr)
	}
	if stopErr != nil {
		return fmt.Errorf("stop btrunner: %w", stopErr)
	}
	return nil
}

// Generic Helpers

func positiveID(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid positive id: %s", value)
	}
	return id, nil
}
