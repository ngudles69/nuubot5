package executor

import (
	"bytes"
	"strings"
	"testing"

	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestObserverCountsBBOAndRecordsStopLoss(t *testing.T) {
	var output bytes.Buffer
	executor, err := createObserver(
		logging.Create(&output), 1, 1,
		signaler.Signal{SignalMS: 1_000, AvailableMS: 2_000, Side: signaler.Long, Price: 100},
		config.Executor{Kind: "observer", StopLossPct: 0.01},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := executor.Start(); err != nil {
		t.Fatal(err)
	}
	var first = market.BBO{TimestampMS: 3_000, Price: 100}
	var second = market.BBO{TimestampMS: 4_000, Price: 99}
	if err := executor.IngestBBO(first); err != nil {
		t.Fatal(err)
	}
	executor.OnBBO(first)
	if err := executor.IngestBBO(second); err != nil {
		t.Fatal(err)
	}
	executor.OnBBO(second)
	if !executor.Terminal() || executor.ExitReason() != "stop_loss" ||
		executor.stats.startMS != 3_000 || executor.stats.endMS != 4_000 ||
		executor.stats.ingestBBOCount != 2 || executor.stats.onBBOCount != 2 {
		t.Fatalf("unexpected observer state: %+v", executor.stats)
	}
	if err := executor.Stop("completed"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(
		output.String(),
		"ingest_bbo_count=2 on_bbo_count=2 runs=0 stop_reason=stop_loss",
	) {
		t.Fatalf("missing observer counters in stop log: %s", output.String())
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
