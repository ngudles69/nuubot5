package signaler

import (
	"fmt"
	"time"

	"nuubot/internal/ohlcv"
	"nuubot/internal/toolkit/logging"
)

type macross struct {
	state          signalerState
	signalInterval ohlcv.Interval
	regimeInterval ohlcv.Interval
	fastPeriod     int
	slowPeriod     int
	regimePeriod   int
}

// Section 1 - Program Flow

// Init prepares Macross and its complete Signal history.
func (m *macross) Init(
	log *logging.Logger,
	cfg Config,
	symbol string,
	source string,
	start time.Time,
	end time.Time,
) error {
	// configure macross
	var err = m.configure(cfg)
	if err != nil {
		return err
	}

	// load macross data
	var loaded []Series
	var rows int
	loaded, rows, err = loadSeries(source, start, end, m.requirements())
	if err != nil {
		return err
	}

	// calculate macross signals
	var packages []Package
	packages, err = m.Calculate(symbol, loaded)
	if err != nil {
		return fmt.Errorf("calculate macross signals: %w", err)
	}
	return m.state.init(log, cfg.Kind, symbol, packages, len(loaded), rows)
}

// Stop stops Macross.
func (m *macross) Stop() {
	m.state.Stop()
}

// Section 2 - Domain Helpers

// Signals returns Macross Signal history.
func (m *macross) Signals(symbol string, atMS uint64, count int) []Package {
	return m.state.Signals(symbol, atMS, count)
}

// Calculate calculates one Signal package for every admitted signal bar.
func (m *macross) Calculate(symbol string, loaded []Series) ([]Package, error) {
	// find rows
	var signalBars, err = findRows(loaded, m.signalInterval)
	if err != nil {
		return nil, err
	}
	var regimeBars *Series
	regimeBars, err = findRows(loaded, m.regimeInterval)
	if err != nil {
		return nil, err
	}
	var signalDuration time.Duration
	signalDuration, err = m.signalInterval.Duration()
	if err != nil {
		return nil, err
	}
	var regimeDuration time.Duration
	regimeDuration, err = m.regimeInterval.Duration()
	if err != nil {
		return nil, err
	}
	var signalDurationMS = uint64(signalDuration.Milliseconds())
	var regimeDurationMS = uint64(regimeDuration.Milliseconds())

	// calculate emas
	var fast = ema(signalBars.Close, m.fastPeriod)
	var slow = ema(signalBars.Close, m.slowPeriod)
	var regime = ema(regimeBars.Close, m.regimePeriod)

	// align regime
	var aligned = make([]float64, len(signalBars.Close))
	var ready = make([]bool, len(signalBars.Close))
	var regimeRow int
	var latest float64
	var hasLatest bool
	for row := range signalBars.StartMS {
		var signalBoundary = signalBars.StartMS[row] + signalDurationMS
		for regimeRow < len(regimeBars.StartMS) &&
			regimeBars.StartMS[regimeRow]+regimeDurationMS <= signalBoundary {
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
	var packages = make(
		[]Package,
		0,
		max(0, len(signalBars.Close)-signalBars.PriorRows),
	)
	for row := signalBars.PriorRows; row < len(signalBars.Close); row++ {
		var regimeName = "unknown"
		if ready[row] && signalBars.Close[row] > aligned[row] {
			regimeName = "bull"
		} else if ready[row] && signalBars.Close[row] < aligned[row] {
			regimeName = "bear"
		}

		var action = NoAction
		if ready[row] && row+1 >= m.slowPeriod {
			if fast[row] > slow[row] && regimeName == "bull" {
				action = StartCycle
			} else {
				action = StopCycle
			}
		}

		var signalPackage Package
		signalPackage, err = CreatePackage(
			symbol,
			signalBars.StartMS[row]+signalDurationMS,
			action,
			regimeName,
			0,
			map[string]any{
				"bar_start_ms": float64(signalBars.StartMS[row]),
				"signal_price": signalBars.Close[row],
				"fast_ma":      fast[row],
				"slow_ma":      slow[row],
				"regime_ma":    aligned[row],
			},
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, signalPackage)
	}
	return packages, nil
}

func (m *macross) configure(cfg Config) error {
	// parse intervals
	var signalInterval, err = ohlcv.ParseInterval(cfg.SignalTimeframe)
	if err != nil {
		return err
	}
	var regimeInterval ohlcv.Interval
	regimeInterval, err = ohlcv.ParseInterval(cfg.RegimeTimeframe)
	if err != nil {
		return err
	}

	// validate config
	if signalInterval == regimeInterval ||
		cfg.FastMA <= 0 ||
		cfg.FastMA >= cfg.SlowMA ||
		cfg.RegimeEMA <= 0 {
		return fmt.Errorf("invalid macross config")
	}
	m.signalInterval = signalInterval
	m.regimeInterval = regimeInterval
	m.fastPeriod = cfg.FastMA
	m.slowPeriod = cfg.SlowMA
	m.regimePeriod = cfg.RegimeEMA
	return nil
}

func (m *macross) requirements() []Requirement {
	return []Requirement{
		{Interval: m.signalInterval, PriorRows: m.slowPeriod + 10},
		{Interval: m.regimeInterval, PriorRows: m.regimePeriod + 10},
	}
}

// Section 3 - Generic Helpers

func ema(values []float64, period int) []float64 {
	var result = make([]float64, len(values))
	if len(values) == 0 {
		return result
	}
	var alpha = 2 / float64(period+1)
	result[0] = values[0]
	for index := 1; index < len(values); index++ {
		result[index] = alpha*values[index] + (1-alpha)*result[index-1]
	}
	return result
}
