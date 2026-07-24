package main

import (
	"errors"
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

	// open server log
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

	// open bot log
	var botLog, botLogErr = logging.OpenBot(sweepID, botID)
	if botLogErr != nil {
		log.Error(fmt.Sprintf("logging.OpenBot() failed: %v", botLogErr))
		os.Exit(1)
	}
	log = botLog

	// create btrunner
	var runner btrunner.BtRunner

	// initialize btrunner
	err = runner.Init(log, sweepID, botID)
	if err != nil {
		log.Error(fmt.Sprintf("btrunner.Init() failed: %v", err))
		os.Exit(1)
	}

	// start btrunner
	err = runner.Start()
	if err != nil {
		err = errors.Join(err, runner.Stop())
		log.Error(fmt.Sprintf("btrunner.Start() failed: %v", err))
		os.Exit(1)
	}

	// loop btrunner
	var loopErr = runner.Loop()

	// stop btrunner
	var stopErr = runner.Stop()
	if loopErr != nil {
		log.Error(fmt.Sprintf("btrunner.Loop() failed: %v", errors.Join(loopErr, stopErr)))
		os.Exit(1)
	}
	if stopErr != nil {
		log.Error(fmt.Sprintf("btrunner.Stop() failed: %v", stopErr))
		os.Exit(1)
	}

	// log result
	log.Info(fmt.Sprintf("btrunner completed successfully in %s", time.Since(started)))
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
