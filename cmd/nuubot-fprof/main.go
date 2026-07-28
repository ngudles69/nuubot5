// Command nuubot-fprof runs exact A/B/C function profiling.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"nuubot/internal/fprof"
)

// Section 1 - Program Flow

func main() {
	// Step 1: parse profile request
	var sweepID = flag.Uint64("sweep", 0, "Sweep ID")
	var botID = flag.Uint64("bot", 0, "Bot ID")
	var top = flag.Int("top", 100, "maximum function rows in text report")
	flag.Parse()
	if *sweepID == 0 || *botID == 0 || *top <= 0 || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: nuubot-fprof -sweep ID -bot ID [-top N]")
		os.Exit(2)
	}

	// Step 2: run function profile
	var root, err = os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "get working directory:", err)
		os.Exit(1)
	}
	var result fprof.SessionReport
	result, err = fprof.Run(fprof.Config{
		Root:    root,
		SweepID: *sweepID,
		BotID:   *botID,
		Top:     *top,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Step 3: print final report
	var reportPath = filepath.Join(result.Session, "report.txt")
	var payload []byte
	payload, err = os.ReadFile(reportPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read function profile report:", err)
		os.Exit(1)
	}
	fmt.Print(string(payload))
	fmt.Println()
	fmt.Println("report=" + reportPath)
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
