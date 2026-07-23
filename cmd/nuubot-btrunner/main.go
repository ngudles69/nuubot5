package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"nuubot5/internal/btrunner"
	"nuubot5/internal/common"
	"nuubot5/internal/config"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Parse identity.
	if len(args) != 2 {
		return fmt.Errorf("usage: nuubot-btrunner <sweep_id> <bot_id>")
	}
	sweepID, err := positiveID(args[0])
	if err != nil {
		return err
	}
	botID, err := positiveID(args[1])
	if err != nil {
		return err
	}

	// Load installation.
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	cfg, err := config.Load(filepath.Join(root, "config.toml"))
	if err != nil {
		return err
	}

	// Start one logger.
	logDir := config.Rooted(root, cfg.Paths.Logs)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create log directory %s: %w", logDir, err)
	}
	logPath := filepath.Join(logDir, fmt.Sprintf("nuubot5-bot-%d-%d.log", sweepID, botID))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log %s: %w", logPath, err)
	}
	defer file.Close()
	logger := common.NewLogger(io.MultiWriter(os.Stdout, file))

	// Run lifecycle.
	runner, err := btrunner.New(logger, root, cfg, sweepID, botID)
	if err != nil {
		return err
	}
	if err := runner.Start(); err != nil {
		return err
	}
	runErr := runner.Run()
	stopErr := runner.Stop()
	if runErr != nil {
		return runErr
	}
	return stopErr
}

func positiveID(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid positive ID: %s", value)
	}
	return id, nil
}
