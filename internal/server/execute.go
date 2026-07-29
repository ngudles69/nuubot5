// Package server owns the Nuubot Server application lifecycle.
package server

import (
	"context"
	"fmt"
	"io"

	"nuubot/internal/server/webserver"
	"nuubot/internal/toolkit/logging"
)

// Options defines one Server execution request.
type Options struct {
	Address string
	Output  io.Writer
}

// Section 1 - Program Flow

// Execute runs one Server until cancellation or WebServer failure.
func Execute(caller context.Context, log *logging.Logger, options Options) error {
	// Step 1: validate Server options
	if caller == nil || log == nil || options.Address == "" || options.Output == nil {
		return fmt.Errorf("execute server requires complete options")
	}

	// Step 2: create WebServer
	var web, err = webserver.Create(log, options.Address, options.Output)
	if err != nil {
		return fmt.Errorf("execute server: %w", err)
	}

	// Step 3: run WebServer
	log.Info("server execute started")
	err = web.Run(caller)
	if err != nil {
		return fmt.Errorf("execute server: %w", err)
	}

	// Step 4: log execute completed
	log.Info("server execute completed")
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
