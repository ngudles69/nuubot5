# BtSweep Package

Status: Template validation and expansion implemented. Persistence and execution unimplemented.
Covers: `internal/btsweep/**`
Purpose: Load one Sweep template and deterministically generate complete validated Bot Configs.

## Boundary

`btsweep.Load` receives one Sweep template path.

It returns:

- absolute source and referenced Bot template paths;
- exact BotSpec ID;
- replay symbol and absolute, clean tick path;
- ordered date ranges;
- exact generated Config TOML and SHA-256;
- one-based deterministic Bot numbers.

It writes no database, creates no record ID, and launches no process.

`cmd/nuubot-bt-sweep` remains the approved `Under Construction.` placeholder.

## Program Flow

```text
Load
  resolve source path
  decode sweep template
  validate sweep template
  load bot template
  validate parameter dimensions
  expand generated bots
```

Each combination decodes a fresh Bot map, applies values, encodes complete TOML,
and calls `botspec.Validate` with the template's exact `bot_spec`.

## Sweep Template

One file contains:

```toml
[sweep]
doc = "Operator description."
template = "../../bots/macross_grid_v1.toml"
symbol = "BTC"
ticks = "D:/workspace/data/binance/parquet/spot/monthly/klines/BTCUSDT/1s"

[[sweep.date_ranges]]
name = "BTCUSDT-2026-Q1"
start = "2026-03-01"
end = "2026-06-01"

[sweep.parameters.executors.grid]
levels = [30, 50]
```

Rules:

- One Sweep template references one Bot template.
- `sweep.doc` must contain non-whitespace text.
- Relative Bot-template and tick paths resolve from the Sweep source directory.
- Date ranges require unique nonempty names and increasing `YYYY-MM-DD` dates.
- Date-range order is preserved.
- The parameters table may contain zero dimensions.
- Zero dimensions emit one unchanged Bot Config per date range.
- Every present parameter value is an explicit nonempty list.
- Parameter paths are sorted before Cartesian expansion.
- Parameter list order is preserved.
- Range syntax and scalar parameter shorthand are rejected.
- Unknown Sweep fields and generated record IDs are rejected.

## Bot Template

One Bot template is one complete scalar Config.

Top-level `executors` and `risks` remain BotSpec-owned arrays of tables.
Parameter arrays inside Bot fields are rejected.

The unchanged template must pass exact `botspec.Validate` before expansion.
Every generated Config must pass the same validation after expansion.
Every generated Executor symbol must equal the Sweep replay symbol.

Ordinary nested parameter paths resolve directly.

An Executor array-of-table uses its stable Config `id` as the selector:

```toml
[sweep.parameters.executors.grid]
levels = [30, 50]
```

`grid` selects the one `[[executors]]` row with `id = "grid"`.
Missing or duplicate selectors fail.

A parameter path must exist in the Bot template and belong to the exact
BotSpec's recognized scalar path catalogue. Ignored extra BotConfig fields
cannot become Sweep dimensions.

Top-level `id`, `sweep_id`, and `bot_id` are generated record identities and are
invalid in Bot templates. Executor `id` is Config identity, not a generated
record ID.

## Determinism

Date range is the outer expansion order.

Sorted parameter path is the Cartesian dimension order. Earlier paths change
more slowly. Values retain template order.

Bot numbers start at one and increase across all date ranges.

BurntSushi TOML encodes each complete Config. SHA-256 hashes those exact emitted
bytes.

## Deferred

- Immutable Sweep and Bot record creation.
- Global Bot ID allocation and optional Sweep grouping IDs.
- Database writes and idempotent unchanged-template reuse.
- Bounded BtBot workers, execution, cancellation, and aggregation.
- `nuubot-bt-sweep` command behavior.
- `nuubot-cli create sweep -f <abc.toml>` implementation.
