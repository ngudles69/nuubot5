package botspec

import (
	"nuubot/internal/executor"
	"nuubot/internal/signaler"
)

// ControllerSpec contains one Bot's Controller specification.
type ControllerSpec struct {
	MaxCycles uint64
}

// RiskSpec contains one Bot's Risk specification.
type RiskSpec struct {
	Kind string
}

// Spec contains one validated and shaped Bot specification.
type Spec struct {
	ID         string
	Controller ControllerSpec
	Signaler   signaler.Config
	Risks      []RiskSpec
	Executors  []executor.Spec
}

// Section 1 - Program Flow

// Build validates and shapes exact BotConfig TOML into one BotSpec.
func Build(botSpecID, configTOML string) (Spec, error) {
	var spec, err = build(botSpecID, configTOML)
	if err != nil {
		return Spec{}, err
	}
	spec.ID = botSpecID
	return spec, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
