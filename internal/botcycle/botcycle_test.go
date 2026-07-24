package botcycle

import (
	"bytes"
	"errors"
	"testing"

	"nuubot/internal/config"
	"nuubot/internal/market"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestBotCycleDispatchesObserverBBO(t *testing.T) {
	var signal = cycleSignal(t, true)
	var cycle Control
	var err = cycle.Init(
		logging.Create(&bytes.Buffer{}),
		1,
		nil,
		signal,
		Inputs{},
		[]config.Executor{{Kind: "observer", StopLossPct: "0.01"}},
	)
	if err != nil {
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
	completed, err = cycle.Run(4_000)
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
	var cycle Control
	var err = cycle.Init(
		logging.Create(&bytes.Buffer{}),
		1,
		nil,
		signal,
		Inputs{},
		[]config.Executor{{Kind: "observer", StopLossPct: "0.01"}},
	)
	if !errors.Is(err, ErrRejected) {
		t.Fatalf("actual error %v, expected admission rejection", err)
	}
}

// Section 2 - Domain Helpers

func cycleSignal(t *testing.T, enterLong bool) signaler.Package {
	t.Helper()
	var signal, err = signaler.CreatePackage(
		"BTC",
		2_000,
		enterLong,
		false,
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
