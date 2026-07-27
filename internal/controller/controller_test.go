package controller

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/bot"
	"nuubot/internal/botcycle"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/risk"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

type testSignaler struct {
	packages []signaler.Package
}

func (s *testSignaler) Signals(_ string, atMS uint64, _ int) []signaler.Package {
	if len(s.packages) == 0 || s.packages[0].TimestampMS() > atMS {
		return nil
	}
	return s.packages
}

func (s *testSignaler) Stop() {}

type testRisk struct{}

func (testRisk) Assess(risk.Input) risk.Decision { return risk.Allow }
func (testRisk) Stop()                           {}

type decisionRisk struct {
	decision    risk.Decision
	assessments int
}

func (r *decisionRisk) Assess(risk.Input) risk.Decision {
	r.assessments++
	return r.decision
}

func (*decisionRisk) Stop() {}

type testCycle struct {
	reconResults []botcycle.ReconResult
	reconErrors  []error
	onReconErr   error
	runErr       error
	reconciles   int
	onRecons     int
	runs         int
}

func (c *testCycle) Reconcile(uint64, bool) (botcycle.ReconResult, error) {
	var index = c.reconciles
	c.reconciles++
	return c.reconResults[index], c.reconErrors[index]
}

func (c *testCycle) OnRecon(uint64) error {
	c.onRecons++
	return c.onReconErr
}

func (c *testCycle) Run(uint64) (bool, error) {
	c.runs++
	return false, c.runErr
}

func (*testCycle) Stop(string) (string, error)   { return "test", nil }
func (*testCycle) IngestBBO(market.BBO) error    { return nil }
func (*testCycle) OnBBO(market.BBO)              {}
func (*testCycle) Result() botcycle.Result       { return botcycle.Result{} }
func (*testCycle) Telemetry() botcycle.Telemetry { return botcycle.Telemetry{} }

// Section 1 - Program Flow

func TestControllerReconFailureBarrier(t *testing.T) {
	var reconErr = errors.New("Account Recon failed")
	var cycle = &testCycle{
		reconResults: []botcycle.ReconResult{
			{Failed: true, MaxConsecutiveFailures: 1},
			{Failed: true, MaxConsecutiveFailures: 2},
			{Failed: true, MaxConsecutiveFailures: 3},
		},
		reconErrors: []error{reconErr, reconErr, reconErr},
	}
	var policy = &decisionRisk{decision: risk.Allow}
	var controller = barrierController(cycle, policy)
	for failure := 1; failure <= 2; failure++ {
		var stop, err = controller.Run(uint64(failure) * 1_000)
		if err != nil || stop {
			t.Fatalf("failure %d returned stop=%t error=%v", failure, stop, err)
		}
	}
	if policy.assessments != 0 || cycle.onRecons != 0 || cycle.runs != 0 {
		t.Fatalf(
			"first two failures ran decisions risk=%d on_recon=%d run=%d",
			policy.assessments,
			cycle.onRecons,
			cycle.runs,
		)
	}
	var stop, err = controller.Run(3_000)
	if err == nil || !errors.Is(err, reconErr) || stop {
		t.Fatalf("third failure returned stop=%t error=%v", stop, err)
	}
	if policy.assessments != 0 || cycle.onRecons != 0 || cycle.runs != 0 {
		t.Fatalf(
			"third failure ran decisions risk=%d on_recon=%d run=%d",
			policy.assessments,
			cycle.onRecons,
			cycle.runs,
		)
	}
}

func TestControllerResumesAfterReconRecovery(t *testing.T) {
	var reconErr = errors.New("Account Recon failed")
	var cycle = &testCycle{
		reconResults: []botcycle.ReconResult{
			{Failed: true, MaxConsecutiveFailures: 1},
			{Failed: true, MaxConsecutiveFailures: 2},
			{},
		},
		reconErrors: []error{reconErr, reconErr, nil},
	}
	var policy = &decisionRisk{decision: risk.Allow}
	var controller = barrierController(cycle, policy)
	for pass := 1; pass <= 3; pass++ {
		var stop, err = controller.Run(uint64(pass) * 1_000)
		if err != nil || stop {
			t.Fatalf("pass %d returned stop=%t error=%v", pass, stop, err)
		}
	}
	if policy.assessments != 1 || cycle.onRecons != 1 || cycle.runs != 1 {
		t.Fatalf(
			"recovered Recon did not resume decisions risk=%d on_recon=%d run=%d",
			policy.assessments,
			cycle.onRecons,
			cycle.runs,
		)
	}
}

func TestControllerKeepsNonReconErrorsFatal(t *testing.T) {
	var executionErr = errors.New("Executor OnRecon failed")
	var cycle = &testCycle{
		reconResults: []botcycle.ReconResult{{}},
		reconErrors:  []error{nil},
		onReconErr:   executionErr,
	}
	var policy = &decisionRisk{decision: risk.Allow}
	var controller = barrierController(cycle, policy)
	var stop, err = controller.Run(1_000)
	if err == nil || !errors.Is(err, executionErr) || stop {
		t.Fatalf("ordinary execution failure returned stop=%t error=%v", stop, err)
	}
	if policy.assessments != 1 || cycle.onRecons != 1 || cycle.runs != 0 {
		t.Fatalf(
			"unexpected decision calls risk=%d on_recon=%d run=%d",
			policy.assessments,
			cycle.onRecons,
			cycle.runs,
		)
	}
}

func TestControllerReusesCurrentStartActionAfterCycleCompletes(t *testing.T) {
	var start, err = signaler.CreatePackage(
		"BTC",
		1_000,
		signaler.StartCycle,
		"bull",
		0,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var controller Controller
	err = controller.Init(logging.Create(&bytes.Buffer{}), bot.Definition{
		Identity: bot.Identity{
			SweepID: 1, BotID: 2, BotSpecID: "test",
			ConfigTOML: "bot_spec = \"test\"\n", ConfigHash: "hash",
		},
		SignalSymbol: "BTC",
		MaxCycles:    2,
		Signaler:     &testSignaler{packages: []signaler.Package{start}},
		Risks:        []risk.Risk{testRisk{}},
		Executors: []executor.Spec{{
			ID: "observer", Kind: "observer", Side: executor.Long,
			Resource:    executor.Resource{Symbol: "BTC"},
			StopLossPct: decimal.RequireFromString("0.01"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = controller.Start(); err != nil {
		t.Fatal(err)
	}
	if err = controller.IngestBBO(market.BBO{
		Symbol: "BTC", TimestampMS: 2_000, Price: 100,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = controller.Run(2_000); err != nil {
		t.Fatal(err)
	}
	if err = controller.IngestBBO(market.BBO{
		Symbol: "BTC", TimestampMS: 2_500, Price: 100,
	}); err != nil {
		t.Fatal(err)
	}
	if err = controller.IngestBBO(market.BBO{
		Symbol: "BTC", TimestampMS: 3_000, Price: 99,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = controller.Run(3_000); err != nil {
		t.Fatal(err)
	}
	if _, err = controller.Run(4_000); err != nil {
		t.Fatal(err)
	}
	var firstTelemetry = controller.Telemetry()
	var secondTelemetry = controller.Telemetry()
	if !reflect.DeepEqual(firstTelemetry, secondTelemetry) {
		t.Fatal("Controller telemetry changed owned state")
	}
	if err = controller.Stop("test"); err != nil {
		t.Fatal(err)
	}
	if actual := len(controller.Result().Cycles); actual != 2 {
		t.Fatalf("actual cycles %d, expected 2", actual)
	}
}

func TestControllerAssessesEveryRiskBeforeActing(t *testing.T) {
	var start, err = signaler.CreatePackage(
		"BTC",
		1_000,
		signaler.StartCycle,
		"bull",
		0,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var cycleRisk = &decisionRisk{decision: risk.Allow}
	var controllerRisk = &decisionRisk{decision: risk.Allow}
	var controller Controller
	err = controller.Init(logging.Create(&bytes.Buffer{}), bot.Definition{
		Identity: bot.Identity{
			SweepID: 1, BotID: 2, BotSpecID: "test",
			ConfigTOML: "bot_spec = \"test\"\n", ConfigHash: "hash",
		},
		SignalSymbol: "BTC",
		MaxCycles:    2,
		Signaler:     &testSignaler{packages: []signaler.Package{start}},
		Risks:        []risk.Risk{cycleRisk, controllerRisk},
		Executors: []executor.Spec{{
			ID: "observer", Kind: "observer", Side: executor.Long,
			Resource:    executor.Resource{Symbol: "BTC"},
			StopLossPct: decimal.RequireFromString("0.01"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = controller.Start(); err != nil {
		t.Fatal(err)
	}
	if err = controller.IngestBBO(market.BBO{
		Symbol: "BTC", TimestampMS: 2_000, Price: 100,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = controller.Run(2_000); err != nil {
		t.Fatal(err)
	}

	cycleRisk.decision = risk.StopCycle
	controllerRisk.decision = risk.StopController
	var stop bool
	stop, err = controller.Run(3_000)
	if err != nil {
		t.Fatal(err)
	}
	if !stop || cycleRisk.assessments != 2 || controllerRisk.assessments != 2 {
		t.Fatalf(
			"stop=%t cycle_assessments=%d controller_assessments=%d",
			stop,
			cycleRisk.assessments,
			controllerRisk.assessments,
		)
	}
	if err = controller.Stop("parent_stop"); err != nil {
		t.Fatal(err)
	}
	if actual := controller.Result().ExitReason; actual != "risk" {
		t.Fatalf("exit reason %q, expected risk", actual)
	}
}

func TestStopCycleRiskBlocksCycleStart(t *testing.T) {
	var start, err = signaler.CreatePackage(
		"BTC",
		1_000,
		signaler.StartCycle,
		"bull",
		0,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var policy = &decisionRisk{decision: risk.StopCycle}
	var controller Controller
	err = controller.Init(logging.Create(&bytes.Buffer{}), bot.Definition{
		Identity: bot.Identity{
			SweepID: 1, BotID: 2, BotSpecID: "test",
			ConfigTOML: "bot_spec = \"test\"\n", ConfigHash: "hash",
		},
		SignalSymbol: "BTC",
		MaxCycles:    2,
		Signaler:     &testSignaler{packages: []signaler.Package{start}},
		Risks:        []risk.Risk{policy},
		Executors: []executor.Spec{{
			ID: "observer", Kind: "observer", Side: executor.Long,
			Resource:    executor.Resource{Symbol: "BTC"},
			StopLossPct: decimal.RequireFromString("0.01"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = controller.Start(); err != nil {
		t.Fatal(err)
	}
	if err = controller.IngestBBO(market.BBO{
		Symbol: "BTC", TimestampMS: 2_000, Price: 100,
	}); err != nil {
		t.Fatal(err)
	}
	var stop bool
	stop, err = controller.Run(2_000)
	if err != nil {
		t.Fatal(err)
	}
	if stop {
		t.Fatal("StopCycle requested Controller stop")
	}
	if err = controller.Stop("test"); err != nil {
		t.Fatal(err)
	}
	if actual := len(controller.Result().Cycles); actual != 0 {
		t.Fatalf("actual cycles %d, expected 0", actual)
	}
}

// Section 2 - Domain Helpers

func barrierController(cycle cycleControl, policy risk.Risk) Controller {
	return Controller{
		definition: bot.Definition{
			Signaler: &testSignaler{},
			Risks:    []risk.Risk{policy},
		},
		cycle:          cycle,
		lastRisk:       make([]risk.Decision, 1),
		resourceEquity: make(map[executor.Resource]decimal.Decimal),
		started:        true,
	}
}

// Section 3 - Generic Helpers
