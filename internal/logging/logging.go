package logging

import (
	"io"
	"log/slog"
)

// New returns the process logger.
func New(output io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(output, nil))
}
