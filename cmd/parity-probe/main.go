package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"nuubot/internal/parity"
	"nuubot/internal/toolkit/logging"
)

const logName = "parity-probe.log"

// Section 1 - Program Flow

func main() {
	var started = time.Now()

	// open parity log
	var log, err = logging.Open(logName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to open log file:", err)
		os.Exit(1)
	}

	// parse input
	var input parity.Input
	input, err = parity.ParseInput(os.Args[1:])
	if err != nil {
		log.Error(fmt.Sprintf("parity.ParseInput() failed: %v", err))
		os.Exit(1)
	}

	// initialize parity probe
	var root string
	root, err = os.Getwd()
	if err != nil {
		log.Error(fmt.Sprintf("os.Getwd() failed: %v", err))
		os.Exit(1)
	}
	var probe parity.Probe
	err = probe.Init(log, root, input)
	if err != nil {
		log.Error(fmt.Sprintf("parity.Probe.Init() failed: %v", err))
		os.Exit(1)
	}

	// run parity probe
	err = probe.Run(context.Background())
	if err != nil {
		log.Error(fmt.Sprintf("parity.Probe.Run() failed: %v", err))
		os.Exit(1)
	}

	// log result
	log.Info(fmt.Sprintf("parity-probe completed successfully in %s", time.Since(started)))
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
