# Nuubot5 Structure Suggestions v2

Date: 2026-07-24

Status: Design recommendations for user review. No implementation is approved by
this document.

This report combines the four-project research from `suggestions_v1.md` with
the refined Nuubot5 model discussed afterward.

Where v1 conflicts with v2, v2 supersedes it. The main correction is:

```text
"BotSpec" = prebuilt executable Go composition
BotSpecID = stable Config-facing name selecting one BotSpec
BotConfig = user-controlled parameter data
BotRevision = immutable BotSpec version + BotConfig snapshot
Controller = running Bot brain, renamed from Runtime
BotCycle = one coordinated activation of the configured Executors
```

`BotSpec` remains a working name. It is quoted where useful because
`BotProgram` may ultimately describe executable code more precisely. Do not
rename it until the user approves the final vocabulary.

# Summary

## What to Take from Each Project

| Project | Take | Reject |
|---|---|---|
| [BackNRun](https://github.com/raykavin/backnrun-go) | One-screen composition: context, log, strategy, settings, exchange, bot, run | Global logger, strategy-to-broker coupling, duplicated strategy interfaces, oversized `Bot.Run` |
| [GoBacktest](https://github.com/gobacktest/gobacktest) | Clear `Data -> Signal -> Order -> Fill` vocabulary | Archived implementation, mutable setter construction, fluent tree DSL, swallowed loop errors |
| [GoCryptoTrader](https://github.com/thrasher-corp/gocryptotrader) | Decimal money, multi-asset concepts, explicit data/signal/order/fill/result boundaries, reporting vocabulary | Giant setup path, reflection, plugins, giant interfaces, giant configuration |
| [Indicator](https://github.com/cinar/indicator) | Small constructors, readable indicator composition, focused strategy tests | Channel-first historical calculations and repeated reporting methods |
| Nuubot5 | Reconciliation-first control, exact replay proof, direct synchronous flow, BotCycle-owned Executors, Account/Venue/Ledger correctness | Broad setup context, ambient Bot behavior, single-symbol events, entry-only signals, no-op Risk, fragmented results |

The target is not to copy one project. It is:

```text
BackNRun readability
+ Nuubot5 ownership and correctness
+ GoBacktest vocabulary
+ GoCryptoTrader feature completeness where required
+ Indicator authoring simplicity
```

## Desired Composition Example of a "BotSpec"

The user writes BotConfig:

```toml
name = "btc-multifactor-grid"
bot_spec = "macross-grid-hedge"
network = "mainnet"

[controller]
max_concurrent_cycles = 1
max_cycles = 3

[parameters]
minimum_good_indicators = 6
fast_period = 9
slow_period = 21
maximum_drawdown_pct = "10"

[parameters.primary_grid]
position_size = "100"

[parameters.hedge]
position_size = "25"

[accounts.primary]
venue = "hyperliquid"
credential = "main-account"
```

The selected prebuilt BotSpec contains normal Go code:

```go
package macrossgridhedge

type Config struct {
    MinimumGoodIndicators int
    FastPeriod            int
    SlowPeriod            int
    MaximumDrawdownPct    decimal.Decimal
    PrimaryGridSize       decimal.Decimal
    HedgeSize             decimal.Decimal
}

func Build(config Config) (bot.Definition, error) {
    signals, err := newSignaler(config)
    if err != nil {
        return bot.Definition{}, err
    }

    return bot.Definition{
        Signaler: signals,
        Executors: []executor.Spec{
            newPrimaryGrid(config),
            newHedge(config),
        },
        Risks: []risk.Policy{
            newMaximumDrawdown(config.MaximumDrawdownPct),
        },
    }, nil
}

func (s *Signaler) Check(state signaler.State) signaler.Intent {
    good := countTrue(
        s.fastEMA.Bullish(),
        s.slowEMA.Bullish(),
        s.rsi.Healthy(),
        s.volume.Confirmed(),
        s.volatility.Acceptable(),
        s.trend.Strong(),
        s.marketBreadth.Positive(),
        s.funding.Acceptable(),
    )

    momentum := s.rsi.Momentum() || s.volume.Momentum()

    if good >= s.minimumGood && momentum && state.BTC.Bullish() {
        return signaler.StartCycle
    }
    if state.BTC.Bearish() || good < 3 {
        return signaler.StopCycle
    }
    return signaler.NoAction
}
```

This is an API sketch, not an approved signature.

The important contract is:

```text
BotSpecID
  -> matching BotSpec code
  + user BotConfig
  -> admitted BotDefinition
  -> Controller
  -> BotCycles
  -> Executors
```

BotConfig does not become a trading-language DSL. Complex indicator logic stays
in readable, testable Go.

`bot_spec` selects the only valid schema for the remaining BotConfig:

```text
bot_spec = "macross-grid-hedge"
  -> moving-average/grid/hedge Config fields are valid
  -> RSI-only Config fields are invalid

bot_spec = "rsi-trade"
  -> RSI Config fields are valid
  -> moving-average-only Config fields are invalid
```

Admission resolves BotSpecID first, then strictly decodes the selected BotSpec's
typed Config. Changing only `bot_spec` while leaving incompatible parameters
must fail with unknown or missing fields.

Each BotSpec therefore has its own intended Config file and example:

```text
configs/macross-grid-hedge.toml
  -> bot_spec = "macross-grid-hedge"
  -> botspecs/macrossgridhedge
  -> moving-average parameters

configs/rsi-grid-hedge.toml
  -> bot_spec = "rsi-grid-hedge"
  -> botspecs/rsigridhedge
  -> RSI parameters
```

The files may share common envelope fields such as Controller limits, Account
references, and network. Their Bot-specific parameter sections are not
interchangeable.

The same BotSpec catalogue can power a BotConfig Template Generator:

```text
CLI / Web / AI
  -> request template for BotSpecID
  -> Template Generator
  -> matching BotSpec's canonical template
  -> editable BotConfig
  -> same BotSpec validates the edited Config
```

The clean backtest surface becomes:

```go
func main() {
    ctx := context.Background()
    log := logging.New(...)

    config, err := botconfig.Load(...)
    if err != nil {
        log.Fatal(err)
    }

    definition, err := botspec.Build(config.BotSpecID, config)
    if err != nil {
        log.Fatal(err)
    }

    controller, err := controller.New(log, definition)
    if err != nil {
        log.Fatal(err)
    }

    result, err := backtest.Run(ctx, config.Replay, controller)
    if err != nil {
        log.Fatal(err)
    }

    log.Info(result.Summary())
}
```

Production uses the same composition inside `BotManager.StartBot`.

## Architecture Overview

```text
User / AI
  |-- selects BotSpec
  |-- supplies BotConfig
  `-- sends create/start/stop/clone commands
            |
            v
        nuubot-server
            |
        BotManager
            |
      immutable BotRevision
            |
      +-----+-------------------+
      |                         |
    Runner                    BtRunner
    live feed                 replay files
    WallClock                 TickClock
      |                         |
      +--------> Controller <---+
                    |
          +---------+----------+
          |         |          |
       Signaler    Risks     Accounts
          |                    |
          v                    v
       lifecycle           Venue / Ledger
         intent
          |
          v
     active BotCycles [0..max_concurrent_cycles]
          |
       Executors
```

Sweep ownership remains outside the Bot:

```text
SweepPlan
  -> BotConfig variants
  -> SweepManager
  -> BtRunner per variant
  -> BacktestResults
  -> ranking
```

### Ownership

| Owner | Owns |
|---|---|
| User | BotSpec selection, BotConfig values, and lifecycle commands |
| BotSpecID | Stable Config value mapping to one premade BotSpec |
| BotSpec | Executable algorithm, typed parameters, component composition, data requirements |
| BotRevision | Exact BotSpec ID/version/hash plus BotConfig snapshot/hash |
| BotManager | Runner processes and externally requested Bot lifecycle |
| Runner | Live feed/subscriptions, WallClock, one Controller |
| BtRunner | ReplayReader, TickClock, exact replay proof, one Controller |
| Controller | BotConfig snapshot, Signaler, Risks, Accounts, active BotCycles, cumulative state |
| BotCycle | One admitted activation and its Executors |
| Executor | One execution policy |
| Account | Venue access and one Ledger |
| Venue | External or simulated execution truth |
| Ledger | Trades, Orders, Fills, and local evidence |
| SweepManager | BotConfig variation and backtest coordination |

### Invariants to Preserve

- Backtest and live use the same Controller, BotCycle, Executor, Account, and
  Ledger behavior.
- Controller is synchronous. Runner and BtRunner serialize events into it.
- Reconciliation completes before Risk and execution decisions.
- Signaler calculates intent. It never calls Executors.
- Controller accepts or rejects Signaler intent.
- BotCycle admits every Executor before any Executor trades.
- Account owns Venue and Ledger truth.
- Failed replay never publishes a successful result.
- User configuration is immutable for one Controller generation.
- No generic event bus, dependency injection container, reflection registry, or
  Go plugin is introduced.

# Targeted Recommendations

## Recommendation #1 - Define BotSpec, BotSpecID, BotConfig, BotDefinition, and BotRevision

### What

- **Now:** `internal/datastore.BotSpec` contains one symbol, tick path, replay
  dates, and optional run dates. Strategy, Executors, and Risks come from an
  environment-selected TOML file.
- **New:** BotSpecID is the stable Config-facing name of one prebuilt BotSpec.
  BotSpec is Go code. BotConfig is spec-specific user data. BotDefinition is the
  constructed component set. BotRevision immutably binds exact code and
  configuration.

### Why

The same Bot identity can currently select different trading behavior through
ambient configuration. Complex trading logic also cannot be represented safely
as TOML without creating a new programming language.

The new boundary makes code expressive and runs reproducible.

### How (Changes)

1. Use the approved names BotSpec, BotSpecID, BotConfig, BotDefinition, and
   BotRevision.
2. Define BotSpecID as a stable string such as:

   ```toml
   bot_spec = "macross-grid-hedge"
   ```

3. Map each BotSpecID to exactly one compiled BotSpec builder.
4. Resolve BotSpecID before decoding its spec-specific Config body.
5. Reject unknown, missing, or incompatible fields for the selected BotSpec.
6. Give each BotSpec its own canonical Config example file.
7. Rename the current datastore `BotSpec` to a name matching its actual replay
   input role.
8. Define one small BotSpec entrypoint conceptually equivalent to:

   ```go
   Build(config) (bot.Definition, error)
   ```

9. Let each BotSpec own its typed BotConfig model and validation.
10. Persist:
   - BotSpecID;
   - BotSpec version;
   - source or artifact hash;
   - BotConfig snapshot;
   - BotConfig hash.
11. Keep BotSpec Go source in the source tree, for example:

    ```text
    internal/botspecs/macrossgridhedge/
    internal/botspecs/rsigridhedge/
    ```

12. Do not place Go source in mutable `workspace/**`.
13. Build BotSpecs through one explicit compiled catalogue initially.
14. Let the running program list and select BotSpecs from that compiled
    catalogue.
15. Require rebuild and redeployment after adding or changing BotSpec Go source.
16. Start with a direct switch or map. Add build-time catalogue generation only
    when manual maintenance becomes a demonstrated problem.
17. Keep BotSpec selection out of `internal/executor/executor.go`.
18. Let `executor.go` select only individual Executor implementations requested
    by the already-built BotDefinition.
19. Do not use Go plugins, reflection registration, runtime `.go` loading,
    on-host compilation, or an
    expression DSL.
20. Treat code changes as a new BotSpec version behind the stable BotSpecID.
21. Treat parameter changes as a new BotRevision.
22. Make clone copy BotSpecID, pinned BotSpec version/hash, and BotConfig only.
23. If runtime-installable BotSpecs later become necessary, use separately
    compiled, hashed, process-isolated Runner artifacts behind a narrow
    protocol. Do not load raw Go source into the Server.

### Completion Proof

- The same BotRevision always resolves the same code and parameter values.
- `bot_spec = "macross-grid-hedge"` cannot admit RSI-only parameters.
- Macross and RSI variants use separate canonical Config files and typed
  decoders.
- Ambient `NUUBOT_CONFIG` cannot change Bot trading behavior.
- A BotSpec can contain arbitrary multi-indicator Go logic.
- The running program can list every compiled BotSpec and reject unknown IDs.
- Adding an uncompiled `.go` file cannot silently change a running Server.
- Clone copies no runtime, Account, Ledger, Order, Fill, or credential state.

## Recommendation #2 - Hard-Rename Runtime to Controller

### What

- **Now:** `Runtime` is documented as the running Bot. It owns signal
  consumption, Risk checks, BotCycle admission, stop intent, and results.
- **New:** `Controller` explicitly names the Bot brain controlling execution
  lifecycle.

### Why

`Runtime` is generic infrastructure language. `Controller` explains the actual
responsibility: control signals, Risks, BotCycles, limits, draining, and
shutdown.

This also distinguishes internal trading control from `BotManager`, which owns
external Runner lifecycles.

### How (Changes)

1. Perform one direct hard rename:
   - package `runtime` to `controller`;
   - type `Runtime` to `Controller`;
   - `Runtime.Run` to `Controller.Run`;
   - Runtime config/result/store names to Controller equivalents;
   - source comments, errors, logs, tests, scripts, and wiki pages.
2. Do not add a Controller wrapper around Runtime.
3. Preserve `Run` as one synchronous control pass.
4. Preserve BtRunner `Loop` as repeated replay.
5. Preserve exact reconciliation-first operation order.
6. Remove Controller's dependency on `setup.Context`.
7. Construct Controller from one narrow admitted BotDefinition.
8. Thread caller context from command/Server through setup and Runner/BtRunner.
9. Keep Runner and BtRunner names unchanged.

### Completion Proof

- No active source or canonical documentation refers to the old Runtime owner.
- Controller tests construct it without setup, datastore, files, or
  credentials.
- Existing replay behavior and proof counts remain unchanged.
- No compatibility alias or parallel Runtime path remains.

## Recommendation #3 - Make Controller Own Cycle Capacity and Lifetime

### What

- **Now:** Controller has one `cycle *BotCycle` and `max_cycles`. Signals are
  skipped while the cycle exists.
- **New:** Controller owns a bounded active-cycle collection and two explicit
  limits:

  ```toml
  [controller]
  max_concurrent_cycles = 1
  max_cycles = 3
  ```

### Why

This makes Controller the authoritative execution brain while allowing future
parallel BotCycles without changing BotCycle or Executor ownership.

It also gives precise lifetime behavior:

- capacity controls simultaneous cycles;
- `max_cycles` controls total admitted cycles;
- Risk may stop earlier.

### How (Changes)

1. Replace the single active pointer with an ordered collection keyed by
   CycleID.
2. Require `max_concurrent_cycles >= 1`.
3. Count one cycle only after complete successful BotCycle admission.
4. Do not count rejected admissions.
5. Count admitted cycles even when they later fail.
6. Reject capacity-blocked signals. Do not queue stale signals.
7. Once the final permitted cycle is admitted:
   - enter `draining`;
   - reject every new start;
   - continue recon and management for active cycles;
   - stop Controller when every active cycle becomes terminal.
8. External stop, Risk stop, fatal error, or Runner shutdown may stop earlier.
9. Reset the cycle budget only for a new Controller generation unless recovery
   explicitly resumes the old generation.
10. Record started, rejected, active, closed, failed, and capacity-rejected
    cycle counts.

### Completion Proof

- `max_concurrent_cycles = 1` preserves current sequential behavior.
- No run admits more than `max_cycles`.
- No point exceeds the concurrency limit.
- Final admitted cycles finish or stop gracefully before Controller shutdown.
- Capacity rejection is visible in structured results.

## Recommendation #4 - Replace Entry Direction with Lifecycle Intent

### What

- **Now:** Signaler packages contain `enter_long`, `enter_short`,
  `close_long`, and `close_short`. Controller consumes only entry booleans.
  Every current Executor derives the same side from the triggering Signal.
- **New:** Signaler emits typed Controller lifecycle intent. Each ExecutorSpec
  owns its own symbol, side, role, account reference, and parameters.

### Why

A signal must be able to start a long grid and short hedge together. One global
long/short trigger cannot express opposing or multi-symbol Executors.

Signaler should calculate. Controller should control. Executor should execute.

### How (Changes)

1. Define the minimum typed actions:
   - `NoAction`;
   - `StartCycle`;
   - `StopCycle` or `StopAllCycles`, depending on approved concurrency
     semantics.
2. Do not add generic `Pause` until its trading meaning is approved.
3. If needed later, prefer `SuspendEntries`:
   - block new exposure;
   - continue reconciliation;
   - preserve protective orders;
   - preserve exits and Risk.
4. Keep signal timestamp and useful diagnostic fields.
5. Keep execution-driving values typed.
6. Restrict arbitrary custom fields to telemetry unless a typed contract is
   approved.
7. Move long/short direction from Signal into each ExecutorSpec.
8. Never let Signaler call Controller, BotCycle, Executor, Account, or Venue.
9. Controller consumes each signal timestamp once and records its decision.

### Completion Proof

- One StartCycle intent can admit opposing Executors.
- Controller may reject intent because of Risk, capacity, draining, or state.
- Signaler contains no execution or Account dependency.
- Every ignored or rejected intent has a structured reason.

## Recommendation #5 - Keep BotCycle as the Coordinated Executor Unit

### What

- **Now:** BotCycle creates multiple Executors, rolls back earlier Executors
  when later admission fails, dispatches capabilities, and completes when all
  Executors become terminal.
- **New:** Preserve BotCycle, but make its coordinated multi-Executor contract
  explicit.

### Why

`ExecutorGroup` sounds like a static collection. `BotCycle` correctly means one
time-bounded activation that may repeat throughout one Controller generation.

This is the key Nuubot5 composition primitive.

### How (Changes)

1. Keep BotCycle as the owner of one admitted set of Executors.
2. Give every BotCycle a durable CycleID and sequence number.
3. Require all Executors to initialize before any Venue mutation.
4. Preserve rollback when any Executor rejects or fails admission.
5. Make the default fatal policy:
   - any Executor error stops the entire BotCycle;
   - every sibling Executor receives the same parent stop reason.
6. Define normal completion policy explicitly before implementation:
   - recommended default: one Executor terminal decision closes the coordinated
     cycle;
   - alternative: all Executors must independently complete.
7. Stop Executors in deterministic reverse ownership order.
8. Preserve reconciliation as one complete barrier.
9. Return one immutable BotCycleResult containing:
   - Cycle identity;
   - triggering intent;
   - Executor identities and results;
   - Account references;
   - start/end timestamps;
   - stop reason;
   - failure evidence.
10. Do not create an additional ExecutorGroup layer.

### Completion Proof

- DCA+hedge, grid+hedge, opposing grids, and multi-leg compositions admit as
  one BotCycle.
- No Executor trades after partial group admission.
- One fatal Executor cannot leave sibling Executors unmanaged.
- BotCycle result explains every member outcome.

## Recommendation #6 - Give Risk Controller-Wide State

### What

- **Now:** `Risk.AssessStop() bool` receives no state. Controller discards
  reconciled Account snapshots before Risk. BalancedRisk never requests exit.
- **New:** Controller builds one immutable BotSnapshot across active and
  completed cycles and passes it into every Risk policy.

### Why

The overnight control requirement depends on cumulative truth:

```text
Cycle 1 loses 6%
Cycle 2 loses 4%
next signal arrives
maximum Bot drawdown is 10%
Controller must reject or stop
```

Per-Executor or per-cycle Risk cannot enforce that Bot-level rule.

### How (Changes)

1. Define a small immutable BotSnapshot containing:
   - Bot/Revision/Generation identity;
   - active Cycle summaries;
   - completed Cycle results;
   - deduplicated Account snapshots;
   - starting/current/peak equity;
   - realized and unrealized PnL;
   - current and maximum drawdown;
   - exposure by Account and symbol;
   - current timestamp.
2. Change Risk from a no-input boolean to a decision over BotSnapshot.
3. Begin with only necessary decisions:
   - continue;
   - reject new cycle;
   - stop active cycles;
   - stop Controller.
4. Reconcile before building BotSnapshot.
5. Evaluate Risk before Signaler admission and before Executor actions.
6. Persist only the minimum Risk state needed for generation recovery.
7. Remove BalancedRisk from active defaults until it has real behavior.
8. Do not add a RiskManager package until shared behavior proves one is needed.

### Completion Proof

- A maximum-drawdown Risk blocks the next cycle after cumulative loss reaches
  its threshold.
- Risk sees one coherent post-reconciliation snapshot.
- Shared Accounts are counted once.
- A Risk decision records policy, threshold, observed value, action, and time.

## Recommendation #7 - Make Accounts Named Controller-Generation Resources

### What

- **Now:** Each Account is privately owned by one Executor and created inside
  one BotCycle.
- **New:** BotDefinition declares named Accounts once. Controller owns their
  generation lifecycle. Executors reference Accounts by stable ID through
  narrow execution calls.

### Why

Composed or concurrent Executors may use the same physical Hyperliquid account.
Separate Account/Ledger objects would duplicate balances, margin, positions,
reconciliation, and available equity.

Accounts must also retain cumulative truth across sequential BotCycles.

### How (Changes)

1. Declare Accounts once in BotConfig/BotDefinition.
2. Resolve credential references only during selected live Account admission.
3. Create one Account object per unique Account ID per Controller generation.
4. Let each ExecutorSpec reference one or more Account IDs.
5. Keep Venue and Ledger privately owned by Account.
6. Let Executors submit narrow intents through their admitted Account
   references.
7. Move reconciliation ownership to Controller -> Accounts:
   - reconcile each unique Account once;
   - build one barrier;
   - evaluate Risk;
   - deliver accepted snapshots to relevant BotCycles/Executors.
8. Keep Controller from reaching into Ledger, Trade, Order, or Fill.
9. Include CycleID and ExecutorID on every submitted trading intent so shared
   Account evidence remains attributable.
10. Hardcut the old per-Executor Account lifecycle after parity proof. Do not
    preserve both paths.

### Completion Proof

- Two Executors sharing `accounts.primary` observe one balance and one Ledger.
- Concurrent cycles cannot double-count the same Account equity.
- Reconciliation queries each physical Account once per barrier.
- Every Trade, Order, and Fill remains attributable to Bot, Generation, Cycle,
  and Executor.

## Recommendation #8 - Add Symbol-Qualified Multi-Asset Data

### What

- **Now:** datastore BotSpec has one symbol. `market.BBO` has timestamp and
  price but no symbol. Controller retains one latest BBO. Signaler initializes
  from one replay source.
- **New:** BotSpec declares its market-data requirements. Runner and BtRunner
  deliver equivalent symbol-qualified events into Controller.

### Why

Pairs trading, cross-market filters, multi-symbol grids, and BTC-regime checks
cannot be correct with one implicit symbol.

### How (Changes)

1. Add symbol identity to BBO, bars, snapshots, and other market events.
2. Replace Controller's single latest BBO with symbol-keyed admitted state.
3. Let BotSpec/Signaler declare required symbols, event types, intervals, and
   warmup.
4. BtRunner loads and deterministically merges required historical streams.
5. Runner asks DataEngine for equivalent live subscriptions.
6. Define and document deterministic ordering for equal timestamps.
7. Preserve no-lookahead behavior for closed bars.
8. Reject BotDefinition admission when required data is unavailable.
9. Keep market decoding and WebSocket/REST clients outside Controller.
10. Test one two-symbol BotSpec through both simulated event delivery and
    historical replay.

### Completion Proof

- A BotSpec may evaluate BTC while trading another symbol.
- A pairs BotCycle receives correct symbol-specific prices.
- Replay order is deterministic across repeated runs.
- Controller code is independent of file, WebSocket, and REST transports.

## Recommendation #9 - Make Admission Typed and Reproducible

### What

- **Now:** setup returns a broad Context containing config, credentials,
  datastore Bot data, metadata, and result paths. Raw decimal strings and
  tagged-union config survive deep into construction. Backtests may refresh
  current metadata.
- **New:** ingress converts external data once into one admitted, typed,
  immutable BotRevision and run input.

### Why

Construction is difficult to read and the same historical run may depend on
ambient config, unnecessary credentials, current network state, or changed
instrument rules.

### How (Changes)

1. Separate:
   - AppConfig: server paths, logging, database, timeouts;
   - BotSpecID: selected premade BotSpec name;
   - BotConfig: user-selected values valid only for that BotSpec;
   - BacktestInput: replay range and data identity;
   - credential references: no raw secrets in BotConfig.
2. Decode the common Config envelope only far enough to resolve BotSpecID.
3. Dispatch to that BotSpec's typed Config decoder.
4. Validate only that BotSpec's supported configuration options.
5. Reject unknown, missing, and cross-kind fields.
6. Parse decimal and duration values once at ingress.
7. Materialize defaults into the admitted BotConfig snapshot.
8. Remove Controller's dependency on setup.
9. Load live credentials only when constructing a live Account.
10. Require no private credentials for simulator backtests.
11. Pin metadata snapshot/version/hash for historical execution.
12. Include BotSpecID, BotSpec, BotConfig, replay data, metadata, and build hashes in the
    run identity.
13. Derive timeouts from caller context, never from a new background context.

### Completion Proof

- Controller accepts only typed trusted values.
- Simulator backtests run without credential files.
- Repeating identical admitted inputs yields identical effective composition.
- Result identifies every behavior-affecting input.

## Recommendation #10 - Return One First-Class Result Hierarchy

### What

- **Now:** Controller result contains BotCycles, Account results are nested,
  replay proof remains private, and publishers create separate artifacts.
- **New:** One immutable result hierarchy supports notebooks, reports, sweeps,
  AI evaluation, and persistence.

### Why

The Python-like experience should be:

```text
result = run(botRevision)
```

Users should not scrape logs or know internal publishers.

### How (Changes)

1. Define ControllerResult:
   - Bot/Revision/Generation identity;
   - final lifecycle state;
   - cumulative Bot metrics;
   - Risk decisions;
   - BotCycleResults;
   - Account results.
2. Define BacktestResult as ControllerResult plus:
   - replay identity and range;
   - data and metadata hashes;
   - exact replay proof;
   - simulated and elapsed time;
   - terminal publication status.
3. Preserve Executor identity in Cycle results.
4. Preserve Bot/Cycle/Executor identity in Trades, Orders, and Fills.
5. Make text reports, plots, publishers, and Sweep ranking consume Result.
6. Keep reporting out of BotSpec and Signaler.
7. Publish only after successful Controller stop and replay proof.
8. Keep failed/partial results explicitly non-successful.

### Completion Proof

- One returned object answers identity, lifecycle, performance, Risk, cycle,
  Account, Trade, Order, Fill, and proof questions.
- A notebook requires no internal package access.
- A failed replay cannot appear as a completed result.
- Report code imports no concrete BotSpec.

## Recommendation #11 - Put HTTP and CLI Above BotManager

### What

- **Now:** `nuubot-server`, `nuubot-cli`, and `nuubot-runner` are approved
  placeholders. BtRunner is the implemented process.
- **New:** PocketBase custom routes call BotManager. CLI and GUI use the same
  HTTP commands. BotManager owns Runner generations.

### Why

One authoritative Server path prevents CLI, GUI, and internal callers from
starting Bots differently.

### How (Changes)

1. Expose:
   - `CreateBot`;
   - `CloneBot`;
   - `StartBot`;
   - `StopBot`;
   - `GetBot`;
   - `ListBots`.
2. Suggested routes:

   ```text
   POST /api/v1/mainnet/bots
   POST /api/v1/mainnet/bots/{id}/start
   POST /api/v1/mainnet/bots/{id}/stop
   POST /api/v1/mainnet/bots/{id}/clone
   GET  /api/v1/mainnet/bots/{id}
   GET  /api/v1/mainnet/bots
   ```

3. CLI reads the user file and calls HTTP. It does not import Controller.
4. Server owns BotConfig decoding and canonical validation.
5. BotManager owns active Runner handles and cancellation.
6. Derive Bot lifetime context from Server/BotManager context, never the HTTP
   request context.
7. Persist BotRevision and generation.
8. Make repeated start/stop idempotent.
9. Reject stale-generation mutations.
10. Clone only BotSpec reference and BotConfig.
11. Use PocketBase realtime for status updates when the GUI needs them.
12. Keep trading mutations behind custom routes, not generic collection CRUD.

### Completion Proof

- CLI and GUI invoke the same Server operations.
- One Bot ID cannot have duplicate active generations.
- Server restart can identify desired and observed Bot state.
- Stale commands cannot mutate a newer generation.

## Recommendation #12 - Put Sweeps and AI Authoring Above BotSpec

### What

- **Now:** BtRunner executes one stored Bot while strategy/config ownership is
  fragmented.
- **New:** SweepManager varies BotConfig for one BotSpec. AI may draft new
  BotSpec code and BotConfig, but every artifact passes normal build,
  validation, test, and backtest gates.

### Why

Parameter search and algorithm search are different:

```text
same BotSpec + many BotConfigs = parameter sweep
different BotSpec versions = algorithm search
```

Keeping them separate preserves reproducibility and makes AI output auditable.

### How (Changes)

1. Let each BotSpec declare tunable parameters, types, units, ranges, and
   constraints.
2. Define SweepPlan as:
   - base BotRevision;
   - parameter space;
   - dataset;
   - objective metrics;
   - resource limits.
3. SweepManager creates immutable BotConfig variants.
4. Each variant runs through normal BtRunner and returns BacktestResult.
5. Rank only declared Result metrics.
6. Provide canonical BotSpec examples and an authoring contract for AI.
7. AI-generated deliverables should include:
   - Go BotSpec code;
   - typed BotConfig;
   - example config;
   - focused behavior tests;
   - stated assumptions and data requirements.
8. Compile and test AI code before any backtest.
9. Do not execute arbitrary generated code inside the production Server.
10. Start with compiled-in BotSpecs. Consider isolated compiled Runner artifacts
    only when dynamic deployment becomes a proven requirement.
11. Add the approved template endpoint for CLI and Web use. Add broader schema
    or component-catalogue endpoints only when a proven consumer needs them.

### Completion Proof

- Sweep changes parameter data without changing BotSpec source.
- Every result records exact BotSpec and BotConfig identity.
- AI cannot bypass Controller, Risk, Account, or Venue contracts.
- Search output can be reproduced from stored artifacts.

## Recommendation #13 - Restore Documentation Truth and Remove False Surface

### What

- **Now:** PROJECT, ARCHITECTURE, DESIGN, logic pages, coding examples, and
  implementation disagree about completed packages and BtRunner APIs. Some
  interfaces advertise unused or proof-only capabilities.
- **New:** One current vocabulary and ownership model appears in source,
  canonical wiki, root README, examples, and logs.

### Why

A clean public composition surface cannot remain clean when documentation
describes different owners or obsolete calls. AI authoring especially depends
on unambiguous canonical contracts.

### How (Changes)

1. Approve vocabulary before implementation:
   - BotSpecID;
   - BotSpec working name;
   - BotConfig;
   - BotDefinition;
   - BotRevision;
   - Controller;
   - BotCycle;
   - Executor.
2. Reconcile PROJECT, ARCHITECTURE, DESIGN, HANDOFF references, package pages,
   concept pages, and coding examples.
3. Mark legacy logic pages historical or migrate their remaining truth.
4. Add a short root README after APIs stabilize:
   - purpose;
   - architecture diagram;
   - one BotSpec;
   - one BotConfig;
   - one backtest;
   - one Result;
   - canonical wiki links.
5. Remove unused Executor Signaler context if still unused.
6. Remove Observer's proof-only BBO ingestion capability if behavior remains
   unchanged.
7. Keep replay proof at BtRunner/Simulator ownership boundaries.
8. Update design and source together for every ownership hardcut.
9. Do not document no-op Risk as protection.

### Completion Proof

- Every documented public call exists.
- No implemented package is labelled unimplemented.
- Ownership diagrams match source.
- A human or AI can locate one canonical BotSpec authoring example.
- No interface exists solely to make architecture appear more flexible.

## Recommendation #14 - Add a BotConfig Template Generator

### What

- **Now:** Users must manually create BotSpec-specific Config files and know
  every valid section and field.
- **New:** One small Template Generator accepts BotSpecID and returns that
  BotSpec's canonical editable BotConfig through CLI or HTTP.

### Why

The selected BotSpec already identifies one typed Config schema. It can also
identify one safe canonical template.

This improves human, Web, CLI, and AI authoring without weakening strict
BotSpec-specific validation.

### How (Changes)

1. Make each BotSpec own:
   - its typed Config model;
   - its strict validation;
   - one canonical `config.example.toml`.
2. Preserve human-readable comments, field ordering, units, and safe starting
   values in the canonical template.
3. Prefer safe simulator defaults.
4. Include `bot_spec` in every generated file.
5. Include credential references only. Never include secrets.
6. Add one small domain utility conceptually equivalent to:

   ```go
   Generate(botSpecID string) ([]byte, error)
   ```

7. Resolve templates through the same explicit BotSpec catalogue used for
   BotSpec construction.
8. Do not place this utility in Controller, Executor, Signaler, or generic
   toolkit code.
9. Add CLI operations:

   ```text
   nuubot-cli bot kinds
   nuubot-cli bot template macross-grid-hedge
   nuubot-cli bot template macross-grid-hedge --output bot.toml
   nuubot-cli bot validate bot.toml
   ```

10. Refuse to overwrite an existing output file unless the user explicitly
    requests replacement.
11. Add Server routes:

    ```text
    GET  /api/v1/bot-specs
    GET  /api/v1/bot-specs/{botSpecID}/template
    POST /api/v1/bot-specs/{botSpecID}/validate
    ```

12. Let the Web interface:
    - select BotSpec;
    - fetch its template;
    - edit the Config;
    - validate it;
    - create the Bot only after successful validation.
13. Let AI use the same template and validation endpoints.
14. Test every canonical template through:
    - BotSpecID match;
    - strict typed decode;
    - unknown-field rejection;
    - semantic validation;
    - successful BotDefinition construction.
15. Do not generate templates through reflection initially. Struct tags cannot
    preserve good comments, examples, ordering, or recommended values.

### Completion Proof

- Every listed BotSpec returns one canonical template.
- Generated Config contains the requested BotSpecID.
- The unchanged generated template passes strict validation and builds its
  BotDefinition.
- A Macross template cannot validate as an RSI BotSpec.
- CLI, Web, and AI receive identical template bytes.
- Unknown BotSpecIDs fail without creating a file.
- Existing files are preserved unless replacement is explicit.

# Recommended Implementation Order

This order minimizes parallel paths and protects existing replay proof:

1. Approve terminology and ownership contracts in design.
2. Define BotSpec/BotConfig/BotDefinition/BotRevision without changing trading
   behavior.
3. Hard-rename Runtime to Controller and remove setup dependency.
4. Move current behavior through the new composition surface.
5. Add Controller cycle capacity, draining, and typed lifecycle intent.
6. Make BotCycle coordinated failure/completion behavior explicit.
7. Introduce Controller-wide BotSnapshot and one real Risk policy.
8. Resolve Controller-generation Account ownership before concurrent or shared
   Account execution.
9. Add symbol-qualified events and multi-source Runner/BtRunner parity.
10. Add the first-class Result hierarchy.
11. Build BotManager/Server/CLI control.
12. Add SweepPlan and AI authoring/catalog surfaces from proven needs.
13. Add the BotConfig Template Generator to CLI and Server API.
14. Reconcile README and canonical documentation after each approved hardcut.

Each approved recommendation should become a separate plan or a deliberately
grouped ownership hardcut. Do not implement all recommendations as one change.

# Explicit Non-Goals

Do not add:

- a BotSpec DSL;
- Go plugins;
- reflection-based registration;
- reflection-generated Config templates;
- a dependency injection container;
- a generic event bus;
- a second backtest trading engine;
- goroutine-first indicator calculation;
- a giant common strategy interface;
- a shallow Controller wrapper around Runtime;
- a separate ExecutorGroup object;
- a no-op Risk advertised as protection;
- live config mutation inside one Controller generation;
- compatibility bridges after an approved hardcut;
- a new optimizer when SweepManager owns parameter variation.

# Top Recommendation

Approve this core vocabulary and ownership first:

```text
User selects BotSpec and supplies its BotConfig
  -> immutable BotRevision
  -> Runner or BtRunner
  -> Controller
      -> Signaler
      -> Controller-wide Risks
      -> named Accounts
      -> bounded active BotCycles
          -> coordinated Executors
  -> first-class Result
```

Then make the first implementation change a hardcut that introduces the
BotSpec/BotConfig/BotDefinition boundary and renames Runtime to Controller
without changing proven replay behavior.
