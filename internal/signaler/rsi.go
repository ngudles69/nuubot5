package signaler

import (
	"fmt"
	"time"

	"nuubot/internal/config"
	"nuubot/internal/ohlcv"
	"nuubot/internal/toolkit/logging"
)

type rsi struct {
	state        signalerState
	interval     ohlcv.Interval
	rsiPeriod    int
	volumePeriod int
}

// Section 1 - Program Flow

// Init prepares RSI and its complete Signal history.
func (r *rsi) Init(
	log *logging.Logger,
	cfg config.Signaler,
	symbol string,
	source string,
	start time.Time,
	end time.Time,
) error {
	// configure rsi
	var err = r.configure(cfg)
	if err != nil {
		return err
	}

	// load rsi data
	var loaded []Series
	var rows int
	loaded, rows, err = loadSeries(source, start, end, r.requirements())
	if err != nil {
		return err
	}

	// calculate rsi signals
	var packages []Package
	packages, err = r.Calculate(symbol, loaded)
	if err != nil {
		return fmt.Errorf("calculate rsi signals: %w", err)
	}
	return r.state.init(log, cfg.Kind, symbol, packages, len(loaded), rows)
}

// Stop stops RSI.
func (r *rsi) Stop() {
	r.state.Stop()
}

// Section 2 - Domain Helpers

// Signals returns RSI Signal history.
func (r *rsi) Signals(symbol string, atMS uint64, count int) []Package {
	return r.state.Signals(symbol, atMS, count)
}

// Calculate calculates one Signal package for every admitted signal bar.
func (r *rsi) Calculate(symbol string, loaded []Series) ([]Package, error) {
	// find rows
	var data, err = findRows(loaded, r.interval)
	if err != nil {
		return nil, err
	}

	// calculate indicators
	var rsiValues = relativeStrength(data.Close, r.rsiPeriod)
	var volumeAverage = simpleMovingAverage(data.Volume, r.volumePeriod)
	var readyAfter = max(r.rsiPeriod, r.volumePeriod)

	// calculate signals
	var packages = make([]Package, 0, max(0, len(data.Close)-data.PriorRows))
	var previous string
	for row := data.PriorRows; row+1 < len(data.Close); row++ {
		var volumeRatio float64
		if volumeAverage[row] > 0 {
			volumeRatio = data.Volume[row] / volumeAverage[row]
		}
		var oversold = rsiValues[row] <= 30
		var overbought = rsiValues[row] >= 70
		var side string
		if row+1 >= readyAfter && volumeRatio > 1 {
			if oversold {
				side = Long
			} else if overbought {
				side = Short
			}
		}
		var enterLong = side == Long && side != previous
		var enterShort = side == Short && side != previous

		var signalPackage Package
		signalPackage, err = CreatePackage(
			symbol,
			data.StartMS[row+1],
			enterLong,
			enterShort,
			false,
			false,
			"neutral",
			0,
			map[string]any{
				"bar_start_ms": float64(data.StartMS[row]),
				"signal_price": data.Close[row],
				"rsi":          rsiValues[row],
				"volume_ratio": volumeRatio,
				"oversold":     oversold,
				"overbought":   overbought,
			},
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, signalPackage)
		previous = side
	}
	return packages, nil
}

func (r *rsi) configure(cfg config.Signaler) error {
	// parse interval
	var interval, err = ohlcv.ParseInterval(cfg.SignalTimeframe)
	if err != nil {
		return err
	}

	// validate config
	if cfg.RSIPeriod <= 0 || cfg.VolumePeriod <= 0 {
		return fmt.Errorf("invalid rsi config")
	}
	r.interval = interval
	r.rsiPeriod = cfg.RSIPeriod
	r.volumePeriod = cfg.VolumePeriod
	return nil
}

func (r *rsi) requirements() []Requirement {
	var prior = max(r.rsiPeriod, r.volumePeriod) + 10
	return []Requirement{{Interval: r.interval, PriorRows: prior}}
}

// Section 3 - Generic Helpers

func relativeStrength(values []float64, period int) []float64 {
	var result = make([]float64, len(values))
	if len(values) == 0 {
		return result
	}
	var alpha = 2 / float64(period+1)
	var upEMA = 0.1
	var downEMA = 0.1
	result[0] = 50
	for index := 1; index < len(values); index++ {
		var up float64
		var down float64
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

func simpleMovingAverage(values []float64, period int) []float64 {
	var result = make([]float64, len(values))
	var window = make([]float64, period)
	var sum float64
	for index, value := range values {
		var position = index % period
		sum += value - window[position]
		window[position] = value
		result[index] = sum / float64(min(index+1, period))
	}
	return result
}
