package replay

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"nuubot5/internal/common"
	"nuubot5/internal/market"

	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

const batchSize = 65_536

type Reader struct {
	log         *common.Logger
	files       []string
	nextFile    int
	file        *file.Reader
	records     pqarrow.RecordReader
	times       *array.Int64
	prices      *array.Float64
	nextRow     int
	rows        int
	startUS     uint64
	endUS       uint64
	lastMS      uint64
	hasLast     bool
	filesLoaded uint64
	ticksLoaded uint64
	firstMS     uint64
	failed      bool
	stopped     bool
}

func NewReader(logger *common.Logger, ticksPath string, start, end time.Time) (*Reader, error) {
	files := monthFiles(ticksPath, start, end)
	for _, path := range files {
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("replay parquet not found: %s", path)
		}
	}
	logger.Info("tickreader", "init files=%d batch_size=%d", len(files), batchSize)
	return &Reader{
		log:     logger,
		files:   files,
		startUS: uint64(start.UnixMicro()),
		endUS:   uint64(end.UnixMicro()),
	}, nil
}

func (r *Reader) Next() (market.BBO, bool, error) {
	for {
		if r.nextRow >= r.rows {
			if err := r.readBatch(); err != nil {
				r.failed = true
				return market.BBO{}, false, err
			}
			if r.rows == 0 {
				return market.BBO{}, false, nil
			}
		}

		closeTime := r.times.Value(r.nextRow)
		price := r.prices.Value(r.nextRow)
		r.nextRow++
		if closeTime < 0 {
			r.failed = true
			return market.BBO{}, false, fmt.Errorf("close_time_us must be non-negative")
		}
		closeUS := uint64(closeTime)
		if closeUS < r.startUS || closeUS >= r.endUS {
			continue
		}
		tick, err := admitTick(r.lastMS, r.hasLast, closeUS, price)
		if err != nil {
			r.failed = true
			return market.BBO{}, false, err
		}
		r.lastMS = tick.TimestampMS
		r.hasLast = true
		r.ticksLoaded++
		if r.firstMS == 0 {
			r.firstMS = tick.TimestampMS
		}
		return tick, true, nil
	}
}

func (r *Reader) Stop() error {
	if r.stopped {
		return nil
	}
	r.stopped = true
	err := r.closeFile()
	status := "success"
	if r.failed || err != nil {
		status = "failed"
	}
	r.log.Info(
		"tickreader", "stop status=%s files=%d/%d ticks=%d first_ts_ms=%d last_ts_ms=%d",
		status, r.filesLoaded, len(r.files), r.ticksLoaded, r.firstMS, r.lastMS,
	)
	return err
}

func (r *Reader) readBatch() error {
	for {
		if r.records == nil {
			ready, err := r.openFile()
			if err != nil {
				return err
			}
			if !ready {
				r.rows = 0
				return nil
			}
		}
		if r.records.Next() {
			record := r.records.RecordBatch()
			timeFields := record.Schema().FieldIndices("close_time_us")
			priceFields := record.Schema().FieldIndices("close")
			if len(timeFields) != 1 || len(priceFields) != 1 {
				return fmt.Errorf("Parquet requires unique close_time_us and close columns")
			}
			times, ok := record.Column(timeFields[0]).(*array.Int64)
			if !ok {
				return fmt.Errorf("close_time_us must be Int64")
			}
			prices, ok := record.Column(priceFields[0]).(*array.Float64)
			if !ok {
				return fmt.Errorf("close must be Float64")
			}
			if times.NullN() != 0 || prices.NullN() != 0 || times.Len() != prices.Len() {
				return fmt.Errorf("Parquet BBO contains null or unequal columns")
			}
			r.times = times
			r.prices = prices
			r.nextRow = 0
			r.rows = times.Len()
			return nil
		}
		if err := r.records.Err(); err != nil {
			return fmt.Errorf("read Parquet record batch: %w", err)
		}
		if err := r.closeFile(); err != nil {
			return err
		}
	}
}

func (r *Reader) openFile() (bool, error) {
	if r.nextFile >= len(r.files) {
		return false, nil
	}
	path := r.files[r.nextFile]
	r.nextFile++
	parquetFile, err := file.OpenParquetFile(path, false)
	if err != nil {
		return false, fmt.Errorf("open Parquet %s: %w", path, err)
	}
	schema := parquetFile.MetaData().Schema
	timeIndex := schema.ColumnIndexByName("close_time_us")
	priceIndex := schema.ColumnIndexByName("close")
	if timeIndex < 0 || priceIndex < 0 {
		parquetFile.Close()
		return false, fmt.Errorf("Parquet requires close_time_us and close columns")
	}
	arrowReader, err := pqarrow.NewFileReader(
		parquetFile,
		pqarrow.ArrowReadProperties{BatchSize: batchSize},
		memory.NewGoAllocator(),
	)
	if err != nil {
		parquetFile.Close()
		return false, fmt.Errorf("create Arrow reader %s: %w", path, err)
	}
	records, err := arrowReader.GetRecordReader(
		context.Background(),
		[]int{timeIndex, priceIndex},
		nil,
	)
	if err != nil {
		parquetFile.Close()
		return false, fmt.Errorf("create Parquet record reader %s: %w", path, err)
	}
	r.file = parquetFile
	r.records = records
	r.filesLoaded++
	return true, nil
}

func (r *Reader) closeFile() error {
	if r.records != nil {
		r.records.Release()
		r.records = nil
	}
	r.times = nil
	r.prices = nil
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

func admitTick(lastMS uint64, hasLast bool, closeUS uint64, price float64) (market.BBO, error) {
	fraction := closeUS % 1_000_000
	if fraction < 999_000 || fraction > 999_999 {
		return market.BBO{}, fmt.Errorf("1s close_time_us must end in 999000..=999999: %d", closeUS)
	}
	seconds := closeUS / 1_000_000
	if seconds >= math.MaxUint64/1000 {
		return market.BBO{}, fmt.Errorf("close_time_us normalization overflow")
	}
	timestampMS := (seconds + 1) * 1000
	if hasLast && (lastMS > math.MaxUint64-1000 || timestampMS != lastMS+1000) {
		return market.BBO{}, fmt.Errorf("1s sequence expected %d, received %d", lastMS+1000, timestampMS)
	}
	return market.NewBBO(timestampMS, price)
}

func monthFiles(path string, start, end time.Time) []string {
	market := filepath.Base(filepath.Dir(path))
	month := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	files := make([]string, 0, 3)
	for month.Before(end) {
		name := fmt.Sprintf("%s-1s-%04d-%02d.parquet", market, month.Year(), month.Month())
		files = append(files, filepath.Join(path, name))
		month = month.AddDate(0, 1, 0)
	}
	return files
}
