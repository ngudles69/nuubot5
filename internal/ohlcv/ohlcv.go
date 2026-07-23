package ohlcv

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/metadata"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

type Interval string

const (
	Second1 Interval = "1s"
	Hour1   Interval = "1h"
	Hour4   Interval = "4h"
)

type Row struct {
	StartMS uint64
	Open    float64
	High    float64
	Low     float64
	Close   float64
	Volume  float64
}

type Data struct {
	Interval Interval
	StartMS  []uint64
	Open     []float64
	High     []float64
	Low      []float64
	Close    []float64
	Volume   []float64
}

type Reader struct {
	interval Interval
	duration time.Duration
	startUS  int64
	endUS    int64
	expected uint64
	files    []string
	nextFile int
	file     *file.Reader
	records  pqarrow.RecordReader
	starts   *array.Int64
	opens    *array.Float64
	highs    *array.Float64
	lows     *array.Float64
	closes   *array.Float64
	volumes  *array.Float64
	nextRow  int
	rows     int
	count    uint64
	firstMS  uint64
	lastMS   uint64
	complete bool
	closed   bool
}

// Section 1 - Program Flow

// Open creates one streaming six-column OHLCV reader.
func Open(source string, interval Interval, start, end time.Time) (*Reader, error) {
	var duration, err = interval.Duration()
	if err != nil {
		return nil, err
	}
	if !start.Before(end) || end.Sub(start)%duration != 0 {
		return nil, fmt.Errorf("invalid %s range: %s..%s", interval, start, end)
	}
	var path = filepath.Join(filepath.Dir(source), string(interval))
	var files = monthFiles(path, interval, start, end)
	for _, filePath := range files {
		var info os.FileInfo
		info, err = os.Stat(filePath)
		if err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("OHLCV parquet not found: %s", filePath)
		}
	}
	return &Reader{
		interval: interval,
		duration: duration,
		startUS:  start.UnixMicro(),
		endUS:    end.UnixMicro(),
		expected: uint64(end.Sub(start) / duration),
		files:    files,
	}, nil
}

// Load returns one complete validated OHLCV range.
func Load(source string, interval Interval, start, end time.Time) (Data, error) {
	var data = Data{Interval: interval}
	var reader, err = Open(source, interval, start, end)
	if err != nil {
		return data, err
	}
	defer reader.Close()

	for {
		var row Row
		var ok bool
		row, ok, err = reader.Next()
		if err != nil {
			return Data{}, err
		}
		if !ok {
			break
		}
		data.StartMS = append(data.StartMS, row.StartMS)
		data.Open = append(data.Open, row.Open)
		data.High = append(data.High, row.High)
		data.Low = append(data.Low, row.Low)
		data.Close = append(data.Close, row.Close)
		data.Volume = append(data.Volume, row.Volume)
	}
	err = reader.Close()
	if err != nil {
		return Data{}, err
	}
	return data, nil
}

// Next returns the next validated OHLCV row.
func (r *Reader) Next() (Row, bool, error) {
	for {
		if r.nextRow == r.rows {
			var err = r.readBatch()
			if err != nil {
				return Row{}, false, err
			}
			if r.rows == 0 {
				err = r.verify()
				return Row{}, false, err
			}
		}

		var index = r.nextRow
		r.nextRow++
		var openUS = r.starts.Value(index)
		if openUS < r.startUS || openUS >= r.endUS {
			continue
		}
		var row, err = r.admit(index)
		if err != nil {
			return Row{}, false, err
		}
		return row, true, nil
	}
}

// Close releases the active Parquet reader and file.
func (r *Reader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	return r.closeFile()
}

// Section 2 - Domain Helpers

func ParseInterval(value string) (Interval, error) {
	switch Interval(value) {
	case Second1, Hour1, Hour4:
		return Interval(value), nil
	default:
		return "", fmt.Errorf("unknown OHLCV interval: %s", value)
	}
}

func (interval Interval) Duration() (time.Duration, error) {
	switch interval {
	case Second1:
		return time.Second, nil
	case Hour1:
		return time.Hour, nil
	case Hour4:
		return 4 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown OHLCV interval: %s", interval)
	}
}

func (r *Reader) admit(index int) (Row, error) {
	var openUS = r.starts.Value(index)
	var durationMS = uint64(r.duration.Milliseconds())
	var startMS, err = normalizeStart(openUS, durationMS)
	if err != nil {
		return Row{}, err
	}
	var row = Row{
		StartMS: startMS,
		Open:    r.opens.Value(index),
		High:    r.highs.Value(index),
		Low:     r.lows.Value(index),
		Close:   r.closes.Value(index),
		Volume:  r.volumes.Value(index),
	}
	if !valid(row.Open, row.High, row.Low, row.Close, row.Volume) {
		return Row{}, fmt.Errorf("invalid OHLCV row start_ts_ms=%d", row.StartMS)
	}
	if r.count > 0 && row.StartMS != r.lastMS+durationMS {
		return Row{}, fmt.Errorf(
			"OHLCV sequence expected %d, received %d",
			r.lastMS+durationMS,
			row.StartMS,
		)
	}
	if r.count == 0 {
		r.firstMS = row.StartMS
	}
	r.lastMS = row.StartMS
	r.count++
	return row, nil
}

func (r *Reader) verify() error {
	if r.complete {
		return nil
	}
	r.complete = true
	var expectedFirst = uint64(r.startUS / 1000)
	var expectedLast = uint64(r.endUS/1000) - uint64(r.duration.Milliseconds())
	if r.count != r.expected || r.firstMS != expectedFirst || r.lastMS != expectedLast {
		return fmt.Errorf(
			"%s range mismatch rows=%d/%d first=%d/%d last=%d/%d",
			r.interval,
			r.count,
			r.expected,
			r.firstMS,
			expectedFirst,
			r.lastMS,
			expectedLast,
		)
	}
	return nil
}

func (r *Reader) readBatch() error {
	for {
		if r.records == nil {
			var ready, err = r.openFile()
			if err != nil {
				return err
			}
			if !ready {
				r.rows = 0
				return nil
			}
		}
		if r.records.Next() {
			return r.setBatch(r.records.RecordBatch())
		}
		var err = r.records.Err()
		if err != nil {
			return fmt.Errorf("read OHLCV record batch: %w", err)
		}
		err = r.closeFile()
		if err != nil {
			return err
		}
	}
}

func (r *Reader) openFile() (bool, error) {
	if r.nextFile == len(r.files) {
		return false, nil
	}
	var path = r.files[r.nextFile]
	r.nextFile++
	var parquetFile, err = file.OpenParquetFile(path, false)
	if err != nil {
		return false, fmt.Errorf("open OHLCV parquet %s: %w", path, err)
	}
	var names = []string{"open_time_us", "open", "high", "low", "close", "volume"}
	var columns = make([]int, len(names))
	for index, name := range names {
		columns[index] = parquetFile.MetaData().Schema.ColumnIndexByName(name)
		if columns[index] < 0 {
			parquetFile.Close()
			return false, fmt.Errorf("OHLCV parquet %s missing %s", path, name)
		}
	}
	var groups []int
	groups, err = overlappingRowGroups(parquetFile, columns[0], r.startUS, r.endUS)
	if err != nil {
		parquetFile.Close()
		return false, fmt.Errorf("select OHLCV row groups %s: %w", path, err)
	}
	var reader *pqarrow.FileReader
	reader, err = pqarrow.NewFileReader(
		parquetFile,
		pqarrow.ArrowReadProperties{BatchSize: 122_880},
		memory.NewGoAllocator(),
	)
	if err != nil {
		parquetFile.Close()
		return false, fmt.Errorf("create OHLCV arrow reader %s: %w", path, err)
	}
	var records pqarrow.RecordReader
	records, err = reader.GetRecordReader(context.Background(), columns, groups)
	if err != nil {
		parquetFile.Close()
		return false, fmt.Errorf("create OHLCV record reader %s: %w", path, err)
	}
	r.file = parquetFile
	r.records = records
	return true, nil
}

func (r *Reader) setBatch(record arrow.Record) error {
	var err error
	r.starts, err = intColumn(record, "open_time_us")
	if err != nil {
		return err
	}
	r.opens, err = floatColumn(record, "open")
	if err != nil {
		return err
	}
	r.highs, err = floatColumn(record, "high")
	if err != nil {
		return err
	}
	r.lows, err = floatColumn(record, "low")
	if err != nil {
		return err
	}
	r.closes, err = floatColumn(record, "close")
	if err != nil {
		return err
	}
	r.volumes, err = floatColumn(record, "volume")
	if err != nil {
		return err
	}
	if r.starts.NullN()+r.opens.NullN()+r.highs.NullN()+
		r.lows.NullN()+r.closes.NullN()+r.volumes.NullN() != 0 {
		return fmt.Errorf("OHLCV parquet contains null")
	}
	r.nextRow = 0
	r.rows = r.starts.Len()
	return nil
}

func (r *Reader) closeFile() error {
	if r.records != nil {
		r.records.Release()
		r.records = nil
	}
	r.starts = nil
	r.opens = nil
	r.highs = nil
	r.lows = nil
	r.closes = nil
	r.volumes = nil
	if r.file == nil {
		return nil
	}
	var err = r.file.Close()
	r.file = nil
	if err != nil {
		return fmt.Errorf("close OHLCV parquet: %w", err)
	}
	return nil
}

func overlappingRowGroups(
	parquetFile *file.Reader,
	startColumn int,
	startUS int64,
	endUS int64,
) ([]int, error) {
	var groups = make([]int, 0, parquetFile.NumRowGroups())
	for index := 0; index < parquetFile.NumRowGroups(); index++ {
		var chunk, err = parquetFile.MetaData().RowGroup(index).ColumnChunk(startColumn)
		if err != nil {
			return nil, err
		}
		var set bool
		set, err = chunk.StatsSet()
		if err != nil {
			return nil, err
		}
		if !set {
			groups = append(groups, index)
			continue
		}
		var stats metadata.TypedStatistics
		stats, err = chunk.Statistics()
		if err != nil {
			return nil, err
		}
		var times, ok = stats.(*metadata.Int64Statistics)
		if !ok {
			return nil, fmt.Errorf("open_time_us statistics must be int64")
		}
		if !times.HasMinMax() || times.Max() >= startUS && times.Min() < endUS {
			groups = append(groups, index)
		}
	}
	return groups, nil
}

func monthFiles(path string, interval Interval, start, end time.Time) []string {
	var symbol = filepath.Base(filepath.Dir(path))
	var month = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	var files = make([]string, 0, 4)
	for month.Before(end) {
		files = append(files, filepath.Join(path, fmt.Sprintf(
			"%s-%s-%04d-%02d.parquet",
			symbol, interval, month.Year(), month.Month(),
		)))
		month = month.AddDate(0, 1, 0)
	}
	return files
}

func normalizeStart(openUS int64, durationMS uint64) (uint64, error) {
	if openUS < 0 || openUS%1000 != 0 {
		return 0, fmt.Errorf("open_time_us must be a non-negative whole millisecond: %d", openUS)
	}
	var startMS = uint64(openUS) / 1000
	if startMS%durationMS != 0 {
		return 0, fmt.Errorf("open_time_us is not interval aligned: %d", openUS)
	}
	return startMS, nil
}

func valid(open, high, low, close, volume float64) bool {
	var values = [...]float64{open, high, low, close, volume}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return open > 0 && low > 0 && volume >= 0 &&
		high >= math.Max(open, close) && low <= math.Min(open, close)
}

// Section 3 - Generic Helpers

func intColumn(record arrow.Record, name string) (*array.Int64, error) {
	var indices = record.Schema().FieldIndices(name)
	if len(indices) != 1 {
		return nil, fmt.Errorf("OHLCV parquet requires unique %s", name)
	}
	var column, ok = record.Column(indices[0]).(*array.Int64)
	if !ok {
		return nil, fmt.Errorf("%s must be int64", name)
	}
	return column, nil
}

func floatColumn(record arrow.Record, name string) (*array.Float64, error) {
	var indices = record.Schema().FieldIndices(name)
	if len(indices) != 1 {
		return nil, fmt.Errorf("OHLCV parquet requires unique %s", name)
	}
	var column, ok = record.Column(indices[0]).(*array.Float64)
	if !ok {
		return nil, fmt.Errorf("%s must be float64", name)
	}
	return column, nil
}
