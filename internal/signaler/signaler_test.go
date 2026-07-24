package signaler

import (
	"testing"

	"nuubot/internal/config"
	"nuubot/internal/ohlcv"
)

// Section 1 - Program Flow

func TestMacrossUsesOnlyClosedRegimeBars(t *testing.T) {
	strategy, err := createMacross(config.Signaler{
		SignalTimeframe: "1h", RegimeTimeframe: "4h",
		FastMA: 2, SlowMA: 3, RegimeEMA: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	loaded := []Series{
		testRows(ohlcv.Hour1, []float64{10, 9, 8, 9, 10, 10}, []uint64{10, 11, 12, 13, 14, 15}),
		testRows(ohlcv.Hour4, []float64{5, 5, 5}, []uint64{0, 1, 2}),
	}
	signals, err := strategy.Calculate(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 1 || signals[0].AvailableMS != 15 || signals[0].Side != Long {
		t.Fatalf("unexpected signals: %+v", signals)
	}

	loaded[1].StartMS[2] = 16
	signals, err = strategy.Calculate(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 0 {
		t.Fatalf("future regime bar produced signals: %+v", signals)
	}
}

func TestRSIRequiresVolumeConfirmation(t *testing.T) {
	strategy, err := createRSI(config.Signaler{SignalTimeframe: "1h", RSIPeriod: 2, VolumePeriod: 2})
	if err != nil {
		t.Fatal(err)
	}
	data := testRows(ohlcv.Hour1, []float64{100, 90, 80, 80}, []uint64{1, 2, 3, 4})
	data.Volume = []float64{1, 1, 2, 0}
	signals, err := strategy.Calculate([]Series{data})
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 1 || signals[0].AvailableMS != 4 || signals[0].Side != Long {
		t.Fatalf("unexpected signals: %+v", signals)
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
