package signaler

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"nuubot/internal/config"
	"nuubot/internal/ohlcv"
	"nuubot/internal/toolkit/logging"
)

const (
	Long  = "long"
	Short = "short"
)

// Package contains one immutable, timestamped Signal result.
type Package struct {
	symbol       string
	timestampMS  uint64
	enterLong    bool
	enterShort   bool
	closeLong    bool
	closeShort   bool
	regime       string
	riskScore    float64
	customFields map[string]any
}

type Requirement struct {
	Interval  ohlcv.Interval
	PriorRows int
}

type Series struct {
	ohlcv.Data
	PriorRows int
}

// Signaler exposes passive Signal history.
type Signaler interface {
	Signals(string, uint64, int) []Package
	Stop()
}

type implementation interface {
	Signaler
	Init(*logging.Logger, config.Signaler, string, string, time.Time, time.Time) error
}

type signalerState struct {
	log      *logging.Logger
	kind     string
	symbol   string
	packages []Package
	stopped  bool
}

// Section 1 - Program Flow

// Create selects and initializes the configured Signaler.
func Create(
	log *logging.Logger,
	cfg config.Signaler,
	symbol string,
	source string,
	start time.Time,
	end time.Time,
) (Signaler, error) {
	// select signaler
	var selected implementation
	switch cfg.Kind {
	case "macross":
		selected = &macross{}
	case "rsi":
		selected = &rsi{}
	default:
		return nil, fmt.Errorf("unknown signaler: %s", cfg.Kind)
	}

	// initialize signaler
	var err = selected.Init(log, cfg, symbol, source, start, end)
	if err != nil {
		return nil, err
	}
	return selected, nil
}

// CreatePackage creates one validated flat Signal package.
func CreatePackage(
	symbol string,
	timestampMS uint64,
	enterLong bool,
	enterShort bool,
	closeLong bool,
	closeShort bool,
	regime string,
	riskScore float64,
	custom map[string]any,
) (Package, error) {
	// validate standard fields
	if symbol == "" || timestampMS == 0 {
		return Package{}, fmt.Errorf("signal package requires symbol and timestamp")
	}
	if enterLong && enterShort {
		return Package{}, fmt.Errorf("signal package cannot enter long and short")
	}
	if math.IsNaN(riskScore) || math.IsInf(riskScore, 0) ||
		riskScore < 0 || riskScore > 100 {
		return Package{}, fmt.Errorf("signal package risk_score must be between 0 and 100")
	}

	// validate custom fields
	var customFields = make(map[string]any, len(custom))
	for name, value := range custom {
		switch name {
		case "symbol", "timestamp_ms", "enter_long", "enter_short",
			"close_long", "close_short", "regime", "risk_score":
			return Package{}, fmt.Errorf("signal package custom field is reserved: %s", name)
		}
		customFields[name] = value
	}

	// create package
	return Package{
		symbol:       symbol,
		timestampMS:  timestampMS,
		enterLong:    enterLong,
		enterShort:   enterShort,
		closeLong:    closeLong,
		closeShort:   closeShort,
		regime:       regime,
		riskScore:    riskScore,
		customFields: customFields,
	}, nil
}

// Section 2 - Domain Helpers

// Signals returns the last available Signal packages in chronological order.
func (s *signalerState) Signals(symbol string, atMS uint64, count int) []Package {
	if s.stopped || symbol != s.symbol || count <= 0 {
		return nil
	}
	var end = sort.Search(len(s.packages), func(index int) bool {
		return s.packages[index].TimestampMS() > atMS
	})
	var start = max(0, end-count)
	return s.packages[start:end]
}

// Stop stops Signal admission and reports final statistics.
func (s *signalerState) Stop() {
	if s.stopped {
		return
	}
	// stop signaler
	s.stopped = true
	s.log.Info(fmt.Sprintf(
		"signaler stopped kind=%s symbol=%s signal_packages=%d",
		s.kind,
		s.symbol,
		len(s.packages),
	))
}

// MarshalJSON returns the Package's flat JSON representation.
func (p Package) MarshalJSON() ([]byte, error) {
	var fields = map[string]any{
		"symbol":       p.symbol,
		"timestamp_ms": p.timestampMS,
		"enter_long":   p.enterLong,
		"enter_short":  p.enterShort,
		"close_long":   p.closeLong,
		"close_short":  p.closeShort,
		"regime":       p.regime,
		"risk_score":   p.riskScore,
	}
	for name, value := range p.customFields {
		fields[name] = value
	}
	return json.Marshal(fields)
}

// Symbol returns the Package symbol.
func (p Package) Symbol() string {
	return p.symbol
}

// TimestampMS returns the Package availability timestamp.
func (p Package) TimestampMS() uint64 {
	return p.timestampMS
}

// EnterLong reports the standard long-entry trigger.
func (p Package) EnterLong() bool {
	return p.enterLong
}

// EnterShort reports the standard short-entry trigger.
func (p Package) EnterShort() bool {
	return p.enterShort
}

// Bool returns one Boolean field.
func (p Package) Bool(name string) (bool, bool) {
	switch name {
	case "enter_long":
		return p.enterLong, true
	case "enter_short":
		return p.enterShort, true
	case "close_long":
		return p.closeLong, true
	case "close_short":
		return p.closeShort, true
	default:
		var value, ok = p.customFields[name].(bool)
		return value, ok
	}
}

// Number returns one numeric field.
func (p Package) Number(name string) (float64, bool) {
	switch name {
	case "timestamp_ms":
		return float64(p.timestampMS), true
	case "risk_score":
		return p.riskScore, true
	default:
		var value, ok = p.customFields[name].(float64)
		return value, ok
	}
}

// Text returns one text field.
func (p Package) Text(name string) (string, bool) {
	switch name {
	case "symbol":
		return p.symbol, true
	case "regime":
		return p.regime, true
	default:
		var value, ok = p.customFields[name].(string)
		return value, ok
	}
}

func (s *signalerState) init(
	log *logging.Logger,
	kind string,
	symbol string,
	packages []Package,
	timeframes int,
	rows int,
) error {
	// validate packages
	for index, signalPackage := range packages {
		if signalPackage.Symbol() != symbol ||
			signalPackage.TimestampMS() == 0 ||
			(index > 0 &&
				packages[index-1].TimestampMS() >= signalPackage.TimestampMS()) {
			return fmt.Errorf("signaler produced invalid package order")
		}
	}

	// initialize signaler
	s.log = log
	s.kind = kind
	s.symbol = symbol
	s.packages = packages
	log.Info(fmt.Sprintf(
		"signaler initialized kind=%s symbol=%s timeframes=%d rows_loaded=%d "+
			"signal_packages=%d",
		kind,
		symbol,
		timeframes,
		rows,
		len(packages),
	))
	return nil
}

func loadSeries(
	source string,
	start time.Time,
	end time.Time,
	requirements []Requirement,
) ([]Series, int, error) {
	var loaded = make([]Series, 0, len(requirements))
	var rowCount int
	for _, requirement := range requirements {
		var duration, err = requirement.Interval.Duration()
		if err != nil {
			return nil, 0, fmt.Errorf("resolve signaler interval: %w", err)
		}
		var loadStart = start.Add(-duration * time.Duration(requirement.PriorRows))

		// load ohlcv
		var rows ohlcv.Data
		rows, err = ohlcv.Load(source, requirement.Interval, loadStart, end)
		if err != nil {
			return nil, 0, fmt.Errorf("load signaler OHLCV: %w", err)
		}
		loaded = append(loaded, Series{Data: rows, PriorRows: requirement.PriorRows})
		rowCount += len(rows.Close)
	}
	return loaded, rowCount, nil
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
