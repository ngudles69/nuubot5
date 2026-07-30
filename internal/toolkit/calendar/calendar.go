// Package calendar resolves canonical calendar period labels.
package calendar

import (
	"fmt"
	"strconv"
	"time"
)

// Section 1 - Program Flow

// ResolvePeriod resolves one canonical label into a UTC half-open range.
func ResolvePeriod(label string) (time.Time, time.Time, error) {
	switch {
	case len(label) == 4:
		return resolveYear(label)
	case len(label) == 7 && label[4:6] == "-H":
		return resolveHalf(label)
	case len(label) == 7 && label[4:6] == "-Q":
		return resolveQuarter(label)
	case len(label) == 8 && label[4:6] == "-M":
		return resolveMonth(label)
	case len(label) == 8 && label[4:6] == "-W":
		return resolveWeek(label)
	case len(label) == len(time.DateOnly):
		return resolveDay(label)
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period label %q", label)
	}
}

// Section 2 - Domain Helpers

func resolveYear(label string) (time.Time, time.Time, error) {
	var year, err = parseYear(label)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	var start = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(1, 0, 0), nil
}

func resolveHalf(label string) (time.Time, time.Time, error) {
	var year, err = parseYear(label[:4])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	var half, valid = parseNumber(label[6:], 1, 2)
	if !valid {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period label %q", label)
	}
	var month = time.Month(1 + (half-1)*6)
	var start = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 6, 0), nil
}

func resolveQuarter(label string) (time.Time, time.Time, error) {
	var year, err = parseYear(label[:4])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	var quarter, valid = parseNumber(label[6:], 1, 4)
	if !valid {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period label %q", label)
	}
	var month = time.Month(1 + (quarter-1)*3)
	var start = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 3, 0), nil
}

func resolveMonth(label string) (time.Time, time.Time, error) {
	var year, err = parseYear(label[:4])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	var month, valid = parseNumber(label[6:], 1, 12)
	if !valid {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period label %q", label)
	}
	var start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0), nil
}

func resolveWeek(label string) (time.Time, time.Time, error) {
	var year, err = parseYear(label[:4])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	var week, valid = parseNumber(label[6:], 1, 53)
	if !valid {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period label %q", label)
	}
	var januaryFourth = time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	var weekday = (int(januaryFourth.Weekday()) + 6) % 7
	var start = januaryFourth.AddDate(0, 0, -weekday+(week-1)*7)
	var actualYear, actualWeek = start.ISOWeek()
	if actualYear != year || actualWeek != week {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period label %q", label)
	}
	return start, start.AddDate(0, 0, 7), nil
}

func resolveDay(label string) (time.Time, time.Time, error) {
	var start, err = time.Parse(time.DateOnly, label)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period label %q", label)
	}
	return start, start.AddDate(0, 0, 1), nil
}

// Section 3 - Generic Helpers

func parseYear(value string) (int, error) {
	var year, err = strconv.Atoi(value)
	if err != nil || year < 1 || year > 9999 {
		return 0, fmt.Errorf("invalid period label %q", value)
	}
	return year, nil
}

func parseNumber(value string, minimum, maximum int) (int, bool) {
	var number, err = strconv.Atoi(value)
	return number, err == nil && number >= minimum && number <= maximum
}
