package live

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nuubot/internal/toolkit/logging"
)

// Options defines one standalone Live Run request.
type Options struct {
	SweepID uint64
	BotID   uint64
}

// Section 1 - Program Flow

// Execute runs one complete standalone Live Run.
func Execute(caller context.Context, options Options) error {
	// Step 1: validate Run options
	if caller == nil || options.SweepID == 0 || options.BotID == 0 {
		return fmt.Errorf("execute Live Run requires complete options")
	}
	var started = time.Now()

	// Step 2: open Bot log
	var log, err = logging.OpenBotLog(options.SweepID, options.BotID)
	if err != nil {
		return fmt.Errorf("open Bot log: %w", err)
	}

	// Step 3: initialize Live Run
	var run Run
	err = run.Init(caller, log, options.SweepID, options.BotID)
	if err != nil {
		return fmt.Errorf("initialize Live Run: %w", err)
	}

	// Step 4: start Live Run
	err = run.Start()
	if err != nil {
		return errors.Join(fmt.Errorf("start Live Run: %w", err), run.Stop())
	}

	// Step 5: loop Live Run
	var loopErr = run.Loop()

	// Step 6: stop Live Run
	var stopErr = run.Stop()
	if loopErr != nil {
		return errors.Join(fmt.Errorf("loop Live Run: %w", loopErr), stopErr)
	}
	if stopErr != nil {
		return fmt.Errorf("stop Live Run: %w", stopErr)
	}

	// Step 7: log Run completed
	log.Info(fmt.Sprintf("runner completed successfully in %s", time.Since(started)))
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
