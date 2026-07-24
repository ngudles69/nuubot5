package logging

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"time"
)

const (
	// ServerLog receives failures before program identity is established.
	ServerLog       = "server.log"
	logDir          = "workspace/logs"
	timestampFormat = "2006-Jan-02 15:04:05"
)

// Logger writes complete messages using the Nuubot log format.
type Logger struct {
	output *stdlog.Logger
}

// Section 1 - Program Flow

// Create constructs one process logger.
func Create(output io.Writer) *Logger {
	// create logger
	return &Logger{output: stdlog.New(output, "", 0)}
}

// Open returns an append-only file logger.
func Open(name string) (*Logger, error) {
	// create log directory
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory %s: %w", logDir, err)
	}
	// open log file
	path := filepath.Join(logDir, name)
	output, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log %s: %w", path, err)
	}
	// return logger
	return Create(output), nil
}

// OpenBot opens one identity-bound Bot logger.
func OpenBot(sweepID, botID uint64) (*Logger, error) {
	// open bot log
	return Open(fmt.Sprintf("bot_%d_%d.log", sweepID, botID))
}

// Debug writes one debug message.
func (l *Logger) Debug(message string) {
	l.write("DEBUG", message)
}

// Info writes one informational message.
func (l *Logger) Info(message string) {
	l.write("INFO", message)
}

// Warning writes one warning message.
func (l *Logger) Warning(message string) {
	l.write("WARNING", message)
}

// Error writes one error message.
func (l *Logger) Error(message string) {
	l.write("ERROR", message)
}

// Critical writes one critical message.
func (l *Logger) Critical(message string) {
	l.write("CRITICAL", message)
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers

func (l *Logger) write(level, message string) {
	// write record
	l.output.Printf(
		"%s [%5s] %s",
		time.Now().Format(timestampFormat),
		level,
		message,
	)
}
