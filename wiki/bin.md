# Nuubot Binaries

## Purpose

This page owns executable names, roles, implementation status, and internal package boundaries.

Command packages remain thin wrappers. Internal packages own application behavior.

## Trading Runners

```text
Executable          Scope       Internal owner      Status
nuubot-bt-sweep     Backtest    internal/btsweep    Command placeholder; template admission implemented
nuubot-bt-bot       Backtest    internal/btbot      Implemented
nuubot-runner       Live        internal/runner     Placeholder
```

### `nuubot-bt-sweep`

The target command creates and later runs one immutable backtest Sweep.

Target creation flow:

```text
load Sweep template
load referenced Bot template
validate and expand variants
create immutable Sweep and Bot records
return generated IDs
```

Future execution launches one `nuubot-bt-bot` process per generated Bot.

The current command prints `Under Construction.`.

`internal/btsweep` independently implements template loading, validation, and
deterministic expansion. It writes no database and launches no process.

Immutable Sweep and Bot creation, ID reuse, execution, cancellation, and
aggregation remain unimplemented.

### `nuubot-bt-bot`

Runs one stored Bot through one bounded historical replay.

```text
cmd/nuubot-bt-bot
`-- internal/btbot
    |-- Replay Reader
    |-- TickClock
    `-- Controller
```

The current database boundary still supplies Sweep ID and Bot ID. The target Bot ID is globally unique; Sweep ID is optional grouping provenance.

Performance profiling is enabled only through the explicit `-pp <prefix>` command option.

### `nuubot-runner`

Runs one standalone live, testnet, paper, or Simulator Bot.

The current command prints `Under Construction.`. `internal/runner` is not implemented.

Future live admission reads the immutable Bot network. An operator network argument may confirm that value but never override it.

## Administration and Reporting

```text
Executable          Purpose                                      Status
nuubot-server       Optional application server and supervision  Placeholder
nuubot-cli          Operator commands                             Placeholder
nuubot-report       Render and aggregate backtest reports         Implemented
parity-probe        Compare selected Venue responses              Implemented
```

The approved placeholder commands print `Under Construction.` until their implementation is authorized.

The only documented future Sweep import command is:

```text
nuubot-cli create sweep -f <abc.toml>
```

That CLI behavior is unimplemented.

## System Testing

`stest.sh` is the only system-test entrypoint.

```sh
./stest.sh -bot <bot_id>
./stest.sh -sweep <sweep_id>
./stest.sh -bot <bot_id> -runs 10
./stest.sh -bot <bot_id> -pp
```

`-bot` runs one globally unique Bot. The current datastore resolves its Sweep ID for `nuubot-bt-bot` compatibility.

`-sweep` runs every Bot grouped under one Sweep in deterministic order.

`-pp` enables performance profiling and requires one run.

## Naming Contract

- `bt` means backtest.
- `sweep` means one immutable grouping and generation definition.
- `bot` means one exact immutable Bot configuration.
- `runner` without `bt` means live execution.
- Executable names use visible hyphens between concepts.
- Old `nuubot-btrunner` and `BtRunner` compatibility names are prohibited.
