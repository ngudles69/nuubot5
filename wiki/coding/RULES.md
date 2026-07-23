# Mandatory Coding Rules

These rules are literal requirements.

Exceptions require explicit prior user approval and an `AGENTS.md` Key
Decision.

## 1. Priority

Code MUST satisfy:

1. correct behavior;
2. readable intent;
3. simple idiomatic Go;
4. exact proof; and
5. measured performance.

Hypothetical extensibility is prohibited.

## 2. Dependency Order

Stop at the first adequate solution:

1. existing Nuubot5 code;
2. Go standard library;
3. existing approved dependency;
4. approved maintained pure-Go library; and
5. minimum custom implementation.

Standard-library packages require no separate approval.

An approved dependency requires no new approval for its existing purpose.

New and upgraded external dependencies require explicit prior user approval.

An external dependency MUST:

- remove meaningful implementation or maintenance;
- be maintained and license-compatible;
- support Windows and Linux;
- preserve straightforward builds and tests;
- be pure Go; and
- solve a current requirement.

Adding or upgrading a dependency MUST run:

1. `go mod tidy`;
2. focused affected-path tests; and
3. `govulncheck ./...`.

Non-standard Go requires explicit prior user approval.

Non-standard Go includes:

- CGO and native C/C++ bindings;
- `unsafe`;
- handwritten assembly;
- `//go:linkname`;
- runtime and compiler internals;
- Go plugins;
- dynamically loaded native libraries; and
- non-standard compilers or toolchains.

Standard Go build tags, including canonical `noasm`, remain allowed.

Wrappers MUST NOT reproduce a library API.

A wrapper MUST own project policy, stable configuration, or a domain boundary.

## 3. Scope

Every changed line MUST serve the confirmed task.

Adjacent code MUST NOT be reformatted, renamed, or refactored.

A new abstraction MUST own a real behavioral, external, or domain boundary.

Interfaces MUST be consumer-owned and minimal.

Interfaces MUST NOT exist only for mocking.

Ordinary constructors MUST return concrete types.

An approved configuration-selected factory is allowed to return its
consumer-owned interface.

Placeholder packages, files, interfaces, configuration, adapters, and
factories are prohibited.

New dead code MUST be removed.

Pre-existing unrelated dead code MUST be reported, not removed.

## 4. Ownership and Lifecycle

Every mutable component MUST have one owner.

A parent MUST control only direct children.

Components MUST exchange values or narrow intent calls.

Components MUST NOT expose internal mutable state.

Lifecycle methods MUST remain ordered:

```text
New
Init
Start
Loop or Run
Stop
```

`Loop` MUST repeat work until its stop condition.

`Run` MUST execute one operation.

Repeated iteration MUST belong inside `Loop`.

`Pass` MUST execute one timer-driven control pass.

Empty lifecycle phases MUST be omitted.

`Stop` MUST be safe after successful `Start`.

Valid repeated shutdown paths MUST make `Stop` idempotent.

Event and input producers MUST stop before their consumers.

Remaining children MUST unwind in reverse ownership order.

Primary work errors MUST NOT be hidden by shutdown errors.

## 5. Errors

Nuubot5 MUST use standard `error`, `errors`, and `fmt`.

Custom error-construction helpers and packages are prohibited.

Operational failures MUST return `error`.

Leaf errors MUST name the failed operation and relevant identity.

Internal errors MUST wrap only at:

- exported package operations;
- lifecycle operations;
- Domain Helpers; and
- executable or background-task boundaries.

Unexported same-owner functions MUST return received errors unchanged.

Domain Helpers remain wrap boundaries.

```go
bars, err := readBars(path)
if err != nil {
	return fmt.Errorf("load signaler bars: %w", err)
}
```

Allowed internal wraps MUST use `%w`.

`%w` exposes the wrapped error to the caller.

Third-party errors MUST use `%v` or translation.

Third-party `%w` requires a documented `errors.Is` or `errors.As` contract.

Exposed sentinel and typed errors MUST be documented.

Error text MUST be lowercase, concise, and unpunctuated.

Returned errors MUST be logged exactly once.

Logging ownership belongs to the executable, request, or background-task
boundary.

Lower components MUST NOT log returned error values.

Lower components are allowed to log useful terminal statistics.

`panic` MUST NOT handle operational failures.

Retries, skips, repairs, defaults, and fallbacks require explicit recovery
contracts.

External input MUST be validated before mutation, persistence, or external
calls.

An identity-bearing executable MUST:

```go
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
		log.Error(fmt.Sprintf("parseInput() failed: %v", err))
		os.Exit(1)
	}

	var botLog, botLogErr = logging.OpenBot(sweepID, botID)
	if botLogErr != nil {
		log.Error(fmt.Sprintf("logging.OpenBot() failed: %v", botLogErr))
		os.Exit(1)
	}
	log = botLog

	var runner *btrunner.BtRunner
	runner, err = btrunner.Init(log, sweepID, botID)
	if err != nil {
		log.Error(fmt.Sprintf("btrunner.Init() failed: %v", err))
		os.Exit(1)
	}
	err = runner.Start()
	if err != nil {
		err = errors.Join(err, runner.Stop())
		log.Error(fmt.Sprintf("btrunner.Start() failed: %v", err))
		os.Exit(1)
	}
	var loopErr = runner.Loop()
	var stopErr = runner.Stop()
	if loopErr != nil {
		log.Error(fmt.Sprintf("btrunner.Loop() failed: %v", errors.Join(loopErr, stopErr)))
		os.Exit(1)
	}
	if stopErr != nil {
		log.Error(fmt.Sprintf("btrunner.Stop() failed: %v", stopErr))
		os.Exit(1)
	}
	log.Info(fmt.Sprintf("btrunner completed successfully in %s", time.Since(started)))
}
```

## 6. Logging

Nuubot5 MUST use `internal/toolkit/logging.Logger`.

`internal/toolkit/logging.Open` MUST own directories, filenames, append-only opening, and record formatting.

No executable or component may write operational output to stdout or stderr after a logger exists.

If `server.log` cannot open, the executable MUST report that failure to stderr and exit nonzero.

Logging MUST be configured once at each executable or server boundary.

Components MUST receive explicit `*logging.Logger` values named `log`.

Components MUST NOT open log files or configure handlers.

Every log call MUST receive exactly one complete message string.

Every line MUST use `YYYY-MMM-DD HH:MM:SS [LEVEL] message`.

Levels MUST be uppercase and right-aligned to a minimum width of five.

Terminal logs MUST prove completion, locate failures, measure work, or report
domain results.

Components MUST report their own statistics.

Parents MUST NOT collect child statistics only to re-log them.

Required:

```go
var message = fmt.Sprintf(
	"runtime stopped ticks_accepted=%d passes=%d",
	stats.ticks,
	stats.passes,
)
log.Info(message)
```

Prohibited:

```go
log.Info("runtime stopped", "ticks_accepted", stats.ticks)
```

## 7. Concurrency

Synchronous code MUST remain synchronous until concurrency is required.

Every goroutine MUST have an owner, stop condition, and error path.

Long-lived goroutines MUST accept `context.Context`.

Channels MUST carry events or ownership transfer.

Channels MUST NOT hide shared mutable state.

A mutex MUST protect one stated invariant.

A mutex MUST NOT compensate for unclear ownership.

WebSocket readers MUST publish typed events or update owned feed state.

Runtime decisions MUST remain in Runtime's timed synchronous pass.

BBO checks MUST use the configured fast cadence.

Dirty reconciliation MUST use the configured slower cadence.

`wiki/logic/runner.md` owns cadence defaults.

## 8. Data and Safety

Boundary data MUST be validated once and converted into trusted Go types.

Timestamps, identities, prices, quantities, and venue outcomes MUST NOT be
invented.

Parquet replay timestamps MUST retain admission validation.

Approved non-standard Go requires documented invariants and focused proof.

Secrets MUST NOT enter source, wiki, logs, tests, handoff, or prompts.

## 9. Proof

Every task MUST define observable success before implementation.

A bug fix MUST leave the smallest failing check.

A refactor MUST prove behavior before and after.

Execution changes MUST pass focused tests and one real operator path.

Invalid input MUST be proven to fail as specified.

Full stability runs require explicit user approval.

Exit zero MUST NOT prove replay completion alone.

Semantic counts and completion markers MUST also match.

Completion MUST report proof run and proof omitted.

## 10. Authority

The user MUST approve edits and consequential commands before execution.

The user MUST approve new dependencies, upgrades, non-standard Go, commits,
pushes, and external effects.

Nuubot4 is reference material only.

Nuubot5 owns its implementation and proof.

## 11. Current Source State

The implemented BtRunner path follows these rules.

New and changed source MUST preserve this contract.
