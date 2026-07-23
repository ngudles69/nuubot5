package bars

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

// Timeframe identifies a supported bar interval.
type Timeframe string

const (
	// Hour1 identifies one-hour bars.
	Hour1 Timeframe = "1h"
	// Hour4 identifies four-hour bars.
	Hour4 Timeframe = "4h"
)

// Requirement describes one required bar series.
type Requirement struct {
	Timeframe Timeframe
	PriorBars int
}

// Data contains one validated OHLCV series.
type Data struct {
	Timeframe Timeframe
	PriorBars int
	StartMS   []uint64
	EndMS     []uint64
	Open      []float64
	High      []float64
	Low       []float64
	Close     []float64
	Volume    []float64
}

// Section 1 - Program Flow

// Load returns all required validated bars.
func Load(
	logger *slog.Logger,
	ticksPath string,
	start time.Time,
	end time.Time,
	requirements []Requirement,
) ([]Data, error) {
	log := logger.With("component", "bars")
	marketPath := filepath.Dir(ticksPath)
	loaded := make([]Data, 0, len(requirements))
	for _, requirement := range requirements {
		data, err := loadTimeframe(
			filepath.Join(marketPath, string(requirement.Timeframe)),
			start,
			end,
			requirement,
		)
		if err != nil {
			return nil, fmt.Errorf("load %s bars: %w", requirement.Timeframe, err)
		}
		log.Info(
			"bars loaded",
			"event", "load",
			"status", "success",
			"timeframe", data.Timeframe,
			"rows", len(data.Close),
			"prior_bars", data.PriorBars,
		)
		loaded = append(loaded, data)
	}
	return loaded, nil
}

// Section 2 - Domain Helpers

// ParseTimeframe returns a supported Timeframe.
func ParseTimeframe(value string) (Timeframe, error) {
	switch Timeframe(value) {
	case Hour1, Hour4:
		return Timeframe(value), nil
	default:
		return "", fmt.Errorf("unknown timeframe: %s", value)
	}
}

func loadTimeframe(path string, start, end time.Time, requirement Requirement) (Data, error) {
	duration := requirement.Timeframe.duration()
	warmup := start.Add(-duration * time.Duration(requirement.PriorBars))
	data := Data{Timeframe: requirement.Timeframe, PriorBars: requirement.PriorBars}
	for _, path := range monthFiles(path, warmup, end) {
		if err := readFile(path, warmup, end, duration, &data); err != nil {
			return Data{}, err
		}
	}

	expected := int(end.Sub(warmup) / duration)
	expectedStart := uint64(warmup.UnixMilli())
	expectedEnd := uint64(end.UnixMilli())
	if len(data.Close) != expected || len(data.StartMS) == 0 ||
		data.StartMS[0] != expectedStart || data.EndMS[len(data.EndMS)-1] != expectedEnd {
		return Data{}, fmt.Errorf(
			"%s bars range mismatch rows=%d/%d first=%d/%d last=%d/%d",
			requirement.Timeframe,
			len(data.Close), expected,
			first(data.StartMS), expectedStart,
			last(data.EndMS), expectedEnd,
		)
	}
	return data, nil
}

func readFile(path string, start, end time.Time, duration time.Duration, data *Data) error {
	parquetFile, err := file.OpenParquetFile(path, false)
	if err != nil {
		return fmt.Errorf("open bars parquet %s: %v", path, err)
	}
	defer parquetFile.Close()

	names := []string{"open_time_us", "open", "high", "low", "close", "volume", "close_time_us"}
	columns := make([]int, len(names))
	for index, name := range names {
		columns[index] = parquetFile.MetaData().Schema.ColumnIndexByName(name)
		if columns[index] < 0 {
			return fmt.Errorf("bars parquet %s missing %s", path, name)
		}
	}
	reader, err := pqarrow.NewFileReader(
		parquetFile,
		pqarrow.ArrowReadProperties{BatchSize: 65_536},
		memory.NewGoAllocator(),
	)
	if err != nil {
		return fmt.Errorf("create bars arrow reader %s: %v", path, err)
	}
	records, err := reader.GetRecordReader(context.Background(), columns, nil)
	if err != nil {
		return fmt.Errorf("create bars record reader %s: %v", path, err)
	}
	defer records.Release()

	startUS := start.UnixMicro()
	endUS := end.UnixMicro()
	durationMS := uint64(duration.Milliseconds())
	for records.Next() {
		record := records.RecordBatch()
		starts, err := intColumn(record, "open_time_us")
		if err != nil {
			return err
		}
		ends, err := intColumn(record, "close_time_us")
		if err != nil {
			return err
		}
		opens, err := floatColumn(record, "open")
		if err != nil {
			return err
		}
		highs, err := floatColumn(record, "high")
		if err != nil {
			return err
		}
		lows, err := floatColumn(record, "low")
		if err != nil {
			return err
		}
		closes, err := floatColumn(record, "close")
		if err != nil {
			return err
		}
		volumes, err := floatColumn(record, "volume")
		if err != nil {
			return err
		}
		if starts.NullN()+ends.NullN()+opens.NullN()+highs.NullN()+lows.NullN()+closes.NullN()+volumes.NullN() != 0 {
			return fmt.Errorf("bars parquet contains null OHLCV")
		}

		for row := 0; row < int(record.NumRows()); row++ {
			openUS := starts.Value(row)
			if openUS < startUS || openUS >= endUS {
				continue
			}
			if openUS < 0 || ends.Value(row) < 0 {
				return fmt.Errorf("bars timestamps must be non-negative")
			}
			startMS := uint64(openUS) / 1000
			endMS, err := normalizeClose(uint64(ends.Value(row)))
			if err != nil {
				return err
			}
			open := opens.Value(row)
			high := highs.Value(row)
			low := lows.Value(row)
			close := closes.Value(row)
			volume := volumes.Value(row)
			if endMS != startMS+durationMS || !validOHLCV(open, high, low, close, volume) {
				return fmt.Errorf("invalid OHLCV bar start_ts_ms=%d", startMS)
			}
			if len(data.StartMS) > 0 && startMS != data.StartMS[len(data.StartMS)-1]+durationMS {
				return fmt.Errorf("bars sequence expected %d, received %d", data.StartMS[len(data.StartMS)-1]+durationMS, startMS)
			}
			data.StartMS = append(data.StartMS, startMS)
			data.EndMS = append(data.EndMS, endMS)
			data.Open = append(data.Open, open)
			data.High = append(data.High, high)
			data.Low = append(data.Low, low)
			data.Close = append(data.Close, close)
			data.Volume = append(data.Volume, volume)
		}
	}
	if err := records.Err(); err != nil {
		return fmt.Errorf("read bars parquet %s: %v", path, err)
	}
	return nil
}

func (timeframe Timeframe) duration() time.Duration {
	if timeframe == Hour1 {
		return time.Hour
	}
	return 4 * time.Hour
}

func monthFiles(path string, start, end time.Time) []string {
	market := filepath.Base(filepath.Dir(path))
	month := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	files := make([]string, 0, 4)
	for month.Before(end) {
		files = append(files, filepath.Join(path, fmt.Sprintf(
			"%s-%s-%04d-%02d.parquet",
			market, filepath.Base(path), month.Year(), month.Month(),
		)))
		month = month.AddDate(0, 1, 0)
	}
	return files
}

func intColumn(record arrow.Record, name string) (*array.Int64, error) {
	indices := record.Schema().FieldIndices(name)
	if len(indices) != 1 {
		return nil, fmt.Errorf("bars parquet requires unique %s", name)
	}
	column, ok := record.Column(indices[0]).(*array.Int64)
	if !ok {
		return nil, fmt.Errorf("%s must be Int64", name)
	}
	return column, nil
}

func floatColumn(record arrow.Record, name string) (*array.Float64, error) {
	indices := record.Schema().FieldIndices(name)
	if len(indices) != 1 {
		return nil, fmt.Errorf("bars parquet requires unique %s", name)
	}
	column, ok := record.Column(indices[0]).(*array.Float64)
	if !ok {
		return nil, fmt.Errorf("%s must be Float64", name)
	}
	return column, nil
}

func normalizeClose(closeUS uint64) (uint64, error) {
	seconds := closeUS / 1_000_000
	if seconds >= math.MaxUint64/1000 {
		return 0, fmt.Errorf("close_time_us normalization overflow")
	}
	return (seconds + 1) * 1000, nil
}

func validOHLCV(open, high, low, close, volume float64) bool {
	values := [...]float64{open, high, low, close, volume}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return open > 0 && low > 0 && volume >= 0 && high >= math.Max(open, close) && low <= math.Min(open, close)
}

func first(values []uint64) uint64 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}

func last(values []uint64) uint64 {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}

// Section 3 - Generic Helpers
