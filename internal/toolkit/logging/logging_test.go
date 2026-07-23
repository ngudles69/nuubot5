package logging

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

// Section 1 - Program Flow

func TestLoggerFormat(t *testing.T) {
	var output bytes.Buffer
	var log = New(&output)

	log.Debug("debug")
	log.Info("info")
	log.Warning("warning")
	log.Error("error")
	log.Critical("critical")

	var lines = strings.Split(strings.TrimSpace(output.String()), "\n")
	var levels = []string{`DEBUG`, ` INFO`, `WARNING`, `ERROR`, `CRITICAL`}
	for index, level := range levels {
		var pattern = regexp.MustCompile(
			`^\d{4}-[A-Z][a-z]{2}-\d{2} \d{2}:\d{2}:\d{2} \[` +
				level + `\] \w+$`,
		)
		if !pattern.MatchString(lines[index]) {
			t.Fatalf("actual %q, expected Nuubot log format", lines[index])
		}
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
