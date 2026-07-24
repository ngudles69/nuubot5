// Package resultpublisher atomically publishes successful BtRunner evidence.
package resultpublisher

import (
	"fmt"
	"os"
	"path/filepath"

	"nuubot/internal/ledger"
	"nuubot/internal/runtime"
	"nuubot/internal/simulator"
)

// Section 1 - Program Flow

// Publish writes every memory-only Account result to one completed database.
func Publish(result runtime.Result) error {
	var path string
	var accounts int
	for _, cycle := range result.Cycles {
		for _, current := range cycle.Accounts {
			if current.PersistMode != ledger.None {
				continue
			}
			if path == "" {
				path = current.ResultPath
			}
			if current.ResultPath != path {
				return fmt.Errorf("publish result: Accounts use different result paths")
			}
			accounts++
		}
	}
	if accounts == 0 {
		return nil
	}

	// prepare temporary result path
	var err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("publish result: prepare directory: %v", err)
	}
	var partial = path + ".partial"
	err = os.Remove(partial)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("publish result: clear partial file: %v", err)
	}
	var published = false
	defer func() {
		if !published {
			os.Remove(partial)
		}
	}()

	// publish Account children
	for _, cycle := range result.Cycles {
		for _, current := range cycle.Accounts {
			if current.PersistMode != ledger.None {
				continue
			}
			err = ledger.Publish(partial, current.Ledger)
			if err != nil {
				return fmt.Errorf("publish result: %w", err)
			}
			if current.Simulator != nil {
				err = simulator.Publish(partial, *current.Simulator)
				if err != nil {
					return fmt.Errorf("publish result: %w", err)
				}
			}
		}
	}

	// publish completed result
	err = os.Rename(partial, path)
	if err != nil {
		return fmt.Errorf("publish result: replace completed database: %v", err)
	}
	published = true
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
