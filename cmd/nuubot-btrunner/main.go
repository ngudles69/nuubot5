package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"nuubot/internal/btrunner"
	"nuubot/internal/toolkit/logging"
)

const program = "nuubot-btrunner"

// Section 1 - Program Flow

func main() {
	var started = time.Now()
	var log, err = logging.Open(logging.ServerLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to open log file:", err)
		os.Exit(1)
	}

	// Parse input.
	var sweepID, botID uint64
	sweepID, botID, err = parseInput(os.Args[1:])
	if err != nil {
		log.Error("parseInput() failed", "error", err)
		os.Exit(1)
	}

	// Set log to Bot log.
	var botLog, botLogErr = logging.OpenBot(sweepID, botID)
	if botLogErr != nil {
		log.Error("logging.OpenBot() failed", "error", botLogErr)
		os.Exit(1)
	}
	log = botLog

	// Run BtRunner.
	err = btrunner.Run(log, sweepID, botID)
	if err != nil {
		log.Error("btrunner.Run() failed", "duration", time.Since(started), "error", err)
		os.Exit(1)
	}

	// Log result.
	log.Info("btrunner.Run() completed successfully", "duration", time.Since(started))
}

// Section 2 - Domain Helpers

func parseInput(args []string) (uint64, uint64, error) {
	if len(args) != 2 {
		return 0, 0, fmt.Errorf("usage: %s <sweep_id> <bot_id>", program)
	}

	// extract sweepID
	var sweepID, err = positiveID(args[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse sweep id: %w", err)
	}

	// extract botID
	var botID uint64
	botID, err = positiveID(args[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse bot id: %w", err)
	}
	return sweepID, botID, nil
}

func positiveID(value string) (uint64, error) {
	var id, err = strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid positive id: %s", value)
	}
	return id, nil
}

// Section 3 - Generic Helpers
