package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	// ServerLog receives failures before program identity is established.
	ServerLog = "server.log"
	logDir    = "workspace/logs"
)

// Section 1 - Program Flow

// New returns the process logger.
func New(output io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(output, nil))
}

// Open returns an append-only file logger.
func Open(name string) (*slog.Logger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory %s: %w", logDir, err)
	}
	path := filepath.Join(logDir, name)
	output, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log %s: %w", path, err)
	}
	return New(output), nil
}

// OpenBot opens one identity-bound Bot logger.
func OpenBot(sweepID, botID uint64) (*slog.Logger, error) {
	log, err := Open(BotLog(sweepID, botID))
	if err != nil {
		return nil, err
	}
	return log.With(
		"sweep_id", sweepID,
		"bot_id", botID,
	), nil
}

// BotLog returns one Bot log filename.
func BotLog(sweepID, botID uint64) string {
	return fmt.Sprintf("bot_%d_%d.log", sweepID, botID)
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
