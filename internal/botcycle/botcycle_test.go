package botcycle

import (
	"bytes"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"nuubot/internal/account"
	"nuubot/internal/botspec"
	appconfig "nuubot/internal/config"
	"nuubot/internal/executor"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/setup"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
)

type reconExecutor struct {
	snapshot            account.Snapshot
	consecutiveFailures uint64
	reconErr            error
	reconciles          int
	signal              signaler.Package
	status              executor.Status
}

func (e *reconExecutor) OnInit(executor.Context) error { return nil }
func (e *reconExecutor) OnStop(string) error           { return nil }
func (e *reconExecutor) Status() executor.Status       { return e.status }
func (e *reconExecutor) ExitReason() string            { return "" }
func (e *reconExecutor) Telemetry() executor.Telemetry { return executor.Telemetry{} }
func (e *reconExecutor) Result() (executor.Result, error) {
	return executor.Result{}, nil
}
func (e *reconExecutor) Reconcile(uint64, bool) (account.Snapshot, bool, uint64, error) {
	e.reconciles++
	return e.snapshot, e.reconErr == nil, e.consecutiveFailures, e.reconErr
}
func (e *reconExecutor) OnSignal(signal signaler.Package) error {
	e.signal = signal
	return nil
}

// Section 1 - Program Flow

func TestBotCycleReconHasNoPartialSnapshotBarrier(t *testing.T) {
	var first = &reconExecutor{
		consecutiveFailures: 1,
		reconErr:            errors.New("first Recon failed"),
		status:              executor.Running,
	}
	var second = &reconExecutor{
		consecutiveFailures: 2,
		reconErr:            errors.New("second Recon failed"),
		status:              executor.Running,
	}
	var third = &reconExecutor{
		snapshot: account.Snapshot{ExecutorNumber: 3, ObservedMS: 1_000},
		status:   executor.Running,
	}
	var cycle = BotCycle{
		nuubot:    botCycleNuubot(t, nil),
		executors: []executor.Executor{first, second, third},
	}
	var result, err = cycle.AcctRecon(false)
	if err == nil {
		t.Fatal("failed reconciliation barrier returned nil error")
	}
	if !result.Failed || result.MaxConsecutiveFailures != 2 || result.Snapshots != nil {
		t.Fatalf("unexpected reconciliation barrier: %+v", result)
	}
	if first.reconciles != 1 || second.reconciles != 1 || third.reconciles != 1 {
		t.Fatalf(
			"reconcile calls first=%d second=%d third=%d, expected one each",
			first.reconciles,
			second.reconciles,
			third.reconciles,
		)
	}
}

func TestBotCycleDeliversCompleteSignalPackage(t *testing.T) {
	var signal, err = signaler.CreatePackage(
		"BTC",
		2_000,
		signaler.NoAction,
		"bull",
		0,
		map[string]any{"enter_long": true, "abc": true},
	)
	if err != nil {
		t.Fatal(err)
	}
	var running = &reconExecutor{status: executor.Running}
	var configured = &reconExecutor{status: executor.Configured}
	var cycle = BotCycle{
		executors: []executor.Executor{running, configured},
		running:   true,
	}
	if _, err = cycle.Run(signal); err != nil {
		t.Fatal(err)
	}
	var enterLong, exists = running.signal.Bool("enter_long")
	if running.signal.TimestampMS() != 2_000 || !exists || !enterLong {
		t.Fatalf("running Executor received incomplete Signal")
	}
	if configured.signal.TimestampMS() != 0 {
		t.Fatalf("non-running Executor received Signal")
	}
}

func TestBotCycleDispatchesObserverBBO(t *testing.T) {
	var signal = cycleSignal(t, true)
	var cycle BotCycle
	var nuubot = botCycleNuubot(t, []botspec.ExecutorSpec{{
		ID: "observer", Kind: "observer", Side: executor.Long,
		Resource:    botspec.Resource{Symbol: "BTC"},
		StopLossPct: decimal.RequireFromString("0.01"),
	}})
	var err = cycle.Init(
		nuubot,
		1,
		signal,
		Inputs{LatestBBOs: map[string]market.BBO{
			"BTC": {TimestampMS: 2_000, Price: 100},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err = cycle.Start(); err != nil {
		t.Fatal(err)
	}
	var first = market.BBO{TimestampMS: 3_000, Price: 100}
	var second = market.BBO{TimestampMS: 4_000, Price: 99}
	if err = cycle.IngestBBO(first); err != nil {
		t.Fatal(err)
	}
	cycle.OnBBO(first)
	if err = cycle.IngestBBO(second); err != nil {
		t.Fatal(err)
	}
	cycle.OnBBO(second)
	var completed bool
	completed, err = cycle.Run(signal)
	if err != nil {
		t.Fatal(err)
	}
	if !completed {
		t.Fatal("observer stop loss did not complete bot cycle")
	}
	var reason string
	reason, err = cycle.Stop("completed")
	if err != nil {
		t.Fatal(err)
	}
	if reason != "stop_loss" {
		t.Fatalf("actual reason %q, expected stop_loss", reason)
	}
}

func TestBotCycleReturnsAdmissionRejection(t *testing.T) {
	var signal = cycleSignal(t, false)
	var cycle BotCycle
	var nuubot = botCycleNuubot(t, []botspec.ExecutorSpec{{
		ID: "observer", Kind: "observer", Side: "both",
		Resource:    botspec.Resource{Symbol: "BTC"},
		StopLossPct: decimal.RequireFromString("0.01"),
	}})
	var err = cycle.Init(
		nuubot,
		1,
		signal,
		Inputs{},
	)
	if !errors.Is(err, ErrRejected) {
		t.Fatalf("actual error %v, expected admission rejection", err)
	}
}

func TestBotCycleIngestsGridBBOWhileStopping(t *testing.T) {
	var cycle BotCycle
	var nuubot = botCycleNuubot(t, []botspec.ExecutorSpec{gridSpec()})
	var err = cycle.Init(
		nuubot,
		1,
		cycleSignal(t, true),
		Inputs{
			LatestBBOs: map[string]market.BBO{
				"BTC": {TimestampMS: 2_000, Price: 100},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err = cycle.Start(); err != nil {
		t.Fatal(err)
	}
	var boundary = market.BBO{TimestampMS: 3_000, Price: 105}
	if err = cycle.IngestBBO(boundary); err != nil {
		t.Fatal(err)
	}
	cycle.OnBBO(boundary)
	var after = market.BBO{TimestampMS: 4_000, Price: 104}
	if err = cycle.IngestBBO(after); err != nil {
		t.Fatal(err)
	}
	cycle.OnBBO(after)
	if _, err = cycle.Stop("completed"); err != nil {
		t.Fatal(err)
	}
	var result = cycle.Result()
	var accountResult = result.Executors[0].Account
	if accountResult == nil || accountResult.Snapshot.ObservedMS != 4_000 {
		t.Fatalf("unexpected stopping Account result: %+v", accountResult)
	}
}

// Section 2 - Domain Helpers

func cycleSignal(t *testing.T, start bool) signaler.Package {
	t.Helper()
	var action = signaler.NoAction
	if start {
		action = signaler.StartCycle
	}
	var signal, err = signaler.CreatePackage(
		"BTC",
		2_000,
		action,
		"bull",
		0,
		map[string]any{"signal_price": 100.0},
	)
	if err != nil {
		t.Fatal(err)
	}
	return signal
}

func botCycleNuubot(t *testing.T, specs []botspec.ExecutorSpec) *setup.Nuubot {
	t.Helper()
	var log = logging.Create(&bytes.Buffer{})
	var activeClock, err = clock.Create(clock.Tick)
	if err != nil {
		t.Fatalf("create BotCycle test Clock: %v", err)
	}
	if err = activeClock.Init(log, 1_000); err != nil {
		t.Fatalf("initialize BotCycle test Clock: %v", err)
	}
	return &setup.Nuubot{
		Log:   log,
		Clock: activeClock,
		App: appconfig.App{
			Hyperliquid: appconfig.Hyperliquid{MinOrderNotionalUSDC: 11},
		},
		BotSpec: botspec.Spec{Executors: specs},
		Meta: meta.Instrument{
			Network:       "testnet",
			Kind:          "perp",
			Symbol:        "BTC",
			SizeDecimals:  5,
			PriceDecimals: 1,
		},
	}
}

func gridSpec() botspec.ExecutorSpec {
	return botspec.ExecutorSpec{
		ID:   "grid",
		Role: "grid",
		Kind: "grid",
		Side: executor.Long,
		Resource: botspec.Resource{
			Venue:             "simulator",
			Network:           "simnet",
			PhysicalAccountID: "sim",
			Symbol:            "BTC",
		},
		CapitalUSDC:    decimal.NewFromInt(100),
		GridLevels:     10,
		RangePct:       decimal.RequireFromString("0.05"),
		MinExpectedPnL: decimal.Zero,
		FeePct:         decimal.RequireFromString("0.05"),
		PersistMode:    "none",
	}
}

// Section 3 - Generic Helpers
