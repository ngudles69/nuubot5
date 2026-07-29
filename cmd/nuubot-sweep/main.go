package main

import (
	"fmt"
	"os"
	"path/filepath"

	"nuubot/internal/btsweep"
)

const program = "nuubot-sweep"

// Section 1 - Program Flow

func main() {
	// parse input
	var templatePath, err = parseInput(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// create Sweep and Bots
	var created btsweep.Creation
	created, err = btsweep.Create(templatePath, filepath.FromSlash("workspace/db/nuubot.db"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// print created identities
	fmt.Printf("SweepID %d\n", created.SweepID)
	for _, bot := range created.Bots {
		fmt.Printf("BotID %d BotNo %d\n", bot.ID, bot.Number)
	}
}

// Section 2 - Domain Helpers

func parseInput(args []string) (string, error) {
	if len(args) != 2 || args[0] != "-f" || args[1] == "" {
		return "", fmt.Errorf("usage: %s -f <sweep_template>", program)
	}
	return args[1], nil
}

// Section 3 - Generic Helpers
