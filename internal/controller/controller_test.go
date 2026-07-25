package controller

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/bot"
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

// Section 1 - Program Flow

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

// Section 3 - Generic Helpers
