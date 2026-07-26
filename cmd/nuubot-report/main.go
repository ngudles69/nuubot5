package main

import (
	"fmt"
	"os"
	"strconv"

	"nuubot/internal/report"
)

const program = "nuubot-report"

// Section 1 - Program Flow

func main() {
	var err = run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 5 {
		return fmt.Errorf(
			"usage: %s <requested> <sweep_id> <bot_id> <suite_elapsed_ms> <json_path>",
			program,
		)
	}
	var requested, err = positiveInt(args[0])
	if err != nil {
		return fmt.Errorf("parse requested runs: %v", err)
	}
	var sweepID uint64
	sweepID, err = positiveUint64(args[1])
	if err != nil {
		return fmt.Errorf("parse sweep id: %v", err)
	}
	var botID uint64
	botID, err = positiveUint64(args[2])
	if err != nil {
		return fmt.Errorf("parse bot id: %v", err)
	}
	var suiteElapsedMS int64
	suiteElapsedMS, err = strconv.ParseInt(args[3], 10, 64)
	if err != nil || suiteElapsedMS < 0 {
		return fmt.Errorf("parse suite elapsed: invalid non-negative integer")
	}
	var attempts []report.Attempt
	attempts, err = report.ReadAttempts(os.Stdin)
	if err != nil {
		return err
	}
	var suite report.Suite
	suite, err = report.BuildSuite(report.SuiteInput{
		Requested:      requested,
		SweepID:        sweepID,
		BotID:          botID,
		SuiteElapsedMS: suiteElapsedMS,
		Attempts:       attempts,
	})
	if err != nil {
		return err
	}
	err = report.WriteSuiteJSON(args[4], suite)
	if err != nil {
		return err
	}
	err = report.WriteTable(os.Stdout, suite)
	if err != nil {
		return err
	}
	if suite.Status != "pass" {
		return fmt.Errorf("SuiteReport failed")
	}
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers

func positiveInt(value string) (int, error) {
	var result, err = strconv.Atoi(value)
	if err != nil || result <= 0 {
		return 0, fmt.Errorf("invalid positive integer: %s", value)
	}
	return result, nil
}

func positiveUint64(value string) (uint64, error) {
	var result, err = strconv.ParseUint(value, 10, 64)
	if err != nil || result == 0 {
		return 0, fmt.Errorf("invalid positive integer: %s", value)
	}
	return result, nil
}
