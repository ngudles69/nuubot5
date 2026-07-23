package replay

import (
	"fmt"
	"time"

	"nuubot/internal/market"
	"nuubot/internal/ohlcv"
	"nuubot/internal/toolkit/logging"
)

// Reader streams validated BBO values through OHLCV.
type Reader struct {
	log         *logging.Logger
	rows        *ohlcv.Reader
	ticksLoaded uint64
	firstMS     uint64
	lastMS      uint64
	failed      bool
	stopped     bool
}

// Section 1 - Program Flow

// NewReader opens one streaming six-column OHLCV reader.
func NewReader(log *logging.Logger, source string, start, end time.Time) (*Reader, error) {
	var rows, err = ohlcv.Open(source, ohlcv.Second1, start, end)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("tick reader initialized interval=%s", ohlcv.Second1))
	return &Reader{log: log, rows: rows}, nil
}

// Next returns the next validated BBO.
func (r *Reader) Next() (market.BBO, bool, error) {
	var row, ok, err = r.rows.Next()
	if err != nil {
		r.failed = true
		return market.BBO{}, false, err
	}
	if !ok {
		return market.BBO{}, false, nil
	}
	var bbo market.BBO
	bbo, err = market.NewBBO(row.StartMS+1000, row.Close)
	if err != nil {
		r.failed = true
		return market.BBO{}, false, fmt.Errorf("convert OHLCV row to BBO: %w", err)
	}
	r.ticksLoaded++
	if r.firstMS == 0 {
		r.firstMS = bbo.TimestampMS
	}
	r.lastMS = bbo.TimestampMS
	return bbo, true, nil
}

// Stop closes OHLCV and reports final replay statistics.
func (r *Reader) Stop() error {
	if r.stopped {
		return nil
	}
	r.stopped = true
	var err = r.rows.Close()
	r.log.Info(fmt.Sprintf(
		"tick reader stopped ticks=%d first_ts_ms=%d last_ts_ms=%d",
		r.ticksLoaded,
		r.firstMS,
		r.lastMS,
	))
	return err
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
