package signaler

import (
	"encoding/json"
	"math"
	"testing"

	"nuubot/internal/ohlcv"
)

// Section 1 - Program Flow

func TestMacrossUsesOnlyClosedRegimeBars(t *testing.T) {
	var strategy macross
	var err = strategy.configure(Config{
		SignalTimeframe: "1h", RegimeTimeframe: "4h",
		FastMA: 2, SlowMA: 3, RegimeEMA: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	var loaded = []Series{
		testRows(ohlcv.Hour1, []float64{10, 9, 8, 9, 10, 10}, []uint64{10, 11, 12, 13, 14, 15}),
		testRows(ohlcv.Hour4, []float64{5, 5, 5}, []uint64{0, 1, 2}),
	}
	var packages []Package
	packages, err = strategy.Calculate("BTC", loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(packages) != 5 ||
		packages[4].Action() != StartCycle ||
		packages[4].TimestampMS() != 15 {
		t.Fatalf("unexpected packages: %+v", packages)
	}

	loaded[1].StartMS[2] = 16
	packages, err = strategy.Calculate("BTC", loaded)
	if err != nil {
		t.Fatal(err)
	}
	for _, signalPackage := range packages {
		if signalPackage.Action() == StartCycle {
			t.Fatalf("future regime bar produced entry: %+v", signalPackage)
		}
	}
}

func TestRSIRequiresVolumeConfirmation(t *testing.T) {
	var strategy rsi
	var err = strategy.configure(Config{
		SignalTimeframe: "1h",
		RSIPeriod:       2,
		VolumePeriod:    2,
	})
	if err != nil {
		t.Fatal(err)
	}
	var data = testRows(ohlcv.Hour1, []float64{100, 90, 80, 80}, []uint64{1, 2, 3, 4})
	data.Volume = []float64{1, 1, 2, 0}
	var packages []Package
	packages, err = strategy.Calculate("BTC", []Series{data})
	if err != nil {
		t.Fatal(err)
	}
	if len(packages) != 3 ||
		packages[2].Action() != StartCycle ||
		packages[2].TimestampMS() != 4 {
		t.Fatalf("unexpected packages: %+v", packages)
	}
}

func TestPackageHistoryAndFlatJSON(t *testing.T) {
	var first, err = CreatePackage(
		"BTC", 100, NoAction, "bull", 24,
		map[string]any{"vol_spike": 1.3},
	)
	if err != nil {
		t.Fatal(err)
	}
	var second Package
	second, err = CreatePackage(
		"BTC", 200, StartCycle, "bull", 25,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var state = signalerState{symbol: "BTC", packages: []Package{first, second}}
	var actual = state.Signals("BTC", 200, 2)
	if len(actual) != 2 || actual[0].TimestampMS() != 100 || actual[1].TimestampMS() != 200 {
		t.Fatalf("unexpected history: %+v", actual)
	}
	var encoded []byte
	encoded, err = json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["vol_spike"] != 1.3 ||
		decoded["risk_score"] != 24.0 ||
		decoded["action"] != string(NoAction) {
		t.Fatalf("unexpected flat package: %s", encoded)
	}
}

func TestPackageRejectsInvalidRiskScore(t *testing.T) {
	var _, err = CreatePackage(
		"BTC", 100, NoAction, "bull", math.NaN(), nil,
	)
	if err == nil {
		t.Fatal("NaN risk score was accepted")
	}
}

// Section 2 - Domain Helpers

func testRows(interval ohlcv.Interval, closes []float64, starts []uint64) Series {
	return Series{
		Data: ohlcv.Data{
			Interval: interval, StartMS: starts,
			Open: closes, High: closes, Low: closes, Close: closes,
			Volume: make([]float64, len(closes)),
		},
	}
}

// Section 3 - Generic Helpers
