package btsweep

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"nuubot/internal/botspec"
)

type generatedConfig struct {
	Controller struct {
		MaxCycles int64 `toml:"max_cycles"`
	} `toml:"controller"`
	Executors []struct {
		Levels int64 `toml:"levels"`
	} `toml:"executors"`
}

// Section 1 - Program Flow

func TestSystemTemplatesGenerateOneBot(t *testing.T) {
	var root = projectRoot(t)
	var templates = map[string]string{
		"macross_observer_v1.toml": botspec.MacrossObserver,
		"macross_trade_v1.toml":    botspec.MacrossTrade,
		"macross_grid_v1.toml":     botspec.MacrossGrid,
	}
	for name, wantBotSpecID := range templates {
		t.Run(name, func(t *testing.T) {
			var path = filepath.Join(root, "templates", "sweeps", "tests", name)
			var expansion, err = Load(path)
			if err != nil {
				t.Fatalf("load system Sweep template: %v", err)
			}
			if expansion.BotSpecID != wantBotSpecID {
				t.Fatalf("BotSpecID = %q, want %q", expansion.BotSpecID, wantBotSpecID)
			}
			if len(expansion.Bots) != 1 {
				t.Fatalf("generated Bots = %d, want 1", len(expansion.Bots))
			}
			assertGeneratedBot(t, expansion.Bots[0], 1, "2026-03-01..2026-06-01")
		})
	}
}

func TestLoadExpandsParameterFreeSweepDeterministically(t *testing.T) {
	var botTOML = canonicalBotTemplate(t, "macross_grid_v1.toml")
	var sweepTOML = strings.Replace(validSweep(`
[sweep.parameters]
`),
		`periods = [{ start = "2026-03-01", end = "2026-04-01" }]`,
		`periods = [
    { start = "2026-03-01", end = "2026-04-01" },
    { label = "2026-M04" },
]`,
		1,
	)
	var path = writeFixture(t, botTOML, sweepTOML)

	var first, err = Load(path)
	if err != nil {
		t.Fatalf("load parameter-free Sweep: %v", err)
	}
	var second Expansion
	second, err = Load(path)
	if err != nil {
		t.Fatalf("reload parameter-free Sweep: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("repeated parameter-free expansion was not deterministic")
	}
	if len(first.Bots) != 2 {
		t.Fatalf("generated Bots = %d, want 2", len(first.Bots))
	}
	assertGeneratedBot(t, first.Bots[0], 1, "2026-03-01..2026-04-01")
	assertGeneratedBot(t, first.Bots[1], 2, "2026-M04")
	if first.Bots[1].DateRange.Start != "2026-04-01" ||
		first.Bots[1].DateRange.End != "2026-05-01" {
		t.Fatalf("resolved period = %+v", first.Bots[1].DateRange)
	}
	if first.Bots[0].ConfigTOML != first.Bots[1].ConfigTOML {
		t.Fatal("parameter-free date ranges changed Config TOML")
	}
	if first.Bots[0].ConfigHash != first.Bots[1].ConfigHash {
		t.Fatal("parameter-free date ranges changed Config hash")
	}
}

func TestLoadResolvesRelativeTicksFromSweepSource(t *testing.T) {
	var botTOML = canonicalBotTemplate(t, "macross_grid_v1.toml")
	var sweepTOML = strings.Replace(
		validSweep("\n[sweep.parameters]\n"),
		`ticks = "D:/workspace/data/binance/parquet/spot/monthly/klines/BTCUSDT/1s"`,
		`ticks = "ticks/BTCUSDT/1s"`,
		1,
	)
	var path = writeFixture(t, botTOML, sweepTOML)
	var originalDirectory, err = os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err = os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		var restoreErr = os.Chdir(originalDirectory)
		if restoreErr != nil {
			t.Errorf("restore working directory: %v", restoreErr)
		}
	})

	var expansion Expansion
	expansion, err = Load(path)
	if err != nil {
		t.Fatalf("load Sweep outside source directory: %v", err)
	}
	var want = filepath.Join(filepath.Dir(path), "ticks", "BTCUSDT", "1s")
	if expansion.TicksPath != want {
		t.Fatalf("ticks path = %q, want %q", expansion.TicksPath, want)
	}
}

func TestLoadExpandsSortedParametersAndOrderedDates(t *testing.T) {
	var botTOML = canonicalBotTemplate(t, "macross_grid_v1.toml")
	var sweepTOML = strings.Replace(validSweep(`
[sweep.parameters.executors.grid]
levels = [30, 50]

[sweep.parameters.controller]
max_cycles = [1, 2]
`),
		`periods = [{ start = "2026-03-01", end = "2026-04-01" }]`,
		`periods = [
    { start = "2026-03-01", end = "2026-04-01" },
    { label = "2026-M04" },
]`,
		1,
	)
	var path = writeFixture(t, botTOML, sweepTOML)

	var first, err = Load(path)
	if err != nil {
		t.Fatalf("load expanded Sweep: %v", err)
	}
	var second Expansion
	second, err = Load(path)
	if err != nil {
		t.Fatalf("reload expanded Sweep: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("repeated expansion was not deterministic")
	}
	if len(first.Bots) != 8 {
		t.Fatalf("generated Bots = %d, want 8", len(first.Bots))
	}

	var wantRanges = []string{
		"2026-03-01..2026-04-01",
		"2026-03-01..2026-04-01",
		"2026-03-01..2026-04-01",
		"2026-03-01..2026-04-01",
		"2026-M04",
		"2026-M04",
		"2026-M04",
		"2026-M04",
	}
	var wantMaxCycles = []int64{1, 1, 2, 2, 1, 1, 2, 2}
	var wantLevels = []int64{30, 50, 30, 50, 30, 50, 30, 50}
	for index, generated := range first.Bots {
		assertGeneratedBot(t, generated, uint64(index+1), wantRanges[index])
		var config generatedConfig
		if _, err = toml.Decode(generated.ConfigTOML, &config); err != nil {
			t.Fatalf("decode generated Bot %d: %v", generated.Number, err)
		}
		if config.Controller.MaxCycles != wantMaxCycles[index] {
			t.Fatalf("Bot %d max_cycles = %d, want %d", generated.Number, config.Controller.MaxCycles, wantMaxCycles[index])
		}
		if len(config.Executors) != 1 || config.Executors[0].Levels != wantLevels[index] {
			t.Fatalf("Bot %d levels = %+v, want %d", generated.Number, config.Executors, wantLevels[index])
		}
	}
}

func TestLoadRejectsInvalidTemplates(t *testing.T) {
	var validBot = canonicalBotTemplate(t, "macross_grid_v1.toml")
	var tests = []struct {
		name       string
		botTOML    string
		sweepTOML  string
		missingBot bool
		want       string
	}{
		{
			name:      "empty documentation",
			botTOML:   validBot,
			sweepTOML: strings.Replace(validSweep(validParameters()), `doc = "test Sweep"`, `doc = ""`, 1),
			want:      "sweep.doc must not be empty",
		},
		{
			name:      "Executor symbol lacks replay input",
			botTOML:   validBot,
			sweepTOML: strings.Replace(validSweep(validParameters()), `symbol = "BTC"`, `symbol = "ETH"`, 1),
			want:      "executor symbol BTC lacks replay input ETH",
		},
		{
			name:       "missing Bot template",
			botTOML:    validBot,
			sweepTOML:  strings.Replace(validSweep(validParameters()), `template = "bot.toml"`, `template = "missing.toml"`, 1),
			missingBot: true,
			want:       "load bot template",
		},
		{
			name:      "bad date",
			botTOML:   validBot,
			sweepTOML: strings.Replace(validSweep(validParameters()), `start = "2026-03-01"`, `start = "not-a-date"`, 1),
			want:      "invalid period start",
		},
		{
			name:    "duplicate period",
			botTOML: validBot,
			sweepTOML: strings.Replace(
				validSweep(validParameters()),
				`periods = [{ start = "2026-03-01", end = "2026-04-01" }]`,
				`periods = [
    { start = "2026-03-01", end = "2026-04-01" },
    { label = "2026-M03" },
]`,
				1,
			),
			want: "duplicate period",
		},
		{
			name:      "mixed period modes",
			botTOML:   validBot,
			sweepTOML: strings.Replace(validSweep(validParameters()), `{ start = "2026-03-01"`, `{ label = "2026-M03", start = "2026-03-01"`, 1),
			want:      "either label or start and end",
		},
		{
			name:      "incomplete explicit period",
			botTOML:   validBot,
			sweepTOML: strings.Replace(validSweep(validParameters()), `, end = "2026-04-01"`, "", 1),
			want:      "either label or start and end",
		},
		{
			name:      "invalid period label",
			botTOML:   validBot,
			sweepTOML: strings.Replace(validSweep(validParameters()), `{ start = "2026-03-01", end = "2026-04-01" }`, `{ label = "2026-Q5" }`, 1),
			want:      "invalid period label",
		},
		{
			name:    "unknown parameter path",
			botTOML: validBot,
			sweepTOML: validSweep(`
[sweep.parameters.signaler]
unknown = [9]
`),
			want: "missing parameter path",
		},
		{
			name:    "unrecognized existing field",
			botTOML: strings.Replace(validBot, `bot_spec = "macross_grid_bot"`, "bot_spec = \"macross_grid_bot\"\nfuture = 1", 1),
			sweepTOML: validSweep(`
[sweep.parameters]
future = [1]
`),
			want: "not recognized",
		},
		{
			name:    "wrong parameter type",
			botTOML: validBot,
			sweepTOML: validSweep(`
[sweep.parameters.executors.grid]
levels = ["30"]
`),
			want: "value has type",
		},
		{
			name:    "empty parameter list",
			botTOML: validBot,
			sweepTOML: validSweep(`
[sweep.parameters.executors.grid]
levels = []
`),
			want: "list must not be empty",
		},
		{
			name:    "parameter scalar",
			botTOML: validBot,
			sweepTOML: validSweep(`
[sweep.parameters.executors.grid]
levels = 30
`),
			want: "must be an explicit list",
		},
		{
			name:    "range syntax",
			botTOML: validBot,
			sweepTOML: validSweep(`
[sweep.parameters.executors.grid.levels]
start = 30
stop = 50
step = 10
`),
			want: "range syntax is not supported",
		},
		{
			name:      "Bot parameter array",
			botTOML:   strings.Replace(validBot, "levels = 30", "levels = [30]", 1),
			sweepTOML: validSweep(validParameters()),
			want:      "must be scalar",
		},
		{
			name:      "Sweep generated id",
			botTOML:   validBot,
			sweepTOML: strings.Replace(validSweep(validParameters()), "[sweep]\n", "[sweep]\nbot_id = 7\n", 1),
			want:      "unknown fields",
		},
		{
			name:      "Bot generated id",
			botTOML:   "bot_id = 7\n" + validBot,
			sweepTOML: validSweep(validParameters()),
			want:      "generated identity field",
		},
		{
			name:    "missing Executor selector",
			botTOML: validBot,
			sweepTOML: validSweep(`
[sweep.parameters.executors.missing]
levels = [30]
`),
			want: "executor id does not exist",
		},
		{
			name:    "generated Config fails BotSpec",
			botTOML: validBot,
			sweepTOML: validSweep(`
[sweep.parameters.executors.grid]
levels = [2]
`),
			want: "validate generated bot",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var path string
			if test.missingBot {
				path = writeSweepOnly(t, test.sweepTOML)
			} else {
				path = writeFixture(t, test.botTOML, test.sweepTOML)
			}
			var _, err = Load(path)
			if err == nil {
				t.Fatal("invalid Sweep template was accepted")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %q, want text %q", err, test.want)
			}
		})
	}
}

// Section 2 - Domain Helpers

func assertGeneratedBot(t *testing.T, generated Bot, number uint64, dateName string) {
	t.Helper()
	if generated.Number != number {
		t.Fatalf("Bot number = %d, want %d", generated.Number, number)
	}
	if generated.DateRange.Name != dateName {
		t.Fatalf("Bot %d date range = %q, want %q", number, generated.DateRange.Name, dateName)
	}
	var wantHash = fmt.Sprintf("%x", sha256.Sum256([]byte(generated.ConfigTOML)))
	if generated.ConfigHash != wantHash {
		t.Fatalf("Bot %d Config hash = %q, want %q", number, generated.ConfigHash, wantHash)
	}
}

func canonicalBotTemplate(t *testing.T, name string) string {
	t.Helper()
	var path = filepath.Join(projectRoot(t), "templates", "bots", name)
	var content, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read canonical Bot template: %v", err)
	}
	return string(content)
}

func validSweep(parameters string) string {
	return `[sweep]
doc = "test Sweep"
template = "bot.toml"
symbol = "BTC"
ticks = "D:/workspace/data/binance/parquet/spot/monthly/klines/BTCUSDT/1s"

periods = [{ start = "2026-03-01", end = "2026-04-01" }]
` + parameters
}

func validParameters() string {
	return `
[sweep.parameters.executors.grid]
levels = [30]
`
}

// Section 3 - Generic Helpers

func projectRoot(t *testing.T) string {
	t.Helper()
	var _, file, _, valid = runtime.Caller(0)
	if !valid {
		t.Fatal("locate test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func writeFixture(t *testing.T, botTOML, sweepTOML string) string {
	t.Helper()
	var directory = t.TempDir()
	var botPath = filepath.Join(directory, "bot.toml")
	var err = os.WriteFile(botPath, []byte(botTOML), 0o600)
	if err != nil {
		t.Fatalf("write Bot fixture: %v", err)
	}
	return writeSweep(t, directory, sweepTOML)
}

func writeSweepOnly(t *testing.T, sweepTOML string) string {
	t.Helper()
	return writeSweep(t, t.TempDir(), sweepTOML)
}

func writeSweep(t *testing.T, directory, sweepTOML string) string {
	t.Helper()
	var path = filepath.Join(directory, "sweep.toml")
	var err = os.WriteFile(path, []byte(sweepTOML), 0o600)
	if err != nil {
		t.Fatalf("write Sweep fixture: %v", err)
	}
	return path
}
