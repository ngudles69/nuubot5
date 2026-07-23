package signaler

import (
	"nuubot/internal/config"
	"nuubot/internal/ohlcv"
)

type rsi struct {
	interval     ohlcv.Interval
	rsiPeriod    int
	volumePeriod int
}

// Section 1 - Program Flow

func newRSI(cfg config.Signaler) (*rsi, error) {
	interval, err := ohlcv.ParseInterval(cfg.SignalTimeframe)
	if err != nil {
		return nil, err
	}
	return &rsi{interval: interval, rsiPeriod: cfg.RSIPeriod, volumePeriod: cfg.VolumePeriod}, nil
}

func (r *rsi) Requirements() []Requirement {
	prior := max(r.rsiPeriod, r.volumePeriod) + 10
	return []Requirement{{Interval: r.interval, PriorRows: prior}}
}

func (r *rsi) Calculate(loaded []Series) ([]Signal, error) {
	data, err := findRows(loaded, r.interval)
	if err != nil {
		return nil, err
	}
	rsiValues := relativeStrength(data.Close, r.rsiPeriod)
	volumeAverage := simpleMovingAverage(data.Volume, r.volumePeriod)
	ready := max(r.rsiPeriod, r.volumePeriod)
	signals := make([]Signal, 0, 64)
	var previous Side
	for row := data.PriorRows; row+1 < len(data.Close); row++ {
		var side Side
		if row+1 >= ready && data.Volume[row] > volumeAverage[row] {
			if rsiValues[row] <= 30 {
				side = Long
			} else if rsiValues[row] >= 70 {
				side = Short
			}
		}
		if side != "" && side != previous {
			signals = append(signals, Signal{
				SignalMS: data.StartMS[row], AvailableMS: data.StartMS[row+1],
				Side: side, Price: data.Close[row],
			})
		}
		previous = side
	}
	return signals, nil
}

// Section 2 - Domain Helpers

func relativeStrength(values []float64, period int) []float64 {
	result := make([]float64, len(values))
	if len(values) == 0 {
		return result
	}
	alpha := 2 / float64(period+1)
	upEMA, downEMA := 0.1, 0.1
	result[0] = 50
	for index := 1; index < len(values); index++ {
		up, down := 0.0, 0.0
		if values[index] > values[index-1] {
			up = values[index] - values[index-1]
		} else {
			down = values[index-1] - values[index]
		}
		upEMA = alpha*up + (1-alpha)*upEMA
		downEMA = alpha*down + (1-alpha)*downEMA
		result[index] = 100 * upEMA / (upEMA + downEMA)
	}
	return result
}

// Section 3 - Generic Helpers

func simpleMovingAverage(values []float64, period int) []float64 {
	result := make([]float64, len(values))
	window := make([]float64, period)
	sum := 0.0
	for index, value := range values {
		position := index % period
		sum += value - window[position]
		window[position] = value
		result[index] = sum / float64(min(index+1, period))
	}
	return result
}
