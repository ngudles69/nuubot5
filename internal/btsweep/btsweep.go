// Package btsweep loads, validates, and expands backtest Sweep templates.
package btsweep

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"nuubot/internal/botspec"
)

// DateRange identifies one ordered historical replay window.
type DateRange struct {
	Name  string
	Start string
	End   string
}

// Bot contains one deterministic generated Bot Config.
type Bot struct {
	Number     uint64
	DateRange  DateRange
	ConfigTOML string
	ConfigHash string
}

// Expansion contains validated values needed to create one Sweep and its Bots.
type Expansion struct {
	SourcePath   string
	TemplatePath string
	Doc          string
	BotSpecID    string
	Symbol       string
	TicksPath    string
	Bots         []Bot
}

type sweepTemplate struct {
	Sweep sweepSection `toml:"sweep"`
}

type sweepSection struct {
	Doc        string         `toml:"doc"`
	Template   string         `toml:"template"`
	Symbol     string         `toml:"symbol"`
	Ticks      string         `toml:"ticks"`
	DateRanges []dateRangeRow `toml:"date_ranges"`
	Parameters map[string]any `toml:"parameters"`
}

type dateRangeRow struct {
	Name  string `toml:"name"`
	Start string `toml:"start"`
	End   string `toml:"end"`
}

type parameter struct {
	path   []string
	name   string
	values []any
}

var commonParameterPaths = map[string]bool{
	"controller.max_cycles":     true,
	"signaler.fast_ma":          true,
	"signaler.regime_ema":       true,
	"signaler.regime_timeframe": true,
	"signaler.signal_timeframe": true,
	"signaler.slow_ma":          true,
}

var observerParameterPaths = map[string]bool{
	"executors.kind":          true,
	"executors.side":          true,
	"executors.stop_loss_pct": true,
	"executors.symbol":        true,
}

var tradeParameterPaths = map[string]bool{
	"executors.capital_usdc":        true,
	"executors.fee_pct":             true,
	"executors.kind":                true,
	"executors.network":             true,
	"executors.order_size_usdc":     true,
	"executors.persist_mode":        true,
	"executors.physical_account_id": true,
	"executors.role":                true,
	"executors.side":                true,
	"executors.slippage_pct":        true,
	"executors.stop_loss_pct":       true,
	"executors.symbol":              true,
	"executors.take_profit_pct":     true,
	"executors.venue":               true,
}

var gridParameterPaths = map[string]bool{
	"executors.capital_usdc":          true,
	"executors.fee_pct":               true,
	"executors.kind":                  true,
	"executors.levels":                true,
	"executors.min_expected_pnl_usdc": true,
	"executors.network":               true,
	"executors.persist_mode":          true,
	"executors.physical_account_id":   true,
	"executors.recon":                 true,
	"executors.range_pct":             true,
	"executors.role":                  true,
	"executors.side":                  true,
	"executors.slippage_pct":          true,
	"executors.symbol":                true,
	"executors.venue":                 true,
}

// Section 1 - Program Flow

// Load loads, validates, and deterministically expands one Sweep template.
func Load(path string) (Expansion, error) {
	// resolve source path
	var result Expansion
	var sourcePath, err = filepath.Abs(path)
	if err != nil {
		return result, fmt.Errorf("load sweep template: resolve source path: %w", err)
	}
	result.SourcePath = filepath.Clean(sourcePath)

	// decode sweep template
	var raw sweepTemplate
	var metadata toml.MetaData
	metadata, err = toml.DecodeFile(result.SourcePath, &raw)
	if err != nil {
		return result, fmt.Errorf("load sweep template %s: %v", result.SourcePath, err)
	}
	if undecoded := unknownSweepFields(metadata); len(undecoded) != 0 {
		return result, fmt.Errorf("load sweep template %s: unknown fields: %v", result.SourcePath, undecoded)
	}

	// validate sweep template
	var ranges []DateRange
	ranges, err = validateSweep(raw.Sweep)
	if err != nil {
		return result, fmt.Errorf("load sweep template %s: %w", result.SourcePath, err)
	}
	result.Doc = strings.TrimSpace(raw.Sweep.Doc)
	result.Symbol = raw.Sweep.Symbol
	result.TicksPath = resolveSourcePath(result.SourcePath, raw.Sweep.Ticks)

	// load bot template
	result.TemplatePath = resolveSourcePath(result.SourcePath, raw.Sweep.Template)
	var templateTOML []byte
	templateTOML, err = os.ReadFile(result.TemplatePath)
	if err != nil {
		return result, fmt.Errorf("load bot template %s: %w", result.TemplatePath, err)
	}
	var base map[string]any
	base, err = decodeBot(string(templateTOML))
	if err != nil {
		return result, fmt.Errorf("load bot template %s: %w", result.TemplatePath, err)
	}
	result.BotSpecID, err = botSpecID(base)
	if err != nil {
		return result, fmt.Errorf("load bot template %s: %w", result.TemplatePath, err)
	}
	if err = botspec.Validate(result.BotSpecID, string(templateTOML)); err != nil {
		return result, fmt.Errorf("load bot template %s: %w", result.TemplatePath, err)
	}

	// validate parameter dimensions
	var parameters []parameter
	parameters, err = collectParameters(raw.Sweep.Parameters)
	if err != nil {
		return result, fmt.Errorf("load sweep template %s: %w", result.SourcePath, err)
	}
	for _, current := range parameters {
		if err = validateParameter(result.BotSpecID, base, current); err != nil {
			return result, fmt.Errorf("load sweep template %s: %w", result.SourcePath, err)
		}
	}

	// expand generated bots
	result.Bots, err = expandBots(
		string(templateTOML),
		result.BotSpecID,
		result.Symbol,
		ranges,
		parameters,
	)
	if err != nil {
		return result, fmt.Errorf("load sweep template %s: %w", result.SourcePath, err)
	}
	return result, nil
}

// Section 2 - Domain Helpers

func validateSweep(raw sweepSection) ([]DateRange, error) {
	if strings.TrimSpace(raw.Doc) == "" {
		return nil, fmt.Errorf("sweep.doc must not be empty")
	}
	if strings.TrimSpace(raw.Template) == "" {
		return nil, fmt.Errorf("sweep.template must not be empty")
	}
	if strings.TrimSpace(raw.Symbol) == "" {
		return nil, fmt.Errorf("sweep.symbol must not be empty")
	}
	if strings.TrimSpace(raw.Ticks) == "" {
		return nil, fmt.Errorf("sweep.ticks must not be empty")
	}
	if len(raw.DateRanges) == 0 {
		return nil, fmt.Errorf("sweep.date_ranges must not be empty")
	}
	if raw.Parameters == nil {
		return nil, fmt.Errorf("sweep.parameters must be a table")
	}

	var names = make(map[string]bool, len(raw.DateRanges))
	var ranges = make([]DateRange, 0, len(raw.DateRanges))
	for _, row := range raw.DateRanges {
		var name = strings.TrimSpace(row.Name)
		if name == "" {
			return nil, fmt.Errorf("date range name must not be empty")
		}
		if names[name] {
			return nil, fmt.Errorf("duplicate date range name: %s", name)
		}
		names[name] = true

		var start, err = time.Parse(time.DateOnly, row.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid date range %s start: %v", name, err)
		}
		var end time.Time
		end, err = time.Parse(time.DateOnly, row.End)
		if err != nil {
			return nil, fmt.Errorf("invalid date range %s end: %v", name, err)
		}
		if !start.Before(end) {
			return nil, fmt.Errorf("date range %s start must precede end", name)
		}
		ranges = append(ranges, DateRange{Name: name, Start: row.Start, End: row.End})
	}
	return ranges, nil
}

func decodeBot(configTOML string) (map[string]any, error) {
	var raw map[string]any
	var _, err = toml.Decode(configTOML, &raw)
	if err != nil {
		return nil, fmt.Errorf("decode bot template: %v", err)
	}
	for _, key := range []string{"id", "bot_id", "sweep_id"} {
		if _, exists := raw[key]; exists {
			return nil, fmt.Errorf("bot template contains generated identity field: %s", key)
		}
	}
	if err = validateBotValues(raw, nil); err != nil {
		return nil, err
	}
	return raw, nil
}

func validateBotValues(value any, path []string) error {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if err := validateBotValues(child, append(path, key)); err != nil {
				return err
			}
		}
	case []map[string]any:
		if !isBotTableArray(path) {
			return fmt.Errorf("bot template field must be scalar: %s", strings.Join(path, "."))
		}
		for _, child := range current {
			if err := validateBotValues(child, path); err != nil {
				return err
			}
		}
	case []any:
		if !isBotTableArray(path) {
			return fmt.Errorf("bot template field must be scalar: %s", strings.Join(path, "."))
		}
		for _, child := range current {
			var table, valid = child.(map[string]any)
			if !valid {
				return fmt.Errorf("bot template field must be scalar: %s", strings.Join(path, "."))
			}
			if err := validateBotValues(table, path); err != nil {
				return err
			}
		}
	case []string, []int64, []float64, []bool, []time.Time:
		return fmt.Errorf("bot template field must be scalar: %s", strings.Join(path, "."))
	}
	return nil
}

func isBotTableArray(path []string) bool {
	if len(path) != 1 {
		return false
	}
	return path[0] == "executors" || path[0] == "risks"
}

func botSpecID(raw map[string]any) (string, error) {
	var value, exists = raw["bot_spec"]
	if !exists {
		return "", fmt.Errorf("bot template bot_spec must not be empty")
	}
	var botSpecID, valid = value.(string)
	if !valid || botSpecID == "" {
		return "", fmt.Errorf("bot template bot_spec must be a nonempty string")
	}
	return botSpecID, nil
}

func collectParameters(raw map[string]any) ([]parameter, error) {
	var output []parameter
	var err = appendParameters(&output, raw, nil)
	if err != nil {
		return nil, err
	}
	sort.Slice(output, func(left, right int) bool {
		return output[left].name < output[right].name
	})
	return output, nil
}

func appendParameters(output *[]parameter, raw map[string]any, prefix []string) error {
	if hasRangeSyntax(raw) {
		return fmt.Errorf("parameter range syntax is not supported: %s", strings.Join(prefix, "."))
	}
	for key, value := range raw {
		var path = appendPath(prefix, key)
		if table, valid := value.(map[string]any); valid {
			if err := appendParameters(output, table, path); err != nil {
				return err
			}
			continue
		}
		var values, valid = parameterValues(value)
		if !valid {
			return fmt.Errorf("parameter %s must be an explicit list", strings.Join(path, "."))
		}
		if len(values) == 0 {
			return fmt.Errorf("parameter %s list must not be empty", strings.Join(path, "."))
		}
		*output = append(*output, parameter{
			path:   path,
			name:   strings.Join(path, "."),
			values: values,
		})
	}
	return nil
}

func validateParameter(botSpecID string, base map[string]any, current parameter) error {
	var templateValue, schemaPath, err = parameterTarget(base, current.path)
	if err != nil {
		return err
	}
	if !recognizedParameterPath(botSpecID, schemaPath) {
		return fmt.Errorf("parameter path is not recognized by %s: %s", botSpecID, current.name)
	}
	if !isScalar(templateValue) {
		return fmt.Errorf("parameter path does not select a scalar: %s", current.name)
	}
	for _, value := range current.values {
		if !sameScalarType(templateValue, value) {
			return fmt.Errorf(
				"parameter %s value has type %T, want %T",
				current.name,
				value,
				templateValue,
			)
		}
	}
	return nil
}

func parameterTarget(base map[string]any, path []string) (any, string, error) {
	var current any = base
	var schema []string
	for index := 0; index < len(path); {
		switch value := current.(type) {
		case map[string]any:
			var key = path[index]
			var child, exists = value[key]
			if !exists {
				return nil, "", fmt.Errorf("missing parameter path: %s", strings.Join(path, "."))
			}
			schema = append(schema, key)
			current = child
			index++
		case []map[string]any:
			if strings.Join(schema, ".") != "executors" {
				return nil, "", fmt.Errorf("parameter path requires an unsupported table selector: %s", strings.Join(path, "."))
			}
			var selected map[string]any
			var err error
			selected, err = selectExecutor(value, path[index])
			if err != nil {
				return nil, "", fmt.Errorf("parameter %s: %w", strings.Join(path, "."), err)
			}
			current = selected
			index++
		case []any:
			var tables, valid = tableValues(value)
			if !valid || strings.Join(schema, ".") != "executors" {
				return nil, "", fmt.Errorf("parameter path requires an unsupported table selector: %s", strings.Join(path, "."))
			}
			var selected map[string]any
			var err error
			selected, err = selectExecutor(tables, path[index])
			if err != nil {
				return nil, "", fmt.Errorf("parameter %s: %w", strings.Join(path, "."), err)
			}
			current = selected
			index++
		default:
			return nil, "", fmt.Errorf("missing parameter path: %s", strings.Join(path, "."))
		}
	}
	return current, strings.Join(schema, "."), nil
}

func selectExecutor(executors []map[string]any, id string) (map[string]any, error) {
	var selected map[string]any
	for _, current := range executors {
		if current["id"] != id {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("executor id is not unique: %s", id)
		}
		selected = current
	}
	if selected == nil {
		return nil, fmt.Errorf("executor id does not exist: %s", id)
	}
	return selected, nil
}

func recognizedParameterPath(botSpecID, path string) bool {
	if commonParameterPaths[path] {
		return true
	}
	switch botSpecID {
	case botspec.MacrossObserver:
		return observerParameterPaths[path]
	case botspec.MacrossTrade:
		return tradeParameterPaths[path]
	case botspec.MacrossGrid:
		return gridParameterPaths[path]
	default:
		return false
	}
}

func expandBots(
	templateTOML string,
	botSpecID string,
	replaySymbol string,
	ranges []DateRange,
	parameters []parameter,
) ([]Bot, error) {
	var combinations = parameterCombinations(parameters)
	var bots = make([]Bot, 0, len(ranges)*len(combinations))
	for _, dateRange := range ranges {
		for _, combination := range combinations {
			var config, err = decodeBot(templateTOML)
			if err != nil {
				return nil, err
			}
			for index, value := range combination {
				if err = setParameter(config, parameters[index].path, value); err != nil {
					return nil, err
				}
			}
			var configTOML string
			configTOML, err = encodeBot(config)
			if err != nil {
				return nil, err
			}
			if err = botspec.Validate(botSpecID, configTOML); err != nil {
				return nil, fmt.Errorf("validate generated bot %d: %w", len(bots)+1, err)
			}
			if err = botspec.ValidateReplaySymbol(botSpecID, configTOML, replaySymbol); err != nil {
				return nil, fmt.Errorf(
					"validate generated bot %d replay symbol: %w",
					len(bots)+1,
					err,
				)
			}
			bots = append(bots, Bot{
				Number:     uint64(len(bots) + 1),
				DateRange:  dateRange,
				ConfigTOML: configTOML,
				ConfigHash: fmt.Sprintf("%x", sha256.Sum256([]byte(configTOML))),
			})
		}
	}
	return bots, nil
}

func parameterCombinations(parameters []parameter) [][]any {
	var output [][]any
	var current = make([]any, len(parameters))
	appendCombinations(&output, current, parameters, 0)
	return output
}

func appendCombinations(
	output *[][]any,
	current []any,
	parameters []parameter,
	index int,
) {
	if index == len(parameters) {
		var combination = append([]any(nil), current...)
		*output = append(*output, combination)
		return
	}
	for _, value := range parameters[index].values {
		current[index] = value
		appendCombinations(output, current, parameters, index+1)
	}
}

func setParameter(base map[string]any, path []string, replacement any) error {
	var current any = base
	for index := 0; index < len(path)-1; {
		switch value := current.(type) {
		case map[string]any:
			current = value[path[index]]
			index++
		case []map[string]any:
			var selected, err = selectExecutor(value, path[index])
			if err != nil {
				return err
			}
			current = selected
			index++
		case []any:
			var tables, valid = tableValues(value)
			if !valid {
				return fmt.Errorf("missing parameter path: %s", strings.Join(path, "."))
			}
			var selected, err = selectExecutor(tables, path[index])
			if err != nil {
				return err
			}
			current = selected
			index++
		default:
			return fmt.Errorf("missing parameter path: %s", strings.Join(path, "."))
		}
	}
	var table, valid = current.(map[string]any)
	if !valid {
		return fmt.Errorf("missing parameter path: %s", strings.Join(path, "."))
	}
	table[path[len(path)-1]] = replacement
	return nil
}

func encodeBot(raw map[string]any) (string, error) {
	var output bytes.Buffer
	var encoder = toml.NewEncoder(&output)
	var err = encoder.Encode(raw)
	if err != nil {
		return "", fmt.Errorf("encode generated bot config: %v", err)
	}
	return output.String(), nil
}

// Section 3 - Generic Helpers

func unknownSweepFields(metadata toml.MetaData) []toml.Key {
	var output []toml.Key
	for _, key := range metadata.Undecoded() {
		if len(key) >= 2 && key[0] == "sweep" && key[1] == "parameters" {
			continue
		}
		output = append(output, key)
	}
	return output
}

func resolveSourcePath(sourcePath, referencedPath string) string {
	if filepath.IsAbs(referencedPath) {
		return filepath.Clean(referencedPath)
	}
	return filepath.Clean(filepath.Join(filepath.Dir(sourcePath), referencedPath))
}

func hasRangeSyntax(raw map[string]any) bool {
	var _, hasStart = raw["start"]
	var _, hasStop = raw["stop"]
	var _, hasStep = raw["step"]
	return hasStart && hasStop && hasStep
}

func appendPath(path []string, value string) []string {
	var output = make([]string, len(path), len(path)+1)
	copy(output, path)
	return append(output, value)
}

func parameterValues(value any) ([]any, bool) {
	switch values := value.(type) {
	case []any:
		return values, true
	case []string:
		var output = make([]any, len(values))
		for index := range values {
			output[index] = values[index]
		}
		return output, true
	case []int64:
		var output = make([]any, len(values))
		for index := range values {
			output[index] = values[index]
		}
		return output, true
	case []float64:
		var output = make([]any, len(values))
		for index := range values {
			output[index] = values[index]
		}
		return output, true
	case []bool:
		var output = make([]any, len(values))
		for index := range values {
			output[index] = values[index]
		}
		return output, true
	case []time.Time:
		var output = make([]any, len(values))
		for index := range values {
			output[index] = values[index]
		}
		return output, true
	default:
		return nil, false
	}
}

func tableValues(values []any) ([]map[string]any, bool) {
	var output = make([]map[string]any, 0, len(values))
	for _, value := range values {
		var table, valid = value.(map[string]any)
		if !valid {
			return nil, false
		}
		output = append(output, table)
	}
	return output, true
}

func isScalar(value any) bool {
	switch value.(type) {
	case string, bool, int64, float64, time.Time:
		return true
	default:
		return false
	}
}

func sameScalarType(left, right any) bool {
	switch left.(type) {
	case string:
		_, valid := right.(string)
		return valid
	case bool:
		_, valid := right.(bool)
		return valid
	case int64:
		_, valid := right.(int64)
		return valid
	case float64:
		_, valid := right.(float64)
		return valid
	case time.Time:
		_, valid := right.(time.Time)
		return valid
	default:
		return false
	}
}
