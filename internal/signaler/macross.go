package signaler

import (
	"fmt"

	"nuubot5/internal/bars"
	"nuubot5/internal/config"
)

type macross struct {
	signalTimeframe bars.Timeframe
	regimeTimeframe bars.Timeframe
	fastPeriod      int
	slowPeriod      int
	regimePeriod    int
}

func newMacross(cfg config.Signaler) (*macross, error) {
	signalTimeframe, err := bars.ParseTimeframe(cfg.SignalTimeframe)
	if err != nil {
		return nil, err
	}
	regimeTimeframe, err := bars.ParseTimeframe(cfg.RegimeTimeframe)
	if err != nil {
		return nil, err
	}
	if signalTimeframe == regimeTimeframe {
		return nil, fmt.Errorf("macross signal and regime timeframes must differ")
	}
	return &macross{
		signalTimeframe: signalTimeframe,
		regimeTimeframe: regimeTimeframe,
		fastPeriod:      cfg.FastMA,
		slowPeriod:      cfg.SlowMA,
		regimePeriod:    cfg.RegimeEMA,
	}, nil
}

func (m *macross) BarsNeeded() []bars.Requirement {
	return []bars.Requirement{
		{Timeframe: m.signalTimeframe, PriorBars: m.slowPeriod + 10},
		{Timeframe: m.regimeTimeframe, PriorBars: m.regimePeriod + 10},
	}
}

func (m *macross) Calculate(loaded []bars.Data) ([]Signal, error) {
	signalBars, err := findBars(loaded, m.signalTimeframe)
	if err != nil {
		return nil, err
	}
	regimeBars, err := findBars(loaded, m.regimeTimeframe)
	if err != nil {
		return nil, err
	}
	fast := ema(signalBars.Close, m.fastPeriod)
	slow := ema(signalBars.Close, m.slowPeriod)
	regime := ema(regimeBars.Close, m.regimePeriod)

	aligned := make([]float64, len(signalBars.Close))
	ready := make([]bool, len(signalBars.Close))
	regimeRow := 0
	var latest float64
	hasLatest := false
	for row, signalEnd := range signalBars.EndMS {
		for regimeRow < len(regimeBars.EndMS) && regimeBars.EndMS[regimeRow] <= signalEnd {
			if regimeRow+1 >= m.regimePeriod {
				latest = regime[regimeRow]
				hasLatest = true
			}
			regimeRow++
		}
		aligned[row] = latest
		ready[row] = hasLatest
	}

	signals := make([]Signal, 0, 64)
	for row := signalBars.PriorBars; row < len(signalBars.Close); row++ {
		if !ready[row] || row == 0 || row+1 < m.slowPeriod {
			continue
		}
		previous := fast[row-1] - slow[row-1]
		current := fast[row] - slow[row]
		var side Side
		if previous <= 0 && current > 0 && signalBars.Close[row] > aligned[row] {
			side = Long
		} else if previous >= 0 && current < 0 && signalBars.Close[row] < aligned[row] {
			side = Short
		} else {
			continue
		}
		signals = append(signals, Signal{
			SignalMS: signalBars.StartMS[row], AvailableMS: signalBars.EndMS[row],
			Side: side, Price: signalBars.Close[row],
		})
	}
	return signals, nil
}

func ema(values []float64, period int) []float64 {
	result := make([]float64, len(values))
	if len(values) == 0 {
		return result
	}
	alpha := 2 / float64(period+1)
	result[0] = values[0]
	for index := 1; index < len(values); index++ {
		result[index] = alpha*values[index] + (1-alpha)*result[index-1]
	}
	return result
}

func findBars(loaded []bars.Data, timeframe bars.Timeframe) (*bars.Data, error) {
	for index := range loaded {
		if loaded[index].Timeframe == timeframe {
			return &loaded[index], nil
		}
	}
	return nil, fmt.Errorf("signaler missing %s bars", timeframe)
}
