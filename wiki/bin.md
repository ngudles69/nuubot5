# Nuubot Binaries

## Purpose

This page owns executable names, roles, implementation status, and internal package boundaries.

Command packages remain thin wrappers. Internal packages own application behavior.

## Trading Runners

```text
Executable          Scope       Internal owner      Status
nuubot-sweep         Backtest    internal/btsweep    Command placeholder; template admission implemented
nuubot-backtest      Backtest    internal/backtest   Implemented
nuubot-live          Live        internal/live       Non-runnable scaffold
```

### `nuubot-sweep`

The target command creates and later runs one immutable backtest Sweep.

Target creation flow:

```text
load Sweep template
load referenced Bot template
validate and expand variants
create immutable Sweep and Bot records
return generated IDs
```

Future execution launches one `nuubot-backtest` process per generated Bot.

The current command prints `Under Construction.`.

`internal/btsweep` independently implements template loading, validation, and
deterministic expansion. It writes no database and launches no process.

Immutable Sweep and Bot creation, ID reuse, execution, cancellation, and
aggregation remain unimplemented.

### `nuubot-backtest`

Runs one stored Bot through one bounded historical replay.

```text
cmd/nuubot-backtest
`-- internal/backtest.Execute
    |-- internal/runharness.Profile
    |-- Replay Reader
    |-- TickClock
    `-- Controller
```

The current database boundary still supplies Sweep ID and Bot ID. The target Bot ID is globally unique; Sweep ID is optional grouping provenance.

The command only parses arguments, calls `backtest.Execute`, and reports one terminal error.

Performance profiling is enabled only through `-pp <prefix>`. Shared whole-Run profiling mechanics live in `internal/runharness`.

### `nuubot-live`

Runs one standalone live, testnet, paper, or Simulator Bot.

The command only parses arguments, calls `live.Execute`, and reports one terminal error.

The command and `internal/live` lifecycle scaffold are implemented.

The scaffold is not a working live runtime. WebSocket transport, live Setup, and live Signaler input remain unimplemented.

Do not execute the command until those boundaries are implemented and proven.

Future live admission reads the immutable Bot network. An operator network argument may confirm that value but never override it.

## Administration and Reporting

```text
Executable          Purpose                                      Status
nuubot-server       Optional application server and supervision  Placeholder
nuubot-cli          Operator commands                             Placeholder
nuubot-stest-report  Render and aggregate system-test reports     Implemented
nuubot-fprof        Exact A/B/C function profiling                Implemented
parity-probe        Compare selected Venue responses              Implemented
```

`nuubot-fprof -sweep ID -bot ID [-top N]` writes isolated profiles below `workspace/perf/fprofiles`.

It uses temporary Go build overlays and never modifies tracked application source.

The approved Server and CLI placeholder commands print `Under Construction.` until their implementation is authorized.

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

`-bot` runs one globally unique Bot. The current datastore resolves its Sweep ID for `nuubot-backtest` compatibility.

`-sweep` runs every Bot grouped under one Sweep in deterministic order.

`-pp` enables performance profiling and requires one run.

## Naming Contract

- `backtest` means one bounded historical Bot replay.
- `sweep` means one immutable grouping and generation definition.
- `bot` means one exact immutable Bot configuration.
- `live` means standalone live-network execution.
- `stest-report` means system-test suite aggregation and rendering.
- Executable names use visible hyphens between concepts.
- Old command aliases are prohibited.
