package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"nuubot/internal/backtest"
)

const program = "nuubot-bt-bot"

// Section 1 - Program Flow

func main() {
	var options, err = parseArguments(os.Args[1:])
	if err == nil {
		err = backtest.Execute(context.Background(), options)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Section 2 - Domain Helpers

func parseArguments(args []string) (backtest.Options, error) {
	var options = backtest.Options{Output: os.Stdout}
	if len(args) != 2 && len(args) != 4 {
		return options, fmt.Errorf("usage: %s <sweep_id> <bot_id> [-pp profile_prefix]", program)
	}

	// parse Sweep ID
	var err error
	options.SweepID, err = positiveID(args[0])
	if err != nil {
		return options, fmt.Errorf("parse Sweep ID: %w", err)
	}

	// parse Bot ID
	options.BotID, err = positiveID(args[1])
	if err != nil {
		return options, fmt.Errorf("parse Bot ID: %w", err)
	}

	// parse profile prefix
	if len(args) == 4 {
		if args[2] != "-pp" {
			return options, fmt.Errorf("invalid performance profile flag: %s", args[2])
		}
		options.ProfilePrefix = args[3]
		if options.ProfilePrefix == "" {
			return options, fmt.Errorf("profile prefix is empty")
		}
	}
	return options, nil
}

func positiveID(value string) (uint64, error) {
	var id, err = strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid positive ID: %s", value)
	}
	return id, nil
}

// Section 3 - Generic Helpers
