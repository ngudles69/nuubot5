package runreport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Section 1 - Program Flow

// WriteRunJSON writes one compact machine-readable RunReport record.
func WriteRunJSON(output io.Writer, report Run) error {
	var encoder = json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	var err = encoder.Encode(report)
	if err != nil {
		return fmt.Errorf("write RunReport JSON: %v", err)
	}
	return nil
}

// ReadAttempts reads ordered JSON attempt envelopes.
func ReadAttempts(input io.Reader) ([]Attempt, error) {
	var scanner = bufio.NewScanner(input)
	var attempts []Attempt
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var attempt Attempt
		var err = json.Unmarshal(scanner.Bytes(), &attempt)
		if err != nil {
			return nil, fmt.Errorf("read SuiteReport attempt: %v", err)
		}
		attempts = append(attempts, attempt)
	}
	var err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("read SuiteReport attempts: %v", err)
	}
	return attempts, nil
}

// WriteSuiteJSON atomically writes one complete SuiteReport artifact.
func WriteSuiteJSON(path string, report Suite) error {
	if path == "" {
		return fmt.Errorf("write SuiteReport JSON: path is empty")
	}
	var err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("write SuiteReport JSON: prepare directory: %v", err)
	}
	var partial = path + ".partial"
	var output *os.File
	output, err = os.Create(partial)
	if err != nil {
		return fmt.Errorf("write SuiteReport JSON: create partial: %v", err)
	}
	var encoder = json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(report)
	var closeErr = output.Close()
	if err != nil {
		os.Remove(partial)
		return fmt.Errorf("write SuiteReport JSON: encode: %v", err)
	}
	if closeErr != nil {
		os.Remove(partial)
		return fmt.Errorf("write SuiteReport JSON: close: %v", closeErr)
	}
	err = os.Rename(partial, path)
	if err != nil {
		os.Remove(partial)
		return fmt.Errorf("write SuiteReport JSON: publish: %v", err)
	}
	return nil
}

// WriteTable writes standardized SuiteReport tables.
func WriteTable(output io.Writer, report Suite) error {
	var err error
	_, err = fmt.Fprintf(
		output,
		"%dx BtRunner — Sweep %d, Bot %d\n\n"+
			"BotSpec: %-20s Symbol: %s\n"+
			"Status: %-21s Requested: %d\n"+
			"Attempted: %-18d Passed: %d    Failed: %d\n\n",
		report.Requested,
		report.SweepID,
		report.BotID,
		report.BotSpecID,
		report.Symbol,
		strings.ToUpper(report.Status),
		report.Requested,
		report.Attempted,
		report.Passed,
		report.Failed,
	)
	if err != nil {
		return fmt.Errorf("write SuiteReport header: %v", err)
	}
	err = writeMetrics(output, "Timing (ms)", report.Timing, 1, false)
	if err != nil {
		return err
	}
	err = writeMetrics(output, "Memory (MB)", report.Memory, 3, false)
	if err != nil {
		return err
	}
	err = writeMetrics(
		output,
		"Garbage Collection (#)",
		report.GCRuns,
		1,
		false,
	)
	if err != nil {
		return err
	}
	err = writeMetrics(
		output,
		"Garbage Collection Pause (ms)",
		report.GCPause,
		3,
		false,
	)
	if err != nil {
		return err
	}
	err = writeMetrics(
		output,
		"Replay and Execution (#)",
		report.Execution,
		0,
		false,
	)
	if err != nil {
		return err
	}
	err = writeMetrics(
		output,
		"Financial Results (USDC)",
		report.PnL,
		2,
		true,
	)
	if err != nil {
		return err
	}
	return nil
}

// Section 2 - Domain Helpers

func writeMetrics(
	output io.Writer,
	name string,
	metrics []Metric,
	precision int,
	fixed bool,
) error {
	var rows = make([][6]string, 0, len(metrics)+1)
	rows = append(rows, [6]string{
		"Item",
		"#",
		"Cumulative",
		"Avg",
		"Min",
		"Max",
	})
	for _, metric := range metrics {
		rows = append(rows, [6]string{
			metric.Item,
			strconv.Itoa(metric.Samples),
			formatMetric(metric.Cumulative, precision, fixed),
			formatMetric(metric.Average, precision, fixed),
			formatMetric(metric.Minimum, precision, fixed),
			formatMetric(metric.Maximum, precision, fixed),
		})
	}
	var widths [6]int
	for _, row := range rows {
		for index, value := range row {
			if len(value) > widths[index] {
				widths[index] = len(value)
			}
		}
	}
	var _, err = fmt.Fprintln(output, name)
	if err != nil {
		return fmt.Errorf("write SuiteReport section: %v", err)
	}
	for _, row := range rows {
		var samples = center(row[1], widths[1])
		_, err = fmt.Fprintf(
			output,
			"%-*s  %s  %*s  %*s  %*s  %*s\n",
			widths[0],
			row[0],
			samples,
			widths[2],
			row[2],
			widths[3],
			row[3],
			widths[4],
			row[4],
			widths[5],
			row[5],
		)
		if err != nil {
			return fmt.Errorf("write SuiteReport metric: %v", err)
		}
	}
	_, err = fmt.Fprintln(output)
	if err != nil {
		return fmt.Errorf("write SuiteReport section: %v", err)
	}
	return nil
}

// Section 3 - Generic Helpers

func formatMetric(value *float64, precision int, fixed bool) string {
	if value == nil {
		return "—"
	}
	var result = strconv.FormatFloat(*value, 'f', precision, 64)
	if !fixed && strings.Contains(result, ".") {
		result = strings.TrimRight(strings.TrimRight(result, "0"), ".")
	}
	return addThousands(result)
}

func addThousands(value string) string {
	var parts = strings.SplitN(value, ".", 2)
	var integer = parts[0]
	var sign string
	if strings.HasPrefix(integer, "-") {
		sign = "-"
		integer = strings.TrimPrefix(integer, "-")
	}
	var first = len(integer) % 3
	if first == 0 {
		first = 3
	}
	var groups = []string{integer[:first]}
	for index := first; index < len(integer); index += 3 {
		groups = append(groups, integer[index:index+3])
	}
	var result = sign + strings.Join(groups, ",")
	if len(parts) == 2 {
		result += "." + parts[1]
	}
	return result
}

func center(value string, width int) string {
	var left = (width - len(value)) / 2
	var right = width - len(value) - left
	return strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
}
