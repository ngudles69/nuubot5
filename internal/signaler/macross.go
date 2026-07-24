package signaler

import (
	"fmt"

	"nuubot/internal/config"
	"nuubot/internal/ohlcv"
)

type macross struct {
	signalInterval ohlcv.Interval
	regimeInterval ohlcv.Interval
	fastPeriod     int
	slowPeriod     int
	regimePeriod   int
}

// Section 1 - Program Flow

func createMacross(cfg config.Signaler) (*macross, error) {
	// parse intervals
	signalInterval, err := ohlcv.ParseInterval(cfg.SignalTimeframe)
	if err != nil {
		return nil, err
	}
	regimeInterval, err := ohlcv.ParseInterval(cfg.RegimeTimeframe)
	if err != nil {
		return nil, err
	}
	// validate intervals
	if signalInterval == regimeInterval {
		return nil, fmt.Errorf("macross signal and regime timeframes must differ")
	}
	return &macross{
		signalInterval: signalInterval,
		regimeInterval: regimeInterval,
		fastPeriod:     cfg.FastMA,
		slowPeriod:     cfg.SlowMA,
		regimePeriod:   cfg.RegimeEMA,
	}, nil
}

func (m *macross) Requirements() []Requirement {
	// create requirements
	return []Requirement{
		{Interval: m.signalInterval, PriorRows: m.slowPeriod + 10},
		{Interval: m.regimeInterval, PriorRows: m.regimePeriod + 10},
	}
}

func (m *macross) Calculate(loaded []Series) ([]Signal, error) {
	// find rows
	signalBars, err := findRows(loaded, m.signalInterval)
	if err != nil {
		return nil, err
	}
	regimeBars, err := findRows(loaded, m.regimeInterval)
	if err != nil {
		return nil, err
	}
	// calculate emas
	fast := ema(signalBars.Close, m.fastPeriod)
	slow := ema(signalBars.Close, m.slowPeriod)
	regime := ema(regimeBars.Close, m.regimePeriod)

	// align regime
	aligned := make([]float64, len(signalBars.Close))
	ready := make([]bool, len(signalBars.Close))
	regimeRow := 0
	var latest float64
	hasLatest := false
	for row := 0; row+1 < len(signalBars.StartMS); row++ {
		var signalBoundary = signalBars.StartMS[row+1]
		for regimeRow+1 < len(regimeBars.StartMS) &&
			regimeBars.StartMS[regimeRow+1] <= signalBoundary {
			if regimeRow+1 >= m.regimePeriod {
				latest = regime[regimeRow]
				hasLatest = true
			}
			regimeRow++
		}
		aligned[row] = latest
		ready[row] = hasLatest
	}

	// calculate signals
	signals := make([]Signal, 0, 64)
	for row := signalBars.PriorRows; row+1 < len(signalBars.Close); row++ {
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
			SignalMS: signalBars.StartMS[row], AvailableMS: signalBars.StartMS[row+1],
			Side: side, Price: signalBars.Close[row],
		})
	}
	return signals, nil
}

// Section 2 - Domain Helpers

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

func findRows(loaded []Series, interval ohlcv.Interval) (*Series, error) {
	for index := range loaded {
		if loaded[index].Interval == interval {
			return &loaded[index], nil
		}
	}
	return nil, fmt.Errorf("signaler missing %s OHLCV", interval)
}

// Section 3 - Generic Helpers
