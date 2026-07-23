package signaler

import (
	"testing"

	"nuubot5/internal/bars"
	"nuubot5/internal/config"
)

func TestMacrossUsesOnlyClosedRegimeBars(t *testing.T) {
	strategy, err := newMacross(config.Signaler{
		SignalTimeframe: "1h", RegimeTimeframe: "4h",
		FastMA: 2, SlowMA: 3, RegimeEMA: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	loaded := []bars.Data{
		testBars(bars.Hour1, []float64{10, 9, 8, 9, 10}, []uint64{11, 12, 13, 14, 15}),
		testBars(bars.Hour4, []float64{5, 5}, []uint64{1, 2}),
	}
	signals, err := strategy.Calculate(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 1 || signals[0].AvailableMS != 15 || signals[0].Side != Long {
		t.Fatalf("unexpected signals: %+v", signals)
	}

	loaded[1].EndMS[1] = 16
	signals, err = strategy.Calculate(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 0 {
		t.Fatalf("future regime bar produced signals: %+v", signals)
	}
}

func TestRSIRequiresVolumeConfirmation(t *testing.T) {
	strategy, err := newRSI(config.Signaler{SignalTimeframe: "1h", RSIPeriod: 2, VolumePeriod: 2})
	if err != nil {
		t.Fatal(err)
	}
	data := testBars(bars.Hour1, []float64{100, 90, 80}, []uint64{2, 3, 4})
	data.Volume = []float64{1, 1, 2}
	signals, err := strategy.Calculate([]bars.Data{data})
	if err != nil {
		t.Fatal(err)
	}
	if len(signals) != 1 || signals[0].AvailableMS != 4 || signals[0].Side != Long {
		t.Fatalf("unexpected signals: %+v", signals)
	}
}

func testBars(timeframe bars.Timeframe, closes []float64, ends []uint64) bars.Data {
	starts := make([]uint64, len(ends))
	for index, end := range ends {
		starts[index] = end - 1
	}
	return bars.Data{
		Timeframe: timeframe, StartMS: starts, EndMS: ends,
		Open: closes, High: closes, Low: closes, Close: closes,
		Volume: make([]float64, len(closes)),
	}
}
