# Telemetry

Status: Implemented for the initial BtRunner slice.
Covers: Object `Telemetry()` functions, BtRunner collection, Runner
persistence, and API query boundaries.
Purpose: Capture ongoing operational state without coupling observation to
trading behavior.

## User Intent

Telemetry and RunReport are separate features.

Every participating object owns custom private state.

Each object supplies one standalone `Telemetry()` function that reads its
current state.

A parent calls each child's `Telemetry()` once while composing its own
telemetry.

Telemetry must remain detachable.

Removing telemetry should require deleting its functions, one call per
ownership level, root collection, and telemetry persistence.

Telemetry must not scatter update calls through trading behavior.

BtRunner collects telemetry in memory and writes it during one terminal result
publication.

Runner will collect the same telemetry on WallClock time and persist it
periodically.

The existing Controller cadence, currently ten seconds, should drive collection.

The website will call the API.

The API forwards Bot requests to BotManager.

BotManager reads the latest or requested telemetry records from the owning
database.

The website can use current telemetry as a live monitor.

Historical telemetry can support run playback.

The first implementation should be small.

New telemetry should be added only when a real display, analysis, or operational
requirement needs it.

## Separation

Telemetry is ongoing observation.

Telemetry is not terminal Result evidence.

Telemetry is not final performance analysis.

Telemetry is not RunReport.

RunReport may consume telemetry after execution.

Trading decisions must never consume telemetry.

Risk continues to consume its exact immutable RiskInput.

## Terminology

| Term | Meaning |
|---|---|
| State | One object's canonical private mutable operational facts. |
| Telemetry | One immutable observation derived from current State. |
| Sample | One timestamped root telemetry observation. |
| Current | The latest completed sample. |
| Run playback | Displaying a stored run from ordered samples and domain evidence. |
| Historical replay | BtRunner's historical-data loop. It is not run playback. |

## Ownership

Each object owns its operational State.

State remains domain-specific.

There is no shared generic state map.

There is no shadow telemetry state duplicating operational truth.

`context.Context` carries cancellation, deadlines, and request-scoped immutable
values.

Mutable telemetry must not be stored in `context.Context`.

Each `Telemetry()` function reads only its owner and direct children.

Each parent composes owned child telemetry.

Only BtRunner or Runner decides when collection occurs.

Only the top process owner decides persistence.

No Signaler, Risk, BotCycle, Executor, or Account writes telemetry to SQLite.

## Function Contract

Each implemented telemetry function:

- reads current owned State;
- calls each required child `Telemetry()` once;
- returns an independently owned immutable value;
- performs no mutation;
- performs no reconciliation;
- performs no trading decision;
- performs no file, database, network, or logging operation; and
- remains valid while its owner is running or stopped.

Telemetry methods must not return pointers into mutable State.

Missing optional children produce explicit absence.

They do not produce invented zero-valued child state.

An Account remains absent until its first successful reconciliation observation.

Controller retains carried resource equity while that Account is absent.

## Ownership Flow

```text
BtRunner.Telemetry
  -> Controller.Telemetry
     -> active BotCycle.Telemetry, when present
        -> Executor.Telemetry
           -> Account.Telemetry, when present
```

Signaler and Risk telemetry join this flow only when their current state becomes
a required output.

Trade, Order, Fill, and Grid Level evidence remains in terminal Results and
event tables.

Periodic telemetry must not duplicate their complete trees.

## Collection Timing

BtRunner uses the existing registered Controller callback.

It does not register a second competing telemetry timer.

One successful callback runs:

```text
Controller.Run(now_ms)
collect BtRunner.Telemetry(now_ms)
append one sample
```

Collection occurs after reconciliation, Risk, BotCycle, Signaler, and execution
work complete successfully.

Failed Controller work produces no normal sample.

Terminal failure remains explicit in process and result evidence.

BtRunner timestamps samples with TickClock historical time.

Runner will timestamp samples with WallClock time.

Account telemetry carries the collection timestamp and its latest successful
reconciliation observation timestamp.

These timestamps must not be conflated.

A successful terminal stop records one final sample after Controller cleanup.

The terminal sample may share a timestamp with the preceding sample.

Sample sequence, not timestamp, provides unique ordering.

## Initial Scope

The first implementation records only data already required for the first
monitor and RunReport.

The root sample contains:

- sequence;
- historical timestamp;
- terminal flag;
- replay ticks served;
- Controller runs;
- Signal packages read;
- start actions skipped;
- BotCycles started, rejected, closed, and currently active;
- Bot capital;
- Bot balance;
- Bot equity;
- net PnL;
- peak equity;
- current drawdown; and
- maximum drawdown.

Bot balance means Bot equity minus current unrealized PnL.

Fees remain included in the underlying Account and Bot equity.

The initial implementation does not persist complete Signaler, Risk, Executor,
Account, Trade, Order, Fill, or Grid Level trees in every sample.

Those fields or child series are added only when required.

## BtRunner Persistence

BtRunner retains compact samples in memory.

BtRunner performs no periodic telemetry file write.

Successful replay verification and Controller shutdown produce terminal Results.

BtRunner then:

1. collects the final Controller Result;
2. appends the final telemetry sample;
3. samples Go memory;
4. builds RunReport from an import-safe input;
5. publishes Results, telemetry, and RunReport together; and
6. emits the completed compact RunReport JSON.

ResultPublisher writes all evidence into the dedicated backtest SQLite database
during one terminal publication phase.

`.partial` remains hidden until publication succeeds.

Publication failure produces no completed result database.

One terminal publication phase does not imply one physical filesystem write.

SQLite still writes database pages internally.

The initial table is `telemetry_sample`.

It stores typed columns for the initial root sample.

Financial values remain exact decimal text.

Timestamp and sequence columns support ordered range queries.

`sequence` is the primary key.

An index on `timestamp_ms` supports monitor and playback ranges.

Go memory is sampled after Controller shutdown and before RunReport building and
ResultPublisher.

The stored memory metrics are:

- `heap_before_publication_mb`;
- `total_alloc_before_publication_mb`;
- `gc_runs_before_publication`; and
- `gc_pause_before_publication_ms`.

They exclude RunReport construction and terminal publication allocations.

Exact fresh-process elapsed time still includes both operations.

## Runner Persistence

Runner remains unimplemented.

Its future telemetry uses the same snapshot contracts.

Runner collects after its successful Controller cadence.

Runner persists periodically because monitoring must continue while the process
runs.

Live persistence must not block the Controller decision path.

The future writer receives immutable samples and writes bounded batches.

One Runner remains the only writer for its run database.

Server and BotManager are readers.

Live SQLite must support concurrent readers without changing trading ownership.

The exact live batching, WAL, retention, and failure policy remain deferred
until Runner implementation.

## API and Website

The website never reads SQLite directly.

The website calls the API.

The API remains thin and forwards Bot telemetry requests to BotManager.

BotManager resolves the owning run database and performs the query.

Required query shapes are:

- latest completed sample;
- ordered samples between two timestamps; and
- downsampled samples for charts.

The API returns typed data, not rendered terminal tables.

The monitor polls latest telemetry.

Run playback requests an ordered historical range.

Market price data remains owned by its market-data source.

Telemetry does not duplicate the full historical tick dataset.

## Scale

The current three-month baseline performs about 794,880 Controller passes.

One root sample per pass therefore produces about 794,880 samples.

The initial sample must remain compact.

Periodic snapshots must not embed complete child evidence.

The website must not receive every sample for a large chart.

BotManager or its query layer downsamples by requested window and point limit.

Implementation proof records telemetry rows, database size, publication time,
BtRunner elapsed time, loop time, and memory.

## Growth

Telemetry grows additively from real requirements.

Possible later series include:

- per-resource Account equity and balance;
- active position and margin;
- Signaler indicator values at source-bar cadence;
- Risk decisions and inputs when changed;
- Executor state when changed;
- Runner health;
- feed health; and
- live connection state.

Each new series must define owner, source timestamp, collection cadence,
persistence, query need, and retention.

No generic metric registry is required for the initial implementation.

## Removal

Removing telemetry requires:

1. delete telemetry types;
2. delete each implemented `Telemetry()` function;
3. delete one child call per ownership level;
4. delete root collection;
5. delete telemetry persistence and queries; and
6. remove telemetry from RunReport inputs.

Trading decisions, Results, reconciliation, execution, and lifecycle must remain
unchanged.

## Initial Proof

- Telemetry functions do not mutate owned State.
- Unobserved Accounts never create zero-equity samples.
- Maximum drawdown never decreases between ordered samples.
- One successful Controller callback creates one ordered sample.
- Failed Controller work creates no normal sample.
- Terminal cleanup creates the final sample.
- Three-month BtRunner writes telemetry only during terminal publication.
- Result SQLite integrity and foreign keys pass.
- Stored sample count and boundaries match the collection contract.
- Latest and range queries return ordered values.
- Existing trading counts, PnL, and final equity remain unchanged.
- Memory and elapsed-time impact are recorded.
