# RunReport

Status: Implemented for the initial BtBot slice.
Covers: `internal/report`, terminal report building, structured
output, table rendering, and suite aggregation.
Purpose: Build repeatable terminal analysis without embedding reporting inside
trading objects.

## User Intent

Telemetry and RunReport are separate features.

RunReport is terminal.

It must provide a standardized result examined repeatedly after every test.

The report will grow to include execution, timing, memory, PnL, Risk, BtBot,
BotCycle, Executor, and Trade statistics.

Output must use stable tables with multiple columns.

The assistant must not hand-format changing statistics.

Structured data must enter a reusable program that calculates and renders the
output.

The implementation must not be a throwaway formatter.

It should become reusable by test scripts, later API queries, and future GUI
work.

Existing terminal `Result()` hierarchies remain authoritative.

Domain objects must not receive duplicate `Report()` methods.

RunReport should be a top-level bolt-on.

Deleting RunReport should require deleting its package and one terminal call.

The first implementation should use current available results.

New calculations and sections should be added only when required.

## Separation

RunReport consumes:

- immutable terminal Results;
- collected telemetry;
- BtBot terminal timing and memory; and
- suite-owned external process measurements.

RunReport does not collect live telemetry.

RunReport does not own trading State.

RunReport does not reconcile Accounts.

RunReport does not alter Results.

RunReport does not make trading or Risk decisions.

RunReport does not replace ResultPublisher.

ResultPublisher owns durable backtest evidence.

RunReport owns terminal calculation and presentation.

## Ownership

`internal/report` owns:

- report input validation;
- terminal calculations;
- metric definitions;
- aggregation rules;
- structured report values;
- table rendering; and
- JSON rendering.

BtBot owns one run's terminal Result and internal measurements.

The test harness owns fresh-process and suite measurements.

ResultPublisher owns the dedicated SQLite result database.

BotManager will later own API-side report queries.

No Account, Executor, BotCycle, Risk, Signaler, Controller, or Ledger imports
`report`.

`report` must not import `btbot`.

Its input is an import-safe value owned by `report`.

BtBot fills that value from Controller Result, replay proof, telemetry, and
memory.

## Top-Level Flow

One run:

```text
BtBot verifies replay
  -> stop Controller
  -> collect immutable Controller Result
  -> append final telemetry
  -> sample Go memory
  -> build RunReport once from report.Input
  -> publish Result, telemetry, and RunReport atomically
  -> command emits one compact RunReport JSON record
```

One stability suite:

```text
test harness launches N fresh BtBot processes
  -> capture one compact RunReport JSON record per successful process
  -> retain records and exact process measurements in memory
  -> retain explicit failed-attempt envelopes
  -> build SuiteReport once
  -> write one final SuiteReport JSON
  -> render standardized tables
```

The exact terminal call remains one removable integration point.

The command emits compact JSON to standard output.

Normal process logging remains in the Bot log.

The harness must not recover report values from prose logs.

## Existing Results

RunReport reuses:

- `controller.Result`;
- ordered `botcycle.Result` values;
- ordered `executor.Result` values;
- optional `account.Result`;
- Ledger Trades, Orders, and Fills;
- Simulator Orders and Fills;
- Signal decisions;
- Risk decisions; and
- replay proof.

`report.Input` contains:

- Controller Result;
- replay proof;
- ordered telemetry samples; and
- pre-publication Go memory.

It contains no live object pointers.

It does not import or depend on `btbot`.

Missing report facts should first be assessed as terminal domain facts,
telemetry facts, or derived analytics.

Terminal domain facts belong in the owning Result.

Ongoing observations belong in Telemetry.

Derived values belong only in RunReport.

## Initial RunReport

The first RunReport contains:

- Sweep ID;
- Bot ID;
- BotSpec ID;
- configuration hash;
- symbol;
- replay start and end;
- completion status;
- ticks;
- Controller runs;
- Signal packages;
- BotCycles;
- Trades;
- Orders;
- Fills;
- cancellations;
- stop Orders;
- retries;
- completed round trips;
- starting Capital;
- gross PnL;
- fees;
- net PnL;
- ending Capital;
- maximum drawdown;
- historical-data-loop elapsed time;
- heap before terminal publication;
- total allocation before terminal publication;
- GC runs before terminal publication; and
- GC pause before terminal publication.

The initial report does not invent unavailable performance ratios.

Stop Orders count only the canonical `stop` role.

Sharpe, Calmar, CAGR, expectancy, streaks, long-short breakdowns, and other
analytics require separately defined formulas and source data.

## SuiteReport

SuiteReport contains:

- requested runs;
- attempted runs;
- passed runs;
- failed runs;
- Sweep ID;
- Bot ID;
- BotSpec ID;
- symbol;
- suite elapsed time;
- one ordered RunReport summary per attempted run; and
- calculated aggregates.

Exact process launch-to-exit timing belongs to the test harness.

BtBot cannot measure work after its own final instruction.

Suite elapsed starts after build completes and before the first process launch.

It ends after the final attempted process exits.

Report generation time is excluded.

This exclusion applies only to final SuiteReport aggregation and rendering.

Each child RunReport build and publication occurs inside its process.

Exact fresh-process elapsed time therefore includes child report generation and
publication.

## Standard Table

The complete fixed-width report uses:

```text
Nx BtBot — Sweep <sweep_id>, Bot <bot_id>

BotSpec: <bot_spec_id>    Symbol: <symbol>
Status: <status>          Requested: N
Attempted: N              Passed: N    Failed: N

Timing (ms)
Item                  #  Cumulative  Avg  Min  Max
Suite (total)         1
BtBot              N
Historical Data Loop  N

Memory (MB)
Item                                 #  Cumulative  Avg  Min  Max
Heap Before Publication              N           —
Total Allocation Before Publication  N

Garbage Collection (#)
Item     #  Cumulative  Avg  Min  Max
GC Runs  N

Garbage Collection Pause (ms)
Item      #  Cumulative  Avg  Min  Max
GC Pause  N

Replay and Execution (#)
Item                   #  Cumulative  Avg  Min  Max
Ticks                  N
Controller Runs        N
Telemetry Samples      N
Signal Packages        N
Start Actions Skipped  N
BotCycles Started      N
BotCycles Rejected     N
BotCycles Closed       N
Trades                 N
Orders                 N
Fills                  N
Cancellations          N
Stop Orders            N
Submission Retries     N
Completed Round Trips  N

Financial Results (USDC)
Item              #  Cumulative  Avg  Min  Max
Starting Capital  N           —
Gross PnL         N
Fees              N
Net PnL           N
Ending Capital    N           —
Maximum Drawdown  N           —
```

The renderer, not chat prose, aligns and formats this report.

Numeric source values remain unformatted until rendering.

Units appear once in each category heading.

Missing values render as `—`.

Metric rows remain contiguous.

One blank line separates categories.

`Item` is left-aligned.

`#` is centered.

`Cumulative`, `Avg`, `Min`, and `Max` are right-aligned.

USDC values render with exactly two decimal places.

Stored numeric values retain full precision.

Large per-run and per-cycle details use additional tables.

## Aggregation Rules

| Metric kind | Cumulative | Average | Minimum | Maximum |
|---|---:|---:|---:|---:|
| Counter | Yes | Yes | Yes | Yes |
| Duration | Yes | Yes | Yes | Yes |
| Allocation | Yes | Yes | Yes | Yes |
| Gauge | No | Yes | Yes | Yes |
| Ratio | No | Yes | Yes | Yes |
| Drawdown | No | Yes | Yes | Yes |

Heap is a gauge.

Return and drawdown have no cumulative value.

Suite net PnL cumulative means experimental total across independent runs.

It is not portfolio PnL.

GC pause minimum and maximum initially mean per-BtBot cumulative pause.

They do not claim individual GC-event pause distribution.

Every table states the metric level being aggregated.

## Persistence

The dedicated backtest SQLite database stores one calculated run summary.

The summary is built before publication and stored with all evidence during the
same terminal publication.

Detailed Results and telemetry remain queryable beside it.

The stored summary prevents repeated calculation drift.

The report package may rebuild a report from exact stored evidence for
verification.

One suite may write one structured SuiteReport artifact after all attempts.

No per-run CSV or report file is required.

JSON is the canonical external structured report format.

Each successful child sends one compact JSON record through standard output.

The harness wraps it with run number and exact process elapsed time.

The harness keeps these envelopes in memory.

One suite aggregator reads the envelopes after the final process exits.

SQLite remains the canonical detailed backtest evidence.

## Rendering

RunReport has one structured model.

Renderers consume that model.

Initial renderers are:

- standardized terminal tables; and
- indented JSON.

CSV remains deferred until a real flat export requires it.

Table rendering uses the Go standard library.

JSON rendering uses the Go standard library.

No external reporting dependency is required.

Formatting never recalculates metrics.

## API and Website

The future website does not execute the terminal renderer.

It calls the API for typed summary and detail data.

The API forwards Bot report requests to BotManager.

BotManager queries the run database.

Performance cards read the stored run summary.

Equity and balance charts read telemetry ranges.

BotCycle tables read terminal BotCycle and Executor evidence.

Trade markers read Trade, Order, and Fill evidence.

Run playback combines ordered telemetry, lifecycle evidence, and referenced
market data.

## Growth

RunReport grows through explicit new fields and formulas.

Every new metric defines:

- name;
- unit;
- exact source;
- scope;
- formula;
- aggregation behavior;
- persistence; and
- proof.

No generic formula language is required.

No plugin system is required.

No report-specific fields are added to trading State.

If a fact is absent, its correct owner must be established before implementation.

## Failure

A successful RunReport requires successful terminal Result collection.

Failed attempts still appear in SuiteReport with external process status and
elapsed time.

Unavailable internal fields remain explicit.

The report must not convert a failed run into a successful result.

Publication failure fails the completed report.

A process emits no successful RunReport JSON when terminal publication fails.

## Removal

Removing RunReport requires:

1. delete `internal/report`;
2. delete the single terminal build call;
3. delete run-summary persistence;
4. delete report renderers; and
5. remove report invocation from test harnesses.

Telemetry, Results, lifecycle, trading, and detailed SQLite evidence remain
independent.

## Initial Proof

- RunReport calculations use immutable Results and telemetry only.
- Domain packages do not import `report`.
- One run report builds once.
- One suite report builds once.
- RunReport builds without a Go import cycle.
- Memory fields use the documented pre-publication boundary.
- Successful child output contains one machine-readable JSON record.
- Test harnesses consume no prose report fields.
- The generator produces the complete fixed-width report and JSON.
- Stored summary equals the rendered summary.
- Aggregates match raw run values.
- Failed attempts remain failed.
- Existing backtest trading results remain unchanged.
- Removing the integration call leaves trading execution buildable.
