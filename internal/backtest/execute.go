package backtest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"nuubot/internal/report"
	"nuubot/internal/runharness"
	"nuubot/internal/toolkit/logging"
)

// Options defines one standalone Backtest Run request.
type Options struct {
	SweepID       uint64
	BotID         uint64
	ProfilePrefix string
	Output        io.Writer
}

// Section 1 - Program Flow

// Execute runs one complete standalone Backtest.
func Execute(caller context.Context, options Options) error {
	// Step 1: validate Run options
	if caller == nil || options.SweepID == 0 || options.BotID == 0 || options.Output == nil {
		return fmt.Errorf("execute Backtest requires complete options")
	}
	var started = time.Now()

	// Step 2: open Bot log
	var log, err = logging.OpenBotLog(options.SweepID, options.BotID)
	if err != nil {
		return fmt.Errorf("open Bot log: %w", err)
	}

	// Step 3: start whole-Run profiling
	var profile = runharness.NewProfile(options.ProfilePrefix)
	err = profile.Start()
	if err != nil {
		return fmt.Errorf("start whole-Run profile: %w", err)
	}

	// Step 4: initialize Backtest
	var run Run
	err = run.Init(caller, log, options.SweepID, options.BotID)
	if err != nil {
		return errors.Join(fmt.Errorf("initialize Backtest: %w", err), profile.Stop())
	}

	// Step 5: start Backtest
	err = run.Start()
	if err != nil {
		return errors.Join(fmt.Errorf("start Backtest: %w", err), run.Stop(), profile.Stop())
	}

	// Step 6: loop Backtest
	var loopErr = run.Loop()

	// Step 7: stop Backtest
	var stopErr = run.Stop()
	if loopErr != nil {
		return errors.Join(fmt.Errorf("loop Backtest: %w", loopErr), stopErr, profile.Stop())
	}
	if stopErr != nil {
		return errors.Join(fmt.Errorf("stop Backtest: %w", stopErr), profile.Stop())
	}

	// Step 8: write terminal report
	var result Result
	result, err = run.Result()
	if err == nil {
		err = report.WriteRunJSON(options.Output, result.Report)
	}
	if err != nil {
		return errors.Join(fmt.Errorf("write Backtest result: %w", err), profile.Stop())
	}

	// Step 9: stop whole-Run profiling
	err = profile.Stop()
	if err != nil {
		return fmt.Errorf("stop whole-Run profile: %w", err)
	}

	// Step 10: log Run completed
	log.Info(fmt.Sprintf("btbot completed successfully in %s", time.Since(started)))
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
