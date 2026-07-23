# Nuubot5 Design

## Purpose

This page catalogs every current key object and approved future boundary.

Detailed algorithms belong in `wiki/logic/**`.

## Status Values

- **Implemented**: current Nuubot5 source exists and participates in a real path.
- **Stub**: current source exists but deliberately performs incomplete domain behavior.
- **Approved design**: the contract is approved, but Nuubot5 has no implementation.
- **Future**: expected scope without an approved detailed contract.
- **Excluded**: prohibited from the named scope.

## Current Entry and Configuration

### BtRunner command

**Status:** Implemented  
**Source:** `cmd/nuubot-btrunner/main.go`  
**Owner:** Operating system process.  
**Owns:** Arguments, configuration load, log file, logger, and one BtRunner.  
**Responsibility:** Run one Sweep Bot and return one terminal process result.  
**Must not:** Own replay, trading decisions, indicators, or child statistics.  
**Inputs:** Sweep ID, Bot ID, working directory, and `config.toml`.  
**Outputs:** Process exit status, stdout, stderr, and Bot log.  
**Lifecycle:** Parse, load, create, start, run, stop, exit.  
**Details:** [BtRunner Logic](logic/btrunner.md).

### Configuration types

**Status:** Implemented  
**Source:** `internal/config/config.go`  
**Owner:** Command loads `Config`; consumers receive their sections.  
**Owns:** `Config`, `Paths`, `BtRunner`, `Runtime`, `Signaler`, `Executor`, and `Risk` values.  
**Responsibility:** Decode TOML, reject unknown fields, validate supported values, and root relative paths.  
**Must not:** Start components, load Bot rows, or hide runtime state.  
**Inputs:** `config.toml` path.  
**Outputs:** Validated immutable configuration values.  
**Lifecycle:** `Load` once before object construction.

### Setup Context

**Status:** Implemented  
**Source:** `internal/setup/setup.go`  
**Owner:** BtRunner construction.  
**Owns:** Validated `Config` and `BotSpec` values only.  
**Responsibility:** Load one BotSpec and prove its market-data path remains inside shared data.  
**Must not:** Create Runtime children or retain database resources.  
**Inputs:** Root, configuration, Sweep ID, and Bot ID.  
**Outputs:** `setup.Context`.  
**Lifecycle:** `Init` performs one bounded setup operation.

### BotSpec

**Status:** Implemented  
**Source:** `internal/datastore/models.go`, `internal/datastore/sweep.go`  
**Owner:** Setup Context, then BtRunner construction.  
**Owns:** Symbol, tick path, replay range, and optional Bot start and end timestamps.  
**Responsibility:** Represent one validated Bot configuration loaded from SQLite.  
**Must not:** Own a database connection, Runtime, or mutable Bot state.  
**Inputs:** Read-only Sweep Bot JSON row.  
**Outputs:** Trusted `BotSpec`.  
**Lifecycle:** Loaded once and passed by value.

## Current Market Data

### Bars types and loader

**Status:** Implemented  
**Source:** `internal/bars/bars.go`  
**Owner:** BtRunner invokes loading; Signaler retains returned data.  
**Owns:** `Timeframe`, `Requirement`, `Data`, Parquet decoding, validation, and warmup-range assembly.  
**Responsibility:** Return aligned, complete OHLCV bars for requested timeframes.  
**Must not:** Calculate indicators, select strategies, or drive Runtime.  
**Inputs:** Tick path, replay range, and Signaler requirements.  
**Outputs:** Validated `[]bars.Data`.  
**Lifecycle:** Load once before Signaler preparation.  
**Details:** [Signaler Logic](logic/signaler.md).

### BBO

**Status:** Implemented  
**Source:** `internal/market/market.go`  
**Owner:** Passed by value between replay, Runtime, BotCycle, and Executor.  
**Owns:** Normalized timestamp and price only.  
**Responsibility:** Represent one validated best-price replay event.  
**Must not:** Read data, track history, or execute policy.  
**Inputs:** Millisecond timestamp and positive finite price.  
**Outputs:** Immutable `market.BBO` value.  
**Lifecycle:** Construct, pass by value, discard.

### replay.Reader

**Status:** Implemented  
**Source:** `internal/replay/parquet.go`  
**Owner:** BtRunner.  
**Owns:** Monthly file sequence, Arrow record reader, current batch, admission state, and reader statistics.  
**Responsibility:** Stream validated one-second BBO values across the requested range.  
**Must not:** Drive Clock, call Runtime, or calculate bars.  
**Inputs:** Tick directory and bounded replay range.  
**Outputs:** One `market.BBO`, completion, or error per `Next` call.  
**Lifecycle:** `NewReader`, repeated `Next`, idempotent `Stop`.  
**Details:** [BtRunner Logic](logic/btrunner.md).

### TickClock

**Status:** Implemented  
**Source:** `internal/clock/clock.go`  
**Owner:** BtRunner.  
**Owns:** Interval, next due timestamp, tick count, pass count, and stop state.  
**Responsibility:** Convert replay timestamps into bounded Runtime pass decisions.  
**Must not:** Read wall time, call Runtime, or own callbacks.  
**Inputs:** Validated BBO timestamps.  
**Outputs:** Due flag or overflow error.  
**Lifecycle:** `New`, repeated `Advance`, idempotent `Stop`.

## Current Signals

### Signal values

**Status:** Implemented  
**Source:** `internal/signaler/signaler.go`  
**Owner:** Signaler creates them; Runtime and BotCycle consume them.  
**Owns:** Signal timestamp, availability timestamp, side, and price.  
**Responsibility:** Represent one ordered trading signal without lookahead.  
**Must not:** Own bars, execution, or mutable lifecycle.  
**Inputs:** Concrete calculator output.  
**Outputs:** Immutable `signaler.Signal` values.  
**Lifecycle:** Calculate once, release once, pass by value.

### Signaler

**Status:** Implemented  
**Source:** `internal/signaler/signaler.go`  
**Owner:** Runtime.  
**Owns:** Selected calculator, loaded bars, calculated signals, release position, and lifecycle state.  
**Responsibility:** Declare bar needs, calculate once, validate ordering, and release available signals.  
**Must not:** Read Parquet, open BotCycles, or execute orders.  
**Inputs:** Signaler configuration, validated bars, and current Runtime timestamp.  
**Outputs:** Bar requirements and ordered Signal values.  
**Lifecycle:** `New`, `Prepare`, `Start`, repeated `Next`, idempotent `Stop`.  
**Details:** [Signaler Logic](logic/signaler.md).

### Signaler calculator factory

**Status:** Implemented  
**Source:** `internal/signaler/signaler.go`  
**Owner:** Signaler construction.  
**Owns:** Selection between Macross and RSI calculators.  
**Responsibility:** Return the configured calculator behind Signaler's private consumer-owned interface.  
**Must not:** Add speculative implementations or expose calculator types outside the package.  
**Inputs:** `config.Signaler.Kind`.  
**Outputs:** One calculator or error.  
**Lifecycle:** One selection during `signaler.New`.

### Macross calculator

**Status:** Implemented  
**Source:** `internal/signaler/macross.go`  
**Owner:** Signaler.  
**Owns:** Signal timeframe, regime timeframe, EMA periods, and indicator calculation.  
**Responsibility:** Produce filtered EMA crossover Signals from closed signal and regime bars.  
**Must not:** Use unclosed regime bars, release Signals, or own lifecycle.  
**Inputs:** Validated bars and Macross configuration.  
**Outputs:** Ordered `[]Signal`.  
**Lifecycle:** Construct, report requirements, calculate once.  
**Details:** [Signaler Logic](logic/signaler.md).

### RSI calculator

**Status:** Implemented  
**Source:** `internal/signaler/rsi.go`  
**Owner:** Signaler.  
**Owns:** Timeframe, RSI period, volume period, and indicator calculation.  
**Responsibility:** Produce RSI Signals confirmed by volume.  
**Must not:** Release Signals, read files, or own lifecycle.  
**Inputs:** Validated bars and RSI configuration.  
**Outputs:** Ordered `[]Signal`.  
**Lifecycle:** Construct, report requirements, calculate once.  
**Details:** [Signaler Logic](logic/signaler.md).

## Current Risk and Execution

### Risk interface and factory

**Status:** Implemented  
**Source:** `internal/risk/risk.go`  
**Owner:** Runtime owns returned Risks.  
**Owns:** Configuration-selected construction behind Runtime's consumer-owned interface.  
**Responsibility:** Create each configured Risk and expose assessment and stop behavior.  
**Must not:** Query Account, stop Runtime directly, or invent risk state.  
**Inputs:** Risk configuration and ordinal identity.  
**Outputs:** One `risk.Risk` or error.  
**Lifecycle:** Factory creation, repeated `Assess`, `Stop`.  
**Details:** [Risk Logic](logic/risk.md).

### BalancedRisk

**Status:** Stub  
**Source:** `internal/risk/balanced.go`  
**Owner:** Runtime through `risk.Risk`.  
**Owns:** Assessment count and stop state.  
**Responsibility:** Prove the Risk call path while always declining a stop request.  
**Must not:** Claim equity, drawdown, Account, or real risk behavior.  
**Inputs:** Timed Runtime assessment calls.  
**Outputs:** Always `false`; terminal statistics on stop.  
**Lifecycle:** Construct, repeated `Assess`, idempotent `Stop`.  
**Details:** [Risk Logic](logic/risk.md).

### Executor interface and factory

**Status:** Implemented  
**Source:** `internal/executor/executor.go`  
**Owner:** BotCycle owns returned Executors.  
**Owns:** Configuration-selected construction behind BotCycle's consumer-owned interface.  
**Responsibility:** Create configured Executors with one accepted Signal and lifecycle identity.  
**Must not:** Expose concrete implementations to Runtime or add speculative methods.  
**Inputs:** Executor configuration, Signal, cycle number, and executor number.  
**Outputs:** One `executor.Executor` or error.  
**Lifecycle:** Factory creation followed by the interface lifecycle.  
**Details:** [Executor Logic](logic/executor.md).

### ObserverExecutor

**Status:** Implemented  
**Source:** `internal/executor/observer.go`  
**Owner:** BotCycle through `executor.Executor`.  
**Owns:** Signal, entry observation, stop-loss threshold, terminal reason, and execution statistics.  
**Responsibility:** Observe BBO values and close when its configured stop loss triggers.  
**Must not:** Place orders or create Account, Ledger, Trade, Order, Fill, or Simulator state.  
**Inputs:** Signal, stop-loss configuration, BBO values, timed passes, and parent stop reason.  
**Outputs:** Terminal state, exit reason, and stop statistics.  
**Lifecycle:** Construct, `Start`, repeated `OnBBO` and current `MainLoop`, idempotent `Stop`.  
**Details:** [Executor Logic](logic/executor.md).

### BotCycle

**Status:** Implemented  
**Source:** `internal/botcycle/botcycle.go`  
**Owner:** Runtime owns at most one active BotCycle.  
**Owns:** Accepted Signal, configured Executors, cycle statistics, and cycle lifecycle.  
**Responsibility:** Start Executors, distribute BBO values, run bounded passes, detect completion, and stop children.  
**Must not:** Select Signals, decode data, own Runtime Risk, or create future Account state.  
**Inputs:** Signal, Executor configurations, BBO values, pass timestamps, and stop reason.  
**Outputs:** Completion flag, consolidated exit reason, errors, and terminal statistics.  
**Lifecycle:** `New`, `Start`, repeated `OnBBO` and current `MainLoop`, idempotent `Stop`.  
**Details:** [Executor Logic](logic/executor.md).

### Runtime

**Status:** Implemented  
**Source:** `internal/runtime/runtime.go`  
**Owner:** BtRunner.  
**Owns:** Signaler, configured Risks, one active BotCycle, end timestamp, stop reason, and Runtime statistics.  
**Responsibility:** Accept BBO values, release Signals, manage BotCycles, assess Risk, and own graceful stop.  
**Must not:** Read Parquet, drive its Clock, persist results, or claim future Account integration.  
**Inputs:** Runtime configuration, end timestamp, prepared bars, BBO values, and pass timestamps.  
**Outputs:** Stop request, errors, child actions, and terminal statistics.  
**Lifecycle:** `New`, `PrepareBars`, `Start`, repeated `Ingest` and current `MainLoop`, idempotent `Stop`.  
**Details:** [BtRunner Logic](logic/btrunner.md).

### BtRunner

**Status:** Implemented  
**Source:** `internal/btrunner/btrunner.go`  
**Owner:** BtRunner command.  
**Owns:** replay.Reader, TickClock, Runtime, and replay proof statistics.  
**Responsibility:** Build and supervise one bounded historical replay using one configured Sweep Bot.  
**Must not:** Own Signals, execution policy, child statistics, Accounts, or result persistence.  
**Inputs:** Logger, root, configuration, Sweep ID, and Bot ID.  
**Outputs:** Exact replay result, terminal statistics, or error.  
**Lifecycle:** `New`, `Start`, `Run`, idempotent `Stop`.  
**Details:** [BtRunner Logic](logic/btrunner.md).

## Current Common Mechanics

### common.Logger

**Status:** Implemented  
**Source:** `internal/common/common.go`  
**Owner:** Command creates one logger; components receive its pointer.  
**Owns:** Standard `log.Logger` and formatted `Info` output.  
**Responsibility:** Provide current component-prefixed logging.  
**Must not:** Become precedent for new logging code.  
**Inputs:** `io.Writer`, component, format string, and values.  
**Outputs:** Formatted log lines.  
**Lifecycle:** Construct once and share for the process.  
**Details:** This is pre-contract drift.

### common helpers

**Status:** Implemented  
**Source:** `internal/common/common.go`  
**Owner:** Calling packages.  
**Owns:** `StateError` and non-negative millisecond `Duration`.  
**Responsibility:** Centralize two current repeated mechanics.  
**Must not:** Grow into a generic utility package.  
**Inputs:** State labels or unsigned timestamps.  
**Outputs:** One error or duration value.  
**Lifecycle:** Stateless function calls.

### Standard logging target

**Status:** Approved design  
**Source:** No implementation. Target: `internal/logging`.  
**Owner:** Executable or server boundary creates one logger.  
**Owns:** `log/slog` handler configuration only.  
**Responsibility:** Return an explicitly passed structured logger.  
**Must not:** Wrap the full `slog` API or use global default configuration.  
**Inputs:** Output writer.  
**Outputs:** `*slog.Logger`.  
**Lifecycle:** Configure once per process boundary.  
**Details:** Source migration requires separate user confirmation.

## Approved Live Boundaries

### Server composition

**Status:** Approved design  
**Source:** No Nuubot5 implementation.  
**Owner:** Server command.  
**Owns:** Shared engines, managers, clock services, and HTTP application assembly.  
**Responsibility:** Build, start, supervise, and stop server-owned services.  
**Must not:** Contain Bot, trading, Account, or route implementation logic.  
**Inputs:** Server configuration and process resources.  
**Outputs:** Running service or terminal error.  
**Lifecycle:** Compose, start, loop, stop.

### API routes

**Status:** Future  
**Source:** No Nuubot5 implementation.  
**Owner:** HTTP application composition.  
**Owns:** API route registration and request translation.  
**Responsibility:** Validate requests and call manager boundaries.  
**Must not:** Own Bots, Sweeps, Runtime, or persistence engines.  
**Inputs:** HTTP requests.  
**Outputs:** HTTP responses.  
**Lifecycle:** Register routes during server composition.

### Web routes and assets

**Status:** Future  
**Source:** No Nuubot5 implementation.  
**Owner:** HTTP application composition.  
**Owns:** Web routes and static frontend delivery.  
**Responsibility:** Serve the web frontend and translate web requests.  
**Must not:** Own trading state or duplicate API business logic.  
**Inputs:** HTTP requests and built assets.  
**Outputs:** HTTP responses and assets.  
**Lifecycle:** Register routes during server composition.

### BotManager

**Status:** Approved design  
**Source:** No Nuubot5 implementation.  
**Owner:** Server composition.  
**Owns:** Active Runner registry and Runner lifecycle commands.  
**Responsibility:** Start, find, stop, and report live Bot Runners.  
**Must not:** Enter Runtime internals or own Account truth.  
**Inputs:** Validated Bot lifecycle requests.  
**Outputs:** Runner identity, status, or error.  
**Lifecycle:** Start with server, manage Runners, stop all.

### SweepManager

**Status:** Future  
**Source:** No Nuubot5 implementation.  
**Owner:** Server composition.  
**Owns:** Sweep lifecycle and scheduling state.  
**Responsibility:** Coordinate Sweep-level work through approved Runner or BtRunner boundaries.  
**Must not:** Own Runtime decisions or BotCycle children.  
**Inputs:** Validated Sweep requests.  
**Outputs:** Sweep identity, status, or error.  
**Lifecycle:** Start with server, manage Sweeps, stop all.

### Live Runner

**Status:** Approved design  
**Source:** No Nuubot5 implementation.  
**Owner:** BotManager.  
**Owns:** WallClock, BBO feed, user-event feed, and one Runtime.  
**Responsibility:** Supervise one live Bot and translate external events into synchronous Runtime passes.  
**Must not:** Copy Parquet replay or execute policy inside feed goroutines.  
**Inputs:** Bot configuration, clock events, BBO events, user events, and user stop.  
**Outputs:** Runtime events, status, and terminal result.  
**Lifecycle:** `New`, `Init`, `Start`, `Loop`, `Stop`.  
**Details:** [Live Runner Logic](logic/runner.md).

### WallClock

**Status:** Approved design  
**Source:** No Nuubot5 implementation.  
**Owner:** Live Runner.  
**Owns:** Fast BBO-check cadence and slower reconciliation cadence.  
**Responsibility:** Publish bounded pass events from wall time.  
**Must not:** Call Runtime policy or reconcile Accounts itself.  
**Inputs:** Configuration and context.  
**Outputs:** Typed cadence events.  
**Lifecycle:** Construct, start, loop, stop.

### BBO feed

**Status:** Approved design  
**Source:** No Nuubot5 implementation.  
**Owner:** Live Runner.  
**Owns:** BBO WebSocket connection and latest validated BBO state.  
**Responsibility:** Receive responsive market updates for stop and trailing-stop checks.  
**Must not:** Place orders, reconcile, or mutate Runtime policy.  
**Inputs:** Venue WebSocket messages.  
**Outputs:** Validated typed BBO events.  
**Lifecycle:** Connect, read until cancellation, close.

### User-event feed

**Status:** Approved design  
**Source:** No Nuubot5 implementation.  
**Owner:** Live Runner.  
**Owns:** User-event WebSocket connection. Exact state ownership remains unresolved.  
**Responsibility:** Notify the live path that Account or Ledger truth may be stale.  
**Must not:** Reconcile, mutate Ledger evidence, or decide order state.  
**Inputs:** Venue user-event messages.  
**Outputs:** Dirty notification or typed event.  
**Lifecycle:** Connect, read until cancellation, close.  
**Details:** Exact event shape and dirty-state owner remain unresolved.

## Future Account Domain

### Account

**Status:** Approved design  
**Source:** No Nuubot5 implementation. Nuubot3 is the behavior reference.  
**Owner:** Runtime's authoritative active Account list. Executors create and register required Accounts.  
**Owns:** One Ledger and one selected Account-facing Exchange adapter.  
**Responsibility:** Place, cancel, reconcile, and expose coherent local Account state through its Ledger.  
**Must not:** Let Executors call Ledger or Exchange directly.  
**Inputs:** Account configuration, execution requests, BBO values, and reconciliation triggers.  
**Outputs:** Normalized outcomes and immutable coherent snapshots.  
**Lifecycle:** Create, initialize Ledger, initialize adapter, reconcile, close.  
**Details:** Integration remains incomplete. Reference creation ownership is still migrating from BotCycle toward Executor.

### Ledger

**Status:** Approved design  
**Source:** No Nuubot5 implementation. Nuubot3 is the behavior reference.  
**Owner:** Account.  
**Owns:** Trades; each Trade owns Orders; each Order owns Fills.  
**Responsibility:** Maintain canonical local execution evidence reconciled against backend truth.  
**Must not:** Call an Exchange adapter or invent missing external evidence.  
**Inputs:** Validated backend snapshots and local intent batches.  
**Outputs:** Coherent Trade tree and Account snapshots.  
**Lifecycle:** Load or create, mutate transactionally, reconcile, flush, close.

### Trade

**Status:** Future  
**Source:** No Nuubot5 implementation. Nuubot3 is the behavior reference.  
**Owner:** Ledger.  
**Owns:** Orders and Trade-level evidence.  
**Responsibility:** Represent one strategy-managed trade lifecycle and calculate local outcomes from owned evidence.  
**Must not:** Call Exchange, own Account, or infer missing Fills.  
**Inputs:** Trade intent and reconciled Orders.  
**Outputs:** Trade state and calculated result.  
**Lifecycle:** Create, update through Orders, become terminal.

### Order

**Status:** Future  
**Source:** No Nuubot5 implementation. Nuubot3 is the behavior reference.  
**Owner:** Trade.  
**Owns:** Fills and one submitted-order identity.  
**Responsibility:** Record request, backend identity, status, and execution evidence.  
**Must not:** Submit itself or own sibling Orders.  
**Inputs:** Local order intent and normalized backend rows.  
**Outputs:** Order state and owned Fill collection.  
**Lifecycle:** Create before backend call, reconcile, become terminal.

### Fill

**Status:** Future  
**Source:** No Nuubot5 implementation. Nuubot3 is the behavior reference.  
**Owner:** Order.  
**Owns:** One execution match and its evidence.  
**Responsibility:** Preserve one normalized user-side execution fact.  
**Must not:** Aggregate unrelated executions or own lifecycle children.  
**Inputs:** Validated backend fill row.  
**Outputs:** Immutable Fill evidence.  
**Lifecycle:** Create once and retain.

### Account-facing adapter

**Status:** Approved design  
**Source:** No Nuubot5 implementation.  
**Owner:** Account.  
**Owns:** Backend-specific transport and response normalization.  
**Responsibility:** Present one stable Account-facing surface for live, testnet, and simulator backends.  
**Must not:** Mutate Ledger or import domain Trade, Order, or Fill objects.  
**Inputs:** Account requests and backend credentials or simulator configuration.  
**Outputs:** Normalized backend outcomes and snapshots.  
**Lifecycle:** Initialize, serve requests, close.  
**Details:** The canonical name `Venue` or `Exchange` remains unresolved.

### Live Hyperliquid adapter

**Status:** Future  
**Source:** No Nuubot5 implementation. Nuubot3 AsyncHyperliquid is the behavior reference.  
**Owner:** Account through the Account-facing adapter boundary.  
**Owns:** Hyperliquid REST, signing, and backend-specific response translation.  
**Responsibility:** Execute and query live or testnet Hyperliquid truth.  
**Must not:** Use CGO, mutate Ledger, or simulate fills.  
**Inputs:** Credentials, network configuration, and Account requests.  
**Outputs:** Normalized backend responses.  
**Lifecycle:** Initialize clients, serve requests, close.  
**Details:** Any SDK requires separate dependency approval and batch-order proof.

### Simulator

**Status:** Approved design  
**Source:** No Nuubot5 implementation. Nuubot3 Simulator is the behavior reference.  
**Owner:** Account through the Account-facing adapter boundary.  
**Owns:** Exchange-shaped balances, open orders, order history, fill history, and matching state.  
**Responsibility:** Act as authoritative simulated exchange truth behind the same Account-facing surface.  
**Must not:** Store domain Trade, Order, or Fill objects.  
**Inputs:** Account requests and validated BBO values.  
**Outputs:** Exchange-shaped outcomes and snapshots.  
**Lifecycle:** Initialize, ingest BBO, serve requests, persist when configured, close.
