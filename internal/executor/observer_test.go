package executor

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestObserverHandlesBBOAndRecordsStopLoss(t *testing.T) {
	var output bytes.Buffer
	var signal = testSignal(t, true, false)
	var created, err = Create(Context{
		Log:            logging.Create(&output),
		CycleNumber:    1,
		ExecutorNumber: 1,
		Signal:         signal,
		Config:         config.Executor{Kind: "observer", StopLossPct: 0.01},
	})
	if err != nil {
		t.Fatal(err)
	}
	var observer = created.(*observer)
	var first = market.BBO{TimestampMS: 3_000, Price: 100}
	var second = market.BBO{TimestampMS: 4_000, Price: 99}
	if err = observer.IngestBBO(first); err != nil {
		t.Fatal(err)
	}
	observer.OnBBO(first)
	if err = observer.IngestBBO(second); err != nil {
		t.Fatal(err)
	}
	observer.OnBBO(second)
	if observer.Status() != Stopping ||
		observer.ExitReason() != "stop_loss" ||
		observer.stats.startMS != 3_000 ||
		observer.stats.endMS != 4_000 ||
		observer.stats.ingestBBOCount != 2 ||
		observer.stats.onBBOCount != 2 {
		t.Fatalf("unexpected observer state: %+v", observer.stats)
	}
	if err = observer.OnStop("completed"); err != nil {
		t.Fatal(err)
	}
	if observer.Status() != Stopped {
		t.Fatalf("actual status %q, expected %q", observer.Status(), Stopped)
	}
	if !strings.Contains(
		output.String(),
		"ingest_bbo_count=2 on_bbo_count=2 stop_reason=stop_loss",
	) {
		t.Fatalf("missing observer counters in stop log: %s", output.String())
	}
}

func TestObserverRejectsSignalWithoutOneEntry(t *testing.T) {
	var _, err = Create(Context{
		Log:            logging.Create(&bytes.Buffer{}),
		CycleNumber:    1,
		ExecutorNumber: 1,
		Signal:         testSignal(t, false, false),
		Config:         config.Executor{Kind: "observer", StopLossPct: 0.01},
	})
	if !errors.Is(err, ErrRejected) {
		t.Fatalf("actual error %v, expected admission rejection", err)
	}
}

// Section 2 - Domain Helpers

func testSignal(t *testing.T, enterLong, enterShort bool) signaler.Package {
	t.Helper()
	var signal, err = signaler.CreatePackage(
		"BTC",
		2_000,
		enterLong,
		enterShort,
		false,
		false,
		"bull",
		0,
		map[string]any{"signal_price": 100.0},
	)
	if err != nil {
		t.Fatal(err)
	}
	return signal
}

// Section 3 - Generic Helpers
