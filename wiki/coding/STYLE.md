# Prescriptive Go Style

This file owns every explicit Nuubot5 style rule.

It applies to all hand-written `.go` files, including commands, internal
packages, servers, CLI, tests, tools, and build-tagged files.

Package-local and component-local variants are prohibited.

Exceptions require explicit prior user approval and an `AGENTS.md` Key
Decision.

Files under `vendor/**` are excluded.

Generated files MUST contain `// Code generated ... DO NOT EDIT.`.

Generated files MUST change through their generator. They MUST NOT be
hand-edited.

All hand-written Go MUST pass `gofmt`.

The three-section layout and lifecycle vocabulary are Nuubot's mandatory
project-specific style.

Idiomatic Go applies inside this structure.

Style precedence is:

1. explicit Nuubot rules in this file;
2. authoritative idiomatic Go; and
3. the simplest standard-library solution.

External standard-library, library, and protocol contracts retain their
required shape.

## 1. Required File Shape

Every hand-written Go file MUST use:

```text
package
imports
related declarations grouped idiomatically

// Section 1 - Program Flow
New
Init
Start
Loop or Run
Stop
decisions and paths in call order

// Section 2 - Domain Helpers
domain mechanics

// Section 3 - Generic Helpers
file-local domain-neutral mechanics
```

All three section comments MUST remain, including empty sections.

Section comments MUST use exactly:

```go
// Section 1 - Program Flow
// Section 2 - Domain Helpers
// Section 3 - Generic Helpers
```

Data-only and single-purpose files MUST retain all three section comments.

Related declarations MUST be grouped for readability.

A rigid constants-before-types order is prohibited.

`//go:build` and tool directives MUST remain in their required locations.

Tests MUST map into the same sections:

```text
// Section 1 - Program Flow
TestX functions in tested-flow order

// Section 2 - Domain Helpers
domain test builders and assertions

// Section 3 - Generic Helpers
domain-neutral test mechanics
```

## 2. Program Flow

Program Flow MUST reveal:

- ownership;
- lifecycle;
- loops;
- callbacks;
- key decisions;
- selected paths; and
- meaningful work order.

Use one statement per intent:

```go
var err = loadConfig()
if err != nil {
	return err
}
err = startFeeds()
if err != nil {
	return err
}
err = runRuntime()
if err != nil {
	return err
}
```

Fluent program-flow chains are prohibited:

```go
// Prohibited
var result = loadConfig().startFeeds().runRuntime()
```

A library chain is allowed inside a Domain Helper only when its API requires
that chain.

Lifecycle methods MUST remain contiguous:

```go
func NewRunner(...) (*Runner, error)
func (r *Runner) Init() error
func (r *Runner) Start() error
func (r *Runner) Run() error
func (r *Runner) Stop() error
```

Lifecycle rules:

- `NewX` constructs one valid value without background work.
- `Init` performs explicit fallible preparation.
- `Start` begins admission or owned background work.
- `Run` executes one complete bounded job.
- Finite-data iteration belongs inside `Run`.
- `Loop` continuously supervises until stop or cancellation.
- `Pass` executes one timer-driven control pass.
- `Stop` stops admission and releases owned resources.
- Empty lifecycle phases MUST be omitted.

Decision functions MUST follow lifecycle methods.

Path functions MUST follow their decision in branch order.

## 3. Domain Helpers

A Domain Helper MUST hide technical mechanics behind a stable domain contract.

```go
func readParquetFile(
	symbol string,
	interval string,
	start time.Time,
	end time.Time,
) ([]Bar, error)
```

The caller decides that bars are required.

The helper owns approved decoding and returns validated `[]Bar`.

```go
func placeOrder(
	ctx context.Context,
	venue Venue,
	order Order,
) (OrderResult, error)
```

The caller decides to place an order.

The helper owns venue mechanics and returns one stable domain result.

A Domain Helper MUST NOT:

- select program mode;
- start unrelated components;
- hide lifecycle transitions;
- log returned errors; or
- expose library-specific types above its boundary.

## 4. Generic Helpers

A Generic Helper MUST be domain-neutral.

```go
func formatDate(value time.Time) string
func benchmark(started time.Time) time.Duration
```

It MUST remain below Domain Helpers.

A helper used by one file MUST stay there.

A helper with multiple owners MUST move to a concrete common file.

Packages named `utils`, `helpers`, and `misc` are prohibited.

## 5. Naming

Names MUST use `wiki/abbreviations.md`.

Names MUST state intent, not mechanics.

Use the shortest unmistakable name.

Paired operations MUST be symmetrical.

Boolean names MUST read as facts.

Getters MUST use `X`, not `GetX`, unless retrieval performs work.

Ordinary constructors MUST return concrete types.

An approved configuration-selected factory is allowed to return its
consumer-owned interface.

Interfaces MUST be consumer-owned and minimal.

Interfaces MUST represent real behavioral boundaries.

Interfaces MUST NOT exist only for mocking.

Lifecycle names have fixed meanings:

| Name | Required meaning |
|---|---|
| `NewX` | Construct one valid value without background work. |
| `Init` | Perform explicit fallible preparation. |
| `Start` | Begin admission or owned background work. |
| `Run` | Execute one complete bounded job. |
| `Loop` | Supervise continuously until stop or cancellation. |
| `Pass` | Execute one timer-driven control pass. |
| `Stop` | Stop admission and release owned resources. |
| `OnX` | Accept one named event. |

`MainLoop`, `Mainloop`, `Process`, and `Handle` are prohibited when this table
states the intent.

This vocabulary governs Nuubot-owned lifecycle methods.

External contracts retain exact required names.

## 6. Branches and Returns

Ordinary local declarations MUST use `var`.

Later assignment MUST use `=`.

`:=` is allowed only when `var` or `=` is impossible or materially less clear.

Guard clauses MUST reject invalid state early:

```go
func (r *Runner) Start() error {
	if r.started || r.stopped {
		return nuuerrors.StateError("runner", "start")
	}
	var err = r.runtime.Start()
	if err != nil {
		return fmt.Errorf("start runtime: %w", err)
	}
	r.started = true
	return nil
}
```

Nested success paths are prohibited when a guard keeps flow flat.

Values MUST be declared immediately before their first use.

## 7. Comments

Comments MUST explain:

- intent;
- constraints;
- ownership;
- non-obvious ordering; or
- why simpler code is wrong.

Comments MUST NOT narrate syntax.

Every exported declaration MUST have a Go doc comment.

The comment MUST start with the declaration name and state its contract.

Decorative banners and obvious comments are prohibited.

## 8. Complete Normative Example

This file is the required orchestration model.

Real files MUST omit unused paths and helpers.

```go
package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	nuuerrors "nuubot/internal/toolkit/errors"
)

// Mode selects one Runner path.
type Mode string

const (
	// ModeLive selects continuous live supervision.
	ModeLive Mode = "live"
	// ModeBacktest selects finite historical replay.
	ModeBacktest Mode = "backtest"
)

// Config defines one Runner job.
type Config struct {
	Mode     Mode
	Symbol   string
	Interval string
	Start    time.Time
	End      *time.Time
	Venue    Venue
}

// Runner owns one job and its direct inputs.
type Runner struct {
	log     *slog.Logger
	config  Config
	runtime *Runtime
	bars    []Bar
	feed    Feed
	started bool
	stopped bool
}

// Section 1 - Program Flow

// NewRunner constructs one stopped Runner.
func NewRunner(logger *slog.Logger, config Config) (*Runner, error) {
	var err = validateConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	var runtime *Runtime
	runtime, err = NewRuntime(logger, config.End)
	if err != nil {
		return nil, fmt.Errorf("create runtime: %w", err)
	}

	return &Runner{
		log:     logger.With("component", "runner"),
		config:  config,
		runtime: runtime,
	}, nil
}

// Init prepares the selected path.
func (r *Runner) Init(ctx context.Context) error {
	switch r.config.Mode {
	case ModeLive:
		return r.initLive(ctx)
	case ModeBacktest:
		return r.initBacktest(ctx)
	default:
		return fmt.Errorf("initialize runner: unknown mode %q", r.config.Mode)
	}
}

// Start starts Runtime and its input.
func (r *Runner) Start(ctx context.Context) error {
	if r.started || r.stopped {
		return nuuerrors.StateError("runner", "start")
	}

	var err = r.runtime.Start()
	if err != nil {
		return fmt.Errorf("start runtime: %w", err)
	}

	if r.feed != nil {
		var feedErr = r.feed.Start(ctx)
		if feedErr != nil {
			var stopErr = r.runtime.Stop("start_error")
			if stopErr != nil {
				stopErr = fmt.Errorf(
					"stop runtime after feed start failure: %w",
					stopErr,
				)
			}
			return errors.Join(
				fmt.Errorf("start feed: %w", feedErr),
				stopErr,
			)
		}
	}

	r.started = true
	r.log.Info("runner started", "event", "start", "status", "success")
	return nil
}

// Run executes one complete configured job.
func (r *Runner) Run(ctx context.Context) error {
	if !r.started || r.stopped {
		return nuuerrors.StateError("runner", "run")
	}

	switch r.config.Mode {
	case ModeLive:
		return r.loopLive(ctx)
	case ModeBacktest:
		return r.runBacktest(ctx)
	default:
		return fmt.Errorf("run runner: unknown mode %q", r.config.Mode)
	}
}

// Stop releases Runner-owned resources.
func (r *Runner) Stop() error {
	if r.stopped {
		return nil
	}

	r.started = false
	r.stopped = true

	var feedErr error
	if r.feed != nil {
		feedErr = r.feed.Stop()
		if feedErr != nil {
			feedErr = fmt.Errorf("stop feed: %w", feedErr)
		}
	}

	var runtimeErr = r.runtime.Stop("parent_stop")
	if runtimeErr != nil {
		runtimeErr = fmt.Errorf("stop runtime: %w", runtimeErr)
	}

	var err = errors.Join(feedErr, runtimeErr)
	var status = "success"
	if err != nil {
		status = "failed"
	}

	r.log.Info(
		"runner stopped",
		"event", "stop",
		"status", status,
	)
	return err
}

func (r *Runner) initLive(ctx context.Context) error {
	var feed, err = openWebSocketFeed(ctx, r.config.Venue, r.config.Symbol)
	if err != nil {
		return err
	}
	r.feed = feed
	return nil
}

func (r *Runner) initBacktest(ctx context.Context) error {
	var end = time.Now().UTC()
	if r.config.End != nil {
		end = *r.config.End
	}

	var bars, err = readParquetFile(
		ctx,
		r.config.Symbol,
		r.config.Interval,
		r.config.Start,
		end,
	)
	if err != nil {
		return err
	}
	r.bars = bars
	return nil
}

func (r *Runner) loopLive(ctx context.Context) error {
	for {
		var event, err = r.feed.Next(ctx)
		if err != nil {
			return fmt.Errorf("read live feed: %w", err)
		}
		err = r.onEvent(event)
		if err != nil {
			return err
		}
		if r.runtime.Stopped() {
			return nil
		}
	}
}

func (r *Runner) runBacktest(ctx context.Context) error {
	var bar Bar
	for _, bar = range r.bars {
		var err = ctx.Err()
		if err != nil {
			return err
		}
		err = r.onBar(bar)
		if err != nil {
			return err
		}
		if r.runtime.Stopped() {
			return nil
		}
	}
	return r.runtime.Stop("end_date")
}

func (r *Runner) onEvent(event Event) error {
	switch event.Kind {
	case EventBBO:
		return r.runtime.OnBBO(event.BBO)
	case EventUser:
		r.runtime.MarkAccountDirty(event.AccountID)
		return nil
	default:
		return fmt.Errorf("accept feed event: unknown kind %q", event.Kind)
	}
}

func (r *Runner) onBar(bar Bar) error {
	var err = r.runtime.OnBar(bar)
	if err != nil {
		return fmt.Errorf("accept replay bar: %w", err)
	}
	return nil
}

// Section 2 - Domain Helpers

func readParquetFile(
	ctx context.Context,
	symbol string,
	interval string,
	start time.Time,
	end time.Time,
) ([]Bar, error) {
	var path = parquetPath(symbol, interval, start)

	var bars, err = readArrowBars(ctx, path, start, end)
	if err != nil {
		return nil, fmt.Errorf("read parquet %s: %w", path, err)
	}
	err = validateBars(bars, start, end)
	if err != nil {
		return nil, fmt.Errorf("validate parquet %s: %w", path, err)
	}
	return bars, nil
}

func placeOrder(
	ctx context.Context,
	venue Venue,
	order Order,
) (OrderResult, error) {
	switch venue.Kind {
	case VenueHyperliquid:
		var result, err = placeHyperliquidOrder(ctx, venue, order)
		if err != nil {
			return OrderResult{}, fmt.Errorf(
				"place Hyperliquid order: %w",
				err,
			)
		}
		return result, nil
	case VenuePolymarket:
		var market, err = identifyPolymarket(ctx, venue, order.Symbol)
		if err != nil {
			return OrderResult{}, fmt.Errorf(
				"identify Polymarket: %w",
				err,
			)
		}
		var result OrderResult
		result, err = placePolymarketOrder(ctx, venue, market, order)
		if err != nil {
			return OrderResult{}, fmt.Errorf(
				"place Polymarket order: %w",
				err,
			)
		}
		return result, nil
	default:
		return OrderResult{}, fmt.Errorf(
			"place order: unknown venue %q",
			venue.Kind,
		)
	}
}

// Section 3 - Generic Helpers

func formatDate(value time.Time) string {
	return value.UTC().Format(time.DateOnly)
}

func benchmark(started time.Time) time.Duration {
	return time.Since(started)
}
```

## 9. Fixed Errors and Logging

Shared invalid lifecycle state MUST use:

`internal/toolkit/errors.StateError` owns this construction.

```go
func StateError(owner, action string) error {
	return fmt.Errorf("%s cannot %s from current state", owner, action)
}
```

All other errors MUST use `error`, `errors`, and `fmt`.

Custom error frameworks are prohibited.

Internal Nuubot errors MUST use `%w` at boundaries defined in `RULES.md`.

Third-party errors MUST use `%v` or translation unless their contract exposes
them to `errors.Is` or `errors.As`.

`internal/toolkit/logging/logging.go` MUST own logging setup:

```go
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const ServerLog = "server.log"

// New returns the process logger.
func New(output io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(output, nil))
}

// Open returns an append-only file logger.
func Open(name string) (*slog.Logger, error) {
	var err = os.MkdirAll("workspace/logs", 0o755)
	if err != nil {
		return nil, err
	}
	var output *os.File
	output, err = os.OpenFile(
		filepath.Join("workspace/logs", name),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return nil, err
	}
	return New(output), nil
}

// BotLog returns one Bot log filename.
func BotLog(sweepID, botID uint64) string {
	return fmt.Sprintf("bot_%d_%d.log", sweepID, botID)
}
```

Logging MUST return `*slog.Logger`.

Custom Logger wrappers are prohibited.

Identity-bearing executables MUST start with `server.log`, then replace the logger after identity:

```go
package main

import (
	"fmt"
	"os"
	"time"

	"nuubot/internal/btrunner"
	"nuubot/internal/toolkit/logging"
)

const program = "nuubot-btrunner"

func main() {
	var started = time.Now()
	var log, err = logging.Open(logging.ServerLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to open log file:", err)
		os.Exit(1)
	}

	var sweepID, botID uint64
	sweepID, botID, err = parseInput(os.Args[1:])
	if err != nil {
		log.Error("parseInput() failed", "error", err)
		os.Exit(1)
	}

	var botLog, botLogErr = logging.OpenBot(sweepID, botID)
	if botLogErr != nil {
		log.Error("logging.OpenBot() failed", "error", botLogErr)
		os.Exit(1)
	}
	log = botLog

	err = btrunner.Run(log, sweepID, botID)
	if err != nil {
		log.Error("btrunner.Run() failed", "duration", time.Since(started), "error", err)
		os.Exit(1)
	}
	log.Info("btrunner.Run() completed successfully", "duration", time.Since(started))
}
```

Executables MUST NOT write operational output to stdout or stderr after a logger exists.

If `server.log` cannot open, the executable reports that failure to stderr and exits nonzero.

Components MUST receive explicit `*slog.Logger` values.

Components MUST bind `component` with `logger.With`.

Components MUST NOT log returned error values.

Components are allowed to log terminal statistics and failure status before
returning.

Only executable, request, or background-task boundaries log returned errors.

Global `slog` configuration and `slog.SetDefault` are prohibited.

## 10. Tests

Tests MUST use ordinary Go checks and the same three sections.

```go
// Section 1 - Program Flow

func TestRuntimePassStopsAtEndDate(t *testing.T) {
	var actual = runPass(t)
	var expected = "end_date"

	if actual != expected {
		t.Fatalf("actual %q, expected %q", actual, expected)
	}
}

// Section 2 - Domain Helpers

func runPass(t *testing.T) string {
	t.Helper()
	return "end_date"
}
```

Assertions MUST state actual before expected.

Table tests MUST exist only when cases share identical logic.

Test-only interfaces MUST NOT exist solely for mocking.

## 11. Current Source State

The implemented BtRunner path uses this style.

New and changed source MUST preserve this contract.

## 12. Review Test

Hide function bodies before completion.

The remaining declarations MUST reveal:

- ownership;
- lifecycle;
- loops;
- decisions;
- paths;
- domain boundaries; and
- shutdown order.

Otherwise, rewrite the file before review passes.
