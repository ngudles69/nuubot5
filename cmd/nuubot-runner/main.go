package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"nuubot/internal/runner"
	"nuubot/internal/toolkit/logging"
)

const program = "nuubot-runner"

// Section 1 - Program Flow

func main() {
	var started = time.Now()

	// open server.log
	var log, err = logging.Open(logging.ServerLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to open log file:", err)
		os.Exit(1)
	}

	// parse input
	var sweepID, botID uint64
	sweepID, botID, err = parseInput(os.Args[1:])
	if err != nil {
		log.Error(fmt.Sprintf("parseInput() failed: %v", err))
		os.Exit(1)
	}

	// open bot_<identity>.log
	var botLog *logging.Logger
	botLog, err = logging.OpenBotLog(sweepID, botID)
	if err != nil {
		log.Error(fmt.Sprintf("logging.OpenBotLog() failed: %v", err))
		os.Exit(1)
	}
	log = botLog

	// create runner
	var live runner.Runner

	// initialize runner
	err = live.Init(context.Background(), log, sweepID, botID)
	if err != nil {
		log.Error(fmt.Sprintf("runner.Init() failed: %v", err))
		os.Exit(1)
	}

	// start runner
	err = live.Start()
	if err != nil {
		err = errors.Join(err, live.Stop())
		log.Error(fmt.Sprintf("runner.Start() failed: %v", err))
		os.Exit(1)
	}

	// loop runner
	err = live.Loop()

	// stop runner
	var stopErr = live.Stop()
	if err != nil {
		log.Error(fmt.Sprintf("runner.Loop() failed: %v", errors.Join(err, stopErr)))
		os.Exit(1)
	}
	if stopErr != nil {
		log.Error(fmt.Sprintf("runner.Stop() failed: %v", stopErr))
		os.Exit(1)
	}

	// log result
	log.Info(fmt.Sprintf("runner completed successfully in %s", time.Since(started)))
}

// Section 2 - Domain Helpers

func parseInput(args []string) (uint64, uint64, error) {
	if len(args) != 2 {
		return 0, 0, fmt.Errorf("usage: %s <sweep_id> <bot_id>", program)
	}

	// parse sweep id
	var sweepID, err = positiveID(args[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse sweep id: %w", err)
	}

	// parse bot id
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
