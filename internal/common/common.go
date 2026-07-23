package common

import (
	"fmt"
	"io"
	"log"
)

type Logger struct {
	logger *log.Logger
}

func NewLogger(output io.Writer) *Logger {
	return &Logger{logger: log.New(output, "", log.LstdFlags|log.Lmicroseconds)}
}

func (l *Logger) Info(component, format string, args ...any) {
	l.logger.Printf(component+": "+format, args...)
}

func StateError(owner, action string) error {
	return fmt.Errorf("%s cannot %s from current state", owner, action)
}

func Duration(start, end uint64) uint64 {
	if end < start {
		return 0
	}
	return end - start
}
