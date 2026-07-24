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
	stopped     bool
}

// Section 1 - Program Flow

// Init prepares one streaming six-column OHLCV reader.
func (r *Reader) Init(log *logging.Logger, source string, start, end time.Time) error {
	r.log = log

	// open ohlcv
	var err error
	r.rows, err = ohlcv.Open(source, ohlcv.Second1, start, end)
	if err != nil {
		return err
	}

	// initialize reader
	log.Info(fmt.Sprintf("tick reader initialized interval=%s", ohlcv.Second1))
	return nil
}

// Next returns the next validated BBO.
func (r *Reader) Next() (market.BBO, bool, error) {
	// read next ohlcv
	var row, ok, err = r.rows.Next()
	if err != nil {
		return market.BBO{}, false, err
	}
	if !ok {
		return market.BBO{}, false, nil
	}
	// create bbo
	var bbo market.BBO
	bbo, err = market.CreateBBO(row.StartMS+1000, row.Close)
	if err != nil {
		return market.BBO{}, false, fmt.Errorf("convert OHLCV row to BBO: %w", err)
	}
	// record proof
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
	// close ohlcv
	r.stopped = true
	var err = r.rows.Close()

	// report proof
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
