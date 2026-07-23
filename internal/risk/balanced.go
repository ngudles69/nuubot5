package risk

import "nuubot5/internal/common"

type balanced struct {
	log         *common.Logger
	number      int
	assessments uint64
	stopped     bool
}

func newBalanced(log *common.Logger, number int) *balanced {
	log.Info("risk", "init risk=%d kind=balanced", number)
	return &balanced{log: log, number: number}
}

func (r *balanced) Assess() bool {
	r.assessments++
	return false
}

func (r *balanced) Stop() {
	if r.stopped {
		return
	}
	r.stopped = true
	r.log.Info("risk", "stop status=success risk=%d assessments=%d exits_requested=0", r.number, r.assessments)
}
