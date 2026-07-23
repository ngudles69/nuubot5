package executor

import (
	"io"
	"testing"

	"nuubot5/internal/config"
	"nuubot5/internal/logging"
	"nuubot5/internal/market"
	"nuubot5/internal/signaler"
)

// Program Flow

func TestObserverRecordsStopLoss(t *testing.T) {
	executor, err := newObserver(
		logging.New(io.Discard), 1, 1,
		signaler.Signal{SignalMS: 1_000, AvailableMS: 2_000, Side: signaler.Long, Price: 100},
		config.Executor{Kind: "observer", StopLossPct: 0.01},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := executor.Start(); err != nil {
		t.Fatal(err)
	}
	executor.OnBBO(market.BBO{TimestampMS: 3_000, Price: 100})
	executor.OnBBO(market.BBO{TimestampMS: 4_000, Price: 99})
	if !executor.Terminal() || executor.ExitReason() != "stop_loss" ||
		executor.stats.startMS != 3_000 || executor.stats.endMS != 4_000 {
		t.Fatalf("unexpected observer state: %+v", executor.stats)
	}
}
