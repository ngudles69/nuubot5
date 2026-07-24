package risk

import (
	"fmt"

	"nuubot/internal/toolkit/logging"
)

type balanced struct {
	log         *logging.Logger
	number      int
	assessments uint64
	stopped     bool
}

// Section 1 - Program Flow

func createBalanced(log *logging.Logger, number int) *balanced {
	// create risk
	log.Info(fmt.Sprintf("risk initialized risk=%d kind=balanced", number))
	return &balanced{log: log, number: number}
}

func (r *balanced) AssessStop() bool {
	// record assessment
	r.assessments++
	return false
}

func (r *balanced) Stop() {
	if r.stopped {
		return
	}
	// stop risk
	r.stopped = true
	r.log.Info(fmt.Sprintf(
		"risk stopped risk=%d assessments=%d exits_requested=0",
		r.number,
		r.assessments,
	))
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
