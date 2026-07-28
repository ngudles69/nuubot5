package fprof

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	prettytable "github.com/jedib0t/go-pretty/v6/table"
	prettytext "github.com/jedib0t/go-pretty/v6/text"

	"nuubot/internal/report"
)

// Config contains one complete function-profile request.
type Config struct {
	Root    string
	SweepID uint64
	BotID   uint64
	Top     int
}

// Variant contains one A, B, or C execution result.
type Variant struct {
	Name      string         `json:"name"`
	Mode      string         `json:"mode"`
	ElapsedNS int64          `json:"elapsed_ns"`
	Run       report.Run     `json:"run"`
	Profile   *RuntimeReport `json:"profile,omitempty"`
	Stdout    string         `json:"stdout_path"`
	Stderr    string         `json:"stderr_path"`
}

// Overhead contains one calculated runtime difference.
type Overhead struct {
	Name      string  `json:"name"`
	ElapsedNS int64   `json:"elapsed_ns"`
	Percent   float64 `json:"percent"`
}

// FunctionRow contains one rendered timed function result.
type FunctionRow struct {
	Name      string  `json:"name"`
	Calls     uint64  `json:"calls"`
	FlatNS    int64   `json:"flat_ns"`
	FlatPct   float64 `json:"flat_percent"`
	CumNS     int64   `json:"cum_ns"`
	CumPct    float64 `json:"cum_percent"`
	AvgFlatNS float64 `json:"average_flat_ns"`
	AvgCumNS  float64 `json:"average_cum_ns"`
}

// SessionReport contains one complete A/B/C function-profile report.
type SessionReport struct {
	Session     string        `json:"session"`
	SweepID     uint64        `json:"sweep_id"`
	BotID       uint64        `json:"bot_id"`
	SourceFiles int           `json:"source_files"`
	Functions   int           `json:"instrumented_functions"`
	Variants    []Variant     `json:"variants"`
	Overhead    []Overhead    `json:"overhead"`
	Profile     []FunctionRow `json:"profile"`
	BehaviorOK  bool          `json:"behavior_match"`
}

type behavior struct {
	SweepID             uint64
	BotID               uint64
	BotSpecID           string
	ConfigHash          string
	Symbol              string
	FirstMS             uint64
	LastMS              uint64
	Status              string
	Ticks               uint64
	ControllerRuns      uint64
	SignalPackages      uint64
	StartActionsSkipped uint64
	CyclesStarted       uint64
	CyclesRejected      uint64
	CyclesClosed        uint64
	Trades              uint64
	Orders              uint64
	Fills               uint64
	Cancellations       uint64
	StopOrders          uint64
	Retries             uint64
	RoundTrips          uint64
	BotCapital          string
	GrossPnL            string
	Fees                string
	NetPnL              string
	EndingEquity        string
	MaxDrawdown         string
	TelemetrySamples    int
}

// Section 1 - Program Flow

// Run executes one complete A/B/C function-profile session.
func Run(cfg Config) (SessionReport, error) {
	// Step 1: validate profile request
	if cfg.Root == "" || cfg.SweepID == 0 || cfg.BotID == 0 || cfg.Top <= 0 {
		return SessionReport{}, fmt.Errorf("run function profile: invalid Config")
	}
	var root, err = filepath.Abs(cfg.Root)
	if err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: resolve root: %w", err)
	}

	// Step 2: create isolated profile session
	var sessionName = fmt.Sprintf(
		"s%d-b%d-%s",
		cfg.SweepID,
		cfg.BotID,
		time.Now().UTC().Format("20060102T150405Z"),
	)
	var session = filepath.Join(root, "workspace", "perf", "fprofiles", sessionName)
	for _, child := range []string{"binaries", "overlays", "runs", "profiles"} {
		if err = os.MkdirAll(filepath.Join(session, child), 0o755); err != nil {
			return SessionReport{}, fmt.Errorf("run function profile: create session: %w", err)
		}
	}

	// Step 3: generate instrumented build overlay
	var instrumented InstrumentResult
	instrumented, err = GenerateOverlay(root, filepath.Join(session, "overlays"))
	if err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: %w", err)
	}

	// Step 4: build A, B, and C binaries
	var binaries = map[string]string{
		"A": filepath.Join(session, "binaries", executableName("nuubot-bt-bot-a")),
		"B": filepath.Join(session, "binaries", executableName("nuubot-bt-bot-b")),
		"C": filepath.Join(session, "binaries", executableName("nuubot-bt-bot-c")),
	}
	if err = buildBinary(root, binaries["A"], ""); err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: build A: %w", err)
	}
	if err = buildBinary(root, binaries["B"], instrumented.OverlayPath); err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: build B: %w", err)
	}
	if err = buildBinary(root, binaries["C"], instrumented.OverlayPath); err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: build C: %w", err)
	}

	// Step 5: run A, B, and C sequentially
	var variants = []struct {
		name string
		mode string
	}{
		{name: "A", mode: "original"},
		{name: "B", mode: ModeStructural},
		{name: "C", mode: ModeTimed},
	}
	var sessionReport = SessionReport{
		Session:     session,
		SweepID:     cfg.SweepID,
		BotID:       cfg.BotID,
		SourceFiles: instrumented.Files,
		Functions:   instrumented.Functions,
	}
	for _, current := range variants {
		var result Variant
		result, err = runVariant(
			root,
			session,
			binaries[current.name],
			current.name,
			current.mode,
			cfg.SweepID,
			cfg.BotID,
		)
		if err != nil {
			return SessionReport{}, fmt.Errorf("run function profile: run %s: %w", current.name, err)
		}
		sessionReport.Variants = append(sessionReport.Variants, result)
	}

	// Step 6: verify equivalent behavior
	var expected = stableBehavior(sessionReport.Variants[0].Run)
	for _, variant := range sessionReport.Variants[1:] {
		if !reflect.DeepEqual(expected, stableBehavior(variant.Run)) {
			return SessionReport{}, fmt.Errorf("run function profile: %s behavior differs from A", variant.Name)
		}
	}
	sessionReport.BehaviorOK = true

	// Step 7: calculate runtime overhead and function profile
	sessionReport.Overhead = calculateOverhead(sessionReport.Variants)
	var structural = sessionReport.Variants[1].Profile
	var timed = sessionReport.Variants[2].Profile
	sessionReport.Profile, err = calculateFunctions(structural, timed)
	if err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: %w", err)
	}

	// Step 8: write final JSON and text reports
	var jsonPath = filepath.Join(session, "report.json")
	if err = writeJSON(jsonPath, sessionReport); err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: %w", err)
	}
	var textPath = filepath.Join(session, "report.txt")
	if err = writeText(textPath, sessionReport, cfg.Top); err != nil {
		return SessionReport{}, fmt.Errorf("run function profile: %w", err)
	}
	return sessionReport, nil
}

// Section 2 - Domain Helpers

func buildBinary(root, output, overlay string) error {
	var arguments = []string{"build", "-tags", "noasm"}
	if overlay != "" {
		arguments = append(arguments, "-overlay", overlay)
	}
	arguments = append(arguments, "-o", output, "./cmd/nuubot-bt-bot")
	var command = exec.Command(goExecutable(), arguments...)
	command.Dir = root
	command.Env = cleanEnvironment("CGO_ENABLED=0")
	var payload, err = command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(payload)))
	}
	return nil
}

func runVariant(
	root,
	session,
	binary,
	name,
	mode string,
	sweepID,
	botID uint64,
) (Variant, error) {
	// Step 1: prepare variant outputs
	var stdoutPath = filepath.Join(session, "runs", strings.ToLower(name)+".stdout.json")
	var stderrPath = filepath.Join(session, "runs", strings.ToLower(name)+".stderr.txt")
	var profilePath string
	if mode == ModeStructural || mode == ModeTimed {
		profilePath = filepath.Join(session, "profiles", strings.ToLower(name)+".json")
	}
	var stdout, err = os.Create(stdoutPath)
	if err != nil {
		return Variant{}, err
	}
	var stderr *os.File
	stderr, err = os.Create(stderrPath)
	if err != nil {
		stdout.Close()
		return Variant{}, err
	}

	// Step 2: execute variant
	var command = exec.Command(binary, strconv.FormatUint(sweepID, 10), strconv.FormatUint(botID, 10))
	command.Dir = root
	var settings = []string{"CGO_ENABLED=0"}
	if profilePath != "" {
		settings = append(
			settings,
			"NUUBOT_FPROF_MODE="+mode,
			"NUUBOT_FPROF_OUTPUT="+profilePath,
		)
	}
	command.Env = cleanEnvironment(settings...)
	command.Stdout = stdout
	command.Stderr = stderr
	var started = time.Now()
	var runErr = command.Run()
	var elapsed = time.Since(started)
	var stdoutErr = stdout.Close()
	var stderrErr = stderr.Close()
	if runErr != nil {
		return Variant{}, fmt.Errorf("%v; stdout=%s stderr=%s", runErr, stdoutPath, stderrPath)
	}
	if stdoutErr != nil || stderrErr != nil {
		return Variant{}, fmt.Errorf("close variant output: stdout=%v stderr=%v", stdoutErr, stderrErr)
	}

	// Step 3: decode RunReport and function profile
	var run report.Run
	if err = readJSON(stdoutPath, &run); err != nil {
		return Variant{}, fmt.Errorf("decode RunReport: %w", err)
	}
	var profile *RuntimeReport
	if profilePath != "" {
		profile = &RuntimeReport{}
		if err = readJSON(profilePath, profile); err != nil {
			return Variant{}, fmt.Errorf("decode function profile: %w", err)
		}
		profile.Output = profilePath
		if profile.StackError != "" || profile.OpenFrames != 0 {
			return Variant{}, fmt.Errorf(
				"invalid function stack error=%q open_frames=%d",
				profile.StackError,
				profile.OpenFrames,
			)
		}
	}
	return Variant{
		Name:      name,
		Mode:      mode,
		ElapsedNS: elapsed.Nanoseconds(),
		Run:       run,
		Profile:   profile,
		Stdout:    stdoutPath,
		Stderr:    stderrPath,
	}, nil
}

func calculateOverhead(variants []Variant) []Overhead {
	var a = variants[0].ElapsedNS
	var b = variants[1].ElapsedNS
	var c = variants[2].ElapsedNS
	return []Overhead{
		overhead("B - A", b-a, a),
		overhead("C - B", c-b, b),
		overhead("C - A", c-a, a),
	}
}

func calculateFunctions(structural, timed *RuntimeReport) ([]FunctionRow, error) {
	if structural == nil || timed == nil || structural.Mode != ModeStructural || timed.Mode != ModeTimed {
		return nil, fmt.Errorf("function profiles are incomplete")
	}
	if timed.RootNS <= 0 {
		return nil, fmt.Errorf("timed profile root is empty")
	}
	var calls = make(map[string]uint64, len(structural.Functions))
	for _, function := range structural.Functions {
		calls[function.Name] = function.Calls
	}
	var rows = make([]FunctionRow, 0, len(timed.Functions))
	for _, function := range timed.Functions {
		if calls[function.Name] != function.Calls {
			return nil, fmt.Errorf(
				"function calls differ name=%s structural=%d timed=%d",
				function.Name,
				calls[function.Name],
				function.Calls,
			)
		}
		var row = FunctionRow{
			Name:    function.Name,
			Calls:   function.Calls,
			FlatNS:  function.Flat,
			FlatPct: percent(function.Flat, timed.RootNS),
			CumNS:   function.Cum,
			CumPct:  percent(function.Cum, timed.RootNS),
		}
		if function.Calls != 0 {
			row.AvgFlatNS = float64(function.Flat) / float64(function.Calls)
			row.AvgCumNS = float64(function.Cum) / float64(function.Calls)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(left int, right int) bool {
		if rows[left].FlatNS == rows[right].FlatNS {
			return rows[left].Name < rows[right].Name
		}
		return rows[left].FlatNS > rows[right].FlatNS
	})
	return rows, nil
}

func totalFunctions(rows []FunctionRow) FunctionRow {
	var total = FunctionRow{Name: "Total"}
	for _, row := range rows {
		total.FlatNS += row.FlatNS
		total.FlatPct += row.FlatPct
		total.CumNS += row.CumNS
		total.CumPct += row.CumPct
	}
	return total
}

func stableBehavior(run report.Run) behavior {
	return behavior{
		SweepID:             run.SweepID,
		BotID:               run.BotID,
		BotSpecID:           run.BotSpecID,
		ConfigHash:          run.ConfigHash,
		Symbol:              run.Symbol,
		FirstMS:             run.FirstMS,
		LastMS:              run.LastMS,
		Status:              run.Status,
		Ticks:               run.Ticks,
		ControllerRuns:      run.ControllerRuns,
		SignalPackages:      run.SignalPackages,
		StartActionsSkipped: run.StartActionsSkipped,
		CyclesStarted:       run.CyclesStarted,
		CyclesRejected:      run.CyclesRejected,
		CyclesClosed:        run.CyclesClosed,
		Trades:              run.Trades,
		Orders:              run.Orders,
		Fills:               run.Fills,
		Cancellations:       run.Cancellations,
		StopOrders:          run.StopOrders,
		Retries:             run.Retries,
		RoundTrips:          run.RoundTrips,
		BotCapital:          run.BotCapital.String(),
		GrossPnL:            run.GrossPnL.String(),
		Fees:                run.Fees.String(),
		NetPnL:              run.NetPnL.String(),
		EndingEquity:        run.EndingEquity.String(),
		MaxDrawdown:         run.MaxDrawdown.String(),
		TelemetrySamples:    run.TelemetrySamples,
	}
}

func writeJSON(path string, value any) error {
	var file, err = os.Create(path)
	if err != nil {
		return err
	}
	var encoder = json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	var writeErr = encoder.Encode(value)
	var closeErr = file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func writeText(path string, report SessionReport, top int) error {
	var file, err = os.Create(path)
	if err != nil {
		return err
	}
	var output = bufio.NewWriter(file)
	fmt.Fprintf(output, "Nuubot Function Profile\n\n")
	fmt.Fprintf(output, "Session: %s\n", report.Session)
	fmt.Fprintf(output, "Sweep: %d  Bot: %d  Behavior: MATCH\n", report.SweepID, report.BotID)
	fmt.Fprintf(output, "Instrumented: %d files  %d functions\n\n", report.SourceFiles, report.Functions)

	fmt.Fprintln(output, "Runtime")
	var runtimeTable = newReportTable(3, 4, 5, 6, 7)
	runtimeTable.AppendHeader(prettytable.Row{
		"Variant", "Mode", "Elapsed", "GC Runs", "GC Pause", "Heap", "Total Alloc",
	})
	for _, variant := range report.Variants {
		runtimeTable.AppendRow(prettytable.Row{
			variant.Name,
			variant.Mode,
			formatDuration(variant.ElapsedNS),
			variant.Run.Memory.GCRuns,
			formatDuration(int64(variant.Run.Memory.GCPauseMS * 1e6)),
			formatMB(variant.Run.Memory.HeapMB),
			formatMB(variant.Run.Memory.TotalAllocMB),
		})
	}
	fmt.Fprintln(output, runtimeTable.Render())

	fmt.Fprintln(output, "Overhead")
	var overheadTable = newReportTable(2, 3)
	overheadTable.AppendHeader(prettytable.Row{"Comparison", "Elapsed", "Percent"})
	for _, current := range report.Overhead {
		overheadTable.AppendRow(prettytable.Row{
			current.Name,
			formatDuration(current.ElapsedNS),
			fmt.Sprintf("%.2f%%", current.Percent),
		})
	}
	fmt.Fprintln(output, overheadTable.Render())

	fmt.Fprintln(output, "Functions")
	var functionTable = newReportTable(1, 2, 3, 4, 5)
	functionTable.AppendHeader(prettytable.Row{
		"Calls", "Flat (Flat%)", "Cum (Cum%)", "Avg Flat", "Avg Cum", "Function",
	})
	var total = totalFunctions(report.Profile)
	functionTable.AppendRow(prettytable.Row{
		"",
		fmt.Sprintf("%s (%.2f%%)", formatDuration(total.FlatNS), total.FlatPct),
		fmt.Sprintf("%s (%.2f%%)", formatDuration(total.CumNS), total.CumPct),
		"",
		"",
		total.Name,
	})
	functionTable.AppendSeparator()
	var limit = min(top, len(report.Profile))
	for _, current := range report.Profile[:limit] {
		functionTable.AppendRow(prettytable.Row{
			current.Calls,
			fmt.Sprintf("%s (%.2f%%)", formatDuration(current.FlatNS), current.FlatPct),
			fmt.Sprintf("%s (%.2f%%)", formatDuration(current.CumNS), current.CumPct),
			formatDuration(int64(current.AvgFlatNS)),
			formatDuration(int64(current.AvgCumNS)),
			current.Name,
		})
	}
	fmt.Fprintln(output, functionTable.Render())

	var flushErr = output.Flush()
	var closeErr = file.Close()
	if flushErr != nil {
		return flushErr
	}
	return closeErr
}

// Section 3 - Generic Helpers

func newReportTable(rightColumns ...int) prettytable.Writer {
	var result = prettytable.NewWriter()
	var style = prettytable.StyleLight
	style.Format.Header = prettytext.FormatDefault
	style.Options.DrawBorder = false
	style.Options.SeparateColumns = false
	style.Options.SeparateFooter = false
	style.Options.SeparateRows = false
	result.SetStyle(style)
	var columns = make([]prettytable.ColumnConfig, 0, len(rightColumns))
	for _, number := range rightColumns {
		columns = append(columns, prettytable.ColumnConfig{
			Number:      number,
			Align:       prettytext.AlignRight,
			AlignHeader: prettytext.AlignRight,
		})
	}
	result.SetColumnConfigs(columns)
	return result
}

func goExecutable() string {
	return filepath.Join(runtime.GOROOT(), "bin", executableName("go"))
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func cleanEnvironment(settings ...string) []string {
	var blocked = []string{"CGO_ENABLED=", "NUUBOT_FPROF_MODE=", "NUUBOT_FPROF_OUTPUT="}
	var result []string
	for _, value := range os.Environ() {
		var skip bool
		for _, prefix := range blocked {
			if strings.HasPrefix(value, prefix) {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, value)
		}
	}
	return append(result, settings...)
}

func readJSON(path string, value any) error {
	var file, err = os.Open(path)
	if err != nil {
		return err
	}
	var decodeErr = json.NewDecoder(file).Decode(value)
	var closeErr = file.Close()
	if decodeErr != nil {
		return decodeErr
	}
	return closeErr
}

func overhead(name string, difference, baseline int64) Overhead {
	return Overhead{
		Name:      name,
		ElapsedNS: difference,
		Percent:   percent(difference, baseline),
	}
}

func percent(value, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total) * 100
}

func formatDuration(value int64) string {
	var duration = time.Duration(value)
	if duration >= time.Second || duration <= -time.Second {
		return fmt.Sprintf("%.3fs", float64(duration)/float64(time.Second))
	}
	if duration >= time.Millisecond || duration <= -time.Millisecond {
		return fmt.Sprintf("%.3fms", float64(duration)/float64(time.Millisecond))
	}
	if duration >= time.Microsecond || duration <= -time.Microsecond {
		return fmt.Sprintf("%.3fµs", float64(duration)/float64(time.Microsecond))
	}
	return fmt.Sprintf("%dns", value)
}

func formatMB(value float64) string {
	return fmt.Sprintf("%.3f MB", value)
}
