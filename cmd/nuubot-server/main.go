package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"nuubot/internal/server"
	"nuubot/internal/toolkit/logging"
)

const address = "127.0.0.1:9898"

// Section 1 - Program Flow

func main() {
	// Step 1: open Server log
	var log, err = logging.Open(logging.ServerLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to open log file:", err)
		os.Exit(1)
	}

	// Step 2: create signal context
	var caller, stop = signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Step 3: execute Server
	log.Info(fmt.Sprintf("nuubot-server starting address=%s", address))
	err = server.Execute(caller, log, server.Options{
		Address: address,
		Output:  os.Stdout,
	})
	if err != nil {
		log.Error(fmt.Sprintf("server.Execute() failed: %v", err))
		os.Exit(1)
	}

	// Step 4: log program completed
	log.Info("nuubot-server completed successfully")
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
